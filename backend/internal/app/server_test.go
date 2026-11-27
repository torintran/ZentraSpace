package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func TestBrowserFormOriginAndReferrerPolicy(t *testing.T) {
	const origin = "http://localhost:8087"
	password, err := bcrypt.GenerateFromPassword([]byte("synthetic-test-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	renderer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Path string
			View Page
		}
		if r.URL.Path != "/__render" || r.Method != http.MethodPost ||
			json.NewDecoder(r.Body).Decode(&body) != nil || body.Path == "" || body.View.Screen == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Security-Policy", "script-src 'self' 'nonce-test'")
		w.Write([]byte("<html></html>"))
	}))
	t.Cleanup(renderer.Close)
	frontend, err := newFrontendRenderer(renderer.URL)
	if err != nil {
		t.Fatal(err)
	}
	app := &App{
		origin:   origin,
		frontend: frontend,
		accounts: map[string]staffLogin{"owner": {pilotOrg, "owner", password}},
		sessions: map[string]session{},
		limits:   map[string]limiter{},
	}
	handler := app.routes()

	for _, path := range []string{"/login", "/p/test-invite-token"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, origin+path, nil))
		if got := response.Header().Get("Referrer-Policy"); got != "strict-origin" {
			t.Fatalf("%s uses %q; browser forms must retain Origin without leaking the path", path, got)
		}
	}

	for _, tc := range []struct {
		origin string
		status int
	}{
		{"null", http.StatusForbidden},
		{"https://different.example", http.StatusForbidden},
		{origin, http.StatusSeeOther},
	} {
		response := httptest.NewRecorder()
		form := httptest.NewRequest(http.MethodPost, origin+"/login", strings.NewReader("account=owner&password=synthetic-test-password"))
		form.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		form.Header.Set("Origin", tc.origin)
		handler.ServeHTTP(response, form)
		if response.Code != tc.status {
			t.Errorf("Origin %q: got %d, want %d", tc.origin, response.Code, tc.status)
		}
		if tc.origin == origin {
			cookies := response.Result().Cookies()
			if len(cookies) != 1 || !app.sessions[hash(cookies[0].Value)].expires.After(time.Now()) {
				t.Error("successful login did not create a session")
			}
		}
	}
}
