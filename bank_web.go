package main

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	qrcode "github.com/skip2/go-qrcode"
)

type BankScreen struct {
	Account                      BankAccount
	Transfers                    []BankTransfer
	Reviewed                     bool
	Next, Previous               int
	HasNext, HasPrevious         bool
	MockAvailable, MockAutoMatch bool
}
type TransferReview struct {
	Transfer  BankTransfer
	Suggested *Invoice
	Selected  *Invoice
	Invoices  []Invoice
	Existing  []Entry
}

func (a *App) bankOwner(w http.ResponseWriter, r *http.Request) (string, string, bool) {
	s, id, ok := a.staff(w, r)
	if !ok {
		return "", "", false
	}
	if s.role != "owner" {
		http.Error(w, "Bạn không có quyền đối soát.", http.StatusForbidden)
		return "", "", false
	}
	csrf := hash(id + ":csrf")
	if r.Method == http.MethodPost && subtle.ConstantTimeCompare([]byte(r.Form.Get("csrf")), []byte(csrf)) != 1 {
		http.Error(w, "Yêu cầu không hợp lệ.", http.StatusForbidden)
		return "", "", false
	}
	return s.organization, csrf, true
}

func (a *App) bankPage(w http.ResponseWriter, r *http.Request) {
	org, csrf, ok := a.bankOwner(w, r)
	if !ok {
		return
	}
	const pageSize = 20
	query := r.URL.Query()
	offset, _ := strconv.Atoi(query.Get("offset"))
	if offset < 0 || offset > 3000 {
		offset = 0
	}
	screen := &BankScreen{Reviewed: query.Get("status") == "reviewed", Previous: max(0, offset-pageSize), HasPrevious: offset > 0}
	p := Page{Screen: "bank", Role: "owner", CSRF: csrf, Bank: screen}
	if query.Get("saved") == "account" {
		p.Message = "Đã lưu tài khoản nhận tiền."
	} else if query.Get("saved") == "statement" {
		newCount, e1 := strconv.Atoi(query.Get("added"))
		skipped, e2 := strconv.Atoi(query.Get("skipped"))
		if e1 == nil && e2 == nil && newCount >= 0 && skipped >= 0 && newCount <= 150 && skipped <= 150 {
			p.Message = fmt.Sprintf("Đã thêm %d giao dịch, bỏ qua %d dòng đã có. Công nợ chỉ đổi sau khi bạn duyệt.", newCount, skipped)
		}
	} else if query.Get("saved") == "review" {
		p.Message = "Đã đối soát giao dịch."
	}
	err := a.store.book(r.Context(), org, false, func(b *Book) error {
		screen.Account = b.Bank
		screen.MockAvailable = org == pilotOrg && len(a.mockBankSecret) >= 32 &&
			b.Bank.BIN == mockBankBIN && b.Bank.Account == mockBankAccount
		screen.MockAutoMatch = screen.MockAvailable && b.MockBankAutoMatch
		count := 0
		for j := len(b.Transfers) - 1; j >= 0; j-- {
			tr := b.Transfers[j]
			if (tr.InvoiceID != "") != screen.Reviewed {
				continue
			}
			if count >= offset+pageSize {
				screen.HasNext = true
				break
			}
			if count >= offset {
				screen.Transfers = append(screen.Transfers, tr)
			}
			count++
		}
		screen.Next = offset + pageSize
		return nil
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	a.page(w, r, p)
}

func (a *App) bankReview(w http.ResponseWriter, r *http.Request) {
	org, csrf, ok := a.bankOwner(w, r)
	if !ok {
		return
	}
	review := &TransferReview{}
	err := a.store.book(r.Context(), org, false, func(b *Book) error {
		tr := b.transfer(r.PathValue("id"))
		if tr == nil {
			return errMissing
		}
		review.Transfer = *tr
		review.Suggested = b.suggestInvoice(tr.Memo)
		selected := r.URL.Query().Get("invoice")
		if tr.InvoiceID != "" {
			selected = tr.InvoiceID
		} else if selected == "" && review.Suggested != nil {
			selected = review.Suggested.ID
		}
		linked := make(map[string]bool)
		for _, prior := range b.Transfers {
			if prior.InvoiceID == selected {
				linked[prior.EntryKey] = true
			}
		}
		for j := len(b.Invoices) - 1; j >= 0; j-- {
			i := &b.Invoices[j]
			if !i.Issued {
				continue
			}
			review.Invoices = append(review.Invoices, *i)
			if i.ID != selected {
				continue
			}
			review.Selected = i
			for _, entry := range i.Entries {
				if entry.Kind == "Thu" && entry.Method != cashPayment && entry.Amount == tr.Amount && !linked[entry.Key] && !i.Reversed(entry.Key) {
					review.Existing = append(review.Existing, entry)
				}
			}
		}
		return nil
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	a.page(w, r, Page{Screen: "bank-review", Role: "owner", CSRF: csrf, BankReview: review})
}

func (a *App) bankAction(w http.ResponseWriter, r *http.Request) {
	org, _, ok := a.bankOwner(w, r)
	if !ok {
		return
	}
	next := "/bank"
	err := a.store.book(r.Context(), org, true, func(b *Book) error {
		switch r.PathValue("operation") {
		case "mock-settings":
			if org != pilotOrg || len(a.mockBankSecret) < 32 || b.Bank.BIN != mockBankBIN || b.Bank.Account != mockBankAccount {
				return errMissing
			}
			mode := r.Form.Get("mode")
			if mode != "auto" && mode != "manual" {
				return errInvalid
			}
			b.MockBankAutoMatch = mode == "auto"
		case "account":
			if r.Form.Get("confirm") != "yes" {
				return errInvalid
			}
			bank := BankAccount{strings.TrimSpace(r.Form.Get("bin")), strings.TrimSpace(r.Form.Get("account")), strings.TrimSpace(r.Form.Get("holder"))}
			if err := b.setBank(bank); err != nil {
				return inputError{err}
			}
			next += "?saved=account"
		case "import":
			added, skipped, err := b.importStatement(r.Form.Get("csv"))
			if err != nil {
				return inputError{err}
			}
			next += fmt.Sprintf("?saved=statement&added=%d&skipped=%d", added, skipped)
		case "approve":
			if r.Form.Get("confirm") != "yes" {
				return errInvalid
			}
			existing := r.Form.Get("existing")
			if existing == "" {
				return errInvalid
			}
			if existing == "new" {
				existing = ""
			}
			if err := b.reconcile(r.PathValue("id"), r.Form.Get("invoice"), existing); err != nil {
				return inputError{err}
			}
			next += "?status=reviewed&saved=review"
		default:
			return errMissing
		}
		return nil
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (a *App) sharedBill(r *http.Request) (*Invoice, BankAccount, error) {
	route := r.PathValue("id")
	s, _ := a.getSession(r, "zs_reader")
	if s.role != "share:"+route {
		return nil, BankAccount{}, nil
	}
	org, id, ok := a.store.shareTarget(route)
	if !ok || org != s.organization {
		return nil, BankAccount{}, nil
	}
	var invoice *Invoice
	var bank BankAccount
	err := a.store.book(r.Context(), org, false, func(b *Book) error {
		v := b.invite(id)
		if v != nil && !v.Revoked && v.Used && time.Now().Before(v.Expires) {
			invoice = b.invoice(v.InvoiceID)
			bank = b.Bank
		}
		return nil
	})
	return invoice, bank, err
}

func (a *App) shareQR(w http.ResponseWriter, r *http.Request) {
	i, bank, err := a.sharedBill(r)
	if err != nil {
		a.problem(w, r, err)
		return
	}
	if i == nil || !i.Issued {
		http.NotFound(w, r)
		return
	}
	payload, err := vietQR(bank, i.Remaining(), i.PaymentCode)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	png, err := qrcode.Encode(payload, qrcode.Medium, 280)
	if err != nil {
		a.problem(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}
