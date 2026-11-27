package app

import (
	"golang.org/x/crypto/bcrypt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
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
	frontend       *frontendRenderer
	mu             sync.Mutex
	sessions       map[string]session
	limits         map[string]limiter
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
		a.page(w, r, Page{Screen: "login", Status: http.StatusUnauthorized, Error: "Kiểm tra lại tài khoản hoặc mật khẩu."})
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
