package main

import (
	"context"
	"crypto/subtle"
	"encoding/csv"
	"errors"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type session struct {
	role, organization string
	expires            time.Time
}
type limiter struct {
	count int
	until time.Time
}
type App struct {
	store          *Store
	origin         string
	secure         bool
	accounts       map[string]staffLogin
	mockBankSecret string
	pages          *template.Template
	mu             sync.Mutex
	sessions       map[string]session
	limits         map[string]limiter
}

const pilotOrg = "local-pilot"

type RoomView struct {
	Room
	Done       bool
	Lease      LeaseTerms
	HasLease   bool
	Renewed    bool
	Handover   bool
	Adjustment MeterAdjustment
}
type Page struct {
	Screen, Role, Tab, Period, Error, Message, CSRF, InviteID, ShareURL, Code string
	Filter                                                                    string
	BillingSettings                                                           BillingSettings
	Cash                                                                      CashReport
	CashReceipts                                                              map[string]bool
	EntryDate                                                                 string
	Rooms                                                                     []RoomView
	RoomDetail                                                                *Room
	RoomFees                                                                  []RoomFee
	Lease                                                                     LeaseTerms
	LeaseHistory                                                              []LeaseTerms
	LeaseForm                                                                 bool
	RoomForm, LockOpening, LockRates                                          bool
	Dashboard                                                                 Dashboard
	Invoices                                                                  []Invoice
	Invoice                                                                   *Invoice
	ReconciledEntries                                                         map[string]bool
	Invites                                                                   []Invite
	Import                                                                    *ImportPreview
	Status                                                                    int
	Bank                                                                      *BankScreen
	BankReview                                                                *TransferReview
	PaymentBank                                                               BankAccount
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "mock-bank" {
		log.Fatal(runMockBank())
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if len(os.Args) == 2 && os.Args[1] == "migrate" {
		if err := migrate(ctx); err != nil {
			log.Fatal(err)
		}
		return
	}
	store, err := openStore(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer store.pool.Close()
	if len(os.Args) > 1 && (os.Args[1] == "seed-demo" || os.Args[1] == "seed-bank-demo" || os.Args[1] == "seed-lease-demo") {
		if len(os.Args) != 3 || os.Getenv("ZS_LOCAL_HTTP") != "1" {
			log.Fatal("Usage in local HTTP mode: zentraspace seed-demo YYYY-MM")
		}
		if err := store.book(ctx, pilotOrg, true, func(b *Book) error {
			if os.Args[1] == "seed-lease-demo" {
				return b.seedLeaseDemo(os.Args[2])
			}
			if os.Args[1] == "seed-bank-demo" {
				if len(os.Getenv("ZS_MOCK_BANK_SECRET")) < 32 || (b.Bank.Configured() && (b.Bank.BIN != mockBankBIN || b.Bank.Account != mockBankAccount)) {
					return errors.New("Use a dedicated mock-bank stack with a synthetic beneficiary.")
				}
				if err := b.setBank(BankAccount{mockBankBIN, mockBankAccount, "NGAN HANG MO PHONG"}); err != nil {
					return err
				}
			}
			return b.seedDemo(os.Args[2])
		}); err != nil {
			log.Fatal(err)
		}
		log.Printf("Demo data ready for %s", os.Args[2])
		return
	}
	if err := store.book(ctx, pilotOrg, false, func(*Book) error { return nil }); err != nil {
		log.Fatal("Cannot open pilot data. Check the database configuration and encryption key.")
	}
	origin := strings.TrimRight(os.Getenv("ZS_ORIGIN"), "/")
	u, err := url.Parse(origin)
	local := os.Getenv("ZS_LOCAL_HTTP") == "1"
	if err != nil || u.Host == "" || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || u.User != nil ||
		(!local && u.Scheme != "https") || (local && (u.Scheme != "http" || (u.Hostname() != "localhost" && u.Hostname() != "127.0.0.1"))) {
		log.Fatal("Configure an HTTPS origin, or an explicit loopback HTTP origin for local testing.")
	}
	accounts, err := loadStaff()
	if err != nil {
		log.Fatal(err)
	}
	a := &App{store: store, origin: origin, secure: !local, sessions: map[string]session{}, limits: map[string]limiter{}, accounts: accounts}
	if local {
		a.mockBankSecret = os.Getenv("ZS_MOCK_BANK_SECRET")
	}
	a.pages = template.Must(template.New("app").Funcs(template.FuncMap{
		"money": money, "key": token, "clock": func(t time.Time) string { return t.In(vietnam).Format("15:04 · 02/01/2006") },
	}).ParseFS(webFiles, "web/app.html"))
	server := &http.Server{Addr: ":8080", Handler: a.routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 12 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	log.Print("ZentraSpace listening on :8080")
	log.Fatal(server.ListenAndServe())
}

var vietnam = time.FixedZone("ICT", 7*60*60)

func money(n int64) string {
	s := strconv.FormatInt(n, 10)
	sign := ""
	if magnitude, negative := strings.CutPrefix(s, "-"); negative {
		sign, s = "-", magnitude
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return sign + s + " ₫"
}
func currentPeriod() string { return time.Now().In(vietnam).Format("2006-01") }
func periodOf(s string) string {
	if _, err := time.Parse("2006-01", s); err != nil {
		return currentPeriod()
	}
	return s
}
func number(r *http.Request, key string, low, high int64) (int64, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(r.Form.Get(key)), 10, 64)
	if err != nil || n < low || n > high {
		return 0, errInvalid
	}
	return n, nil
}
func (a *App) allow(key string, n int, span time.Duration) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	if len(a.limits) > 2048 {
		for k, v := range a.limits {
			if now.After(v.until) {
				delete(a.limits, k)
			}
		}
	}
	v, exists := a.limits[key]
	if !exists && len(a.limits) > 4096 {
		return false
	}
	if !now.Before(v.until) {
		v = limiter{until: now.Add(span)}
	}
	if v.count >= n {
		return false
	}
	v.count++
	a.limits[key] = v
	return true
}
func (a *App) newSession(role, org string, d time.Duration) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	for k, v := range a.sessions {
		if time.Now().After(v.expires) {
			delete(a.sessions, k)
		}
	}
	id := token()
	a.sessions[hash(id)] = session{role, org, time.Now().Add(d)}
	return id
}
func (a *App) getSession(r *http.Request, name string) (session, string) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return session{}, ""
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	k := hash(cookie.Value)
	s := a.sessions[k]
	if !time.Now().Before(s.expires) {
		delete(a.sessions, k)
		return session{}, ""
	}
	return s, cookie.Value
}
func (a *App) cookie(w http.ResponseWriter, name, value, path string, age int) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: path, MaxAge: age, Secure: a.secure, HttpOnly: true, SameSite: http.SameSiteLaxMode})
}
func (a *App) page(w http.ResponseWriter, p Page) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if p.Status != 0 {
		w.WriteHeader(p.Status)
	}
	if err := a.pages.ExecuteTemplate(w, "app.html", p); err != nil {
		log.Print("template rendering failed: ", err)
	}
}
func (a *App) routes() http.Handler {
	m := http.NewServeMux()
	m.Handle("GET /assets/", http.FileServer(http.FS(webAssets())))
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if a.store.pool.Ping(ctx) != nil {
			http.Error(w, "unavailable", 503)
			return
		}
		w.Write([]byte("ok"))
	})
	m.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) { a.page(w, Page{Screen: "login"}) })
	m.HandleFunc("GET /privacy", func(w http.ResponseWriter, r *http.Request) { a.page(w, Page{Screen: "privacy"}) })
	m.HandleFunc("POST /login", a.login)
	m.HandleFunc("GET /{$}", a.home)
	m.HandleFunc("GET /invoice/{id}", a.invoice)
	m.HandleFunc("GET /export", a.export)
	m.HandleFunc("GET /cash", a.cash)
	m.HandleFunc("GET /cash.csv", a.cash)
	m.HandleFunc("GET /cash/receipt/{key}", a.cashReceipt)
	m.HandleFunc("POST /cash/receipt/{key}", a.cashReceipt)
	m.HandleFunc("GET /rooms/template.csv", a.roomTemplate)
	m.HandleFunc("GET /rooms/{id}", a.roomDetail)
	m.HandleFunc("GET /bank", a.bankPage)
	m.HandleFunc("GET /bank/review/{id}", a.bankReview)
	m.HandleFunc("POST /bank/{operation}", a.bankAction)
	m.HandleFunc("POST /bank/approve/{id}", func(w http.ResponseWriter, r *http.Request) {
		r.SetPathValue("operation", "approve")
		a.bankAction(w, r)
	})
	m.HandleFunc("POST /action/{action}", a.action)
	m.HandleFunc("GET /p/{id}", a.share)
	m.HandleFunc("GET /p/{id}/qr", a.shareQR)
	m.HandleFunc("POST /p/{id}", a.verify)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'self'; script-src 'self'; img-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		// Keep the origin on browser form submissions without disclosing invitation paths.
		w.Header().Set("Referrer-Policy", "strict-origin")
		w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
		w.Header().Set("Cache-Control", "no-store")
		if a.secure {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		}
		if r.Method == http.MethodPost && r.URL.Path == mockBankPath {
			a.mockWebhook(w, r)
			return
		}
		if r.Method == "POST" {
			if r.Header.Get("Origin") != a.origin {
				http.Error(w, "Yêu cầu không hợp lệ.", http.StatusForbidden)
				return
			}
			var err error
			if strings.HasPrefix(r.URL.Path, "/cash/receipt/") {
				r.Body = http.MaxBytesReader(w, r.Body, receiptBytes+4096)
				err = r.ParseMultipartForm(receiptBytes + 4096)
				if r.MultipartForm != nil {
					defer r.MultipartForm.RemoveAll()
				}
			} else {
				r.Body = http.MaxBytesReader(w, r.Body, 16384)
				err = r.ParseForm()
			}
			if err != nil {
				http.Error(w, "Yêu cầu không hợp lệ.", http.StatusBadRequest)
				return
			}
			if _, err := url.ParseQuery(r.URL.RawQuery); err != nil || len(r.URL.Query()) > 0 {
				http.Error(w, "Yêu cầu không hợp lệ.", http.StatusBadRequest)
				return
			}
		}
		m.ServeHTTP(w, r)
	})
}
func ipOf(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
func (a *App) login(w http.ResponseWriter, r *http.Request) {
	if !a.allow("login:"+ipOf(r), 10, time.Minute) {
		http.Error(w, "Hãy thử lại sau một phút.", 429)
		return
	}
	username := strings.ToLower(strings.TrimSpace(r.Form.Get("account")))
	account, known := a.accounts[username]
	if !known {
		account = a.accounts["owner"]
	}
	if bcrypt.CompareHashAndPassword(account.passwordHash, []byte(r.Form.Get("password"))) != nil || !known {
		a.page(w, Page{Screen: "login", Status: http.StatusUnauthorized, Error: "Kiểm tra lại tài khoản hoặc mật khẩu."})
		return
	}
	id := a.newSession(account.role, account.organization, 8*time.Hour)
	a.cookie(w, "zs_staff", id, "/", 8*3600)
	http.Redirect(w, r, "/", 303)
}

func (a *App) staff(w http.ResponseWriter, r *http.Request) (session, string, bool) {
	s, id := a.getSession(r, "zs_staff")
	if s.role != "owner" && s.role != "meter" {
		http.Redirect(w, r, "/login", 303)
		return s, id, false
	}
	return s, id, true
}
func (a *App) home(w http.ResponseWriter, r *http.Request) {
	s, id, ok := a.staff(w, r)
	if !ok {
		return
	}
	tab := r.URL.Query().Get("tab")
	if tab != "rooms" && tab != "meters" && tab != "billing" && tab != "dashboard" {
		tab = "dashboard"
	}
	if s.role == "meter" {
		tab = "meters"
	}
	period := periodOf(r.URL.Query().Get("period"))
	filter := r.URL.Query().Get("filter")
	if (tab != "meters" || filter != "missing") && (tab != "billing" || (filter != "draft" && filter != "due")) {
		filter = ""
	}
	p := Page{Screen: "home", Role: s.role, Tab: tab, Period: period, Filter: filter, CSRF: hash(id + ":csrf")}
	p.RoomForm = tab == "rooms" && r.URL.Query().Get("add") == "1"
	if r.URL.Query().Get("saved") == "1" {
		p.Message = "Đã lưu."
	}
	if err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		if tab == "dashboard" {
			p.Dashboard = b.dashboard(period)
			return nil
		}
		if tab == "billing" {
			p.BillingSettings = b.Billing[period]
		}
		for _, room := range b.Rooms {
			lease := b.leaseAt(room, period)
			var renewed bool
			reading := b.reading(room.ID, period)
			done := reading != nil
			if ((tab == "meters" || tab == "billing") && !lease.ActiveAt(period)) || (filter == "missing" && done) {
				continue
			}
			if s.role == "meter" {
				room = Room{ID: room.ID, Name: room.Name, Start: room.Start, Electric: room.Electric, Water: room.Water}
			} else if tab == "rooms" && room.Start <= period {
				if lease.ActiveAt(period) {
					room.Tenant = lease.Tenant
					renewed = lease.Renewal
				} else {
					room.Tenant = ""
				}
			}
			view := RoomView{Room: room, Done: done, HasLease: lease.ActiveAt(period), Renewed: renewed}
			if tab == "rooms" {
				view.Lease = lease
			}
			if tab == "meters" {
				if reading != nil {
					view.Electric, view.Water = reading.ElectricAfter, reading.WaterAfter
					view.Handover = reading.Adjustment.Handover
					if s.role == "owner" {
						view.Adjustment = reading.Adjustment
					}
				} else if opening, err := b.meterOpening(room, period); err == nil {
					view.Electric, view.Water, view.Handover = opening.Electric, opening.Water, opening.Handover
				}
			}
			p.Rooms = append(p.Rooms, view)
		}
		if s.role == "owner" {
			for _, i := range b.Invoices {
				if i.Period == period && (filter != "draft" || !i.Issued) &&
					(filter != "due" || (i.Issued && i.Remaining() > 0)) {
					p.Invoices = append(p.Invoices, i)
				}
			}
		}
		return nil
	}); err != nil {
		a.problem(w, err)
		return
	}
	a.page(w, p)
}
func (a *App) invoice(w http.ResponseWriter, r *http.Request) {
	s, id, ok := a.staff(w, r)
	if !ok {
		return
	}
	if s.role != "owner" {
		http.Error(w, "Bạn không có quyền xem bảng kê.", 403)
		return
	}
	p := Page{Screen: "invoice", Role: s.role, CSRF: hash(id + ":csrf")}
	err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		p.Invoice = b.invoice(r.PathValue("id"))
		if p.Invoice == nil {
			return errMissing
		}
		p.ReconciledEntries = make(map[string]bool)
		for _, tr := range b.Transfers {
			if tr.InvoiceID == p.Invoice.ID {
				p.ReconciledEntries[tr.EntryKey] = true
			}
		}
		for _, v := range b.Invites {
			if v.InvoiceID == p.Invoice.ID && !v.Revoked && time.Now().Before(v.Expires) {
				p.Invites = append(p.Invites, v)
			}
		}
		return nil
	})
	if err != nil {
		a.problem(w, err)
		return
	}
	a.page(w, p)
}
func plainCSV(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "=") || strings.HasPrefix(s, "+") || strings.HasPrefix(s, "-") || strings.HasPrefix(s, "@") {
		return "'" + s
	}
	return s
}
func (a *App) roomDetail(w http.ResponseWriter, r *http.Request) {
	s, sessionID, ok := a.staff(w, r)
	if !ok {
		return
	}
	if s.role != "owner" {
		http.Error(w, "Bạn không có quyền xem chi tiết phòng.", http.StatusForbidden)
		return
	}
	p := Page{Screen: "room", Role: s.role, CSRF: hash(sessionID + ":csrf"), Period: periodOf(r.URL.Query().Get("period"))}
	p.RoomForm = r.URL.Query().Get("edit") == "1"
	p.LeaseForm = !p.RoomForm && r.URL.Query().Get("contract") == "1"
	if r.URL.Query().Get("saved") == "1" {
		p.Message = "Đã lưu thông tin phòng."
	} else if r.URL.Query().Get("saved") == "lease" {
		p.Message = "Đã lưu điều khoản."
	} else if r.URL.Query().Get("saved") == "fee" {
		p.Message = "Đã lưu khoản phí."
	}
	if err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		p.RoomDetail = b.room(r.PathValue("id"))
		if p.RoomDetail == nil {
			return errMissing
		}
		p.LockOpening, p.LockRates = b.roomLocks(p.RoomDetail.ID)
		p.Lease = b.leaseAt(*p.RoomDetail, p.Period)
		p.LeaseHistory = b.leaseHistory(*p.RoomDetail)
		for _, fee := range b.RoomFees {
			if fee.RoomID == p.RoomDetail.ID {
				p.RoomFees = append(p.RoomFees, fee)
			}
		}
		for _, i := range b.Invoices {
			if i.RoomID == p.RoomDetail.ID && i.Period == p.Period {
				p.Invoice = &i
				break
			}
		}
		return nil
	}); err != nil {
		a.problem(w, err)
		return
	}
	a.page(w, p)
}

func (a *App) roomTemplate(w http.ResponseWriter, r *http.Request) {
	s, _, ok := a.staff(w, r)
	if !ok {
		return
	}
	if s.role != "owner" {
		http.Error(w, "Bạn không có quyền nhập danh sách phòng.", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=mau-nhap-phong.csv")
	io.WriteString(w, roomCSVHeader+"\r\n")
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
		a.problem(w, err)
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

type inputError struct{ error }

func (a *App) problem(w http.ResponseWriter, err error) {
	var user inputError
	if errors.Is(err, errMissing) {
		http.Error(w, errMissing.Error(), 404)
		return
	}
	if errors.As(err, &user) || errors.Is(err, errInvalid) || errors.Is(err, errLocked) {
		a.page(w, Page{Screen: "error", Status: http.StatusUnprocessableEntity, Error: err.Error()})
		return
	}
	log.Printf("request failed (%T)", err)
	http.Error(w, "Không thể xử lý lúc này.", 500)
}
func (a *App) action(w http.ResponseWriter, r *http.Request) {
	s, sessionID, ok := a.staff(w, r)
	if !ok {
		return
	}
	action := r.PathValue("action")
	if subtle.ConstantTimeCompare([]byte(r.Form.Get("csrf")), []byte(hash(sessionID+":csrf"))) != 1 || (s.role != "owner" && action != "read" && action != "logout") {
		http.Error(w, "Bạn không có quyền thực hiện thao tác này.", 403)
		return
	}
	if action == "logout" {
		a.mu.Lock()
		delete(a.sessions, hash(sessionID))
		a.mu.Unlock()
		a.cookie(w, "zs_staff", "", "/", -1)
		http.Redirect(w, r, "/login", 303)
		return
	}
	period := periodOf(r.Form.Get("period"))
	next := "/?tab=billing&saved=1&period=" + period
	p := Page{Screen: "invite", Role: s.role, CSRF: hash(sessionID + ":csrf")}
	if action == "import-preview" {
		var preview ImportPreview
		err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
			result, err := b.importRooms(period, r.Form.Get("csv"), false)
			preview = result
			if err != nil {
				return inputError{err}
			}
			return nil
		})
		if err != nil {
			a.problem(w, err)
			return
		}
		p.Screen, p.Import = "import", &preview
		a.page(w, p)
		return
	}
	err := a.store.book(r.Context(), s.organization, true, func(b *Book) error {
		value := func(key string) string { return strings.TrimSpace(r.Form.Get(key)) }
		id := value("id")
		switch action {
		case "cash-entry":
			amount, err := number(r, "amount", 1, 100000000000)
			if err != nil {
				return err
			}
			entry := CashEntry{Key: value("key"), Kind: value("kind"), Date: value("date"),
				RoomID: value("room"), Category: value("category"), Reference: value("reference"),
				Note: value("note"), Amount: amount}
			if err := b.postCash(entry); err != nil {
				return inputError{err}
			}
			next = "/cash?saved=1&period=" + entry.Date[:7]
		case "room-fee":
			amount, err := number(r, "amount", 1, 100000000)
			kind := value("kind")
			if err != nil || (kind != "once" && kind != "monthly") {
				return errInvalid
			}
			if err := b.addRoomFee(RoomFee{ID: value("key"), RoomID: id, Label: value("label"),
				Start: value("start"), Monthly: kind == "monthly", Amount: amount}); err != nil {
				return inputError{err}
			}
			next = "/rooms/" + id + "?saved=fee&period=" + period + "#room-fees"
		case "room-fee-stop":
			if err := b.stopRoomFee(id, value("fee"), value("end")); err != nil {
				return inputError{err}
			}
			next = "/rooms/" + id + "?saved=fee&period=" + period + "#room-fees"
		case "billing-settings":
			fee, err := number(r, "commonFee", 0, 100000000)
			if err != nil {
				return err
			}
			if err := b.setBilling(period, BillingSettings{DueDate: value("dueDate"), CommonFeeLabel: value("commonFeeLabel"), CommonFee: fee}); err != nil {
				return inputError{err}
			}
		case "lease":
			rent, e1 := number(r, "rent", 0, 100000000)
			electric, e2 := number(r, "electricRate", 0, 100000)
			water, e3 := number(r, "waterRate", 0, 1000000)
			if e1 != nil || e2 != nil || e3 != nil {
				return errInvalid
			}
			terms := LeaseTerms{RoomID: id, Effective: value("effective"), Tenant: value("tenant"),
				Rent: rent, ElectricRate: electric, WaterRate: water, Renewal: value("renewal") == "yes"}
			if value("months") != "" {
				months, err := number(r, "months", 1, 120)
				if err != nil {
					return err
				}
				terms.Months = int(months)
			}
			if err := b.scheduleLease(terms); err != nil {
				return inputError{err}
			}
			next = "/rooms/" + id + "?saved=lease&period=" + terms.Effective
		case "import":
			if _, err := b.importRooms(period, r.Form.Get("csv"), true); err != nil {
				return inputError{err}
			}
			next = "/?tab=rooms&saved=1&period=" + period
		case "room", "room-edit":
			next = "/?tab=rooms&saved=1&period=" + period
			room := Room{Name: value("name"), Tenant: value("tenant"), Start: period}
			if room.Name == "" || len(room.Name) > 60 || len(room.Tenant) > 150 {
				return errInvalid
			}
			var lockOpening, lockRates bool
			if action == "room-edit" {
				current := b.room(id)
				if current == nil {
					return errMissing
				}
				room.ID, room.Start = current.ID, current.Start
				room.LeaseMonths = current.LeaseMonths
				room.Rent, room.ElectricRate, room.WaterRate = current.Rent, current.ElectricRate, current.WaterRate
				room.Electric, room.Water = current.Electric, current.Water
				lockOpening, lockRates = b.roomLocks(id)
				if lockRates {
					room.Tenant = current.Tenant
				}
				if (lockRates && (r.Form.Has("rent") || r.Form.Has("electricRate") || r.Form.Has("waterRate"))) ||
					(lockOpening && (r.Form.Has("electric") || r.Form.Has("water"))) {
					return inputError{errors.New("Phòng đã có chỉ số hoặc bảng kê phát hành. Hãy mở lại phòng trước khi lưu.")}
				}
				next = "/rooms/" + id + "?saved=1&period=" + period
			} else if value("months") != "" {
				months, err := number(r, "months", 1, 120)
				if err != nil {
					return err
				}
				room.LeaseMonths = int(months)
			}
			if !lockRates {
				var e1, e2, e3 error
				room.Rent, e1 = number(r, "rent", 0, 100000000)
				room.ElectricRate, e2 = number(r, "electricRate", 0, 100000)
				room.WaterRate, e3 = number(r, "waterRate", 0, 1000000)
				if e1 != nil || e2 != nil || e3 != nil {
					return errInvalid
				}
			}
			if !lockOpening {
				var e1, e2 error
				room.Electric, e1 = number(r, "electric", 0, 999999)
				room.Water, e2 = number(r, "water", 0, 999999)
				if e1 != nil || e2 != nil {
					return errInvalid
				}
			}
			if err := b.saveRoom(room, value("revision")); err != nil {
				return inputError{err}
			}
		case "read", "read-special":
			next = "/?tab=meters&saved=1&period=" + period
			if value("filter") == "missing" {
				next += "&filter=missing"
			}
			e, e1 := number(r, "electric", 0, 999999)
			v, e2 := number(r, "water", 0, 999999)
			if e1 != nil || e2 != nil {
				return errInvalid
			}
			if action == "read-special" {
				adjustment, err := formMeterAdjustment(r)
				if err != nil {
					return err
				}
				if err := b.readWithAdjustment(id, period, e, v, adjustment); err != nil {
					return inputError{err}
				}
			} else if err := b.read(id, period, e, v); err != nil {
				return inputError{err}
			}
		case "draft":
			b.draft(period)
			if value("filter") == "draft" || value("filter") == "due" {
				next += "&filter=" + value("filter")
			}
		case "issue":
			i := b.invoice(id)
			if i == nil {
				return errMissing
			}
			if i.Revision() != value("revision") {
				return inputError{errors.New("Bảng kê đã thay đổi. Hãy mở lại bảng kê và kiểm tra trước khi phát hành.")}
			}
			if err := b.issue(id); err != nil {
				return inputError{err}
			}
			next = "/invoice/" + id
		case "post":
			amount, err := number(r, "amount", 1, 100000000000)
			if err != nil {
				return err
			}
			entry := Entry{Key: value("key"), Kind: value("kind"), Amount: amount, Note: value("note"),
				Method: value("method"), Reference: value("reference")}
			if err = b.post(id, entry); err != nil {
				return inputError{err}
			}
			next = "/invoice/" + id
		case "reverse":
			if err := b.reverse(id, value("key"), value("original"), value("note")); err != nil {
				return inputError{err}
			}
			next = "/invoice/" + id
		case "invite":
			i := b.invoice(id)
			if i == nil || !i.Issued {
				return errInvalid
			}
			for j := range b.Invites {
				if b.Invites[j].InvoiceID == id {
					b.Invites[j].Revoked = true
				}
			}
			if len(b.Invites) >= 5000 {
				return inputError{errors.New("Đã đạt giới hạn lời mời của bản thử.")}
			}
			p.Code = accessCode()
			v := Invite{ID: token(), InvoiceID: id, CodeHash: hash(p.Code), Expires: time.Now().Add(24 * time.Hour)}
			b.Invites = append(b.Invites, v)
			p.Invoice = i
			route, err := a.store.shareToken(s.organization, v.ID)
			if err != nil {
				return err
			}
			p.ShareURL = a.origin + "/p/" + route
		case "revoke":
			v := b.invite(id)
			if v == nil {
				return errMissing
			}
			v.Revoked = true
			next = "/invoice/" + v.InvoiceID
		default:
			return errMissing
		}
		return nil
	})
	if err != nil {
		a.problem(w, err)
		return
	}
	if p.Code != "" {
		a.page(w, p)
		return
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}
func (a *App) share(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := Page{Screen: "verify", InviteID: id}
	i, bank, err := a.sharedBill(r)
	if err != nil {
		a.problem(w, err)
		return
	}
	if i != nil && i.Issued {
		p.Screen, p.Invoice = "shared", i
		if _, err := vietQR(bank, i.Remaining(), i.PaymentCode); err == nil {
			p.PaymentBank = bank
		}
	}
	a.page(w, p)
}
func (a *App) verify(w http.ResponseWriter, r *http.Request) {
	route := r.PathValue("id")
	if !a.allow("verify:"+ipOf(r), 30, time.Minute) {
		http.Error(w, "Chưa thể xác minh. Hãy thử lại sau.", 429)
		return
	}
	org, id, ok := a.store.shareTarget(route)
	if !ok {
		a.page(w, Page{Screen: "verify", InviteID: route, Error: "Mã không hợp lệ hoặc lời mời đã hết hạn. Hãy liên hệ người quản lý."})
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
		a.problem(w, err)
		return
	}
	if !verified {
		a.page(w, Page{Screen: "verify", InviteID: route, Error: "Mã không hợp lệ hoặc lời mời đã hết hạn. Hãy liên hệ người quản lý."})
		return
	}
	a.cookie(w, "zs_reader", a.newSession("share:"+route, org, 30*time.Minute), "/p/"+route, 1800)
	http.Redirect(w, r, "/p/"+route, 303)
}
