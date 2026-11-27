package app

import (
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"
	"unicode"
)

const (
	expense        = "expense"
	depositIn      = "deposit_in"
	depositOut     = "deposit_out"
	depositApplied = "deposit_applied"
)

type CashEntry struct {
	Key, Kind, Date, RoomID, Category, Reference, Note string
	Amount                                             int64
}

func (e CashEntry) KindLabel() string {
	switch e.Kind {
	case expense:
		return "Chi phí"
	case depositIn:
		return "Nhận cọc"
	case depositOut:
		return "Hoàn cọc"
	case depositApplied:
		return "Cấn cọc vào bảng kê"
	}
	return e.Kind
}

func (b *Book) postCash(e CashEntry) error {
	_, keyErr := hex.DecodeString(e.Key)
	if _, err := time.Parse("2006-01-02", e.Date); err != nil || keyErr != nil || len(e.Key) != 48 ||
		e.Amount <= 0 || e.Amount > 100000000000 || len(e.Category) > 120 || len(e.Reference) > 160 || len(e.Note) > 240 ||
		strings.ContainsFunc(e.Category+e.Reference+e.Note, unicode.IsControl) {
		return errInvalid
	}
	for _, existing := range b.CashEntries {
		if existing.Key == e.Key {
			if existing == e {
				return nil
			}
			return errors.New("Thao tác đã được ghi với thông tin khác.")
		}
	}
	switch e.Kind {
	case expense:
		if e.Category == "" || e.RoomID != "" {
			return errInvalid
		}
	case depositIn, depositOut, depositApplied:
		room := b.room(e.RoomID)
		if room == nil || e.Category != "" || e.Date < room.Start+"-01" {
			return errInvalid
		}
		if e.Kind == depositOut || e.Kind == depositApplied {
			if e.Note == "" {
				return errors.New("Nhập lý do hoàn cọc.")
			}
			days := make(map[string]int64)
			days[e.Date] = -e.Amount
			for _, posted := range b.CashEntries {
				if posted.RoomID == e.RoomID {
					days[posted.Date] += posted.depositDelta()
				}
			}
			var dates []string
			for day := range days {
				dates = append(dates, day)
			}
			sort.Strings(dates)
			var balance int64
			for _, day := range dates {
				balance += days[day]
				if balance < 0 {
					return errors.New("Số tiền hoàn vượt số cọc đang giữ tại một ngày ghi sổ.")
				}
			}
		}
	default:
		return errInvalid
	}
	b.CashEntries = append(b.CashEntries, e)
	return nil
}

func (e CashEntry) depositDelta() int64 {
	switch e.Kind {
	case depositIn:
		return e.Amount
	case depositOut, depositApplied:
		return -e.Amount
	}
	return 0
}

func (e CashEntry) Income() bool { return e.Kind == depositIn || e.Kind == "Thu" }
func (e CashEntry) CanAttach() bool {
	return e.Kind == expense || e.Kind == depositIn || e.Kind == depositOut
}

type CashRow struct {
	CashEntry
	RoomName string
}

type DepositRoom struct {
	ID, Name string
	Held     int64
}

type CashReport struct {
	RentNet, Expense, Received, Returned, Applied, Held int64
	Rows                                                []CashRow
	Rooms                                               []DepositRoom
}

func (b *Book) cashReport(period string) CashReport {
	var report CashReport
	rooms := make(map[string]int)
	for _, r := range b.Rooms {
		if r.Start <= period {
			rooms[r.ID] = len(report.Rooms)
			report.Rooms = append(report.Rooms, DepositRoom{ID: r.ID, Name: b.roomLabel(r)})
		}
	}
	for _, e := range b.CashEntries {
		if e.Date > period+"-31" {
			continue
		}
		report.Held += e.depositDelta()
		row := CashRow{CashEntry: e}
		if j, ok := rooms[e.RoomID]; ok {
			report.Rooms[j].Held += e.depositDelta()
			row.RoomName = report.Rooms[j].Name
		}
		if !strings.HasPrefix(e.Date, period) {
			continue
		}
		switch e.Kind {
		case expense:
			report.Expense += e.Amount
		case depositIn:
			report.Received += e.Amount
		case depositOut:
			report.Returned += e.Amount
		case depositApplied:
			report.Applied += e.Amount
		}
		report.Rows = append(report.Rows, row)
	}
	bankDates := make(map[[2]string]string)
	for _, t := range b.Transfers {
		if t.InvoiceID != "" {
			bankDates[[2]string{t.InvoiceID, t.EntryKey}] = t.Date
		}
	}
	for _, invoice := range b.Invoices {
		if !invoice.Issued {
			continue
		}
		for _, e := range invoice.Entries {
			date := e.Date.In(vietnam).Format("2006-01-02")
			if linked := bankDates[[2]string{invoice.ID, e.Key}]; linked != "" {
				date = linked
			}
			if !strings.HasPrefix(date, period) {
				continue
			}
			switch e.Kind {
			case "Thu":
				report.RentNet += e.Amount
			case "Hoàn", "Đảo thu":
				report.RentNet -= e.Amount
			default:
				continue
			}
			report.Rows = append(report.Rows, CashRow{
				CashEntry: CashEntry{Key: e.Key, Date: date, Kind: e.Kind, Amount: e.Amount, Reference: e.Reference, Note: e.Note},
				RoomName:  invoice.RoomName,
			})
		}
	}
	sort.SliceStable(report.Rows, func(i, j int) bool { return report.Rows[i].Date > report.Rows[j].Date })
	return report
}
