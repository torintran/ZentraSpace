package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
	key  []byte
}

func aead(master []byte, org string) (cipher.AEAD, error) {
	kdf := hmac.New(sha256.New, master)
	kdf.Write([]byte("zentraspace/book/v1:" + org))
	block, err := aes.NewCipher(kdf.Sum(nil))
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
func encrypt(master []byte, org string, plaintext []byte) ([]byte, error) {
	gcm, err := aead(master, org)
	if err != nil {
		return nil, err
	}
	nonce := randBytes(gcm.NonceSize())
	return gcm.Seal(nonce, nonce, plaintext, []byte(org)), nil
}
func decrypt(master []byte, org string, ciphertext []byte) ([]byte, error) {
	gcm, err := aead(master, org)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("invalid ciphertext")
	}
	return gcm.Open(nil, ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():], []byte(org))
}

func dbURL(user, password string) string {
	return (&url.URL{Scheme: "postgres", User: url.UserPassword(user, password), Host: os.Getenv("DB_HOST") + ":5432", Path: "/zentraspace", RawQuery: "sslmode=disable"}).String()
}

func migrate(ctx context.Context) error {
	runtime := os.Getenv("DB_RUNTIME_PASSWORD")
	adminPassword := os.Getenv("DB_ADMIN_PASSWORD")
	if runtime == "" || adminPassword == "" || os.Getenv("DB_HOST") == "" {
		return errors.New("database configuration required")
	}
	admin, err := pgx.Connect(ctx, dbURL("postgres", adminPassword))
	if err != nil {
		return errors.New("database bootstrap connection failed")
	}
	defer admin.Close(ctx)
	var exists bool
	if err = admin.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='zentra_runtime')").Scan(&exists); err != nil {
		return err
	}
	if !exists {
		if _, err = admin.Exec(ctx, "CREATE ROLE zentra_runtime LOGIN NOSUPERUSER NOBYPASSRLS"); err != nil {
			return err
		}
	}
	runtimeBytes, decodeErr := hex.DecodeString(runtime)
	if decodeErr != nil || len(runtimeBytes) != 32 {
		return errors.New("invalid database runtime password")
	}
	if _, err = admin.Exec(ctx, fmt.Sprintf("ALTER ROLE zentra_runtime PASSWORD '%s'", hex.EncodeToString(runtimeBytes))); err != nil {
		return errors.New("database role setup failed")
	}
	_, err = admin.Exec(ctx, `
 CREATE TABLE IF NOT EXISTS books (
  organization text PRIMARY KEY,
  sealed bytea NOT NULL
 );
 ALTER TABLE books ENABLE ROW LEVEL SECURITY;
 ALTER TABLE books FORCE ROW LEVEL SECURITY;
 DROP POLICY IF EXISTS organization_scope ON books;
 CREATE POLICY organization_scope ON books TO zentra_runtime
 USING (organization = current_setting('app.organization', true))
 WITH CHECK (organization = current_setting('app.organization', true));
 REVOKE ALL ON books FROM PUBLIC;
 GRANT SELECT, INSERT, UPDATE ON books TO zentra_runtime;`)
	return err
}

func openStore(ctx context.Context) (*Store, error) {
	master, err := hex.DecodeString(os.Getenv("ZS_MASTER_KEY"))
	if err != nil || len(master) != 32 {
		return nil, errors.New("ZS_MASTER_KEY must be 32 bytes in hex")
	}
	runtime := os.Getenv("DB_RUNTIME_PASSWORD")
	if runtime == "" || os.Getenv("DB_HOST") == "" {
		return nil, errors.New("runtime database configuration required")
	}
	pool, err := pgxpool.New(ctx, dbURL("zentra_runtime", runtime))
	if err != nil {
		return nil, errors.New("runtime database configuration failed")
	}
	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.New("runtime database connection failed")
	}
	return &Store{pool: pool, key: master}, nil
}

func (s *Store) book(ctx context.Context, org string, write bool, fn func(*Book) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT set_config('app.organization', $1, true)", org); err != nil {
		return err
	}
	query := "SELECT sealed FROM books WHERE organization=$1"
	if write {
		empty, err := encrypt(s.key, org, []byte("{}"))
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "INSERT INTO books VALUES ($1,$2) ON CONFLICT DO NOTHING", org, empty); err != nil {
			return err
		}
		query += " FOR UPDATE"
	}
	var encrypted []byte
	var book Book
	err = tx.QueryRow(ctx, query, org).Scan(&encrypted)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	if err == nil {
		raw, err := decrypt(s.key, org, encrypted)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(raw, &book); err != nil {
			return err
		}
	}
	if err = fn(&book); err != nil {
		return err
	}
	if write {
		raw, err := json.Marshal(book)
		if err != nil {
			return err
		}
		sealed, err := encrypt(s.key, org, raw)
		if err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, "UPDATE books SET sealed=$2 WHERE organization=$1", org, sealed); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}
