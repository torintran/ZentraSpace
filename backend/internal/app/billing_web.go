package app

import (
	"encoding/csv"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (a *App) invoice(w http.ResponseWriter, r *http.Request) {
	s, id, ok := a.staff(w, r)
	if !ok {
		return
	}
	if s.role != "owner" {
		http.Error(w, "Bạn không có quyền xem bảng kê.", 403)
		return
	}
	settings, err := a.smtpView(r.Context(), false)
	if err != nil {
		a.problem(w, r, err)
		return
	}
	p := Page{Screen: "invoice", Role: s.role, CSRF: hash(id + ":csrf"), SMTP: settings}
	err = a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		p.Invoice = b.invoice(r.PathValue("id"))
		if p.Invoice == nil {
			return errMissing
		}
		if settings.Configured && p.Invoice.Issued {
			p.VerificationEmail = b.LeaseEmails[p.Invoice.RoomID+"/"+p.Invoice.LeaseEffective]
		}
		p.ReconciledEntries = make(map[string]bool)
		for _, tr := range b.Transfers {
			if tr.InvoiceID == p.Invoice.ID {
				p.ReconciledEntries[tr.EntryKey] = true
			}
		}
		for _, v := range b.Invites {
			if v.InvoiceID == p.Invoice.ID && !v.Revoked && !v.PendingDelivery && time.Now().Before(v.Expires) {
				p.Invites = append(p.Invites, v)
			}
		}
		return nil
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	a.page(w, r, p)
}

func (a *App) export(w http.ResponseWriter, r *http.Request) {
	s, _, ok := a.staff(w, r)
	if !ok {
		return
	}
	if s.role != "owner" {
		http.Error(w, "Bạn không có quyền xuất báo cáo.", 403)
		return
	}
	period := periodOf(r.URL.Query().Get("period"))
	var rows = [][]string{{"Phòng", "Kỳ", "Tổng bảng kê (VND)", "Đã thu ròng (VND)", "Còn thiếu (VND)", "Thu dư (VND)", "Điều chỉnh giảm (VND)", "Phải thu sau điều chỉnh (VND)"}}
	err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		for _, i := range b.Invoices {
			if i.Issued && i.Period == period {
				c, reduced := i.totals()
				rows = append(rows, []string{plainCSV(i.RoomName), i.Period, strconv.FormatInt(i.Total, 10), strconv.FormatInt(c, 10),
					strconv.FormatInt(i.Remaining(), 10), strconv.FormatInt(i.Credit(), 10), strconv.FormatInt(reduced, 10), strconv.FormatInt(i.Total-reduced, 10)})
			}
		}
		return nil
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=cong-no-"+period+".csv")
	w.Write([]byte("\xef\xbb\xbf"))
	csvOut := csv.NewWriter(w)
	if err = csvOut.WriteAll(rows); err != nil {
		log.Print("csv write failed")
	}
}
