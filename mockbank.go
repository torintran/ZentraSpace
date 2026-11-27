package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	mockBankBIN     = "970407"
	mockBankAccount = "0000000000000"
	mockBankPath    = "/webhooks/mock-bank"
)

type mockTransfer struct {
	Reference string `json:"transaction_id"`
	Date      string `json:"date"`
	Account   string `json:"account"`
	BIN       string `json:"bin"`
	Direction string `json:"direction"`
	Amount    int64  `json:"amount_vnd"`
	Memo      string `json:"memo"`
}

func (e mockTransfer) valid() bool {
	_, err := time.Parse("2006-01-02", e.Date)
	return err == nil && referencePattern.MatchString(e.Reference) && e.Amount > 0 &&
		e.Amount <= 100000000000 && len(e.Memo) <= 300 && !strings.ContainsFunc(e.Memo, unicode.IsControl) &&
		binPattern.MatchString(e.BIN) && accountPattern.MatchString(e.Account)
}

func mockSignature(secret, stamp string, body []byte) []byte {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stamp + "."))
	mac.Write(body)
	return mac.Sum(nil)
}

func (a *App) mockWebhook(w http.ResponseWriter, r *http.Request) {
	if a.secure || len(a.mockBankSecret) < 32 {
		http.NotFound(w, r)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	stamp := r.Header.Get("X-ZS-Mock-Timestamp")
	seconds, timeErr := strconv.ParseInt(stamp, 10, 64)
	age := time.Since(time.Unix(seconds, 0))
	signature, signatureErr := hex.DecodeString(r.Header.Get("X-ZS-Mock-Signature"))
	if timeErr != nil || age < -5*time.Minute || age > 5*time.Minute || signatureErr != nil ||
		!hmac.Equal(signature, mockSignature(a.mockBankSecret, stamp, body)) {
		http.Error(w, "invalid webhook signature", http.StatusUnauthorized)
		return
	}
	var event mockTransfer
	if json.Unmarshal(body, &event) != nil || !event.valid() || event.Direction != "credit" {
		http.Error(w, "invalid incoming transaction", http.StatusUnprocessableEntity)
		return
	}
	if event.BIN != mockBankBIN || event.Account != mockBankAccount {
		http.Error(w, "unknown mock destination", http.StatusUnprocessableEntity)
		return
	}
	var added int
	var matched bool
	err = a.store.book(r.Context(), pilotOrg, true, func(b *Book) error {
		if b.Bank.BIN != event.BIN || b.Bank.Account != event.Account {
			return inputError{errors.New("mock destination is not configured")}
		}
		var statement bytes.Buffer
		statement.WriteString(statementHeader + "\n")
		writer := csv.NewWriter(&statement)
		if err := writer.WriteAll([][]string{{event.Reference, event.Date, strconv.FormatInt(event.Amount, 10), event.Memo}}); err != nil {
			return err
		}
		count, _, err := b.importStatement(statement.String())
		added = count
		if err != nil {
			return inputError{err}
		}
		if count == 1 && b.MockBankAutoMatch {
			transfer := &b.Transfers[len(b.Transfers)-1]
			invoice := b.suggestInvoice(transfer.Memo)
			if invoice != nil && invoice.Remaining() == transfer.Amount {
				ambiguous := false
				for _, entry := range invoice.Entries {
					if entry.Kind == "Thu" && entry.Method != cashPayment && entry.Amount == transfer.Amount &&
						!strings.HasPrefix(entry.Key, "bank-") && !invoice.Reversed(entry.Key) {
						ambiguous = true
						break
					}
				}
				if !ambiguous {
					if err := b.reconcile(transfer.ID, invoice.ID, ""); err != nil {
						return err
					}
					matched = true
				}
			}
		}
		return nil
	})
	if err != nil {
		var invalid inputError
		if errors.As(err, &invalid) {
			http.Error(w, invalid.Error(), http.StatusConflict)
		} else {
			http.Error(w, "transaction unavailable", http.StatusServiceUnavailable)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"created": added == 1, "matched": matched})
}

func runMockBank() error {
	secret, origin := os.Getenv("ZS_MOCK_BANK_SECRET"), os.Getenv("ZS_MOCK_BANK_ORIGIN")
	if len(secret) < 32 || origin == "" {
		return errors.New("Configure the mock bank secret and origin.")
	}
	client := &http.Client{Timeout: 5 * time.Second}
	mux := http.NewServeMux()
	frontend, err := newFrontendRenderer(os.Getenv("ZS_FRONTEND_URL"))
	if err != nil {
		return err
	}
	app := &App{frontend: frontend}
	mux.Handle("GET /_next/static/", frontend.assets())
	mux.Handle("GET /assets/", http.FileServer(http.FS(webAssets())))
	render := func(w http.ResponseWriter, r *http.Request, event mockTransfer, result string, status int) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Referrer-Policy", "strict-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		app.page(w, r, Page{Screen: "mock-bank", MockEvent: &event, Message: result, Status: status})
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, mockTransfer{Reference: "MOCK-" + token()[:12], Date: time.Now().In(vietnam).Format("2006-01-02"), Amount: 300000}, "", http.StatusOK)
	})
	mux.HandleFunc("POST /transfer", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Origin") != origin {
			http.Error(w, "invalid origin", http.StatusForbidden)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "invalid transfer", http.StatusBadRequest)
			return
		}
		amount, _ := strconv.ParseInt(r.Form.Get("amount"), 10, 64)
		event := mockTransfer{Reference: strings.TrimSpace(r.Form.Get("reference")), Date: r.Form.Get("date"),
			Account: mockBankAccount, BIN: mockBankBIN, Direction: "credit", Amount: amount, Memo: strings.TrimSpace(r.Form.Get("memo"))}
		if !event.valid() {
			render(w, r, event, "Kiểm tra mã, ngày, số tiền và nội dung giao dịch.", http.StatusUnprocessableEntity)
			return
		}
		body, _ := json.Marshal(event)
		stamp := strconv.FormatInt(time.Now().Unix(), 10)
		request, err := http.NewRequestWithContext(r.Context(), http.MethodPost, "http://web:8080"+mockBankPath, bytes.NewReader(body))
		if err != nil {
			render(w, r, event, "Không tạo được webhook.", http.StatusBadGateway)
			return
		}
		request.Header.Set("Content-Type", "application/json")
		request.Header.Set("X-ZS-Mock-Timestamp", stamp)
		request.Header.Set("X-ZS-Mock-Signature", hex.EncodeToString(mockSignature(secret, stamp, body)))
		response, err := client.Do(request)
		if err != nil {
			render(w, r, event, "ZentraSpace chưa phản hồi. Có thể gửi lại giao dịch này.", http.StatusBadGateway)
			return
		}
		defer response.Body.Close()
		var result struct{ Created, Matched bool }
		if response.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(response.Body, 1024)).Decode(&result) != nil {
			render(w, r, event, fmt.Sprintf("ZentraSpace từ chối webhook (HTTP %d).", response.StatusCode), http.StatusBadGateway)
		} else if result.Matched {
			render(w, r, event, "Đã tự đối soát. Mở bảng kê để kiểm tra số dư.", http.StatusOK)
		} else if result.Created {
			render(w, r, event, "Đã nhận giao dịch. Mở Đối soát ngân hàng để duyệt.", http.StatusOK)
		} else {
			render(w, r, event, "Giao dịch đã tồn tại. Không ghi trùng.", http.StatusOK)
		}
	})
	log.Print("Mock bank listening on :8080")
	return (&http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout: 10 * time.Second, WriteTimeout: 12 * time.Second, IdleTimeout: 30 * time.Second}).ListenAndServe()
}
