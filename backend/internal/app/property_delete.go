package app

import (
	"errors"
	"net/http"
	"net/url"
	"slices"

	"github.com/jackc/pgx/v5"
)

func (b *Book) deleteSite(id string) error {
	if !slices.ContainsFunc(b.Sites, func(site Site) bool { return site.ID == id }) {
		return errMissing
	}
	if slices.ContainsFunc(b.Rooms, func(room Room) bool { return room.SiteID == id }) {
		return errors.New("Dãy trọ vẫn còn phòng. Hãy chuyển phòng sang dãy khác hoặc xóa các phòng đủ điều kiện trước.")
	}
	if slices.ContainsFunc(b.SiteMeterReadings, func(reading SiteMeterReading) bool { return reading.SiteID == id }) {
		return errors.New("Dãy trọ đã có lịch sử công tơ tổng.")
	}
	b.Sites = slices.DeleteFunc(b.Sites, func(site Site) bool { return site.ID == id })
	return nil
}

func (b *Book) roomDeleteReason(room Room) string {
	id := room.ID
	switch {
	case room.Tenant != "" || slices.ContainsFunc(b.Leases, func(item LeaseTerms) bool { return item.RoomID == id }):
		return "Phòng đã có người thuê hoặc hợp đồng."
	case slices.ContainsFunc(b.Readings, func(item Reading) bool { return item.RoomID == id }):
		return "Phòng đã có chỉ số điện nước."
	case slices.ContainsFunc(b.Invoices, func(item Invoice) bool { return item.RoomID == id }):
		return "Phòng đã có bảng kê."
	case slices.ContainsFunc(b.CashEntries, func(item CashEntry) bool { return item.RoomID == id }):
		return "Phòng đã có giao dịch thu chi hoặc cọc."
	case slices.ContainsFunc(b.Residents, func(item Resident) bool { return item.RoomID == id }):
		return "Phòng đã có lịch sử người ở."
	default:
		return ""
	}
}

func (b *Book) deleteUnusedRoom(id, revision string) (string, error) {
	room := b.room(id)
	if room == nil {
		return "", errMissing
	}
	if reason := b.roomDeleteReason(*room); reason != "" {
		return "", inputError{errors.New(reason)}
	}
	if room.Revision() != revision {
		return "", inputError{errors.New("Thông tin phòng đã thay đổi. Hãy mở lại phòng trước khi xóa.")}
	}
	siteID := room.SiteID
	b.Rooms = slices.DeleteFunc(b.Rooms, func(item Room) bool { return item.ID == id })
	b.RoomFees = slices.DeleteFunc(b.RoomFees, func(item RoomFee) bool { return item.RoomID == id })
	b.PolicyAssignments = slices.DeleteFunc(b.PolicyAssignments, func(item PolicyAssignment) bool { return item.RoomID == id })
	return siteID, nil
}

func (a *App) deleteRoom(w http.ResponseWriter, r *http.Request, org, period string) {
	if r.Form.Get("confirm") != "yes" {
		a.problem(w, r, inputError{errors.New("Hãy xác nhận xóa phòng.")})
		return
	}
	id := r.Form.Get("id")
	var siteID string
	err := a.store.bookTransaction(r.Context(), org, true, func(b *Book, tx pgx.Tx) error {
		var err error
		siteID, err = b.deleteUnusedRoom(id, r.Form.Get("revision"))
		if err != nil {
			return err
		}
		var hasDocuments bool
		if err := tx.QueryRow(r.Context(), "SELECT EXISTS (SELECT 1 FROM lease_documents WHERE organization=$1 AND room_id=$2)", org, id).Scan(&hasDocuments); err != nil {
			return err
		}
		if hasDocuments {
			return inputError{errors.New("Phòng đã có tệp hợp đồng.")}
		}
		_, err = tx.Exec(r.Context(), "DELETE FROM room_photos WHERE organization=$1 AND room_id=$2", org, id)
		return err
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	query := url.Values{"tab": {"rooms"}, "period": {period}, "saved": {"room-deleted"}}
	if siteID != "" {
		query.Set("site", siteID)
	}
	http.Redirect(w, r, "/?"+query.Encode(), http.StatusSeeOther)
}
