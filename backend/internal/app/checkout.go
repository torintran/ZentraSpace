package app

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type Checkout struct {
	RoomID, Effective, Date, Revision, Note string
	Applied, Refunded                       int64
}

type CheckoutQuote struct {
	Date, Effective, Revision   string
	Invoice                     Invoice
	Held, EarlierDue, Remaining int64
}

func (b *Book) checkoutDate(roomID, effective string) string {
	for _, c := range b.Checkouts {
		if c.RoomID == roomID && c.Effective == effective {
			return c.Date
		}
	}
	return ""
}

func (b *Book) quoteCheckout(room Room, date string, now time.Time) (CheckoutQuote, error) {
	day, err := time.Parse(time.DateOnly, date)
	if err != nil || date > now.In(vietnam).Format(time.DateOnly) {
		return CheckoutQuote{}, errors.New("Chọn ngày trả phòng đến hôm nay.")
	}
	period := day.Format("2006-01")
	lease := b.leaseAt(room, period)
	if !lease.ActiveAt(period) || lease.EndedOn != "" || lease.Tenant == "" {
		return CheckoutQuote{}, errors.New("Kỳ này không có hợp đồng để trả phòng.")
	}
	q := CheckoutQuote{Date: date, Effective: lease.Effective}
	for _, i := range b.Invoices {
		if i.RoomID != room.ID {
			continue
		}
		if i.Period == period {
			q.Invoice = i
		}
		if i.Issued && i.Period < period {
			q.EarlierDue += i.Remaining()
		}
		if i.Period > period && b.leaseAt(room, i.Period).Effective == lease.Effective {
			return CheckoutQuote{}, errors.New("Hợp đồng đã có bảng kê ở kỳ sau ngày trả. Hãy kiểm tra kỳ trả phòng.")
		}
	}
	for _, reading := range b.Readings {
		if reading.RoomID == room.ID && reading.Period > period && b.leaseAt(room, reading.Period).Effective == lease.Effective {
			return CheckoutQuote{}, errors.New("Hợp đồng đã có chỉ số ở kỳ sau ngày trả.")
		}
	}
	for _, resident := range b.Residents {
		if resident.RoomID == room.ID && resident.Effective == lease.Effective &&
			(resident.JoinedOn > date || resident.LeftOn > date) {
			return CheckoutQuote{}, errors.New("Có mốc người ở sau ngày trả. Hãy kiểm tra lịch sử người ở trước.")
		}
	}
	for _, e := range b.CashEntries {
		if e.RoomID == room.ID && e.Date <= date {
			q.Held += e.depositDelta()
		}
	}
	if !q.Invoice.Issued {
		reading := b.reading(room.ID, period)
		if reading == nil {
			return CheckoutQuote{}, errors.New("Hãy ghi chỉ số bàn giao cuối kỳ trước khi trả phòng.")
		}
		lease.EndedOn = date
		q.Invoice = b.calculateInvoiceWithLease(room, *reading, q.Invoice.ID, lease)
	}
	q.Remaining = q.Invoice.Remaining()
	data, err := json.Marshal(q)
	if err != nil {
		return CheckoutQuote{}, err
	}
	q.Revision = hash(string(data))
	return q, nil
}

func (b *Book) closeCheckout(c Checkout, now time.Time) error {
	room := b.room(c.RoomID)
	if room == nil {
		return errMissing
	}
	for _, existing := range b.Checkouts {
		if existing.RoomID == c.RoomID && existing.Effective == c.Effective {
			if existing == c {
				return nil
			}
			return errors.New("Hợp đồng đã có biên bản trả phòng.")
		}
	}
	q, err := b.quoteCheckout(*room, c.Date, now)
	if err != nil {
		return err
	}
	if c.Revision != q.Revision || c.Effective != q.Effective {
		return errors.New("Số liệu đã thay đổi. Hãy xem lại bản tất toán trước khi xác nhận.")
	}
	c.Note = strings.TrimSpace(c.Note)
	if c.Note == "" || len(c.Note) > 160 || c.Applied < 0 || c.Refunded < 0 ||
		c.Applied > q.Invoice.Remaining() || c.Applied > q.Held || c.Refunded > q.Held-c.Applied {
		return errors.New("Kiểm tra lý do, số cọc hoàn và số cọc cấn vào bảng kê cuối.")
	}
	for _, e := range []CashEntry{
		{Kind: depositApplied, Amount: c.Applied}, {Kind: depositOut, Amount: c.Refunded},
	} {
		if e.Amount == 0 {
			continue
		}
		e.Key = hash(c.Revision + ":" + e.Kind)[:48]
		e.Date, e.RoomID, e.Note = c.Date, room.ID, c.Note
		if err := b.postCash(e); err != nil {
			return err
		}
	}
	if !q.Invoice.Issued {
		if q.Invoice.ID == "" {
			q.Invoice.ID = token()
			b.Invoices = append(b.Invoices, q.Invoice)
		} else {
			*b.invoice(q.Invoice.ID) = q.Invoice
		}
		if err := b.issue(q.Invoice.ID); err != nil {
			return err
		}
	}
	if c.Applied > 0 {
		i := b.invoice(q.Invoice.ID)
		i.Entries = append(i.Entries, Entry{Key: hash(c.Revision + ":credit"), Kind: "Cấn cọc", Amount: c.Applied, Date: now.UTC(), Note: c.Note})
	}
	for i := range b.Residents {
		r := &b.Residents[i]
		if r.RoomID == c.RoomID && r.Effective == c.Effective && r.LeftOn == "" {
			r.LeftOn = c.Date
		}
	}
	b.Checkouts = append(b.Checkouts, c)
	return nil
}
