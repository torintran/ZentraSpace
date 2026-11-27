package app

import (
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const (
	creditOut         = "Chuyển tiền dư"
	creditIn          = "Cấn tiền dư"
	creditOutReversed = "Đảo chuyển tiền dư"
	creditInReversed  = "Đảo cấn tiền dư"
)

type CreditTransfer struct {
	Key, SourceID, TargetID, Reverses, Note string
	Amount                                  int64
	Date                                    time.Time
}

type CreditTarget struct {
	ID, Period string
	Remaining  int64
}
type CreditTransferView struct {
	CreditTransfer
	TargetPeriod string
	Reversed     bool
}

func (b *Book) creditOptions(source *Invoice) ([]CreditTarget, []CreditTransferView) {
	var targets []CreditTarget
	var transfers []CreditTransferView
	if source.transferableCredit() > 0 {
		for index := range b.Invoices {
			target := &b.Invoices[index]
			if sameCreditLease(source, target) && target.Remaining() > 0 {
				targets = append(targets, CreditTarget{ID: target.ID, Period: target.Period, Remaining: target.Remaining()})
			}
		}
	}
	reversed := make(map[string]bool)
	for _, transfer := range b.CreditTransfers {
		if transfer.Reverses != "" {
			reversed[transfer.Reverses] = true
		}
	}
	for _, transfer := range b.CreditTransfers {
		if transfer.SourceID != source.ID || transfer.Reverses != "" {
			continue
		}
		if target := b.invoice(transfer.TargetID); target != nil {
			transfers = append(transfers, CreditTransferView{CreditTransfer: transfer, TargetPeriod: target.Period, Reversed: reversed[transfer.Key]})
		}
	}
	sort.SliceStable(targets, func(i, j int) bool { return targets[i].Period < targets[j].Period })
	return targets, transfers
}

func (i Invoice) transferredCredit() int64 {
	var moved int64
	for _, e := range i.Entries {
		switch e.Kind {
		case creditOut:
			moved += e.Amount
		case creditOutReversed:
			moved -= e.Amount
		}
	}
	return moved
}

func (i Invoice) refundableCash() int64 {
	return max(0, i.cashCollected()-i.transferredCredit())
}

func (i Invoice) transferableCredit() int64 {
	if !i.Issued || i.LeaseEffective == "" {
		return 0
	}
	return min(i.Credit(), i.refundableCash())
}

func sameCreditLease(source, target *Invoice) bool {
	return source != nil && target != nil && source.Issued && target.Issued && source.ID != target.ID &&
		source.RoomID == target.RoomID && source.LeaseEffective != "" && source.LeaseEffective == target.LeaseEffective && source.Period < target.Period
}

func (b *Book) postCreditTransfer(transfer CreditTransfer, now time.Time) error {
	if len(transfer.Key) != 48 || utf8.RuneCountInString(transfer.Note) > 160 || strings.ContainsFunc(transfer.Note, unicode.IsControl) {
		return errInvalid
	}
	if _, err := hex.DecodeString(transfer.Key); err != nil {
		return errInvalid
	}
	var original *CreditTransfer
	for index := range b.CreditTransfers {
		saved := &b.CreditTransfers[index]
		if saved.Key == transfer.Key {
			transfer.Date = saved.Date
			if *saved == transfer {
				return nil
			}
			return errors.New("Thao tác đã ghi với thông tin khác.")
		}
		if transfer.Reverses != "" && saved.Reverses == transfer.Reverses {
			return errors.New("Khoản chuyển này đã được đảo.")
		}
		if saved.Key == transfer.Reverses {
			original = saved
		}
	}
	source, target := b.invoice(transfer.SourceID), b.invoice(transfer.TargetID)
	if !sameCreditLease(source, target) {
		return errors.New("Chỉ chuyển tiền dư sang bảng kê kỳ sau của cùng hợp đồng.")
	}
	if transfer.Amount <= 0 || transfer.Amount > 100000000000 {
		return errInvalid
	}
	out, in := creditOut, creditIn
	if transfer.Reverses == "" {
		if transfer.Amount > source.transferableCredit() || transfer.Amount > target.Remaining() {
			return errors.New("Số tiền vượt khoản dư hoặc số còn phải trả của bảng kê nhận.")
		}
	} else {
		if original == nil || original.Reverses != "" || original.SourceID != transfer.SourceID || original.TargetID != transfer.TargetID || original.Amount != transfer.Amount || transfer.Note == "" {
			return errInvalid
		}
		out, in = creditOutReversed, creditInReversed
	}
	transfer.Date = now.UTC()
	key := "credit-" + transfer.Key
	source.Entries = append(source.Entries, Entry{Key: key, Kind: out, Amount: transfer.Amount, Date: transfer.Date, Note: transfer.Note, Reference: target.Period})
	target.Entries = append(target.Entries, Entry{Key: key, Kind: in, Amount: transfer.Amount, Date: transfer.Date, Note: transfer.Note, Reference: source.Period})
	b.CreditTransfers = append(b.CreditTransfers, transfer)
	return nil
}
