package app

import (
	"crypto/subtle"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	qrcode "github.com/skip2/go-qrcode"
	"golang.org/x/text/unicode/norm"
)

type BankScreen struct {
	Account                      BankAccount
	Options                      []BankOption
	BankName                     string
	Transfers                    []BankTransfer
	Reviewed                     bool
	Next, Previous               int
	HasNext, HasPrevious         bool
	MockAvailable, MockAutoMatch bool
	Search                       string
	Matches                      int
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
	search := strings.TrimSpace(query.Get("q"))
	if len([]rune(search)) > 120 {
		search = string([]rune(search)[:120])
	}
	words := strings.Fields(bankSearchText(search))
	screen := &BankScreen{Reviewed: query.Get("status") == "reviewed", Search: search}
	options, err := a.store.bankOptions(r.Context())
	if err != nil {
		a.problem(w, r, err)
		return
	}
	screen.Options = options
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
	err = a.store.book(r.Context(), org, false, func(b *Book) error {
		screen.Account = b.Bank
		for _, option := range options {
			if option.BIN == b.Bank.BIN {
				screen.BankName = option.Name
				break
			}
		}
		screen.MockAvailable = org == pilotOrg && len(a.mockBankSecret) >= 32 &&
			b.Bank.BIN == mockBankBIN && b.Bank.Account == mockBankAccount
		screen.MockAutoMatch = screen.MockAvailable && b.MockBankAutoMatch
		matches := make([]BankTransfer, 0)
		for j := len(b.Transfers) - 1; j >= 0; j-- {
			tr := b.Transfers[j]
			if (tr.InvoiceID != "") != screen.Reviewed || !matchesBankSearch(tr, words) {
				continue
			}
			matches = append(matches, tr)
		}
		screen.Matches = len(matches)
		offset = min(offset/pageSize, max(0, len(matches)-1)/pageSize) * pageSize
		end := min(len(matches), offset+pageSize)
		screen.Transfers = matches[offset:end]
		screen.HasNext, screen.HasPrevious = end < len(matches), offset > 0
		screen.Next, screen.Previous = end, max(0, offset-pageSize)
		return nil
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	a.page(w, r, p)
}

func bankSearchText(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case unicode.Is(unicode.Mn, r):
			return -1
		case r == 'đ':
			return 'd'
		default:
			return r
		}
	}, norm.NFD.String(strings.ToLower(s)))
}

func matchesBankSearch(tr BankTransfer, words []string) bool {
	if len(words) == 0 {
		return true
	}
	text := bankSearchText(tr.Reference + " " + tr.Memo + " " + tr.Date + " " + strconv.FormatInt(tr.Amount, 10))
	for _, word := range words {
		if !strings.Contains(text, word) {
			return false
		}
	}
	return true
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
				if !tr.ManualCash && entry.Kind == "Thu" && entry.Method != cashPayment && entry.Amount == tr.Amount && !linked[entry.Key] && !i.Reversed(entry.Key) {
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
	var bin string
	if r.PathValue("operation") == "account" {
		bin = strings.TrimSpace(r.Form.Get("bank"))
		if bin == "" {
			bin = strings.TrimSpace(r.Form.Get("bin"))
		}
		options, err := a.store.bankOptions(r.Context())
		if err != nil {
			a.problem(w, r, err)
			return
		}
		if !slices.ContainsFunc(options, func(option BankOption) bool { return option.BIN == bin }) {
			a.problem(w, r, inputError{fmt.Errorf("Hãy chọn ngân hàng trong danh sách.")})
			return
		}
	}
	err := a.store.book(r.Context(), org, true, func(b *Book) error {
		switch r.PathValue("operation") {
		case "cash-receipt":
			amount, err := number(r, "amount", 1, 100000000000)
			if err != nil {
				return err
			}
			if err := b.receiveCash(r.Form.Get("key"), r.Form.Get("date"), strings.TrimSpace(r.Form.Get("reference")), strings.TrimSpace(r.Form.Get("note")), amount); err != nil {
				return inputError{err}
			}
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
			bank := BankAccount{bin, strings.TrimSpace(r.Form.Get("account")), strings.TrimSpace(r.Form.Get("holder"))}
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

type sharedInvoice struct {
	invoice  *Invoice
	personal *PersonalBill
}

func (a *App) sharedBill(r *http.Request) (sharedInvoice, BankAccount, error) {
	route := r.PathValue("id")
	s, _ := a.getSession(r, "zs_reader")
	if s.role != "share:"+route {
		return sharedInvoice{}, BankAccount{}, nil
	}
	org, id, ok := a.store.shareTarget(route)
	if !ok || org != s.organization {
		return sharedInvoice{}, BankAccount{}, nil
	}
	var result sharedInvoice
	var bank BankAccount
	err := a.store.book(r.Context(), org, false, func(b *Book) error {
		v := b.invite(id)
		if v != nil && v.Used && v.usable(time.Now(), a.secure) {
			if invoice := b.invoice(v.InvoiceID); invoice != nil {
				if v.ResidentShare != 0 {
					result.personal = invoice.personalBill(v.ResidentShare)
				} else {
					result.invoice, bank = invoice, b.Bank
				}
			}
		}
		return nil
	})
	return result, bank, err
}

func (a *App) shareQR(w http.ResponseWriter, r *http.Request) {
	shared, bank, err := a.sharedBill(r)
	if err != nil {
		a.problem(w, r, err)
		return
	}
	i := shared.invoice
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
