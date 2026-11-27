package app

import (
	"context"
	"net/netip"
	"testing"
)

func TestArchivePaginationClampsPage(t *testing.T) {
	events := make([]RoomEvent, 29)
	page, pages := 999, 0
	if got := paginateArchive(events, &page, &pages); len(got) != 5 || page != 3 || pages != 3 {
		t.Fatalf("last page: %d entries, %d/%d", len(got), page, pages)
	}
	if got := paginateArchive(events[:0], &page, &pages); len(got) != 0 || page != 1 || pages != 1 {
		t.Fatalf("empty page: %d entries, %d/%d", len(got), page, pages)
	}
}

func TestTimelineFiltersByEventPeriodAndKind(t *testing.T) {
	room := Room{ID: "r", Start: "2026-10", Tenant: "Khách trước"}
	b := Book{Checkouts: []Checkout{{RoomID: "r", Effective: "2026-10", Date: "2026-11-15"}}}
	if got := b.roomTimeline(room, "2026-11", "checkout"); len(got) != 1 || got[0].Date != "2026-11-15" {
		t.Fatalf("unexpected checkout history: %+v", got)
	}
	if got := b.roomTimeline(room, "2026-10", "checkout"); len(got) != 0 {
		t.Fatalf("checkout outside filtered month: %+v", got)
	}
}

func TestSMTPFromUIRejectsInternalDestinations(t *testing.T) {
	for _, example := range []string{"127.0.0.1", "::ffff:127.0.0.1", "10.2.3.4", "169.254.169.254", "100.64.0.1", "fd00::1"} {
		if smtpPublicIP(netip.MustParseAddr(example)) {
			t.Fatalf("allowed %s", example)
		}
	}
	m := &verificationSMTP{address: "127.0.0.1:465", publicOnly: true}
	if _, err := m.destination(context.Background()); err == nil {
		t.Fatal("accepted local SMTP address from UI")
	}
}
