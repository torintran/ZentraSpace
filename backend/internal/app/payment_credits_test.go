package app

import (
	"strings"
	"testing"
	"time"
)

func TestCreditTransferAndReversalDoNotCreateCash(t *testing.T) {
	b := Book{Invoices: []Invoice{
		{ID: "source", RoomID: "r", LeaseEffective: "2026-10", Period: "2026-11", Total: 100, Issued: true, Entries: []Entry{{Key: "cash-receipt", Kind: "Thu", Amount: 160}}},
		{ID: "target", RoomID: "r", LeaseEffective: "2026-10", Period: "2026-12", Total: 80, Issued: true},
		{ID: "other", RoomID: "r", LeaseEffective: "2027-01", Period: "2027-01", Total: 80, Issued: true},
	}}
	now := time.Date(2026, 12, 30, 0, 0, 0, 0, vietnam)
	move := CreditTransfer{Key: strings.Repeat("a", 48), SourceID: "source", TargetID: "other", Amount: 40}
	if err := b.postCreditTransfer(move, now); err == nil {
		t.Fatal("transferred deposit across leases")
	}
	move.TargetID = "target"
	if err := b.postCreditTransfer(move, now); err != nil {
		t.Fatal(err)
	}
	if err := b.postCreditTransfer(move, now); err != nil {
		t.Fatal(err)
	}
	if len(b.CreditTransfers) != 1 || b.Invoices[0].transferableCredit() != 20 || b.Invoices[1].Remaining() != 40 {
		t.Fatal("incorrect credit after move")
	}
	if b.Invoices[0].cashCollected() != 160 || b.Invoices[1].cashCollected() != 0 {
		t.Fatal("allocation changed cash flow")
	}
	if err := b.post("source", Entry{Key: "cash-refund", Kind: "Hoàn", Amount: 121, Note: "Thử hoàn"}); err == nil {
		t.Fatal("allocated cash was refunded")
	}
	undo := move
	undo.Key, undo.Reverses, undo.Note = strings.Repeat("b", 48), move.Key, "Sửa giao dịch"
	if err := b.postCreditTransfer(undo, now); err != nil {
		t.Fatal(err)
	}
	if err := b.postCreditTransfer(undo, now); err != nil {
		t.Fatal(err)
	}
	if len(b.CreditTransfers) != 2 || b.Invoices[0].transferableCredit() != 60 || b.Invoices[1].Remaining() != 80 {
		t.Fatal("incorrect credit after reversal")
	}
	undo.Key = strings.Repeat("c", 48)
	if err := b.postCreditTransfer(undo, now); err == nil {
		t.Fatal("duplicate reversal accepted")
	}
}
