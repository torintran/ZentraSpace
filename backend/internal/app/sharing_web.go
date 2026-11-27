package app

import (
	"net/http"
	"time"
)

func (a *App) share(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := Page{Screen: "verify", InviteID: id}
	i, bank, err := a.sharedBill(r)
	if err != nil {
		a.problem(w, r, err)
		return
	}
	if i != nil && i.Issued {
		p.Screen, p.Invoice = "shared", i
		if _, err := vietQR(bank, i.Remaining(), i.PaymentCode); err == nil {
			p.PaymentBank = bank
		}
	}
	a.page(w, r, p)
}
func (a *App) verify(w http.ResponseWriter, r *http.Request) {
	route := r.PathValue("id")
	if !a.allow("verify:"+ipOf(r), 30, time.Minute) {
		http.Error(w, "Chưa thể xác minh. Hãy thử lại sau.", 429)
		return
	}
	org, id, ok := a.store.shareTarget(route)
	if !ok {
		a.page(w, r, Page{Screen: "verify", InviteID: route, Error: "Mã không hợp lệ hoặc lời mời đã hết hạn. Hãy liên hệ người quản lý."})
		return
	}
	verified := false
	err := a.store.book(r.Context(), org, true, func(b *Book) error {
		if v := b.invite(id); v != nil {
			verified = v.verify(r.Form.Get("code"), time.Now())
		}
		return nil
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	if !verified {
		a.page(w, r, Page{Screen: "verify", InviteID: route, Error: "Mã không hợp lệ hoặc lời mời đã hết hạn. Hãy liên hệ người quản lý."})
		return
	}
	a.cookie(w, "zs_reader", a.newSession(session{role: "share:" + route, organization: org}, 30*time.Minute), "/p/"+route, 1800)
	http.Redirect(w, r, "/p/"+route, 303)
}
