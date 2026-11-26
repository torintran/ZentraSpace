package main

import (
	"encoding/base32"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const vietQRMax = 499999999
const statementHeader = "ma_giao_dich,ngay,so_tien,noi_dung"

type BankAccount struct {
	BIN, Account, Holder string
}
type BankTransfer struct {
	ID, AccountKey, AccountSuffix, Reference, Date, Memo string
	Amount                                               int64
	InvoiceID, EntryKey                                  string
}

var (
	binPattern         = regexp.MustCompile("^[0-9]{6}$")
	accountPattern     = regexp.MustCompile("^[0-9]{4,19}$")
	referencePattern   = regexp.MustCompile("^[A-Za-z0-9][A-Za-z0-9./_-]{0,79}$")
	paymentCodePattern = regexp.MustCompile("^ZS[0-9A-HJKMNP-TV-Z]{8}$")
	memoPattern        = regexp.MustCompile("^[A-Z0-9]{1,25}$")
	crockford          = base32.NewEncoding("0123456789ABCDEFGHJKMNPQRSTVWXYZ").WithPadding(base32.NoPadding)
)

func (bank BankAccount) Configured() bool {
	return binPattern.MatchString(bank.BIN) && accountPattern.MatchString(bank.Account) && bank.Holder != ""
}
func (bank BankAccount) Suffix() string {
	if len(bank.Account) < 4 {
		return ""
	}
	return bank.Account[len(bank.Account)-4:]
}

func (b *Book) setBank(bank BankAccount) error {
	if !bank.Configured() || utf8.RuneCountInString(bank.Holder) > 80 || strings.ContainsFunc(bank.Holder, unicode.IsControl) {
		return errors.New("Nhập mã BIN 6 số, số tài khoản 4–19 chữ số và tên chủ tài khoản.")
	}
	if b.Bank.Configured() && (b.Bank.BIN != bank.BIN || b.Bank.Account != bank.Account) {
		b.MockBankAutoMatch = false
		for j := range b.Invites {
			b.Invites[j].Revoked = true
		}
	}
	b.Bank = bank
	for j := range b.Invoices {
		if b.Invoices[j].Issued {
			b.ensurePaymentCode(&b.Invoices[j])
		}
	}
	return nil
}

func (b *Book) ensurePaymentCode(i *Invoice) {
	for i.PaymentCode == "" {
		code := "ZS" + crockford.EncodeToString(randBytes(5))
		used := false
		for _, other := range b.Invoices {
			if other.PaymentCode == code {
				used = true
				break
			}
		}
		if !used {
			i.PaymentCode = code
		}
	}
}

func (b *Book) suggestInvoice(memo string) *Invoice {
	words := strings.FieldsFunc(strings.ToUpper(memo), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	var code string
	for _, word := range words {
		if !paymentCodePattern.MatchString(word) {
			continue
		}
		if code != "" && code != word {
			return nil
		}
		code = word
	}
	if code == "" {
		return nil
	}
	for j := range b.Invoices {
		i := &b.Invoices[j]
		if i.Issued && i.PaymentCode == code {
			return i
		}
	}
	return nil
}

func (b *Book) importStatement(data string) (added, skipped int, err error) {
	if !b.Bank.Configured() {
		return 0, 0, errors.New("Hãy lưu tài khoản nhận tiền trước khi nhập sao kê.")
	}
	accountKey := hash(b.Bank.BIN + ":" + b.Bank.Account)
	input := csv.NewReader(strings.NewReader(strings.TrimPrefix(data, "\ufeff")))
	input.FieldsPerRecord = 4
	header, err := input.Read()
	if err != nil || strings.Join(header, ",") != statementHeader {
		return 0, 0, errors.New("Dòng đầu cần theo đúng mẫu CSV sao kê.")
	}
	refs := make(map[string]BankTransfer)
	for _, tr := range b.Transfers {
		if tr.AccountKey == accountKey {
			refs[tr.Reference] = tr
		}
	}
	var additions []BankTransfer
	for row := 2; ; row++ {
		cells, err := input.Read()
		if err == io.EOF {
			break
		}
		if err != nil || row > 151 {
			return 0, 0, fmt.Errorf("Dòng %d: kiểm tra định dạng hoặc giới hạn 150 dòng.", row)
		}
		for j := range cells {
			cells[j] = strings.TrimSpace(cells[j])
		}
		_, dateErr := time.Parse("2006-01-02", cells[1])
		amount, amountErr := strconv.ParseInt(cells[2], 10, 64)
		if !referencePattern.MatchString(cells[0]) || dateErr != nil || amountErr != nil || amount <= 0 || amount > 100000000000 || len(cells[3]) > 300 || strings.ContainsFunc(cells[3], unicode.IsControl) {
			return 0, 0, fmt.Errorf("Dòng %d: kiểm tra mã giao dịch, ngày YYYY-MM-DD, số tiền vào nguyên dương và nội dung.", row)
		}
		if saved, exists := refs[cells[0]]; exists {
			if saved.Date != cells[1] || saved.Amount != amount || saved.Memo != cells[3] {
				return 0, 0, fmt.Errorf("Dòng %d: mã giao dịch đã tồn tại với dữ liệu khác.", row)
			}
			skipped++
			continue
		}
		tr := BankTransfer{ID: token(), AccountKey: accountKey, AccountSuffix: b.Bank.Suffix(), Reference: cells[0], Date: cells[1], Amount: amount, Memo: cells[3]}
		refs[tr.Reference] = tr
		additions = append(additions, tr)
	}
	if len(b.Transfers)+len(additions) > 3000 {
		return 0, 0, errors.New("Bản thử hỗ trợ tối đa 3.000 dòng sao kê.")
	}
	b.Transfers = append(b.Transfers, additions...)
	return len(additions), skipped, nil
}

func (b *Book) transfer(id string) *BankTransfer {
	for j := range b.Transfers {
		if b.Transfers[j].ID == id {
			return &b.Transfers[j]
		}
	}
	return nil
}

func (b *Book) reconcile(transferID, invoiceID, existing string) error {
	tr, i := b.transfer(transferID), b.invoice(invoiceID)
	if tr == nil || i == nil || !i.Issued {
		return errMissing
	}
	key := existing
	if key == "" {
		key = "bank-" + tr.ID
	}
	if tr.InvoiceID != "" {
		if tr.InvoiceID == invoiceID && tr.EntryKey == key {
			return nil
		}
		return errors.New("Giao dịch đã được đối soát vào một bút toán khác.")
	}
	if existing != "" {
		eligible := false
		for _, e := range i.Entries {
			if e.Key == key && e.Kind == "Thu" && e.Method != cashPayment && e.Amount == tr.Amount {
				eligible = true
				break
			}
		}
		if !eligible || i.Reversed(key) {
			return errors.New("Khoản đã ghi phải là khoản thu cùng số tiền trên bảng kê này.")
		}
		for _, other := range b.Transfers {
			if other.InvoiceID == invoiceID && other.EntryKey == key {
				return errors.New("Khoản thu này đã ghép với giao dịch khác.")
			}
		}
	} else if err := b.addEntry(invoiceID, Entry{Key: key, Kind: "Thu", Amount: tr.Amount, Method: transferPayment, Reference: tr.Reference}); err != nil {
		return err
	}
	tr.InvoiceID, tr.EntryKey = invoiceID, key
	return nil
}

func vietQR(bank BankAccount, amount int64, memo string) (string, error) {
	if !bank.Configured() || amount <= 0 || amount > vietQRMax || !memoPattern.MatchString(memo) {
		return "", errInvalid
	}
	beneficiary := tlv("00", bank.BIN) + tlv("01", bank.Account)
	napas := tlv("00", "A000000727") + tlv("01", beneficiary) + tlv("02", "QRIBFTTA")
	data := "000201010212" + tlv("38", napas) + "5303704" + tlv("54", strconv.FormatInt(amount, 10)) + "5802VN" + tlv("62", tlv("08", memo)) + "6304"
	return fmt.Sprintf("%s%04X", data, crc16(data)), nil
}

func tlv(id, value string) string { return fmt.Sprintf("%s%02d%s", id, len(value), value) }

func crc16(data string) uint16 {
	crc := uint16(0xffff)
	for j := range len(data) {
		crc ^= uint16(data[j]) << 8
		for range 8 {
			if crc&0x8000 != 0 {
				crc = crc<<1 ^ 0x1021
			} else {
				crc <<= 1
			}
		}
	}
	return crc
}
