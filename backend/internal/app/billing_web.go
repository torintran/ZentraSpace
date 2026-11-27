package app

import (
	"encoding/csv"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	p := Page{Screen: "invoice", Role: s.role, CSRF: hash(id + ":csrf"), SMTP: settings, ManualInvite: !a.secure}
	if r.URL.Query().Get("saved") == "credit" {
		p.Message = "Đã cập nhật khoản tiền dư."
	}
	err = a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		p.Invoice = b.invoice(r.PathValue("id"))
		if p.Invoice == nil {
			return errMissing
		}
		if room := b.room(p.Invoice.RoomID); room != nil && !p.Invoice.Issued {
			p.MeterReviewHeld = b.siteMeterHold(room.SiteID, p.Invoice.Period)
		}
		if settings.Configured && p.Invoice.Issued {
			p.VerificationEmail = b.verificationAddress(p.Invoice, 0)
			for j := range p.Invoice.ResidentShares {
				p.ResidentEmailReady = append(p.ResidentEmailReady, validEmail(b.verificationAddress(p.Invoice, j+1)))
			}
		}
		if p.Invoice.Issued {
			p.AvailableCredit = p.Invoice.transferableCredit()
			p.CreditTargets, p.CreditTransfers = b.creditOptions(p.Invoice)
		}
		p.ReconciledEntries = make(map[string]bool)
		for _, tr := range b.Transfers {
			if tr.InvoiceID == p.Invoice.ID {
				p.ReconciledEntries[tr.EntryKey] = true
			}
		}
		for _, v := range b.Invites {
			if v.InvoiceID == p.Invoice.ID && v.usable(time.Now(), a.secure) {
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
	pending := r.URL.Query().Get("view") == "pending"
	filename := "cong-no-" + period + ".csv"
	var rows = [][]string{{"Phòng", "Kỳ", "Tổng bảng kê (VND)", "Đã tất toán ròng (VND)", "Còn thiếu (VND)", "Thu dư (VND)", "Điều chỉnh giảm (VND)", "Phải thu sau điều chỉnh (VND)", "Tiền dư đã chuyển (VND)", "Tiền dư có thể chuyển (VND)"}}
	err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		if pending {
			filename = "khoan-nhan-cho-doi-soat-" + period + ".csv"
			rows = [][]string{{"Ngày giao dịch", "Mã giao dịch", "Đuôi tài khoản nhận", "Số tiền (VND)", "Nội dung", "Phương thức"}}
			for _, transfer := range b.Transfers {
				if transfer.InvoiceID == "" && strings.HasPrefix(transfer.Date, period+"-") {
					method := "Chuyển khoản"
					if transfer.ManualCash {
						method = "Tiền mặt"
					}
					rows = append(rows, []string{transfer.Date, plainCSV(transfer.Reference), plainCSV(transfer.AccountSuffix),
						strconv.FormatInt(transfer.Amount, 10), plainCSV(transfer.Memo), method})
				}
			}
			return nil
		}
		for _, i := range b.Invoices {
			if i.Issued && i.Period == period {
				c, reduced := i.totals()
				rows = append(rows, []string{plainCSV(i.RoomName), i.Period, strconv.FormatInt(i.Total, 10), strconv.FormatInt(c, 10),
					strconv.FormatInt(i.Remaining(), 10), strconv.FormatInt(i.Credit(), 10), strconv.FormatInt(reduced, 10), strconv.FormatInt(i.Total-reduced, 10), strconv.FormatInt(i.transferredCredit(), 10), strconv.FormatInt(i.transferableCredit(), 10)})
			}
		}
		return nil
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename="+filename)
	w.Write([]byte("\xef\xbb\xbf"))
	csvOut := csv.NewWriter(w)
	if err = csvOut.WriteAll(rows); err != nil {
		log.Print("csv write failed")
	}
}
