package app

import (
	"testing"
	"time"
)

func TestMidmonthLeaseRequiresHandoverAndProratesBothBounds(t *testing.T) {
	room := Room{ID: "r", Name: "P1", Start: "2028-02", Prepared: true}
	b := Book{Rooms: []Room{room}}
	lease := LeaseTerms{RoomID: room.ID, Effective: "2028-02", StartOn: "2028-02-15", Tenant: "Khách", Rent: 2900000, ElectricRate: 4000, WaterRate: 15000}
	if err := b.scheduleLease(lease); err != nil {
		t.Fatal(err)
	}
	if err := b.scheduleLease(lease); err != nil {
		t.Fatal(err)
	}
	if err := b.read(room.ID, "2028-02", 110, 22); err == nil {
		t.Fatal("meter reading accepted without handover")
	}
	if err := b.readWithAdjustment(room.ID, "2028-02", 110, 22, MeterAdjustment{Handover: true, OpeningElectric: 100, OpeningWater: 20, Reason: "Bàn giao"}); err != nil {
		t.Fatal(err)
	}
	b.draft("2028-02")
	invoice := b.Invoices[0]
	if invoice.Rent != 1500000 || invoice.Total != 1570000 || invoice.RentDays != 15 || invoice.RentMonthDays != 29 {
		t.Fatalf("draft: %+v", invoice)
	}
	now := time.Date(2028, 2, 25, 0, 0, 0, 0, vietnam)
	if _, err := b.quoteCheckout(room, "2028-02-14", now); err == nil {
		t.Fatal("checkout before move-in accepted")
	}
	quote, err := b.quoteCheckout(room, "2028-02-20", now)
	if err != nil || quote.Invoice.Rent != 600000 || quote.Invoice.RentDays != 6 || quote.Invoice.Total != 670000 {
		t.Fatalf("quote: %+v, %v", quote, err)
	}
}

func TestMidmonthLeaseCannotReplaceAnExistingOccupant(t *testing.T) {
	b := Book{Rooms: []Room{{ID: "r", Start: "2028-01", Tenant: "Người đang thuê"}}}
	if err := b.scheduleLease(LeaseTerms{RoomID: "r", Effective: "2028-02", StartOn: "2028-02-15", Tenant: "Khác"}); err == nil {
		t.Fatal("overlapping mid-month lease accepted")
	}
}
