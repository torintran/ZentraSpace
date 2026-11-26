package main

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type Room struct {
	ID, Name, Tenant, Start                        string
	Rent, ElectricRate, WaterRate, Electric, Water int64
}
type Reading struct {
	RoomID, Period                                         string
	ElectricBefore, WaterBefore, ElectricAfter, WaterAfter int64
}
type Invoice struct {
	ID, RoomID, RoomName, Period                                      string
	Rent, ElectricRate, WaterRate, ElectricAmount, WaterAmount, Total int64
	Reading                                                           Reading
	Issued                                                            bool
	Entries                                                           []Entry
	PaymentCode                                                       string
}
type Entry struct {
	Key, Kind, Note string
	Amount          int64
	Date            time.Time
}
type Invite struct {
	ID, InvoiceID, CodeHash string
	Expires                 time.Time
	Attempts                int
	Used, Revoked           bool
}
type Book struct {
	Rooms     []Room
	Readings  []Reading
	Invoices  []Invoice
	Invites   []Invite
	Bank      BankAccount
	Transfers []BankTransfer
}

var (
	errMissing = errors.New("Không tìm thấy dữ liệu.")
	errLocked  = errors.New("Kỳ này đã phát hành. Hãy dùng điều chỉnh có lý do.")
	errInvalid = errors.New("Vui lòng kiểm tra lại thông tin.")
)

func hash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func token() string { return hex.EncodeToString(randBytes(24)) }

func randBytes(size int) []byte {
	value := make([]byte, size)
	if _, err := rand.Read(value); err != nil {
		panic(err)
	}
	return value
}

func accessCode() string {
	n, err := rand.Int(rand.Reader, big.NewInt(100000000))
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%08d", n.Int64())
}

func (v *Invite) verify(code string, now time.Time) bool {
	if v.Used || v.Revoked || v.Attempts >= 5 || !now.Before(v.Expires) {
		return false
	}
	v.Attempts++
	if len(code) != 8 || subtle.ConstantTimeCompare([]byte(hash(code)), []byte(v.CodeHash)) != 1 {
		return false
	}
	v.Used = true
	v.CodeHash = ""
	return true
}

func (b *Book) room(id string) *Room {
	for i := range b.Rooms {
		if b.Rooms[i].ID == id {
			return &b.Rooms[i]
		}
	}
	return nil
}
func (b *Book) invoice(id string) *Invoice {
	for i := range b.Invoices {
		if b.Invoices[i].ID == id {
			return &b.Invoices[i]
		}
	}
	return nil
}
func (b *Book) invite(id string) *Invite {
	for i := range b.Invites {
		if b.Invites[i].ID == id {
			return &b.Invites[i]
		}
	}
	return nil
}
func (b *Book) reading(room, period string) *Reading {
	for i := range b.Readings {
		if b.Readings[i].RoomID == room && b.Readings[i].Period == period {
			return &b.Readings[i]
		}
	}
	return nil
}

func (b *Book) read(id, period string, electric, water int64) error {
	room := b.room(id)
	if room == nil {
		return errMissing
	}
	start, err := time.Parse("2006-01", period)
	if err != nil || period < room.Start {
		return errInvalid
	}
	beforeE, beforeW := room.Electric, room.Water
	if period != room.Start {
		previous := start.AddDate(0, -1, 0).Format("2006-01")
		var last *Invoice
		for j := range b.Invoices {
			i := &b.Invoices[j]
			if i.RoomID == id && i.Period == previous && i.Issued {
				last = i
				break
			}
		}
		if last == nil {
			return errors.New("Hãy phát hành bảng kê của kỳ trước để chốt chỉ số đầu kỳ.")
		}
		beforeE, beforeW = last.Reading.ElectricAfter, last.Reading.WaterAfter
	}
	if electric < beforeE || water < beforeW || electric > 999999 || water > 999999 {
		return errors.New("Chỉ số mới phải từ chỉ số đầu kỳ đến 999.999.")
	}
	for j := range b.Invoices {
		i := &b.Invoices[j]
		if i.RoomID == id && i.Period == period && i.Issued {
			return errLocked
		}
	}
	r := Reading{id, period, beforeE, beforeW, electric, water}
	if prior := b.reading(id, period); prior != nil {
		*prior = r
	} else {
		b.Readings = append(b.Readings, r)
	}
	for j := range b.Invoices {
		i := &b.Invoices[j]
		if i.RoomID == id && i.Period == period && !i.Issued {
			*i = calculate(*room, r, i.ID)
		}
	}
	return nil
}

func calculate(r Room, reading Reading, id string) Invoice {
	electric := (reading.ElectricAfter - reading.ElectricBefore) * r.ElectricRate
	water := (reading.WaterAfter - reading.WaterBefore) * r.WaterRate
	return Invoice{ID: id, RoomID: r.ID, RoomName: r.Name, Period: reading.Period, Rent: r.Rent,
		ElectricRate: r.ElectricRate, WaterRate: r.WaterRate, ElectricAmount: electric,
		WaterAmount: water, Total: r.Rent + electric + water, Reading: reading}
}

func (b *Book) draft(period string) int {
	if _, err := time.Parse("2006-01", period); err != nil {
		return 0
	}
	n := 0
	for _, r := range b.Rooms {
		reading := b.reading(r.ID, period)
		if reading == nil {
			continue
		}
		exists := false
		for _, i := range b.Invoices {
			if i.RoomID == r.ID && i.Period == period {
				exists = true
				break
			}
		}
		if !exists {
			b.Invoices = append(b.Invoices, calculate(r, *reading, token()))
			n++
		}
	}
	return n
}

func (b *Book) issue(id string) error {
	i := b.invoice(id)
	if i == nil {
		return errMissing
	}
	i.Issued = true
	if b.Bank.Configured() {
		b.ensurePaymentCode(i)
	}
	return nil
}

func (i Invoice) totals() (collected, reduced int64) {
	for _, e := range i.Entries {
		switch e.Kind {
		case "Thu":
			collected += e.Amount
		case "Hoàn":
			collected -= e.Amount
		case "Giảm":
			reduced += e.Amount
		}
	}
	return
}
func (i Invoice) Remaining() int64 { c, d := i.totals(); return max(0, i.Total-d-c) }
func (i Invoice) Credit() int64    { c, d := i.totals(); return max(0, c-(i.Total-d)) }

func (b *Book) post(id, key, kind string, amount int64, note string) error {
	if strings.HasPrefix(key, "bank-") {
		return errInvalid
	}
	return b.addEntry(id, key, kind, amount, note)
}

func (b *Book) addEntry(id, key, kind string, amount int64, note string) error {
	i := b.invoice(id)
	if i == nil {
		return errMissing
	}
	if !i.Issued || len(key) < 8 || len(key) > 80 || amount <= 0 || amount > 100000000000 || len(note) > 240 {
		return errInvalid
	}
	for _, e := range i.Entries {
		if e.Key == key {
			if e.Kind == kind && e.Amount == amount && e.Note == note {
				return nil
			}
			return errors.New("Thao tác đã được ghi với thông tin khác.")
		}
	}
	collected, reduced := i.totals()
	switch kind {
	case "Thu":
		if collected+amount > 1000000000000 {
			return errInvalid
		}
	case "Giảm":
		if note == "" || amount > i.Total-reduced {
			return errInvalid
		}
	case "Hoàn":
		if note == "" || amount > collected {
			return errInvalid
		}
	default:
		return errInvalid
	}
	i.Entries = append(i.Entries, Entry{Key: key, Kind: kind, Amount: amount, Note: note, Date: time.Now().UTC()})
	return nil
}
