package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func (a *App) routes() http.Handler {
	m := http.NewServeMux()
	m.Handle("GET /assets/", http.FileServer(http.FS(webAssets())))
	m.Handle("GET /_next/static/", a.frontend.assets())
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if a.store.pool.Ping(ctx) != nil {
			http.Error(w, "unavailable", 503)
			return
		}
		w.Write([]byte("ok"))
	})
	m.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) { a.page(w, r, Page{Screen: "login"}) })
	m.HandleFunc("GET /privacy", func(w http.ResponseWriter, r *http.Request) { a.page(w, r, Page{Screen: "privacy"}) })
	m.HandleFunc("POST /login", a.login)
	m.HandleFunc("GET /{$}", a.home)
	m.HandleFunc("GET /invoice/{id}", a.invoice)
	m.HandleFunc("GET /export", a.export)
	m.HandleFunc("GET /cash", a.cash)
	m.HandleFunc("GET /cash.csv", a.cash)
	m.HandleFunc("GET /cash/receipt/{key}", a.cashReceipt)
	m.HandleFunc("POST /cash/receipt/{key}", a.cashReceipt)
	m.HandleFunc("GET /rooms/template.csv", a.roomTemplate)
	m.HandleFunc("GET /rooms/template.xlsx", a.roomTemplate)
	m.HandleFunc("GET /rooms/{id}", a.roomDetail)
	m.HandleFunc("GET /rooms/{id}/photo", a.roomPhoto)
	m.HandleFunc("POST /rooms/{id}/photo", a.roomPhoto)
	m.HandleFunc("GET /rooms/{id}/contracts/{effective}/file", a.leaseDocument)
	m.HandleFunc("POST /rooms/{id}/contracts/{effective}/file", a.leaseDocument)
	m.HandleFunc("POST /rooms/{id}/contracts/{effective}/attachments", a.leaseDocument)
	m.HandleFunc("GET /rooms/{id}/contracts/{effective}/attachments/{document}", a.leaseDocument)
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
			} else if strings.HasPrefix(r.URL.Path, "/rooms/") && strings.HasSuffix(r.URL.Path, "/photo") {
				r.Body = http.MaxBytesReader(w, r.Body, roomPhotoBytes+4096)
				err = r.ParseMultipartForm(roomPhotoBytes + 4096)
				if r.MultipartForm != nil {
					defer r.MultipartForm.RemoveAll()
				}
			} else if strings.HasPrefix(r.URL.Path, "/rooms/") && (strings.HasSuffix(r.URL.Path, "/file") || strings.HasSuffix(r.URL.Path, "/attachments")) {
				r.Body = http.MaxBytesReader(w, r.Body, leaseDocumentBytes+4096)
				err = r.ParseMultipartForm(leaseDocumentBytes + 4096)
				if r.MultipartForm != nil {
					defer r.MultipartForm.RemoveAll()
				}
			} else if r.URL.Path == "/action/import-file-preview" {
				r.Body = http.MaxBytesReader(w, r.Body, roomImportUploadLimit+4096)
				err = r.ParseMultipartForm(roomImportUploadLimit + 4096)
				if r.MultipartForm != nil {
					defer r.MultipartForm.RemoveAll()
				}
			} else if r.URL.Path == "/action/import" || r.URL.Path == "/action/import-preview" {
				r.Body = http.MaxBytesReader(w, r.Body, 3*roomImportTextLimit+4096)
				err = r.ParseForm()
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

type inputError struct{ error }

func (a *App) problem(w http.ResponseWriter, r *http.Request, err error) {
	var user inputError
	if errors.Is(err, errMeterAccess) {
		http.Error(w, errMeterAccess.Error(), http.StatusForbidden)
		return
	}
	if errors.Is(err, errMissing) {
		http.Error(w, errMissing.Error(), 404)
		return
	}
	if errors.As(err, &user) || errors.Is(err, errInvalid) || errors.Is(err, errLocked) {
		a.page(w, r, Page{Screen: "error", Status: http.StatusUnprocessableEntity, Error: err.Error()})
		return
	}
	log.Printf("request failed (%T)", err)
	http.Error(w, "Không thể xử lý lúc này.", 500)
}
