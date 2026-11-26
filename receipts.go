package main

import (
	"context"
	"crypto/subtle"
	"errors"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"
)

const receiptBytes = 2 << 20

func receiptFormat(data []byte) string {
	switch http.DetectContentType(data) {
	case "application/pdf":
		return "pdf"
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpg"
	}
	return ""
}

func (s *Store) withReceiptTx(ctx context.Context, org string, run func(pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, "SELECT set_config('app.organization', $1, true)", org); err != nil {
		return err
	}
	if err := run(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) receiptKeys(ctx context.Context, org string) (map[string]bool, error) {
	keys := make(map[string]bool)
	err := s.withReceiptTx(ctx, org, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, "SELECT entry_key FROM cash_receipts WHERE organization=$1", org)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var key string
			if err := rows.Scan(&key); err != nil {
				return err
			}
			keys[key] = true
		}
		return rows.Err()
	})
	return keys, err
}

func (a *App) cashReceipt(w http.ResponseWriter, r *http.Request) {
	s, id, ok := a.staff(w, r)
	if !ok {
		return
	}
	if s.role != "owner" {
		http.Error(w, "Bạn không có quyền xem chứng từ.", http.StatusForbidden)
		return
	}
	key := r.PathValue("key")
	ctx := r.Context()
	org := s.organization
	scope := "!cash-receipt:" + org + ":" + key
	if r.Method == http.MethodGet {
		var sealed []byte
		err := a.store.withReceiptTx(ctx, org, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, "SELECT sealed FROM cash_receipts WHERE organization=$1 AND entry_key=$2", org, key).Scan(&sealed)
		})
		if errors.Is(err, pgx.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			a.problem(w, err)
			return
		}
		data, err := decrypt(a.store.key, scope, sealed)
		if err != nil {
			a.problem(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=chung-tu."+receiptFormat(data))
		w.Write(data)
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.Form.Get("csrf")), []byte(hash(id+":csrf"))) != 1 {
		http.Error(w, "Yêu cầu không hợp lệ.", http.StatusForbidden)
		return
	}
	file, _, err := r.FormFile("receipt")
	if err != nil {
		a.problem(w, inputError{errors.New("Chọn tệp chứng từ.")})
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, receiptBytes+1))
	if err != nil || len(data) > receiptBytes || receiptFormat(data) == "" {
		a.problem(w, inputError{errors.New("Chứng từ phải là PDF, JPG hoặc PNG, tối đa 2 MB.")})
		return
	}
	var period string
	if err := a.store.book(ctx, org, false, func(b *Book) error {
		for _, entry := range b.CashEntries {
			if entry.Key == key {
				period = entry.Date[:7]
				return nil
			}
		}
		return errMissing
	}); err != nil {
		a.problem(w, err)
		return
	}
	sealed, err := encrypt(a.store.key, scope, data)
	if err != nil {
		a.problem(w, err)
		return
	}
	if err := a.store.withReceiptTx(ctx, org, func(tx pgx.Tx) error {
		result, err := tx.Exec(ctx, "INSERT INTO cash_receipts VALUES ($1,$2,$3) ON CONFLICT DO NOTHING", org, key, sealed)
		if err == nil && result.RowsAffected() != 1 {
			return inputError{errors.New("Giao dịch đã có chứng từ.")}
		}
		return err
	}); err != nil {
		a.problem(w, err)
		return
	}
	http.Redirect(w, r, "/cash?saved=receipt&period="+period, http.StatusSeeOther)
}
