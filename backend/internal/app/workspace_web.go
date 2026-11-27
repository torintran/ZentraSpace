package app

import (
	"net/http"
)

func (a *App) home(w http.ResponseWriter, r *http.Request) {
	s, id, ok := a.staff(w, r)
	if !ok {
		return
	}
	tab := r.URL.Query().Get("tab")
	if tab != "rooms" && tab != "meters" && tab != "billing" && tab != "dashboard" {
		tab = "dashboard"
	}
	if s.role == "meter" {
		tab = "meters"
	}
	period := periodOf(r.URL.Query().Get("period"))
	filter := r.URL.Query().Get("filter")
	if (tab != "meters" || filter != "missing") && (tab != "billing" || (filter != "draft" && filter != "due")) {
		filter = ""
	}
	p := Page{Screen: "home", Role: s.role, Tab: tab, Period: period, Filter: filter, CSRF: hash(id + ":csrf")}
	p.RoomForm = tab == "rooms" && r.URL.Query().Get("add") == "1"
	if r.URL.Query().Get("saved") == "1" {
		p.Message = "Đã lưu."
	}
	if err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		if tab == "dashboard" {
			p.Dashboard = b.dashboard(period)
			return nil
		}
		if tab == "billing" {
			p.BillingSettings = b.Billing[period]
		}
		for _, room := range b.Rooms {
			lease := b.leaseAt(room, period)
			var renewed bool
			reading := b.reading(room.ID, period)
			done := reading != nil
			if ((tab == "meters" || tab == "billing") && !lease.ActiveAt(period)) || (filter == "missing" && done) {
				continue
			}
			if s.role == "meter" {
				room = Room{ID: room.ID, Name: room.Name, Start: room.Start, Electric: room.Electric, Water: room.Water}
			} else if tab == "rooms" && room.Start <= period {
				if lease.ActiveAt(period) {
					room.Tenant = lease.Tenant
					renewed = lease.Renewal
				} else {
					room.Tenant = ""
				}
			}
			view := RoomView{Room: room, Done: done, HasLease: lease.ActiveAt(period), Renewed: renewed}
			if tab == "rooms" {
				view.Lease = lease
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
					(filter != "due" || (i.Issued && i.Remaining() > 0)) {
					p.Invoices = append(p.Invoices, i)
				}
			}
		}
		return nil
	}); err != nil {
		a.problem(w, r, err)
		return
	}
	a.page(w, r, p)
}
