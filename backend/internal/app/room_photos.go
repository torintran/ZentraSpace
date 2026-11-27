package app

import (
	"bytes"
	"context"
	"crypto/subtle"
	"errors"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"
)

const roomPhotoBytes = 2 << 20

func (s *Store) photoRooms(ctx context.Context, org string) (map[string]bool, error) {
	rooms := make(map[string]bool)
	err := s.withOrganizationTx(ctx, org, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, "SELECT room_id FROM room_photos WHERE organization=$1", org)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var room string
			if err := rows.Scan(&room); err != nil {
				return err
			}
			rooms[room] = true
		}
		return rows.Err()
	})
	return rooms, err
}

func encodeRoomPhoto(data []byte) ([]byte, error) {
	cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || (format != "jpeg" && format != "png") || cfg.Width < 1 || cfg.Height < 1 ||
		int64(cfg.Width)*int64(cfg.Height) > 16000000 {
		return nil, errInvalid
	}
	source, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, errInvalid
	}
	width, height := cfg.Width, cfg.Height
	if longest := max(width, height); longest > 1280 {
		width, height = max(1, width*1280/longest), max(1, height*1280/longest)
	}
	resized := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := range height {
		for x := range width {
			resized.Set(x, y, source.At(x*cfg.Width/width, y*cfg.Height/height))
		}
	}
	var photo bytes.Buffer
	err = jpeg.Encode(&photo, resized, &jpeg.Options{Quality: 82})
	return photo.Bytes(), err
}

func (a *App) roomPhoto(w http.ResponseWriter, r *http.Request) {
	session, sessionID, ok := a.staff(w, r)
	if !ok {
		return
	}
	if session.role != "owner" {
		http.Error(w, "Bạn không có quyền xem ảnh phòng.", http.StatusForbidden)
		return
	}
	ctx, org, id := r.Context(), session.organization, r.PathValue("id")
	scope := "!room-photo:" + org + ":" + id
	if r.Method == http.MethodGet {
		var sealed []byte
		err := a.store.withOrganizationTx(ctx, org, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, "SELECT sealed FROM room_photos WHERE organization=$1 AND room_id=$2", org, id).Scan(&sealed)
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
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write(data)
		return
	}
	if subtle.ConstantTimeCompare([]byte(r.Form.Get("csrf")), []byte(hash(sessionID+":csrf"))) != 1 {
		http.Error(w, "Yêu cầu không hợp lệ.", http.StatusForbidden)
		return
	}
	if err := a.store.book(ctx, org, false, func(b *Book) error {
		if b.room(id) == nil {
			return errMissing
		}
		return nil
	}); err != nil {
		a.problem(w, r, err)
		return
	}
	file, _, err := r.FormFile("photo")
	if err != nil {
		a.problem(w, r, inputError{errors.New("Chọn ảnh phòng để tải lên.")})
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, roomPhotoBytes+1))
	if err != nil || len(data) > roomPhotoBytes {
		a.problem(w, r, inputError{errors.New("Ảnh phòng tối đa 2 MB.")})
		return
	}
	photo, err := encodeRoomPhoto(data)
	if err != nil {
		a.problem(w, r, inputError{errors.New("Dùng ảnh JPG hoặc PNG hợp lệ, tối đa 16 megapixel.")})
		return
	}
	sealed, err := encrypt(a.store.key, scope, photo)
	if err == nil {
		err = a.store.withOrganizationTx(ctx, org, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO room_photos (organization,room_id,sealed) VALUES ($1,$2,$3)
				ON CONFLICT (organization,room_id) DO UPDATE SET sealed=EXCLUDED.sealed`, org, id, sealed)
			return err
		})
	}
	if err != nil {
		a.problem(w, r, err)
		return
	}
	http.Redirect(w, r, "/rooms/"+id+"?saved=photo&period="+periodOf(r.Form.Get("period"))+"#room-profile", http.StatusSeeOther)
}
