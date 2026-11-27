package app

import (
	"context"
	"crypto/subtle"
	"errors"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
)

const leaseDocumentBytes = 8 << 20

func (s *Store) leaseDocumentPeriods(ctx context.Context, org, room string) (map[string]bool, error) {
	periods := make(map[string]bool)
	err := s.withOrganizationTx(ctx, org, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, "SELECT effective FROM lease_documents WHERE organization=$1 AND room_id=$2", org, room)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var period string
			if err := rows.Scan(&period); err != nil {
				return err
			}
			periods[period] = true
		}
		return rows.Err()
	})
	return periods, err
}

func (a *App) leaseDocument(w http.ResponseWriter, r *http.Request) {
	session, sessionID, ok := a.staff(w, r)
	if !ok {
		return
	}
	if session.role != "owner" {
		http.Error(w, "Bạn không có quyền truy cập hợp đồng.", http.StatusForbidden)
		return
	}
	roomID, effective := r.PathValue("id"), r.PathValue("effective")
	if _, err := time.Parse("2006-01", effective); err != nil {
		http.NotFound(w, r)
		return
	}
	ctx, org := r.Context(), session.organization
	scope := "!lease-document:" + org + ":" + roomID + ":" + effective
	if r.Method == http.MethodGet {
		var sealed []byte
		err := a.store.withOrganizationTx(ctx, org, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, "SELECT sealed FROM lease_documents WHERE organization=$1 AND room_id=$2 AND effective=$3", org, roomID, effective).Scan(&sealed)
		})
		if errors.Is(err, pgx.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			a.problem(w, r, err)
			return
		}
		data, err := decrypt(a.store.key, scope, sealed)
		if err != nil {
			a.problem(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=hop-dong-"+effective+"."+attachmentFormat(data))
		w.Write(data)
		return
	}

	if subtle.ConstantTimeCompare([]byte(r.Form.Get("csrf")), []byte(hash(sessionID+":csrf"))) != 1 {
		http.Error(w, "Yêu cầu không hợp lệ.", http.StatusForbidden)
		return
	}
	file, _, err := r.FormFile("document")
	if err != nil {
		a.problem(w, r, inputError{errors.New("Chọn tệp hợp đồng.")})
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, leaseDocumentBytes+1))
	if err != nil || len(data) > leaseDocumentBytes || attachmentFormat(data) == "" {
		a.problem(w, r, inputError{errors.New("Hợp đồng phải là PDF, JPG hoặc PNG, tối đa 8 MB.")})
		return
	}
	sealed, err := encrypt(a.store.key, scope, data)
	if err != nil {
		a.problem(w, r, err)
		return
	}
	if err := a.store.bookTransaction(ctx, org, true, func(b *Book, tx pgx.Tx) error {
		room := b.room(roomID)
		if room == nil || !slices.ContainsFunc(b.leaseHistory(*room), func(terms LeaseTerms) bool {
			return terms.Effective == effective
		}) {
			return errMissing
		}
		result, err := tx.Exec(ctx, "INSERT INTO lease_documents (organization,room_id,effective,sealed) VALUES ($1,$2,$3,$4) ON CONFLICT DO NOTHING", org, roomID, effective, sealed)
		if err == nil && result.RowsAffected() != 1 {
			return inputError{errors.New("Hợp đồng đã có tệp đính kèm.")}
		}
		return err
	}); err != nil {
		a.problem(w, r, err)
		return
	}
	next := "/rooms/" + roomID + "?period=" + effective + "&saved=document"
	if r.FormValue("archive") == "1" {
		next += "&history=1&kind=contracts&archivePeriod=" + effective
	}
	http.Redirect(w, r, next+"#room-contracts", http.StatusSeeOther)
}
