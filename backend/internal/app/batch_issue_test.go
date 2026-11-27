package app

import "testing"

func TestIssueBatchIsAtomicAndCanBeRetried(t *testing.T) {
	const period = "2026-11"
	b := Book{
		Rooms: []Room{{ID: "r1", Start: period}, {ID: "r2", Start: period}},
		Invoices: []Invoice{
			{ID: "i1", RoomID: "r1", Period: period},
			{ID: "i2", RoomID: "r2", Period: period},
		},
	}
	first := "i1:" + b.Invoices[0].Revision()
	second := "i2:" + b.Invoices[1].Revision()
	if err := b.issueBatch(period, []string{first, "i2:stale"}); err == nil || b.Invoices[0].Issued {
		t.Fatal("a stale selection must prevent the entire batch")
	}
	if err := b.issueBatch(period, []string{first, first}); err == nil || b.Invoices[0].Issued {
		t.Fatal("duplicate selections must prevent the entire batch")
	}
	for range 2 {
		if err := b.issueBatch(period, []string{first, second}); err != nil {
			t.Fatal(err)
		}
	}
	if !b.Invoices[0].Issued || !b.Invoices[1].Issued || b.Invoices[0].IssuedOn != b.Invoices[1].IssuedOn {
		t.Fatal("expected both invoices to be issued in the batch")
	}
}
