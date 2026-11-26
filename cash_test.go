package main

import "testing"

func TestCashReportAndBackdatedRefund(t *testing.T) {
	b := Book{Rooms: []Room{{ID: "room", Name: "P1", Start: "2026-11"}}}
	receive := CashEntry{Key: token(), Kind: depositIn, Date: "2026-11-20", RoomID: "room", Amount: 1000}
	refund := CashEntry{Key: token(), Kind: depositOut, Date: "2026-12-01", RoomID: "room", Amount: 600, Note: "Trả cọc"}
	cost := CashEntry{Key: token(), Kind: expense, Date: "2026-12-02", Category: "Bảo trì", Amount: 300}
	for _, e := range []CashEntry{receive, receive, refund, cost} {
		if err := b.postCash(e); err != nil {
			t.Fatal(err)
		}
	}
	if err := b.postCash(CashEntry{Key: token(), Kind: depositOut, Date: "2026-11-21", RoomID: "room", Amount: 500, Note: "Điều chỉnh"}); err == nil {
		t.Fatal("backdated refund would overdraw deposits after another refund")
	}
	first, second := b.cashReport("2026-11"), b.cashReport("2026-12")
	if first.Held != 1000 || first.Received != 1000 || second.Held != 400 || second.Expense != 300 ||
		second.Received != 0 || second.Returned != 600 || second.RentNet != 0 || len(second.Rows) != 2 {
		t.Fatalf("unexpected cash reports: %+v, %+v", first, second)
	}
}
