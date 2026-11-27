package main

import (
	"sort"
	"time"
)

type Dashboard struct {
	Rooms, Missing, Drafts, Outstanding, Issued, Settled int
	Billed, Collected, Due, Credit                       int64
	RoomStatus                                           []dashboardRoom
	Trend                                                []dashboardMonth
	TrendMax                                             int64
}

type dashboardRoom struct {
	Name, Status, Label, InvoiceID string
}

type dashboardMonth struct {
	Period, Label string
	Issued        int
	Collected     int64
	Due           int64
}

func (d Dashboard) ReadingsDone() int { return d.Rooms - d.Missing }

func (b *Book) dashboard(period string) Dashboard {
	var d Dashboard
	month, _ := time.Parse("2006-01", period)
	months := make(map[string]int)
	for i := -5; i <= 0; i++ {
		m := month.AddDate(0, i, 0)
		key := m.Format("2006-01")
		months[key] = len(d.Trend)
		d.Trend = append(d.Trend, dashboardMonth{Period: key, Label: m.Format("01/2006")})
	}
	bills := make(map[string]Invoice)
	readings := make(map[string]bool)
	for _, reading := range b.Readings {
		if reading.Period == period {
			readings[reading.RoomID] = true
		}
	}
	for _, invoice := range b.Invoices {
		if invoice.Period == period {
			bills[invoice.RoomID] = invoice
			if invoice.Issued {
				collected, reduced := invoice.totals()
				d.Issued++
				d.Billed += invoice.Total - reduced
				d.Collected += collected
				d.Due += invoice.Remaining()
				d.Credit += invoice.Credit()
				if invoice.Remaining() == 0 {
					d.Settled++
				} else {
					d.Outstanding++
				}
			} else {
				d.Drafts++
			}
		}
		if index, ok := months[invoice.Period]; ok && invoice.Issued {
			m := &d.Trend[index]
			m.Issued++
			m.Collected += invoice.NetCollected()
			m.Due += invoice.Remaining()
		}
	}
	for _, m := range d.Trend {
		d.TrendMax = max(d.TrendMax, m.Collected, m.Due)
	}
	for _, room := range b.Rooms {
		if !b.leaseAt(room, period).ActiveAt(period) {
			continue
		}
		d.Rooms++
		if !readings[room.ID] {
			d.Missing++
		}
		cell := dashboardRoom{Name: room.Name}
		invoice, hasBill := bills[room.ID]
		switch {
		case hasBill && !invoice.Issued:
			cell.Status, cell.Label = "draft", "Chờ phát hành"
		case hasBill && invoice.Remaining() > 0:
			cell.Status, cell.Label = "due", "Còn phải thu"
		case hasBill:
			cell.Status, cell.Label = "settled", "Đã thu đủ"
		case readings[room.ID]:
			cell.Status, cell.Label = "ready", "Có chỉ số"
		default:
			cell.Status, cell.Label = "missing", "Thiếu chỉ số"
		}
		cell.InvoiceID = invoice.ID
		d.RoomStatus = append(d.RoomStatus, cell)
	}
	sort.SliceStable(d.RoomStatus, func(i, j int) bool { return d.RoomStatus[i].Name < d.RoomStatus[j].Name })
	return d
}
