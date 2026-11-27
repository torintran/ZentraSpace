package app

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
)

type Room struct {
	ID, Name, Tenant, Start                        string
	Rent, ElectricRate, WaterRate, Electric, Water int64
	LeaseMonths                                    int
	Prepared                                       bool `json:",omitempty"`
	SiteID, PolicyID, Description, PolicyFeeLabel  string
	PolicyFee                                      int64
}
type Reading struct {
	RoomID, Period                                         string
	ElectricBefore, WaterBefore, ElectricAfter, WaterAfter int64
	Adjustment                                             MeterAdjustment
}
type Invoice struct {
	ID, RoomID, RoomName, Period                                      string
	LeaseEffective, LeaseTenant                                       string
	DueDate, CommonFeeLabel                                           string
	CommonFee                                                         int64
	RoomFees                                                          []FeeLine `json:",omitempty"`
	Rent, ElectricRate, WaterRate, ElectricAmount, WaterAmount, Total int64
	RentDays, RentMonthDays                                           int `json:",omitempty"`
	Reading                                                           Reading
	Issued                                                            bool
	IssuedOn                                                          string `json:",omitempty"`
	Entries                                                           []Entry
	PaymentCode                                                       string
}
type Entry struct {
	Key, Kind, Note string
	Amount          int64
	Date            time.Time
	Reverses        string `json:",omitempty"`
	Method          string `json:",omitempty"`
	Reference       string `json:",omitempty"`
}

const (
	cashPayment     = "cash"
	transferPayment = "transfer"
)

func (e Entry) MethodLabel() string {
	switch e.Method {
	case cashPayment:
		return "Tiền mặt"
	case transferPayment:
		return "Chuyển khoản"
	}
	return ""
}

type Invite struct {
	ID, InvoiceID, CodeHash string
	Expires                 time.Time
	Attempts                int
	Used, Revoked           bool
	PendingDelivery         bool `json:",omitempty"`
}
type Book struct {
	CreditTransfers    []CreditTransfer
	SiteMeterReadings  []SiteMeterReading
	SiteMeterApprovals []SiteMeterApproval
	PlatformSMTP       *smtpConfiguration `json:",omitempty"`
	Checkouts          []Checkout
	LeaseEmails        map[string]string
	PolicyAssignments  []PolicyAssignment
	Sites              []Site
	Policies           []RoomPolicy
	Residents          []Resident
	CashEntries        []CashEntry
	Rooms              []Room
	RoomFees           []RoomFee
	Leases             []LeaseTerms
	Billing            map[string]BillingSettings
	Readings           []Reading
	Invoices           []Invoice
	Invites            []Invite
	Bank               BankAccount
	MockBankAutoMatch  bool
	Transfers          []BankTransfer
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
	if v.Used || v.Revoked || v.PendingDelivery || v.Attempts >= 5 || !now.Before(v.Expires) {
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

func (r Room) Revision() string {
	data, _ := json.Marshal(r)
	return hash(string(data))
}

func (b *Book) roomLocks(id string) (hasReading, hasIssued bool) {
	for _, r := range b.Readings {
		if r.RoomID == id {
			hasReading = true
			break
		}
	}
	for _, i := range b.Invoices {
		if i.RoomID == id && i.Issued {
			hasIssued = true
			break
		}
	}
	return
}

func (b *Book) saveRoom(room Room, revision string) error {
	for _, existing := range b.Rooms {
		if existing.ID != room.ID && existing.SiteID == room.SiteID && strings.EqualFold(existing.Name, room.Name) {
			return errors.New("Mã phòng đã tồn tại trong dãy này.")
		}
	}
	if err := b.checkSiteCapacity(room); err != nil {
		return err
	}
	if room.ID == "" {
		if len(b.Rooms) >= 150 {
			return errors.New("Bản thử hỗ trợ tối đa 150 phòng.")
		}
		room.ID = token()
		b.Rooms = append(b.Rooms, room)
		return nil
	}
	current := b.room(room.ID)
	if current == nil {
		return errMissing
	}
	if current.Revision() != revision {
		return errors.New("Thông tin phòng đã thay đổi. Hãy mở lại phòng trước khi lưu.")
	}
	*current = room
	for j := range b.Invoices {
		i := &b.Invoices[j]
		if i.RoomID == room.ID && !i.Issued {
			*i = b.calculateInvoice(room, i.Reading, i.ID)
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

func calculate(r Room, reading Reading, id string) Invoice {
	electric := reading.ElectricUsage() * r.ElectricRate
	water := reading.WaterUsage() * r.WaterRate
	return Invoice{ID: id, RoomID: r.ID, RoomName: r.Name, Period: reading.Period, Rent: r.Rent,
		ElectricRate: r.ElectricRate, WaterRate: r.WaterRate, ElectricAmount: electric,
		WaterAmount: water, Total: r.Rent + electric + water, Reading: reading}
}

func (b *Book) draft(period string) int {
	if _, err := time.Parse("2006-01", period); err != nil {
		return 0
	}
	n := 0
	hold := make(map[string]bool)
	for _, site := range b.Sites {
		hold[site.ID] = b.siteMeterHold(site.ID, period)
	}
	for _, r := range b.Rooms {
		if hold[r.SiteID] || !b.leaseAt(r, period).ActiveAt(period) {
			continue
		}
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
			b.Invoices = append(b.Invoices, b.calculateInvoice(r, *reading, token()))
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
	if i.Issued {
		return nil
	}
	room := b.room(i.RoomID)
	if room == nil || !b.leaseAt(*room, i.Period).ActiveAt(i.Period) {
		return errors.New("Phòng không có hợp đồng hiệu lực trong kỳ này.")
	}
	if b.siteMeterHold(room.SiteID, i.Period) {
		return errors.New("Dãy trọ còn chênh lệch công tơ tổng cần duyệt trước khi phát hành.")
	}
	i.Issued = true
	i.IssuedOn = time.Now().In(vietnam).Format(time.DateOnly)
	if b.Bank.Configured() {
		b.ensurePaymentCode(i)
	}
	return nil
}

func (i Invoice) Revision() string {
	value := fmt.Sprintf("%q %q %q %q %q %q %q %q %d %d %d %d %d %d %d %d %d %d %d %#v %#v",
		i.ID, i.RoomID, i.RoomName, i.Period, i.LeaseEffective, i.LeaseTenant, i.DueDate, i.CommonFeeLabel,
		i.CommonFee, i.Rent, i.ElectricRate, i.WaterRate,
		i.ElectricAmount, i.WaterAmount, i.Total, i.Reading.ElectricBefore,
		i.Reading.ElectricAfter, i.Reading.WaterBefore, i.Reading.WaterAfter, i.RoomFees, i.Reading.Adjustment)
	if i.RentMonthDays > 0 {
		value += fmt.Sprintf(" %d/%d", i.RentDays, i.RentMonthDays)
	}
	return hash(value)
}

func (i Invoice) NetCollected() int64 { collected, _ := i.totals(); return collected }
func (i Invoice) NetReduction() int64 { _, reduced := i.totals(); return reduced }

func (i Invoice) totals() (collected, reduced int64) {
	for _, e := range i.Entries {
		switch e.Kind {
		case "Thu", "Cấn cọc", creditIn, creditOutReversed:
			collected += e.Amount
		case "Hoàn", "Đảo thu", creditOut, creditInReversed:
			collected -= e.Amount
		case "Giảm":
			reduced += e.Amount
		case "Đảo giảm":
			reduced -= e.Amount
		}
	}
	return
}
func (i Invoice) Remaining() int64 { c, d := i.totals(); return max(0, i.Total-d-c) }
func (i Invoice) Credit() int64    { c, d := i.totals(); return max(0, c-(i.Total-d)) }

func (i Invoice) cashCollected() int64 {
	amount := i.NetCollected()
	for _, entry := range i.Entries {
		switch entry.Kind {
		case "Cấn cọc", creditIn, creditOutReversed:
			amount -= entry.Amount
		case creditInReversed, creditOut:
			amount += entry.Amount
		}
	}
	return amount
}

func (i Invoice) Reversed(key string) bool {
	for _, e := range i.Entries {
		if e.Reverses == key {
			return true
		}
	}
	return false
}

func (b *Book) reverse(id, key, original, note string) error {
	i := b.invoice(id)
	if i == nil {
		return errMissing
	}
	if !i.Issued || len(key) < 8 || len(key) > 80 || strings.HasPrefix(key, "bank-") || strings.HasPrefix(key, "credit-") || original == "" || note == "" || len(note) > 240 {
		return errInvalid
	}
	var source Entry
	for _, e := range i.Entries {
		if e.Key == key {
			if e.Reverses == original && e.Note == note {
				return nil
			}
			return errors.New("Thao tác đã được ghi với thông tin khác.")
		}
		if e.Key == original {
			source = e
		}
	}
	if i.Reversed(original) {
		return errors.New("Bút toán này đã được đảo.")
	}
	var kind string
	switch source.Kind {
	case "Thu":
		for _, tr := range b.Transfers {
			if tr.InvoiceID == id && tr.EntryKey == original {
				return errors.New("Khoản thu đã đối soát với sao kê. Hãy kiểm tra giao dịch ngân hàng trước khi điều chỉnh.")
			}
		}
		if source.Amount > i.refundableCash() {
			return errors.New("Không thể đảo toàn bộ khoản thu vì số đã thu ròng thấp hơn khoản gốc.")
		}
		kind = "Đảo thu"
	case "Giảm":
		kind = "Đảo giảm"
	default:
		return errInvalid
	}
	i.Entries = append(i.Entries, Entry{Key: key, Kind: kind, Amount: source.Amount, Note: note, Date: time.Now().UTC(), Reverses: original})
	return nil
}

func (b *Book) post(id string, entry Entry) error {
	if strings.HasPrefix(entry.Key, "bank-") || strings.HasPrefix(entry.Key, "credit-") {
		return errInvalid
	}
	return b.addEntry(id, entry)
}

func (b *Book) addEntry(id string, entry Entry) error {
	i := b.invoice(id)
	if i == nil {
		return errMissing
	}
	if !i.Issued || len(entry.Key) < 8 || len(entry.Key) > 80 || entry.Amount <= 0 || entry.Amount > 100000000000 ||
		len(entry.Note) > 240 || len(entry.Reference) > 160 || entry.Reverses != "" {
		return errInvalid
	}
	if entry.Method != "" && (entry.Kind != "Thu" || (entry.Method != cashPayment && entry.Method != transferPayment)) {
		return errInvalid
	}
	for _, e := range i.Entries {
		if e.Key == entry.Key {
			entry.Date = e.Date
			if e == entry {
				return nil
			}
			return errors.New("Thao tác đã được ghi với thông tin khác.")
		}
	}
	collected, reduced := i.totals()
	switch entry.Kind {
	case "Thu":
		if collected+entry.Amount > 1000000000000 {
			return errInvalid
		}
	case "Giảm":
		if entry.Note == "" || entry.Amount > i.Total-reduced {
			return errInvalid
		}
	case "Hoàn":
		if entry.Note == "" || entry.Amount > i.refundableCash() {
			return errInvalid
		}
	default:
		return errInvalid
	}
	entry.Date = time.Now().UTC()
	i.Entries = append(i.Entries, entry)
	return nil
}
