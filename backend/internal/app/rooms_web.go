package app

import (
	"io"
	"net/http"
)

func (a *App) roomDetail(w http.ResponseWriter, r *http.Request) {
	s, sessionID, ok := a.staff(w, r)
	if !ok {
		return
	}
	if s.role != "owner" {
		http.Error(w, "Bạn không có quyền xem chi tiết phòng.", http.StatusForbidden)
		return
	}
	p := Page{Screen: "room", Role: s.role, CSRF: hash(sessionID + ":csrf"), Period: periodOf(r.URL.Query().Get("period"))}
	p.RoomForm = r.URL.Query().Get("edit") == "1"
	p.LeaseForm = !p.RoomForm && r.URL.Query().Get("contract") == "1"
	if r.URL.Query().Get("saved") == "1" {
		p.Message = "Đã lưu thông tin phòng."
	} else if r.URL.Query().Get("saved") == "lease" {
		p.Message = "Đã lưu điều khoản."
	} else if r.URL.Query().Get("saved") == "fee" {
		p.Message = "Đã lưu khoản phí."
	} else if r.URL.Query().Get("saved") == "document" {
		p.Message = "Đã lưu tệp hợp đồng."
	} else if r.URL.Query().Get("saved") == "resident" {
		p.Message = "Đã cập nhật người ở."
	} else if r.URL.Query().Get("saved") == "photo" {
		p.Message = "Đã lưu ảnh phòng."
	}
	if err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		p.RoomDetail = b.room(r.PathValue("id"))
		if p.RoomDetail == nil {
			return errMissing
		}
		p.LockOpening, p.LockRates = b.roomLocks(p.RoomDetail.ID)
		p.Lease = b.leaseAt(*p.RoomDetail, p.Period)
		p.LeaseHistory = b.leaseHistory(*p.RoomDetail)
		p.PolicyMutable = p.RoomDetail.Prepared && len(p.LeaseHistory) == 0 && !p.LockRates
		p.Residents = b.roomResidents(*p.RoomDetail, p.Period)
		p.RoomTimeline = b.roomTimeline(*p.RoomDetail)
		p.Sites, p.Policies = b.siteViews(p.Period), b.Policies
		for _, site := range p.Sites {
			if site.ID == p.RoomDetail.SiteID {
				p.RoomSite = &site
				break
			}
		}
		for _, fee := range b.RoomFees {
			if fee.RoomID == p.RoomDetail.ID {
				p.RoomFees = append(p.RoomFees, fee)
			}
		}
		for _, i := range b.Invoices {
			if i.RoomID == p.RoomDetail.ID && i.Period == p.Period {
				p.Invoice = &i
				break
			}
		}
		return nil
	}); err != nil {
		a.problem(w, r, err)
		return
	}
	if len(p.LeaseHistory) > 0 {
		documents, err := a.store.leaseDocumentPeriods(r.Context(), s.organization, p.RoomDetail.ID)
		if err != nil {
			a.problem(w, r, err)
			return
		}
		p.LeaseDocuments = documents
	}
	photos, err := a.store.photoRooms(r.Context(), s.organization)
	if err != nil {
		a.problem(w, r, err)
		return
	}
	p.HasPhoto = photos[p.RoomDetail.ID]
	a.page(w, r, p)
}

func (a *App) roomTemplate(w http.ResponseWriter, r *http.Request) {
	s, _, ok := a.staff(w, r)
	if !ok {
		return
	}
	if s.role != "owner" {
		http.Error(w, "Bạn không có quyền nhập danh sách phòng.", http.StatusForbidden)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=mau-nhap-phong.csv")
	io.WriteString(w, roomCSVHeader+"\r\n")
}
