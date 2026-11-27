package main

import (
	"encoding/csv"
	"log"
	"net/http"
	"strconv"
	"time"
)

func (a *App) cash(w http.ResponseWriter, r *http.Request) {
	s, id, ok := a.staff(w, r)
	if !ok {
		return
	}
	if s.role != "owner" {
		http.Error(w, "Bạn không có quyền xem sổ thu chi.", http.StatusForbidden)
		return
	}
	period := periodOf(r.URL.Query().Get("period"))
	page := Page{Screen: "cash", Role: s.role, Period: period, CSRF: hash(id + ":csrf"), EntryDate: period + "-01"}
	if period == currentPeriod() {
		page.EntryDate = time.Now().In(vietnam).Format("2006-01-02")
	}
	if r.URL.Query().Get("saved") == "1" {
		page.Message = "Đã ghi giao dịch."
	} else if r.URL.Query().Get("saved") == "receipt" {
		page.Message = "Đã đính kèm chứng từ."
	}
	if err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		page.Cash = b.cashReport(period)
		return nil
	}); err != nil {
		a.problem(w, r, err)
		return
	}
	if r.URL.Path != "/cash.csv" {
		var err error
		page.CashReceipts, err = a.store.receiptKeys(r.Context(), s.organization)
		if err != nil {
			a.problem(w, r, err)
			return
		}
		a.page(w, r, page)
		return
	}
	rows := [][]string{{"Ngày", "Loại", "Phòng", "Danh mục", "Mã chứng từ", "Ghi chú", "Tiền vào (VND)", "Tiền ra (VND)", "Mã bút toán"}}
	for _, e := range page.Cash.Rows {
		in, out := "", ""
		if e.Income() {
			in = strconv.FormatInt(e.Amount, 10)
		} else {
			out = strconv.FormatInt(e.Amount, 10)
		}
		rows = append(rows, []string{e.Date, e.KindLabel(), plainCSV(e.RoomName), plainCSV(e.Category),
			plainCSV(e.Reference), plainCSV(e.Note), in, out, plainCSV(e.Key)})
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=so-thu-chi-"+period+".csv")
	w.Write([]byte("\xef\xbb\xbf"))
	if err := csv.NewWriter(w).WriteAll(rows); err != nil {
		log.Print("cash CSV write failed")
	}
}
