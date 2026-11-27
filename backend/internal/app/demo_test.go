package app

import "testing"

func TestLeaseDemoKeepsTermsAndHistory(t *testing.T) {
	var b Book
	const period = "2026-11"
	if err := b.seedLeaseDemo(period); err != nil {
		t.Fatal(err)
	}
	if len(b.Rooms) != 6 || len(b.Leases) != 2 || len(b.Invoices) < 20 {
		t.Fatal("missing demo rooms, renewals, or billing history")
	}
	if l := b.leaseAt(b.Rooms[3], period); l.Months != 18 || !l.ActiveAt(period) {
		t.Fatal("missing active eighteen-month contract")
	}
	room := b.Rooms[4]
	history := b.leaseHistory(room)
	if len(history) != 2 || history[0].Months != 6 || history[0].Until() != "2026-10" ||
		history[1].Months != 12 || history[1].Effective != period || history[1].Rent <= history[0].Rent {
		t.Fatalf("incorrect renewal history: %+v", history)
	}
	before := b.Invoices[0]
	count := len(b.Invoices)
	if err := b.seedLeaseDemo(period); err != nil {
		t.Fatal(err)
	}
	if len(b.Rooms) != 6 || len(b.Invoices) != count || b.Invoices[0].ID != before.ID || b.Invoices[0].Total != before.Total {
		t.Fatal("re-seeding changed existing records")
	}
}

func TestFixedLeasePreventsOutOfTermBilling(t *testing.T) {
	r := Room{ID: "room", Start: "2026-01", LeaseMonths: 6, Rent: 1000}
	b := Book{Rooms: []Room{r}, Readings: []Reading{{RoomID: r.ID, Period: "2026-07"}}}
	if b.draft("2026-07") != 0 {
		t.Fatal("drafted an invoice after the contract ended")
	}
	terms := LeaseTerms{RoomID: r.ID, Tenant: "Khách thuê mẫu", Effective: "2026-06", Months: 12, Rent: 1100}
	if err := b.scheduleLease(terms); err == nil {
		t.Fatal("scheduled overlapping fixed-duration contracts")
	}
	terms.Effective = "2026-07"
	if err := b.scheduleLease(terms); err != nil {
		t.Fatal(err)
	}
	if b.draft("2026-07") != 1 || b.Invoices[0].Total != 1100 {
		t.Fatal("renewal terms were not applied")
	}
}
