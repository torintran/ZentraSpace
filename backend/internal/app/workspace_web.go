package app

import (
	"net/http"
	"time"
)

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	s, id, ok := a.staff(w, r)
	if !ok {
		return
	}
	tab := r.URL.Query().Get("tab")
	if tab != "rooms" && tab != "meters" && tab != "billing" && tab != "dashboard" && tab != "sites" && tab != "policies" && tab != "settings" {
		tab = "dashboard"
	}
	if s.role == "meter" {
		tab = "meters"
	}
	period := periodOf(r.URL.Query().Get("period"))
	filter := r.URL.Query().Get("filter")
	if (tab != "meters" || filter != "missing") && (tab != "billing" || (filter != "draft" && filter != "due" && filter != "credit")) {
		filter = ""
	}
	p := Page{Screen: "home", Role: s.role, Tab: tab, Period: period, Filter: filter, CSRF: hash(id + ":csrf")}
	if tab == "meters" && s.role == "owner" {
		p.SiteMeterFocus = r.URL.Query().Get("meterSite")
	}
	if tab == "settings" {
		settings, err := a.smtpView(r.Context(), s.platformAdmin)
		if err != nil {
			a.problem(w, r, err)
			return
		}
		p.SMTP = settings
		switch r.URL.Query().Get("saved") {
		case "smtp":
			p.Message = "Đã lưu SMTP cho toàn hệ thống."
		case "smtp-tested":
			p.Message = "Kết nối và xác thực SMTP thành công."
		case "smtp-disabled":
			p.Message = "Đã tắt gửi mã qua SMTP."
		}
		a.page(w, r, p)
		return
	}
	p.RoomForm = tab == "rooms" && r.URL.Query().Get("add") == "1"
	if tab == "rooms" || tab == "sites" {
		p.SelectedSite = r.URL.Query().Get("site")
	}
	p.EditSiteLocation = tab == "sites" && r.URL.Query().Get("location") == "edit"
	p.ShowSiteDetails = tab == "sites" && r.URL.Query().Get("view") == "details"
	p.EditSiteDetails = p.ShowSiteDetails && r.URL.Query().Get("edit") == "1"
	if r.URL.Query().Get("saved") == "1" {
		p.Message = "Đã lưu."
	} else if tab == "policies" && r.URL.Query().Get("saved") == "policy-applied" {
		p.Message = "Đã lên lịch áp dụng mẫu giá và phí."
	} else if tab == "sites" && r.URL.Query().Get("saved") == "location" {
		p.Message = "Đã cập nhật vị trí dãy trọ."
	} else if tab == "sites" && r.URL.Query().Get("saved") == "site-deleted" {
		p.Message = "Đã xóa dãy trọ."
	} else if tab == "rooms" && r.URL.Query().Get("saved") == "room-deleted" {
		p.Message = "Đã xóa phòng."
	}
	if err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		if tab == "dashboard" {
			p.Dashboard = b.dashboard(period)
			return nil
		}
		if tab == "billing" {
			p.BillingSettings = b.Billing[period]
		}
		if s.role == "owner" && (tab == "meters" || tab == "billing") {
			for _, site := range b.Sites {
				check := b.checkSiteMeters(site, period)
				if tab == "meters" || check.NeedsReview {
					p.SiteMeterChecks = append(p.SiteMeterChecks, check)
				}
			}
		}
		if tab == "rooms" || tab == "sites" || tab == "policies" {
			p.Sites = b.siteViews(period)
		}
		if tab == "rooms" || tab == "policies" {
			p.Policies = b.Policies
		}
		if tab == "rooms" && !p.RoomForm {
			p.ExpiringLeases = b.expiringLeases(time.Now(), p.SelectedSite)
		}
		if tab == "sites" || tab == "policies" {
			for _, site := range p.Sites {
				if site.ID == p.SelectedSite && p.EditSiteDetails {
					p.SiteRevision = site.Revision()
					break
				}
			}
			return nil
		}
		for _, room := range b.Rooms {
			if tab == "rooms" && p.SelectedSite != "" && room.SiteID != p.SelectedSite {
				continue
			}
			lease := b.leaseAt(room, period)
			var renewed bool
			reading := b.reading(room.ID, period)
			done := reading != nil
			if ((tab == "meters" || tab == "billing") && !lease.ActiveAt(period)) || (filter == "missing" && done) {
				continue
			}
			if tab == "meters" || tab == "billing" {
				room.Name = b.roomLabel(room)
			}
			if s.role == "meter" {
				room = Room{ID: room.ID, Name: room.Name, Start: room.Start, Electric: room.Electric, Water: room.Water}
			} else if tab == "rooms" {
				if lease.OccupiedIn(period) {
					room.Tenant = lease.Tenant
					renewed = lease.Renewal
				} else {
					room.Tenant = ""
				}
			}
			view := RoomView{Room: room, Done: done, HasLease: lease.ActiveAt(period), Renewed: renewed}
			if tab == "rooms" {
				view.HasLease = lease.OccupiedIn(period)
				view.Lease = lease
				view.People = len(b.roomResidents(room, period))
			}
			if tab == "meters" {
				if reading != nil {
					view.Electric, view.Water = reading.ElectricAfter, reading.WaterAfter
					view.Handover = reading.Adjustment.Handover
					if s.role == "owner" {
						view.Adjustment = reading.Adjustment
					}
				} else if opening, err := b.meterOpening(room, period); err == nil {
					view.Electric, view.Water, view.Handover = opening.Electric, opening.Water, opening.Handover
				}
			}
			p.Rooms = append(p.Rooms, view)
		}
		if s.role == "owner" {
			for _, i := range b.Invoices {
				if i.Period == period && (filter != "draft" || !i.Issued) &&
					(filter != "due" || (i.Issued && i.Remaining() > 0)) &&
					(filter != "credit" || (i.Issued && i.transferableCredit() > 0)) {
					p.Invoices = append(p.Invoices, i)
				}
			}
		}
		return nil
	}); err != nil {
		a.problem(w, r, err)
		return
	}
	if tab == "rooms" && !p.RoomForm && len(p.Rooms) > 0 {
		photos, err := a.store.photoRooms(r.Context(), s.organization)
		if err != nil {
			a.problem(w, r, err)
			return
		}
		for i := range p.Rooms {
			p.Rooms[i].HasPhoto = photos[p.Rooms[i].ID]
		}
	}
	a.page(w, r, p)
}
