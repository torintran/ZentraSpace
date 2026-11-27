package app

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const pilotOrg = "local-pilot"

func Run() {
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
	a.frontend, err = newFrontendRenderer(os.Getenv("ZS_FRONTEND_URL"))
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Addr: ":8080", Handler: a.routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 12 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second, MaxHeaderBytes: 16384}
	log.Print("ZentraSpace listening on :8080")
	log.Fatal(server.ListenAndServe())
}
