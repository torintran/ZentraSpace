package main

import (
	"errors"
	"sort"
	"time"
)

type LeaseTerms struct {
	RoomID, Effective, Tenant     string
	Rent, ElectricRate, WaterRate int64
}

func (b *Book) leaseHistory(room Room) []LeaseTerms {
	history := []LeaseTerms{{room.ID, room.Start, room.Tenant, room.Rent, room.ElectricRate, room.WaterRate}}
	for _, terms := range b.Leases {
		if terms.RoomID == room.ID {
			history = append(history, terms)
		}
	}
	sort.Slice(history, func(i, j int) bool { return history[i].Effective < history[j].Effective })
	return history
}

func (b *Book) leaseAt(room Room, period string) LeaseTerms {
	active := LeaseTerms{room.ID, room.Start, room.Tenant, room.Rent, room.ElectricRate, room.WaterRate}
	for _, terms := range b.Leases {
		if terms.RoomID == room.ID && terms.Effective > active.Effective && terms.Effective <= period {
			active = terms
		}
	}
	return active
}

func (b *Book) calculateInvoice(room Room, reading Reading, id string) Invoice {
	active := b.leaseAt(room, reading.Period)
	room.Rent, room.ElectricRate, room.WaterRate = active.Rent, active.ElectricRate, active.WaterRate
	invoice := calculate(room, reading, id)
	invoice.LeaseEffective, invoice.LeaseTenant = active.Effective, active.Tenant
	settings := b.Billing[reading.Period]
	invoice.DueDate, invoice.CommonFeeLabel, invoice.CommonFee = settings.DueDate, settings.CommonFeeLabel, settings.CommonFee
	invoice.Total += invoice.CommonFee
	for _, fee := range b.RoomFees {
		if fee.RoomID == room.ID && fee.applies(reading.Period) {
			invoice.RoomFees = append(invoice.RoomFees, FeeLine{Label: fee.Label, Amount: fee.Amount})
			invoice.Total += fee.Amount
		}
	}
	return invoice
}

func (b *Book) scheduleLease(terms LeaseTerms) error {
	room := b.room(terms.RoomID)
	if room == nil {
		return errMissing
	}
	if _, err := time.Parse("2006-01", terms.Effective); err != nil || terms.Effective <= room.Start ||
		terms.Tenant == "" || len(terms.Tenant) > 150 ||
		terms.Rent < 0 || terms.Rent > 100000000 ||
		terms.ElectricRate < 0 || terms.ElectricRate > 100000 ||
		terms.WaterRate < 0 || terms.WaterRate > 1000000 {
		return errInvalid
	}
	for _, existing := range b.Leases {
		if existing.RoomID == terms.RoomID && existing.Effective == terms.Effective {
			if existing == terms {
				return nil
			}
			return errors.New("Đã có điều khoản hiệu lực từ kỳ này.")
		}
	}
	for _, invoice := range b.Invoices {
		if invoice.RoomID == terms.RoomID && invoice.Issued && invoice.Period >= terms.Effective {
			return errors.New("Kỳ hiệu lực phải sau tất cả bảng kê đã phát hành của phòng.")
		}
	}
	b.Leases = append(b.Leases, terms)
	for j := range b.Invoices {
		i := &b.Invoices[j]
		if i.RoomID == terms.RoomID && !i.Issued && i.Period >= terms.Effective {
			*i = b.calculateInvoice(*room, i.Reading, i.ID)
		}
	}
	return nil
}
