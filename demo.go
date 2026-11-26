package main

import (
	"errors"
	"fmt"
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
