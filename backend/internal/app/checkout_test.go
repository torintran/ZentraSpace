package app

import (
	"strings"
	"testing"
	"time"
)

func TestCheckoutAppliesDepositAndProratesUnissuedInvoice(t *testing.T) {
	now := time.Date(2026, 11, 20, 0, 0, 0, 0, vietnam)
	b := Book{Rooms: []Room{{ID: "r", Name: "P1", Tenant: "Khách", Start: "2026-11", Rent: 3000000, ElectricRate: 4000, SiteID: "site"}},
		Sites:       []Site{{ID: "site", Name: "Dãy", Capacity: 1}},
		CashEntries: []CashEntry{{Key: strings.Repeat("a", 48), Kind: depositIn, RoomID: "r", Date: "2026-11-01", Amount: 2000000}}}
	if _, err := b.quoteCheckout(b.Rooms[0], "2026-11-15", now); err == nil {
		t.Fatal("checkout accepted without a final reading")
	}
	if err := b.read("r", "2026-11", 10, 2); err != nil {
		t.Fatal(err)
	}
	q, err := b.quoteCheckout(b.Rooms[0], "2026-11-15", now)
	if err != nil || q.Invoice.Rent != 1500000 || q.Remaining != 1540000 {
		t.Fatalf("quote = %+v, %v", q, err)
	}
	c := Checkout{RoomID: "r", Effective: q.Effective, Date: q.Date, Revision: q.Revision, Note: "Đã bàn giao", Applied: 1000000, Refunded: 1000000}
	if err := b.closeCheckout(c, now); err != nil {
		t.Fatal(err)
	}
	if err := b.closeCheckout(c, now); err != nil {
		t.Fatal(err)
	}
	i := b.Invoices[0]
	if !i.Issued || i.Remaining() != 540000 || i.cashCollected() != 0 || len(i.Entries) != 1 || len(b.Checkouts) != 1 {
		t.Fatalf("invoice = %+v", i)
	}
	if b.leaseAt(b.Rooms[0], "2026-12").ActiveAt("2026-12") {
		t.Fatal("room is not vacant")
	}
	last := b.leaseAt(b.Rooms[0], "2026-11")
	if !last.ActiveAt("2026-11") || last.OccupiedIn("2026-11") || len(b.roomResidents(b.Rooms[0], "2026-11")) != 0 || b.siteViews("2026-11")[0].Occupied != 0 {
		t.Fatal("room is still occupied in the checkout month or its final billing period was lost")
	}
	if report := b.cashReport("2026-11"); report.Held != 0 || report.Returned != 1000000 || report.Applied != 1000000 || report.RentNet != 0 {
		t.Fatalf("report = %+v", report)
	}
	if err := b.post(i.ID, Entry{Key: "refund-test", Kind: "Hoàn", Amount: 1, Note: "Không hoàn phần cấn"}); err == nil {
		t.Fatal("deposit applied was refundable as cash")
	}
}

func TestCheckoutPreservesIssuedAmountAndRejectsStalePreview(t *testing.T) {
	now := time.Date(2026, 11, 20, 0, 0, 0, 0, vietnam)
	b := Book{Rooms: []Room{{ID: "r", Tenant: "Khách", Start: "2026-11", Rent: 3000000}}}
	if err := b.read("r", "2026-11", 0, 0); err != nil {
		t.Fatal(err)
	}
	b.draft("2026-11")
	id := b.Invoices[0].ID
	if err := b.issue(id); err != nil {
		t.Fatal(err)
	}
	q, err := b.quoteCheckout(b.Rooms[0], "2026-11-10", now)
	if err != nil || q.Invoice.Rent != 3000000 {
		t.Fatalf("issued quote = %+v, %v", q, err)
	}
	if err := b.post(id, Entry{Kind: "Thu", Key: "external-payment", Amount: 1000}); err != nil {
		t.Fatal(err)
	}
	if err := b.closeCheckout(Checkout{RoomID: "r", Effective: q.Effective, Date: q.Date, Revision: q.Revision, Note: "Bàn giao"}, now); err == nil {
		t.Fatal("stale preview accepted")
	}
}

func TestExpiringLeaseExcludesRenewedAndReturned(t *testing.T) {
	today := time.Date(2026, 11, 27, 0, 0, 0, 0, vietnam)
	b := Book{Rooms: []Room{
		{ID: "soon", Start: "2026-10", Tenant: "A", LeaseMonths: 2},
		{ID: "renewed", Start: "2026-10", Tenant: "B", LeaseMonths: 2},
		{ID: "ended", Start: "2026-10", Tenant: "C", LeaseMonths: 3},
	}, Leases: []LeaseTerms{{RoomID: "renewed", Effective: "2026-12", Tenant: "B", Months: 12}},
		Checkouts: []Checkout{{RoomID: "ended", Effective: "2026-10", Date: "2026-11-20"}}}
	results := b.expiringLeases(today, "")
	if len(results) != 1 || results[0].RoomID != "soon" || results[0].NextPeriod != "2026-12" {
		t.Fatalf("results = %+v", results)
	}
}

func TestSiteEditRevalidatesCapacityAndRevision(t *testing.T) {
	b := Book{Sites: []Site{{ID: "a", Name: "A", Capacity: 2}, {ID: "b", Name: "B", Capacity: 1}}, Rooms: []Room{{SiteID: "a"}, {SiteID: "a"}}}
	revision := b.Sites[0].Revision()
	if err := b.updateSite("a", revision, "B", 2); err == nil {
		t.Fatal("duplicate accepted")
	}
	if err := b.updateSite("a", revision, "A", 1); err == nil {
		t.Fatal("capacity below occupied roster accepted")
	}
	if err := b.updateSite("a", revision, "C", 3); err != nil {
		t.Fatal(err)
	}
	if err := b.updateSite("a", revision, "D", 3); err == nil {
		t.Fatal("stale edit accepted")
	}
}
