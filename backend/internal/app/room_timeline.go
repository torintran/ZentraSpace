package app

import (
	"slices"
	"sort"
	"strings"
)

type RoomEvent struct {
	Date, Kind, Title, Detail, InvoiceID string
	Amount                               int64
}

func (b *Book) roomTimeline(room Room, period, kind string) []RoomEvent {
	events := []RoomEvent{{Date: room.Start, Kind: "room", Title: "Bắt đầu quản lý phòng"}}
	for _, assignment := range b.PolicyAssignments {
		if assignment.RoomID == room.ID {
			events = append(events, RoomEvent{Date: assignment.AppliedOn, Kind: "policy", Title: "Đã lên lịch áp dụng mẫu giá", Detail: assignment.Policy.Name + " · Từ kỳ " + assignment.Effective})
		}
	}
	for _, terms := range b.leaseHistory(room) {
		events = append(events, RoomEvent{Date: terms.Effective, Kind: "lease", Title: "Hợp đồng có hiệu lực", Detail: terms.Tenant})
	}
	for _, c := range b.Checkouts {
		if c.RoomID == room.ID {
			events = append(events, RoomEvent{Date: c.Date, Kind: "checkout", Title: "Trả phòng và tất toán", Detail: c.Note})
		}
	}
	for _, resident := range b.Residents {
		if resident.RoomID != room.ID {
			continue
		}
		events = append(events, RoomEvent{Date: resident.JoinedOn, Kind: "resident", Title: "Người ở vào phòng", Detail: resident.Name})
		if resident.LeftOn != "" {
			events = append(events, RoomEvent{Date: resident.LeftOn, Kind: "resident", Title: "Người ở rời phòng", Detail: resident.Name})
		}
	}
	for _, invoice := range b.Invoices {
		if invoice.RoomID != room.ID || !invoice.Issued {
			continue
		}
		date := invoice.IssuedOn
		if date == "" {
			date = invoice.Period
		}
		events = append(events, RoomEvent{Date: date, Kind: "bill", Title: "Phát hành bảng kê", Detail: "Kỳ " + invoice.Period, Amount: invoice.Total, InvoiceID: invoice.ID})
		for _, entry := range invoice.Entries {
			title := entry.Kind
			if title == "Thu" {
				title = "Thu tiền phòng"
			}
			events = append(events, RoomEvent{Date: entry.Date.In(vietnam).Format("2006-01-02"), Kind: "payment", Title: title, Detail: entry.Note, Amount: entry.Amount, InvoiceID: invoice.ID})
		}
	}
	for _, entry := range b.CashEntries {
		if entry.RoomID == room.ID {
			events = append(events, RoomEvent{Date: entry.Date, Kind: "deposit", Title: entry.KindLabel(), Detail: entry.Note, Amount: entry.Amount})
		}
	}
	events = slices.DeleteFunc(events, func(event RoomEvent) bool {
		return (period != "" && !strings.HasPrefix(event.Date, period)) || (kind != "" && event.Kind != kind)
	})
	sort.SliceStable(events, func(i, j int) bool { return events[i].Date > events[j].Date })
	return events
}
