package main

import (
	"testing"
	"time"
)

func TestBillingAndReplay(t *testing.T) {
	var b Book
	b.Rooms = []Room{{ID: "r", Name: "101", Rent: 2500000, ElectricRate: 3000, WaterRate: 15000, Electric: 10, Water: 20, Start: "2026-11"}}
	if err := b.read("r", "2026-11", 33, 26); err != nil {
		t.Fatal(err)
	}
	if n := b.draft("2026-11"); n != 1 {
		t.Fatal(n)
	}
	if n := b.draft("2026-11"); n != 0 {
		t.Fatal(n)
	}
	i := &b.Invoices[0]
	if i.Total != 2659000 {
		t.Fatal(i.Total)
	}
	if err := b.issue(i.ID); err != nil {
		t.Fatal(err)
	}
	if err := b.issue(i.ID); err != nil {
		t.Fatal(err)
	}
	if err := b.read("r", "2026-11", 35, 26); err == nil {
		t.Fatal("issued reading changed")
	}
	if err := b.post(i.ID, "payment-1", "Thu", 500000, "Tiền mặt"); err != nil {
		t.Fatal(err)
	}
	if err := b.post(i.ID, "payment-1", "Thu", 500000, "Tiền mặt"); err != nil {
		t.Fatal(err)
	}
	if i.Remaining() != 2159000 || len(i.Entries) != 1 {
		t.Fatal(i.Remaining(), i.Entries)
	}
	if err := b.post(i.ID, "payment-1", "Thu", 400000, "Tiền mặt"); err == nil {
		t.Fatal("conflicting replay")
	}
	if err := b.post(i.ID, "discount", "Giảm", 100000, "Điều chỉnh theo thỏa thuận"); err != nil {
		t.Fatal(err)
	}
	if i.Total != 2659000 || i.Remaining() != 2059000 {
		t.Fatal(i.Total, i.Remaining())
	}
}

func TestCatchUpAndOverpayments(t *testing.T) {
	b := Book{Rooms: []Room{{ID: "r", Rent: 1000, Start: "2026-11", Electric: 1, Water: 2}}}
	if err := b.read("r", "2026-12", 5, 5); err == nil {
		t.Fatal("skipped cycle")
	}
	if err := b.read("r", "2026-11", 2, 4); err != nil {
		t.Fatal(err)
	}
	b.draft("2026-11")
	i := &b.Invoices[0]
	if err := b.post(i.ID, "unissued-1", "Thu", 1, ""); err == nil {
		t.Fatal("draft collected")
	}
	if err := b.issue(i.ID); err != nil {
		t.Fatal(err)
	}
	if err := b.post(i.ID, "overpaid-1", "Thu", 1300, "Chuyển khoản"); err != nil {
		t.Fatal(err)
	}
	if i.Remaining() != 0 || i.Credit() != 300 {
		t.Fatal(i.Remaining(), i.Credit())
	}
	if err := b.read("r", "2026-12", 3, 5); err != nil {
		t.Fatal(err)
	}
}

func TestInviteRequiresCodeAndExpires(t *testing.T) {
	now := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)
	v := Invite{CodeHash: hash("12345678"), Expires: now.Add(time.Hour)}
	if v.verify("00000000", now) || !v.verify("12345678", now) || v.verify("12345678", now) {
		t.Fatal("code replay")
	}
	v.Used = false
	if v.verify("12345678", now.Add(time.Hour)) {
		t.Fatal("expired")
	}
}

func TestEncryptedStorage(t *testing.T) {
	key := make([]byte, 32)
	sealed, err := encrypt(key, "org-a", []byte("private figure 12345"))
	if err != nil {
		t.Fatal(err)
	}
	if string(sealed) == "private figure 12345" {
		t.Fatal("plaintext")
	}
	if _, err := decrypt(key, "org-b", sealed); err == nil {
		t.Fatal("cross organization ciphertext")
	}
	opened, err := decrypt(key, "org-a", sealed)
	if err != nil || string(opened) != "private figure 12345" {
		t.Fatal(err, string(opened))
	}
}
