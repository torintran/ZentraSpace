package app

import (
	"testing"
	"time"
)

func TestSiteMeterReviewHoldsOnlyAffectedDraftsAndRecordsActor(t *testing.T) {
	const period = "2026-11"
	b := Book{Sites: []Site{{ID: "affected", Name: "Dãy A"}, {ID: "other", Name: "Dãy B"}}, Rooms: []Room{
		{ID: "p1", Name: "P1", SiteID: "affected", Start: period},
		{ID: "p2", Name: "P2", SiteID: "other", Start: period},
	}}
	for _, room := range b.Rooms {
		if err := b.read(room.ID, period, 10, 4); err != nil {
			t.Fatal(err)
		}
	}
	reading := SiteMeterReading{SiteID: "affected", Period: period, ElectricMeter: "E1", WaterMeter: "W1", ElectricAfter: 15, WaterAfter: 4}
	if err := b.saveSiteReading(reading, ""); err != nil {
		t.Fatal(err)
	}
	check := b.checkSiteMeters(b.Sites[0], period)
	if !check.NeedsReview || check.ElectricDelta != 5 || check.WaterDelta != 0 {
		t.Fatalf("initial check: %+v", check)
	}
	if count := b.draft(period); count != 1 || b.Invoices[0].RoomID != "p2" {
		t.Fatal("draft affected other sites")
	}
	if err := b.read("p1", period, 11, 4); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 27, 0, 0, 0, 0, vietnam)
	if err := b.approveSiteMeters("affected", period, check.Revision, "owner", "Điện khu vực chung", now); err == nil {
		t.Fatal("stale approval accepted")
	}
	check = b.checkSiteMeters(b.Sites[0], period)
	if err := b.approveSiteMeters("affected", period, check.Revision, "owner", "Điện khu vực chung", now); err != nil {
		t.Fatal(err)
	}
	if len(b.SiteMeterApprovals) != 1 || b.SiteMeterApprovals[0].Reviewer != "owner" || b.draft(period) != 1 {
		t.Fatal("approval did not release draft or record reviewer")
	}
	id := b.Invoices[1].ID
	reading.ElectricAfter++
	if err := b.saveSiteReading(reading, check.Revision); err != nil {
		t.Fatal(err)
	}
	if err := b.issue(id); err == nil || len(b.SiteMeterApprovals) != 1 {
		t.Fatal("new discrepancy did not hold issue or preserve prior review")
	}
}
