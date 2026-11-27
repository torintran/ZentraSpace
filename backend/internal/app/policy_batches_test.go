package app

import "testing"

func TestPolicyBatchPreviewAndInvoiceTerms(t *testing.T) {
	room := Room{ID: "first", SiteID: "site", Prepared: true, Start: "2026-11", Rent: 700, PolicyFeeLabel: "Phí cũ", PolicyFee: 3}
	second := Room{ID: "second", SiteID: "site", Prepared: true, Start: "2026-11"}
	occupied := Room{ID: "occupied", SiteID: "site", Start: "2026-11", Tenant: "Khách mẫu"}
	policy := RoomPolicy{ID: "policy", Name: "Mẫu mới", Rent: 900, FeeLabel: "Wi-Fi", Fee: 25}
	book := Book{Sites: []Site{{ID: "site", Name: "Dãy A"}}, Rooms: []Room{room, second, occupied}, Policies: []RoomPolicy{policy}}
	preview, err := book.previewPolicy("site", policy.ID, "2026-12")
	if err != nil || preview.Eligible != 2 || preview.Rooms[2].Blocked == "" || len(book.PolicyAssignments) != 0 {
		t.Fatalf("preview should be read-only and exclude the occupied room: %#v, %v", preview, err)
	}
	book.Rooms[1].Rent++
	if err := book.applyPolicy("site", policy.ID, "2026-12", preview.Revision, "2026-11-27"); err == nil || len(book.PolicyAssignments) != 0 {
		t.Fatal("the old preview should reject the batch after a room price changes")
	}
	preview, err = book.previewPolicy("site", policy.ID, "2026-12")
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := book.applyPolicy("site", policy.ID, "2026-12", preview.Revision, "2026-11-27"); err != nil {
			t.Fatal(err)
		}
	}
	if len(book.PolicyAssignments) != 2 {
		t.Fatal("retry should preserve one assignment per eligible room")
	}
	if book.leaseAt(room, "2026-11").Rent != 700 || book.leaseAt(room, "2026-12").Rent != 900 {
		t.Fatal("the new list price should begin at the selected period")
	}
	if err := book.scheduleLease(LeaseTerms{RoomID: room.ID, Effective: "2026-11", Tenant: "Vào trước", Rent: 700}); err != nil {
		t.Fatal(err)
	}
	if err := book.scheduleLease(LeaseTerms{RoomID: second.ID, Effective: "2026-12", Tenant: "Vào sau", Rent: 900}); err != nil {
		t.Fatal(err)
	}
	book.Policies[0].Fee = 80
	oldLease := book.calculateInvoice(room, Reading{Period: "2026-12"}, "old")
	newLease := book.calculateInvoice(second, Reading{Period: "2026-12"}, "new")
	if oldLease.Total != 703 || newLease.Total != 925 {
		t.Fatalf("billing must use the fee captured on each contract, got %d and %d", oldLease.Total, newLease.Total)
	}
}
