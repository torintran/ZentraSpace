package main

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

func (b *Book) seedDemo(period string) error {
	if _, err := time.Parse("2006-01", period); err != nil {
		return errors.New("Demo period must use YYYY-MM.")
	}
	for index, state := range []string{"paid", "partial", "draft", "missing"} {
		name := fmt.Sprintf("DEMO-%s-P%02d", period, index+1)
		id := hash("demo-room:" + name)[:48]
		if b.room(id) != nil {
			continue
		}
		for _, room := range b.Rooms {
			if strings.EqualFold(room.Name, name) {
				return fmt.Errorf("Demo room code %s is already in use.", name)
			}
		}
		if len(b.Rooms) >= 150 {
			return errors.New("The local pilot supports up to 150 rooms.")
		}
		room := Room{ID: id, Name: name, Tenant: fmt.Sprintf("Người thuê mẫu %d", index+1),
			Start: period, Rent: 2000000 + int64(index)*300000, ElectricRate: 3500, WaterRate: 15000,
			Electric: 100, Water: 20}
		b.Rooms = append(b.Rooms, room)
		if state == "missing" {
			continue
		}
		if err := b.read(id, period, 110, 22); err != nil {
			return err
		}
		invoice := b.calculateInvoice(room, *b.reading(id, period), token())
		b.Invoices = append(b.Invoices, invoice)
		if state == "draft" {
			continue
		}
		if err := b.issue(invoice.ID); err != nil {
			return err
		}
		amount := invoice.Total
		if state == "partial" {
			amount = 1000000
		}
		if err := b.post(invoice.ID, Entry{Key: "demo-payment", Kind: "Thu", Amount: amount, Method: cashPayment, Note: "Giao dịch mô phỏng"}); err != nil {
			return err
		}
	}
	return nil
}

func (b *Book) seedLeaseDemo(period string) error {
	month, err := time.Parse("2006-01", period)
	if err != nil {
		return errors.New("Demo period must use YYYY-MM.")
	}
	cases := []struct {
		age, initial, renewal int
		state                 string
	}{
		{2, 6, 0, "paid"},
		{5, 6, 0, "partial"},
		{3, 12, 0, "draft"},
		{4, 18, 0, "missing"},
		{6, 6, 12, "partial"},
		{12, 12, 6, "paid"},
	}
	for index, sample := range cases {
		name := fmt.Sprintf("HD-MAU-%s-P%02d", period, index+1)
		id := hash("lease-demo:" + name)[:48]
		if b.room(id) != nil {
			continue
		}
		for _, r := range b.Rooms {
			if strings.EqualFold(r.Name, name) {
				return fmt.Errorf("Demo room code %s is already in use.", name)
			}
		}
		if len(b.Rooms) >= 150 {
			return errors.New("The local pilot supports up to 150 rooms.")
		}
		room := Room{ID: id, Name: name, Tenant: fmt.Sprintf("Khách thuê mẫu %02d", index+1),
			Start: month.AddDate(0, -sample.age, 0).Format("2006-01"), LeaseMonths: sample.initial,
			Rent: int64(1700000 + 100000*rand.IntN(17)), ElectricRate: 3500, WaterRate: 15000,
			Electric: int64(50 + rand.IntN(150)), Water: int64(5 + rand.IntN(20))}
		b.Rooms = append(b.Rooms, room)
		if sample.renewal > 0 {
			if err := b.scheduleLease(LeaseTerms{RoomID: id, Effective: period, Tenant: room.Tenant,
				Rent: room.Rent + 100000, ElectricRate: room.ElectricRate, WaterRate: room.WaterRate,
				Months: sample.renewal, Renewal: true}); err != nil {
				return err
			}
		}
		electric, water := room.Electric, room.Water
		for offset := -sample.age; offset <= 0; offset++ {
			if offset == 0 && sample.state == "missing" {
				break
			}
			date := month.AddDate(0, offset, 0)
			billPeriod := date.Format("2006-01")
			electric += int64(18 + rand.IntN(46))
			water += int64(2 + rand.IntN(6))
			if err := b.read(id, billPeriod, electric, water); err != nil {
				return err
			}
			i := b.calculateInvoice(room, *b.reading(id, billPeriod), token())
			b.Invoices = append(b.Invoices, i)
			if offset == 0 && sample.state == "draft" {
				break
			}
			if err := b.issue(i.ID); err != nil {
				return err
			}
			amount := i.Total
			if (offset == 0 && sample.state == "partial") || (offset < 0 && rand.IntN(4) == 0) {
				amount = (i.Total / 2 / 1000) * 1000
			}
			if err := b.post(i.ID, Entry{Key: "lease-demo-payment", Kind: "Thu", Amount: amount,
				Method: cashPayment, Note: "Giao dịch mô phỏng"}); err != nil {
				return err
			}
			b.invoice(i.ID).Entries[0].Date = date.AddDate(0, 0, 20)
		}
	}
	return nil
}
