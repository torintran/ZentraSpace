package app

import "testing"

func TestResidentDaysSnapshotAndRounding(t *testing.T) {
	room := Room{ID: "r", Name: "P1", Start: "2026-11", Tenant: "A", Rent: 101}
	b := Book{Rooms: []Room{room}, Residents: []Resident{
		{ID: "b", RoomID: "r", Effective: "2026-11", Name: "B", JoinedOn: "2026-11-11", LeftOn: "2026-11-21"},
	}}
	if err := b.setResidentPolicy(room, "2026-11", "days"); err != nil {
		t.Fatal(err)
	}
	b.Invoices = []Invoice{b.calculateInvoice(room, Reading{Period: "2026-11"}, "i")}
	got := b.Invoices[0].ResidentShares
	if len(got) != 2 || got[0].Days != 30 || got[1].Days != 10 || got[0].Amount != 76 || got[1].Amount != 25 {
		t.Fatalf("unexpected shares: %+v", got)
	}
	if err := b.issue("i"); err != nil {
		t.Fatal(err)
	}
	b.Residents[0].LeftOn = "2026-11-25"
	b.refreshRoomFeeDrafts(room)
	if b.Invoices[0].ResidentShares[1] != got[1] {
		t.Fatal("issued snapshot changed")
	}
	if invoicePresentation(&b.Invoices[0], true).ResidentShares != nil {
		t.Fatal("shared page leaked residents")
	}
	negative := []ResidentShare{{Days: 30}, {Days: 10}}
	allocateResidentShares(-101, negative)
	if negative[0].Amount != -76 || negative[1].Amount != -25 {
		t.Fatalf("unexpected negative shares: %+v", negative)
	}
}
