package app

import (
	"strings"
	"testing"
)

func TestUnallocatedCash(t *testing.T) {
	b := Book{Invoices: []Invoice{{ID: "invoice", Issued: true, Total: 100000}}}
	key := token()
	for range 2 {
		if err := b.receiveCash(key, "2026-11-20", "PT01", "", 50000); err != nil {
			t.Fatal(err)
		}
	}
	if len(b.Transfers) != 1 || b.Invoices[0].Remaining() != 100000 {
		t.Fatal("an unallocated receipt was duplicated or changed the invoice")
	}
	for range 2 {
		if err := b.reconcile(key, "invoice", ""); err != nil {
			t.Fatal(err)
		}
	}
	if b.Invoices[0].Remaining() != 50000 || b.Invoices[0].Entries[0].Method != cashPayment {
		t.Fatal("allocated receipt was duplicated or lost its payment method")
	}
}

func TestVietQRKnownPayload(t *testing.T) {
	bank := BankAccount{BIN: "970407", Account: "19036688889999", Holder: "TEST"}
	got, err := vietQR(bank, 2500000, "HD0012")
	want := "00020101021238580010A000000727012800069704070114190366888899990208QRIBFTTA5303704540725000005802VN62100806HD00126304A938"
	if err != nil || got != want {
		t.Fatalf("unexpected payload %q, %v", got, err)
	}
	if _, err := vietQR(bank, 500000000, "HD0012"); err == nil {
		t.Fatal("out-of-range VietQR amount accepted")
	}
}

func TestStatementImportAndReview(t *testing.T) {
	b := Book{Bank: BankAccount{BIN: "970407", Account: "123456789", Holder: "TEST"}, Invoices: []Invoice{
		{ID: "invoice-1", Issued: true, PaymentCode: "ZS1234ABCD", Total: 1100},
	}}
	header := "ma_giao_dich,ngay,so_tien,noi_dung\n"
	text := header + "tx-01,2026-11-20,800,THU ZS1234ABCD\n"
	if _, _, err := b.importStatement(text + "tx-02,2026-11-20,0,loi\n"); err == nil || len(b.Transfers) != 0 {
		t.Fatal("invalid CSV modified the statement queue")
	}
	if newCount, _, err := b.importStatement(text); err != nil || newCount != 1 {
		t.Fatal(newCount, err)
	}
	if _, skipped, err := b.importStatement(text); err != nil || skipped != 1 {
		t.Fatal(skipped, err)
	}
	if _, _, err := b.importStatement(strings.Replace(text, ",800,", ",801,", 1)); err == nil {
		t.Fatal("duplicate ID with changed amount accepted")
	}
	tr := b.Transfers[0]
	if p := b.suggestInvoice(tr.Memo); p == nil || p.ID != "invoice-1" || b.Invoices[0].Remaining() != 1100 {
		t.Fatal("suggestion must not record payment")
	}
	if p := b.suggestInvoice(tr.Memo + " ZS5678EFGH"); p != nil {
		t.Fatal("ambiguous references must be reviewed manually")
	}
	if p := b.suggestInvoice("ABCZS1234ABCD"); p != nil {
		t.Fatal("a substring matched a payment code")
	}
	if err := b.reconcile(tr.ID, "invoice-1", ""); err != nil {
		t.Fatal(err)
	}
	if err := b.reconcile(tr.ID, "invoice-1", ""); err != nil {
		t.Fatal(err)
	}
	if i := b.Invoices[0]; i.Remaining() != 300 || len(i.Entries) != 1 {
		t.Fatal(i)
	}
	if err := b.reconcile(tr.ID, "other-invoice", ""); err == nil {
		t.Fatal("reviewed transfer was reassigned")
	}
}

func TestLinkPreviouslyRecordedPayment(t *testing.T) {
	b := Book{Bank: BankAccount{BIN: "970407", Account: "123456789", Holder: "TEST"}, Invoices: []Invoice{{ID: "i", Issued: true, Total: 1000}}}
	if err := b.post("i", Entry{Key: "payment-1", Kind: "Thu", Amount: 300, Note: "Chuyển khoản"}); err != nil {
		t.Fatal(err)
	}
	raw := "ma_giao_dich,ngay,so_tien,noi_dung\nref-a,2026-11-20,300,NO MEMO\nref-b,2026-11-20,300,NO MEMO\n"
	if _, _, err := b.importStatement(raw); err != nil {
		t.Fatal(err)
	}
	if err := b.reconcile(b.Transfers[0].ID, "i", "payment-1"); err != nil {
		t.Fatal(err)
	}
	if err := b.reconcile(b.Transfers[1].ID, "i", "payment-1"); err == nil {
		t.Fatal("one payment mapped to more than one bank transaction")
	}
	if b.Invoices[0].Remaining() != 700 || len(b.Invoices[0].Entries) != 1 {
		t.Fatal("linking recorded a second payment")
	}
	if err := b.reverse("i", "reverse-linked", "payment-1", "Nhập nhầm"); err == nil {
		t.Fatal("reversed a reconciled payment")
	}
	if err := b.post("i", Entry{Key: "payment-2", Kind: "Thu", Amount: 300, Note: "Chuyển khoản"}); err != nil {
		t.Fatal(err)
	}
	if err := b.reverse("i", "reverse-payment-2", "payment-2", "Nhập nhầm"); err != nil {
		t.Fatal(err)
	}
	if err := b.reconcile(b.Transfers[1].ID, "i", "payment-2"); err == nil {
		t.Fatal("matched a reversed payment")
	}
	if err := b.post("i", Entry{Key: "payment-3", Kind: "Thu", Amount: 300, Method: cashPayment}); err != nil {
		t.Fatal(err)
	}
	if err := b.reconcile(b.Transfers[1].ID, "i", "payment-3"); err == nil {
		t.Fatal("matched a cash payment to a bank transfer")
	}
}
