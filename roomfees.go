package main

import (
	"errors"
	"strings"
	"time"
	"unicode"
)

type RoomFee struct {
	ID, RoomID, Label, Start, End string
	Monthly                       bool
	Amount                        int64
}

type FeeLine struct {
	Label  string
	Amount int64
}

func (fee RoomFee) applies(period string) bool {
	if fee.End != "" && period >= fee.End {
		return false
	}
	return period == fee.Start || (fee.Monthly && period > fee.Start)
}

func (b *Book) addRoomFee(fee RoomFee) error {
	room := b.room(fee.RoomID)
	if room == nil {
		return errMissing
	}
	if _, err := time.Parse("2006-01", fee.Start); err != nil || fee.Start < room.Start || len(fee.ID) != 48 ||
		fee.Amount <= 0 || fee.Amount > 100000000 || fee.Label == "" || len(fee.Label) > 160 ||
		strings.ContainsFunc(fee.Label, unicode.IsControl) {
		return errInvalid
	}
	for _, saved := range b.RoomFees {
		if saved.ID == fee.ID {
			saved.End = ""
			if saved == fee {
				return nil
			}
			return errInvalid
		}
	}
	for _, invoice := range b.Invoices {
		if invoice.RoomID == room.ID && invoice.Issued && fee.applies(invoice.Period) {
			return errors.New("Khoản phí trùng kỳ đã phát hành. Chọn kỳ bắt đầu sau bảng kê đã phát hành.")
		}
	}
	b.RoomFees = append(b.RoomFees, fee)
	b.refreshRoomFeeDrafts(*room)
	return nil
}

func (b *Book) stopRoomFee(roomID, feeID, end string) error {
	room := b.room(roomID)
	if room == nil {
		return errMissing
	}
	for j := range b.RoomFees {
		fee := &b.RoomFees[j]
		if fee.ID != feeID || fee.RoomID != roomID {
			continue
		}
		if !fee.Monthly {
			end = fee.Start
		}
		if _, err := time.Parse("2006-01", end); err != nil || end < fee.Start {
			return errInvalid
		}
		if fee.End == end {
			return nil
		}
		if fee.End != "" {
			return errors.New("Khoản phí đã có kỳ dừng.")
		}
		for _, invoice := range b.Invoices {
			if invoice.RoomID == roomID && invoice.Issued && invoice.Period >= end && fee.applies(invoice.Period) {
				return errors.New("Khoản phí đã có trên bảng kê phát hành. Chọn kỳ dừng sau kỳ đó.")
			}
		}
		fee.End = end
		b.refreshRoomFeeDrafts(*room)
		return nil
	}
	return errMissing
}

func (b *Book) refreshRoomFeeDrafts(room Room) {
	for j := range b.Invoices {
		i := &b.Invoices[j]
		if i.RoomID == room.ID && !i.Issued {
			*i = b.calculateInvoice(room, i.Reading, i.ID)
		}
	}
}
