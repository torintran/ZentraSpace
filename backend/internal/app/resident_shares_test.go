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

func TestResidentPaymentBalancesAndReversals(t *testing.T) {
	b := Book{Invoices: []Invoice{{
		ID: "i", Issued: true, Total: 100,
		ResidentShares: []ResidentShare{{Name: "A", Amount: 60}, {Name: "B", Amount: 40}},
	}}}
	payment := Entry{Key: "payment-a", Kind: "Thu", Amount: 25, Method: cashPayment, ResidentShare: 1}
	for _, entry := range []Entry{payment, payment,
		{Key: "shared-pay", Kind: "Thu", Amount: 10},
		{Key: "reduction", Kind: "Giảm", Amount: 5, Note: "B", ResidentShare: 2},
	} {
		if err := b.post("i", entry); err != nil {
			t.Fatal(err)
		}
	}
	payment.ResidentShare = 2
	if err := b.post("i", payment); err == nil {
		t.Fatal("a payment replay changed its resident")
	}
	payment.Key, payment.ResidentShare = "unknown-resident", 3
	if err := b.post("i", payment); err == nil {
		t.Fatal("accepted a nonexistent share")
	}
	if err := b.post("i", Entry{Key: "refund-b", Kind: "Hoàn", Amount: 1, Note: "B", ResidentShare: 2}); err == nil {
		t.Fatal("accepted a resident refund against another resident's payment")
	}
	check := func(a, second, remaining int64) {
		t.Helper()
		view := invoicePresentation(b.invoice("i"), false)
		if view.ResidentBalances[0] != a || view.ResidentBalances[1] != second ||
			view.ResidentUnallocated != 10 || view.Remaining != remaining ||
			a+second-view.ResidentUnallocated != remaining {
			t.Fatalf("unexpected balances: %+v", view)
		}
	}
	check(35, 35, 60)
	if err := b.reverse("i", "reverse-a", "payment-a", "Correction"); err != nil {
		t.Fatal(err)
	}
	check(60, 35, 85)
	shared := invoicePresentation(b.invoice("i"), true)
	if len(shared.Entries) != 0 || len(shared.ResidentBalances) != 0 || shared.ResidentUnallocated != 0 {
		t.Fatal("shared page leaked resident ledger")
	}
}

func TestPersonalShareEmailChangeRevokesOnlyItsInvites(t *testing.T) {
	b := Book{
		LeaseEmails: map[string]string{"r/2026-11": "representative@example.test"},
		Residents: []Resident{
			{ID: "a", RoomID: "r", Effective: "2026-11", Email: "old@example.test"},
			{ID: "b", RoomID: "r", Effective: "2026-11", Email: "other@example.test"},
		},
		Invoices: []Invoice{{ID: "i", RoomID: "r", LeaseEffective: "2026-11", Issued: true,
			ResidentShares: []ResidentShare{{Name: "Đại diện"}, {ResidentID: "a"}, {ResidentID: "b"}}}},
		Invites: []Invite{{InvoiceID: "i"}, {InvoiceID: "i", ResidentShare: 2}, {InvoiceID: "i", ResidentShare: 3}},
	}
	if err := b.setResidentEmail("r", "a", "new@example.test"); err != nil {
		t.Fatal(err)
	}
	i := b.invoice("i")
	if b.verificationAddress(i, 1) != "representative@example.test" ||
		b.verificationAddress(i, 2) != "new@example.test" ||
		b.verificationAddress(i, 3) != "other@example.test" {
		t.Fatal("invites did not resolve to the recorded recipient")
	}
	if b.Invites[0].Revoked || !b.Invites[1].Revoked || b.Invites[2].Revoked {
		t.Fatal("email change revoked invitations outside its resident scope")
	}
}
