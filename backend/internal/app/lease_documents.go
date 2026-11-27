package app

import (
	"context"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

const leaseDocumentBytes = 8 << 20
const leaseDocumentLimit = 12

type LeaseAttachment struct{ ID, Label string }

func leaseAttachmentScope(org, room, period, id string) string {
	scope := "!lease-document:" + org + ":" + room + ":" + period
	if id != "" {
		scope += ":" + id
	}
	return scope
}

func (s *Store) leaseDocumentFiles(ctx context.Context, org, room string, effective []string) (map[string][]LeaseAttachment, error) {
	periods := make(map[string][]LeaseAttachment)
	err := s.withOrganizationTx(ctx, org, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, "SELECT effective, document_id, sealed_label FROM lease_documents WHERE organization=$1 AND room_id=$2 AND effective=ANY($3) ORDER BY created_at, document_id", org, room, effective)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var period, id string
			var sealed []byte
			if err := rows.Scan(&period, &id, &sealed); err != nil {
				return err
			}
			label := "Hợp đồng chính"
			if id != "" {
				data, err := decrypt(s.key, leaseAttachmentScope(org, room, period, id)+":label", sealed)
				if err != nil {
					return err
				}
				label = string(data)
			}
			periods[period] = append(periods[period], LeaseAttachment{ID: id, Label: label})
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
	roomID, effective, documentID := r.PathValue("id"), r.PathValue("effective"), r.PathValue("document")
	if _, err := time.Parse("2006-01", effective); err != nil {
		http.NotFound(w, r)
		return
	}
	ctx, org := r.Context(), session.organization
	if r.Method == http.MethodGet {
		if _, err := hex.DecodeString(documentID); err != nil || (documentID != "" && len(documentID) != 48) {
			http.NotFound(w, r)
			return
		}
		var sealed []byte
		err := a.store.withOrganizationTx(ctx, org, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, "SELECT sealed FROM lease_documents WHERE organization=$1 AND room_id=$2 AND effective=$3 AND document_id=$4", org, roomID, effective, documentID).Scan(&sealed)
		})
		if errors.Is(err, pgx.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			a.problem(w, r, err)
			return
		}
		data, err := decrypt(a.store.key, leaseAttachmentScope(org, roomID, effective, documentID), sealed)
		if err != nil {
			a.problem(w, r, err)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		filename := "hop-dong-" + effective
		if documentID != "" {
			filename = "phu-luc-" + effective + "-" + documentID[:8]
		}
		w.Header().Set("Content-Disposition", "attachment; filename="+filename+"."+attachmentFormat(data))
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
	var sealedLabel []byte
	if strings.HasSuffix(r.URL.Path, "/attachments") {
		label := strings.TrimSpace(r.Form.Get("label"))
		if label == "" || utf8.RuneCountInString(label) > 120 || strings.ContainsFunc(label, unicode.IsControl) {
			a.problem(w, r, inputError{errors.New("Nhập tên phụ lục hợp lệ, tối đa 120 ký tự.")})
			return
		}
		documentID = hash(leaseAttachmentScope(org, roomID, effective, "") + "\x00" + label + "\x00" + hash(string(data)))[:48]
		sealedLabel, err = encrypt(a.store.key, leaseAttachmentScope(org, roomID, effective, documentID)+":label", []byte(label))
		if err != nil {
			a.problem(w, r, err)
			return
		}
	}
	sealed, err := encrypt(a.store.key, leaseAttachmentScope(org, roomID, effective, documentID), data)
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
		var count int
		var replay bool
		if err := tx.QueryRow(ctx, "SELECT count(*), COALESCE(bool_or(document_id=$4),false) FROM lease_documents WHERE organization=$1 AND room_id=$2 AND effective=$3", org, roomID, effective, documentID).Scan(&count, &replay); err != nil {
			return err
		}
		if replay && documentID != "" {
			return nil
		}
		if count >= leaseDocumentLimit {
			return inputError{errors.New("Mỗi hợp đồng lưu tối đa 12 tệp.")}
		}
		result, err := tx.Exec(ctx, "INSERT INTO lease_documents (organization,room_id,effective,document_id,sealed_label,sealed) VALUES ($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING", org, roomID, effective, documentID, sealedLabel, sealed)
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
