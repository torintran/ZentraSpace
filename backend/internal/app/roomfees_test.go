package app

import "testing"

func TestRoomFeesApplyToTheRightPeriod(t *testing.T) {
	room := Room{ID: "room", Name: "P1", Start: "2026-11", Rent: 1000}
	b := Book{Rooms: []Room{room}}
	monthly := RoomFee{ID: token(), RoomID: room.ID, Label: "Wi-Fi", Start: "2026-11", Monthly: true, Amount: 80}
	once := RoomFee{ID: token(), RoomID: room.ID, Label: "Phụ thu", Start: "2026-11", Amount: 35}
	for _, fee := range []RoomFee{monthly, once, monthly} {
		if err := b.addRoomFee(fee); err != nil {
			t.Fatal(err)
		}
	}
	first := b.calculateInvoice(room, Reading{Period: "2026-11"}, token())
	next := b.calculateInvoice(room, Reading{Period: "2026-12"}, token())
	if first.Total != 1115 || len(first.RoomFees) != 2 || next.Total != 1080 {
		t.Fatalf("unexpected fee recurrence: %d and %d", first.Total, next.Total)
	}
	first.Issued = true
	b.Invoices = []Invoice{first, next}
	if err := b.stopRoomFee(room.ID, once.ID, ""); err == nil {
		t.Fatal("cancelled a fee on an issued invoice")
	}
	if err := b.stopRoomFee(room.ID, monthly.ID, "2026-12"); err != nil {
		t.Fatal(err)
	}
	if b.Invoices[0].Total != 1115 || b.Invoices[1].Total != 1000 || b.Invoices[1].Revision() == next.Revision() {
		t.Fatal("stopping the fee did not preserve the issued invoice and refresh the draft")
	}
}
