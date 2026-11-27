package app

import (
	"errors"
	"sort"
	"time"
)

type LeaseTerms struct {
	RoomID, Effective, Tenant     string
	Rent, ElectricRate, WaterRate int64
	Months                        int
	Renewal                       bool
	Phone, SignedOn, StartOn      string
	PolicyFeeLabel                string
	PolicyFee                     int64
	PolicyFeeCaptured             bool
	EndedOn                       string
}

func (l LeaseTerms) startDate() string {
	if l.StartOn != "" {
		return l.StartOn
	}
	return l.Effective + "-01"
}

func (l LeaseTerms) rentDays(period string) (int, int) {
	month, err := time.Parse("2006-01", period)
	if err != nil {
		return 0, 0
	}
	lastDay := month.AddDate(0, 1, -1).Day()
	first, last := 1, lastDay
	if day, err := time.Parse(time.DateOnly, l.StartOn); err == nil && day.Format("2006-01") == period {
		first = day.Day()
	}
	if day, err := time.Parse(time.DateOnly, l.EndedOn); err == nil && day.Format("2006-01") == period {
		last = day.Day()
	}
	return max(0, last-first+1), lastDay
}

func (l LeaseTerms) Until() string {
	if l.EndedOn != "" {
		return l.EndedOn[:7]
	}
	start, err := time.Parse("2006-01", l.Effective)
	if err != nil || l.Months == 0 {
		return ""
	}
	return start.AddDate(0, l.Months-1, 0).Format("2006-01")
}

func (l LeaseTerms) ActiveAt(period string) bool {
	until := l.Until()
	return l.Effective != "" && period >= l.Effective && (until == "" || period <= until)
}

func (l LeaseTerms) OccupiedIn(period string) bool {
	return l.ActiveAt(period) && (l.EndedOn == "" || l.EndedOn[:7] != period)
}

func (r Room) initialLease() LeaseTerms {
	terms := LeaseTerms{RoomID: r.ID, Rent: r.Rent, ElectricRate: r.ElectricRate, WaterRate: r.WaterRate}
	if !r.Prepared {
		terms.Effective, terms.Tenant, terms.Months = r.Start, r.Tenant, r.LeaseMonths
	}
	return terms
}

func (b *Book) leaseHistory(room Room) []LeaseTerms {
	var history []LeaseTerms
	if first := room.initialLease(); first.Effective != "" {
		history = append(history, first)
	}
	for _, terms := range b.Leases {
		if terms.RoomID == room.ID {
			history = append(history, terms)
		}
	}
	sort.Slice(history, func(i, j int) bool { return history[i].Effective < history[j].Effective })
	for i := range history {
		history[i].EndedOn = b.checkoutDate(room.ID, history[i].Effective)
	}
	return history
}

func (b *Book) leaseAt(room Room, period string) LeaseTerms {
	active := room.initialLease()
	for _, terms := range b.Leases {
		if terms.RoomID == room.ID && terms.Effective > active.Effective && terms.Effective <= period {
			active = terms
		}
	}
	if active.Effective == "" {
		policy := b.policyAt(room, period)
		active.Rent, active.ElectricRate, active.WaterRate = policy.Rent, policy.ElectricRate, policy.WaterRate
	}
	active.EndedOn = b.checkoutDate(room.ID, active.Effective)
	return active
}

func (l LeaseTerms) recurringFee(room Room) FeeLine {
	if l.PolicyFeeCaptured {
		return FeeLine{Label: l.PolicyFeeLabel, Amount: l.PolicyFee}
	}
	return FeeLine{Label: room.PolicyFeeLabel, Amount: room.PolicyFee}
}

func (b *Book) calculateInvoice(room Room, reading Reading, id string) Invoice {
	return b.calculateInvoiceWithLease(room, reading, id, b.leaseAt(room, reading.Period))
}

func (b *Book) calculateInvoiceWithLease(room Room, reading Reading, id string, active LeaseTerms) Invoice {
	room.Rent, room.ElectricRate, room.WaterRate = active.Rent, active.ElectricRate, active.WaterRate
	days, monthDays := active.rentDays(reading.Period)
	if days < monthDays {
		room.Rent = (room.Rent*int64(days) + int64(monthDays)/2) / int64(monthDays)
	}
	invoice := calculate(room, reading, id)
	if days < monthDays {
		invoice.RentDays, invoice.RentMonthDays = days, monthDays
	}
	invoice.RoomName = b.roomLabel(room)
	invoice.LeaseEffective, invoice.LeaseTenant = active.Effective, active.Tenant
	settings := b.Billing[reading.Period]
	invoice.DueDate, invoice.CommonFeeLabel, invoice.CommonFee = settings.DueDate, settings.CommonFeeLabel, settings.CommonFee
	invoice.Total += invoice.CommonFee
	if fee := active.recurringFee(room); fee.Amount > 0 {
		invoice.RoomFees = append(invoice.RoomFees, fee)
		invoice.Total += fee.Amount
	}
	for _, fee := range b.RoomFees {
		if fee.RoomID == room.ID && fee.applies(reading.Period) {
			invoice.RoomFees = append(invoice.RoomFees, FeeLine{Label: fee.Label, Amount: fee.Amount})
			invoice.Total += fee.Amount
		}
	}
	b.snapshotResidentShares(&invoice, active)
	return invoice
}

func (b *Book) scheduleLease(terms LeaseTerms) error {
	room := b.room(terms.RoomID)
	if room == nil {
		return errMissing
	}
	if _, err := time.Parse("2006-01", terms.Effective); err != nil || terms.Effective < room.Start ||
		(terms.Effective == room.Start && !room.Prepared) ||
		terms.Tenant == "" || len(terms.Tenant) > 150 || terms.Months < 0 || terms.Months > 120 ||
		len(terms.Phone) > 30 ||
		terms.Rent < 0 || terms.Rent > 100000000 ||
		terms.ElectricRate < 0 || terms.ElectricRate > 100000 ||
		terms.WaterRate < 0 || terms.WaterRate > 1000000 {
		return errInvalid
	}
	if terms.SignedOn != "" {
		if _, err := time.Parse(time.DateOnly, terms.SignedOn); err != nil {
			return errInvalid
		}
	}
	if terms.StartOn != "" {
		day, err := time.Parse(time.DateOnly, terms.StartOn)
		if err != nil || day.Format("2006-01") != terms.Effective {
			return errInvalid
		}
		if day.Day() == 1 {
			terms.StartOn = ""
		}
	}
	for _, existing := range b.Leases {
		if existing.RoomID == terms.RoomID && existing.Effective == terms.Effective {
			terms.PolicyFeeLabel, terms.PolicyFee, terms.PolicyFeeCaptured = existing.PolicyFeeLabel, existing.PolicyFee, existing.PolicyFeeCaptured
			if existing == terms {
				return nil
			}
			return errors.New("Đã có điều khoản hiệu lực từ kỳ này.")
		}
	}
	policy := b.policyAt(*room, terms.Effective)
	if terms.StartOn != "" {
		if terms.Renewal {
			return errors.New("Ký lại hợp đồng bắt đầu từ ngày đầu kỳ. Chọn hợp đồng mới nếu bàn giao giữa tháng.")
		}
		if b.leaseAt(*room, terms.Effective).ActiveAt(terms.Effective) {
			return errors.New("Kỳ bắt đầu đã có hợp đồng. Hoàn tất hợp đồng cũ trước khi tiếp nhận giữa tháng.")
		}
		if b.reading(room.ID, terms.Effective) != nil {
			return errors.New("Kỳ bắt đầu đã có chỉ số. Hãy chọn kỳ khác để ghi mốc bàn giao riêng.")
		}
	}
	terms.PolicyFeeLabel, terms.PolicyFee, terms.PolicyFeeCaptured = policy.FeeLabel, policy.Fee, true
	if terms.Renewal {
		start, _ := time.Parse("2006-01", terms.Effective)
		priorPeriod := start.AddDate(0, -1, 0).Format("2006-01")
		prior := b.leaseAt(*room, priorPeriod)
		if !prior.ActiveAt(priorPeriod) || prior.EndedOn != "" {
			return errors.New("Kỳ ký lại phải nối tiếp một hợp đồng đang hiệu lực.")
		}
	}
	for _, existing := range b.leaseHistory(*room) {
		if existing.Effective < terms.Effective && existing.Until() >= terms.Effective ||
			existing.Effective > terms.Effective && terms.Until() >= existing.Effective {
			return errors.New("Thời hạn bị trùng với hợp đồng đã lên lịch.")
		}
	}
	for _, invoice := range b.Invoices {
		if invoice.RoomID == terms.RoomID && invoice.Issued && invoice.Period >= terms.Effective {
			return errors.New("Kỳ hiệu lực phải sau tất cả bảng kê đã phát hành của phòng.")
		}
	}
	b.Leases = append(b.Leases, terms)
	for j := range b.Invoices {
		i := &b.Invoices[j]
		if i.RoomID == terms.RoomID && !i.Issued && i.Period >= terms.Effective {
			*i = b.calculateInvoice(*room, i.Reading, i.ID)
		}
	}
	return nil
}
