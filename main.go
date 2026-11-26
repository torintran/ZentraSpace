package main

import (
	"context"
	"crypto/subtle"
	"encoding/csv"
	"errors"
	"html/template"
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
	role    string
	expires time.Time
}
type limiter struct {
	count int
	until time.Time
}
type App struct {
	store     *Store
	origin    string
	secure    bool
	passwords map[string][]byte
	pages     *template.Template
	mu        sync.Mutex
	sessions  map[string]session
	limits    map[string]limiter
}

const pilotOrg = "local-pilot"

type RoomView struct {
	Room
	Done bool
}
type Page struct {
	Screen, Role, Tab, Period, Error, Message, CSRF, InviteID, ShareURL, Code string
	Rooms                                                                     []RoomView
	Invoices                                                                  []Invoice
	Invoice                                                                   *Invoice
	Invites                                                                   []Invite
	Import                                                                    *ImportPreview
	Status                                                                    int
	Bank                                                                      *BankScreen
	BankReview                                                                *TransferReview
	PaymentBank                                                               BankAccount
}

func main() {
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
	a := &App{store: store, origin: origin, secure: !local, sessions: map[string]session{}, limits: map[string]limiter{}, passwords: map[string][]byte{}}
	for role, key := range map[string]string{"owner": "ZS_OWNER_PASSWORD", "meter": "ZS_METER_PASSWORD"} {
		secret := os.Getenv(key)
		if len(secret) < 16 || len(secret) > 72 {
			log.Fatal("Configure local account secrets between 16 and 72 bytes.")
		}
		a.passwords[role], err = bcrypt.GenerateFromPassword([]byte(secret), bcrypt.DefaultCost)
		if err != nil {
			log.Fatal(err)
		}
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
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return s + " ₫"
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
func (a *App) newSession(role string, d time.Duration) string {
	a.mu.Lock()
	defer a.mu.Unlock()
	for k, v := range a.sessions {
		if time.Now().After(v.expires) {
			delete(a.sessions, k)
		}
	}
	id := token()
	a.sessions[hash(id)] = session{role, time.Now().Add(d)}
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
	m.HandleFunc("POST /login", a.login)
	m.HandleFunc("GET /{$}", a.home)
	m.HandleFunc("GET /invoice/{id}", a.invoice)
	m.HandleFunc("GET /export", a.export)
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
		if r.Method == "POST" {
			if r.Header.Get("Origin") != a.origin {
				http.Error(w, "Yêu cầu không hợp lệ.", http.StatusForbidden)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, 16384)
			if err := r.ParseForm(); err != nil {
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
	role := strings.TrimSpace(r.Form.Get("account"))
	pw, known := a.passwords[role]
	if !known {
		pw = a.passwords["owner"]
	}
	if bcrypt.CompareHashAndPassword(pw, []byte(r.Form.Get("password"))) != nil || !known {
		a.page(w, Page{Screen: "login", Status: http.StatusUnauthorized, Error: "Kiểm tra lại tài khoản hoặc mật khẩu."})
		return
	}
	id := a.newSession(role, 8*time.Hour)
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
	if tab != "rooms" && tab != "meters" && tab != "billing" {
		tab = "billing"
	}
	if s.role == "meter" {
		tab = "meters"
	}
	period := periodOf(r.URL.Query().Get("period"))
	p := Page{Screen: "home", Role: s.role, Tab: tab, Period: period, CSRF: hash(id + ":csrf")}
	if r.URL.Query().Get("saved") == "1" {
		p.Message = "Đã lưu."
	}
	if err := a.store.book(r.Context(), pilotOrg, false, func(b *Book) error {
		for _, room := range b.Rooms {
			done := b.reading(room.ID, period) != nil
			if s.role == "meter" {
				room = Room{ID: room.ID, Name: room.Name, Start: room.Start, Electric: room.Electric, Water: room.Water}
			}
			if reading := b.reading(room.ID, period); reading != nil {
				room.Electric = reading.ElectricAfter
				room.Water = reading.WaterAfter
			} else if prev, err := time.Parse("2006-01", period); err == nil {
				previous := prev.AddDate(0, -1, 0).Format("2006-01")
				for _, inv := range b.Invoices {
					if inv.RoomID == room.ID && inv.Issued && inv.Period == previous {
						room.Electric = inv.Reading.ElectricAfter
						room.Water = inv.Reading.WaterAfter
						break
					}
				}
			}
			p.Rooms = append(p.Rooms, RoomView{room, done})
		}
		if s.role == "owner" {
			for _, i := range b.Invoices {
				if i.Period == period {
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
	err := a.store.book(r.Context(), pilotOrg, false, func(b *Book) error {
		p.Invoice = b.invoice(r.PathValue("id"))
		if p.Invoice == nil {
			return errMissing
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
	var rows = [][]string{{"Phòng", "Kỳ", "Tổng bảng kê (VND)", "Đã thu ròng (VND)", "Còn thiếu (VND)", "Thu dư (VND)"}}
	err := a.store.book(r.Context(), pilotOrg, false, func(b *Book) error {
		for _, i := range b.Invoices {
			if i.Issued && i.Period == period {
				c, _ := i.totals()
				rows = append(rows, []string{plainCSV(i.RoomName), i.Period, strconv.FormatInt(i.Total, 10), strconv.FormatInt(c, 10), strconv.FormatInt(i.Remaining(), 10), strconv.FormatInt(i.Credit(), 10)})
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
		err := a.store.book(r.Context(), pilotOrg, false, func(b *Book) error {
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
	err := a.store.book(r.Context(), pilotOrg, true, func(b *Book) error {
		value := func(key string) string { return strings.TrimSpace(r.Form.Get(key)) }
		id := value("id")
		switch action {
		case "import":
			if _, err := b.importRooms(period, r.Form.Get("csv"), true); err != nil {
				return inputError{err}
			}
			next = "/?tab=rooms&saved=1&period=" + period
		case "room":
			next = "/?tab=rooms&saved=1&period=" + period
			name, tenant := value("name"), value("tenant")
			if name == "" || len(name) > 60 || len(tenant) > 150 || len(b.Rooms) >= 150 {
				return errInvalid
			}
			for _, r := range b.Rooms {
				if strings.EqualFold(r.Name, name) {
					return inputError{errors.New("Mã phòng đã tồn tại.")}
				}
			}
			rent, e1 := number(r, "rent", 0, 100000000)
			eRate, e2 := number(r, "electricRate", 0, 100000)
			wRate, e3 := number(r, "waterRate", 0, 1000000)
			electric, e4 := number(r, "electric", 0, 999999)
			water, e5 := number(r, "water", 0, 999999)
			if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil {
				return errInvalid
			}
			b.Rooms = append(b.Rooms, Room{token(), name, tenant, period, rent, eRate, wRate, electric, water})
		case "read":
			next = "/?tab=meters&saved=1&period=" + period
			e, e1 := number(r, "electric", 0, 999999)
			v, e2 := number(r, "water", 0, 999999)
			if e1 != nil || e2 != nil {
				return errInvalid
			}
			if err := b.read(id, period, e, v); err != nil {
				return inputError{err}
			}
		case "draft":
			b.draft(period)
		case "issue":
			if err := b.issue(id); err != nil {
				return inputError{err}
			}
			next = "/invoice/" + id
		case "post":
			amount, err := number(r, "amount", 1, 100000000000)
			if err != nil {
				return err
			}
			if err = b.post(id, value("key"), value("kind"), amount, value("note")); err != nil {
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
			p.ShareURL = a.origin + "/p/" + v.ID
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
	id := r.PathValue("id")
	if len(id) != 48 || !a.allow("verify:"+ipOf(r), 30, time.Minute) {
		http.Error(w, "Chưa thể xác minh. Hãy thử lại sau.", 429)
		return
	}
	verified := false
	err := a.store.book(r.Context(), pilotOrg, true, func(b *Book) error {
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
		a.page(w, Page{Screen: "verify", InviteID: id, Error: "Mã không hợp lệ hoặc lời mời đã hết hạn. Hãy liên hệ người quản lý."})
		return
	}
	a.cookie(w, "zs_reader", a.newSession("share:"+id, 30*time.Minute), "/p/"+id, 1800)
	http.Redirect(w, r, "/p/"+id, 303)
}
