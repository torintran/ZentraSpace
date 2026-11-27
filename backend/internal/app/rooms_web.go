package app

import (
	"io"
	"net/http"
	"slices"
	"strconv"
	"time"
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
	p.CheckoutForm = !p.RoomForm && !p.LeaseForm && r.URL.Query().Get("checkout") == "1"
	p.RoomArchive = !p.RoomForm && !p.LeaseForm && !p.CheckoutForm && r.URL.Query().Get("history") == "1"
	if p.RoomArchive {
		p.ArchiveKind = r.URL.Query().Get("kind")
		switch p.ArchiveKind {
		case "contracts", "lease", "bill", "payment", "deposit", "resident", "policy", "checkout":
		default:
			p.ArchiveKind = ""
		}
		if period := r.URL.Query().Get("archivePeriod"); period != "" && periodOf(period) == period {
			p.ArchivePeriod = period
		}
		p.ArchivePage, _ = strconv.Atoi(r.URL.Query().Get("page"))
	}
	now := time.Now()
	p.Today = now.In(vietnam).Format(time.DateOnly)
	month, _ := time.Parse("2006-01", p.Period)
	p.PeriodDate = month.Format(time.DateOnly)
	p.PeriodEnd = month.AddDate(0, 1, -1).Format(time.DateOnly)
	if p.Today >= p.PeriodDate && p.Today <= p.PeriodEnd {
		p.PeriodDate = p.Today
	}
	previewDate := r.URL.Query().Get("date")
	p.CheckoutDate = p.Today
	if previewDate != "" {
		p.CheckoutDate = previewDate
	}
	if p.LeaseForm {
		if _, err := time.Parse("2006-01", r.URL.Query().Get("renew")); err == nil {
			p.RenewPeriod = r.URL.Query().Get("renew")
		}
	}
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
	} else if r.URL.Query().Get("saved") == "checkout" {
		p.Message = "Đã trả phòng và lưu tất toán."
	} else if r.URL.Query().Get("saved") == "lease-email" {
		p.Message = "Đã lưu email xác minh."
	}
	if err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
		p.RoomDetail = b.room(r.PathValue("id"))
		if p.RoomDetail == nil {
			return errMissing
		}
		p.LockOpening, p.LockRates = b.roomLocks(p.RoomDetail.ID)
		p.RoomDeleteReason = b.roomDeleteReason(*p.RoomDetail)
		p.Lease = b.leaseAt(*p.RoomDetail, p.Period)
		if p.CheckoutForm && previewDate != "" {
			quote, err := b.quoteCheckout(*p.RoomDetail, previewDate, now)
			if err != nil {
				p.CheckoutError = err.Error()
			} else {
				p.CheckoutQuote = &quote
			}
		}
		p.RoomHasScheduledPolicy = slices.ContainsFunc(b.PolicyAssignments, func(assignment PolicyAssignment) bool {
			return assignment.RoomID == p.RoomDetail.ID
		})
		p.RoomPolicyFee = p.Lease.recurringFee(*p.RoomDetail)
		if !p.Lease.OccupiedIn(p.Period) {
			policy := b.policyAt(*p.RoomDetail, p.Period)
			p.RoomPolicyFee = FeeLine{Label: policy.FeeLabel, Amount: policy.Fee}
			if p.Lease.EndedOn != "" && p.Lease.EndedOn[:7] == p.Period {
				p.NextLeasePeriod = month.AddDate(0, 1, 0).Format("2006-01")
			}
			p.Lease = LeaseTerms{Rent: policy.Rent, ElectricRate: policy.ElectricRate, WaterRate: policy.WaterRate}
		}
		history := b.leaseHistory(*p.RoomDetail)
		p.LeaseCount = len(history)
		if p.RoomArchive && p.ArchiveKind == "contracts" {
			slices.Reverse(history)
			for _, lease := range history {
				if p.ArchivePeriod == "" || lease.Effective == p.ArchivePeriod {
					p.LeaseHistory = append(p.LeaseHistory, lease)
				}
			}
			p.LeaseHistory = paginateArchive(p.LeaseHistory, &p.ArchivePage, &p.ArchivePages)
		} else if !p.RoomArchive && p.Lease.Effective != "" {
			p.LeaseHistory = []LeaseTerms{p.Lease}
		}
		p.LeaseEmails = make(map[string]string)
		for _, lease := range p.LeaseHistory {
			p.LeaseEmails[lease.Effective] = b.LeaseEmails[p.RoomDetail.ID+"/"+lease.Effective]
		}
		p.PolicyMutable = p.RoomDetail.Prepared && p.LeaseCount == 0 && !p.LockRates
		if p.RoomArchive {
			if p.ArchiveKind != "contracts" {
				p.RoomTimeline = paginateArchive(b.roomTimeline(*p.RoomDetail, p.ArchivePeriod, p.ArchiveKind), &p.ArchivePage, &p.ArchivePages)
			}
			return nil
		}
		p.Residents = b.roomResidents(*p.RoomDetail, p.Period)
		if !p.RoomForm && !p.LeaseForm && !p.CheckoutForm {
			events := b.roomTimeline(*p.RoomDetail, p.Period, "")
			p.RoomTimeline = events[:min(len(events), 8)]
		}
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
		effective := make([]string, 0, len(p.LeaseHistory))
		for _, lease := range p.LeaseHistory {
			effective = append(effective, lease.Effective)
		}
		documents, err := a.store.leaseDocumentFiles(r.Context(), s.organization, p.RoomDetail.ID, effective)
		if err != nil {
			a.problem(w, r, err)
			return
		}
		p.LeaseDocuments = documents
		if len(documents) > 0 && p.RoomDeleteReason == "" {
			p.RoomDeleteReason = "Phòng đã có tệp hợp đồng."
		}
	}
	if p.RoomArchive {
		a.page(w, r, p)
		return
	}
	photos, err := a.store.photoRooms(r.Context(), s.organization)
	if err != nil {
		a.problem(w, r, err)
		return
	}
	p.HasPhoto = photos[p.RoomDetail.ID]
	a.page(w, r, p)
}

func paginateArchive[T any](entries []T, page, pages *int) []T {
	const size = 12
	*pages = max(1, (len(entries)+size-1)/size)
	*page = max(1, min(*page, *pages))
	start := (*page - 1) * size
	return entries[start:min(len(entries), start+size)]
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
