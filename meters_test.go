package main

import (
	"errors"
	"testing"
)

func TestNewLeaseAfterVacancyUsesHandoverReading(t *testing.T) {
	room := Room{ID: "r", Name: "P01", Start: "2026-09", LeaseMonths: 1, Electric: 50, Water: 10, ElectricRate: 3000, WaterRate: 15000}
	b := Book{Rooms: []Room{room}}
	if err := b.scheduleLease(LeaseTerms{RoomID: "r", Effective: "2026-11", Tenant: "Người thuê mới", Rent: 2000000, ElectricRate: 3000, WaterRate: 15000}); err != nil {
		t.Fatal(err)
	}
	if err := b.read("r", "2026-10", 62, 15); err == nil {
		t.Fatal("vacant period accepted a billable reading")
	}
	if err := b.read("r", "2026-11", 70, 18); err == nil {
		t.Fatal("first period after vacancy needs a handover reading")
	}
	adjustment := MeterAdjustment{Handover: true, OpeningElectric: 64, OpeningWater: 16, Reason: "Biên bản nhận phòng"}
	if err := b.readWithAdjustment("r", "2026-11", 70, 18, adjustment); err != nil {
		t.Fatal(err)
	}
	if n := b.draft("2026-11"); n != 1 {
		t.Fatalf("draft count %d", n)
	}
	if i := b.Invoices[0]; i.ElectricAmount != 18000 || i.WaterAmount != 30000 || i.Total != 2048000 {
		t.Fatalf("wrong handover invoice: %+v", i)
	}
	if err := b.readWithAdjustment("r", "2026-12", 75, 20, adjustment); err == nil {
		t.Fatal("handover must not bypass an active unissued period")
	}
	if d := b.dashboard("2026-10"); d.Rooms != 0 || d.Missing != 0 {
		t.Fatalf("vacant room appears in the billing queue: %+v", d)
	}
}

func TestMeterReplacementCountsBothSegmentsAndStaysLocked(t *testing.T) {
	b := Book{Rooms: []Room{{ID: "r", Name: "P01", Start: "2026-11", Rent: 1000000, Electric: 120, Water: 20, ElectricRate: 3000, WaterRate: 15000}}}
	if err := b.read("r", "2026-11", 125, 22); err != nil {
		t.Fatal(err)
	}
	b.draft("2026-11")
	before := b.Invoices[0].Revision()
	adjustment := MeterAdjustment{Electric: MeterSwap{Replaced: true, OldEnd: 130, NewStart: 2}, Reason: "Thay công tơ điện"}
	if err := b.readWithAdjustment("r", "2026-11", 1, 22, adjustment); err == nil {
		t.Fatal("invalid new meter reading was accepted")
	}
	if b.Invoices[0].Revision() != before {
		t.Fatal("failed edit changed the draft")
	}
	if err := b.readWithAdjustment("r", "2026-11", 8, 22, adjustment); err != nil {
		t.Fatal(err)
	}
	if err := b.read("r", "2026-11", 9, 22); err != nil {
		t.Fatal(err)
	}
	i := &b.Invoices[0]
	if i.ElectricAmount != 51000 || i.WaterAmount != 30000 || i.Total != 1081000 || i.Revision() == before {
		t.Fatalf("wrong replacement invoice: %+v", i)
	}
	if err := b.issue(i.ID); err != nil {
		t.Fatal(err)
	}
	if err := b.readWithAdjustment("r", "2026-11", 10, 22, adjustment); !errors.Is(err, errLocked) {
		t.Fatalf("issued meter snapshot not protected: %v", err)
	}
	if err := b.read("r", "2026-12", 12, 23); err != nil {
		t.Fatal(err)
	}
	if r := b.reading("r", "2026-12"); r.ElectricBefore != 9 || r.ElectricUsage() != 3 {
		t.Fatalf("next period must carry new meter closing index: %+v", r)
	}
}
