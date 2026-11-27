package app

import (
	"crypto/subtle"
	"errors"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"time"
)

func (a *App) action(w http.ResponseWriter, r *http.Request) {
	s, sessionID, ok := a.staff(w, r)
	if !ok {
		return
	}
	action := r.PathValue("action")
	if subtle.ConstantTimeCompare([]byte(r.Form.Get("csrf")), []byte(hash(sessionID+":csrf"))) != 1 || (s.role != "owner" && action != "read" && action != "logout") {
		http.Error(w, "Bạn không có quyền thực hiện thao tác này.", 403)
		return
	}
	if action == "logout" {
		a.mu.Lock()
		delete(a.sessions, hash(sessionID))
		a.mu.Unlock()
		a.cookie(w, "zs_staff", "", "/", -1)
		http.Redirect(w, r, "/login", 303)
		return
	}
	period := periodOf(r.Form.Get("period"))
	next := "/?tab=billing&saved=1&period=" + period
	p := Page{Screen: "invite", Role: s.role, CSRF: hash(sessionID + ":csrf")}
	if action == "policy-preview" {
		var preview PolicyPreview
		err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
			var err error
			preview, err = b.previewPolicy(r.Form.Get("scope"), r.Form.Get("policy"), r.Form.Get("effective"))
			if err != nil {
				return inputError{err}
			}
			return nil
		})
		if err != nil {
			a.problem(w, r, err)
			return
		}
		p.Screen, p.Tab, p.Period, p.PolicyPreview = "home", "policies", period, &preview
		a.page(w, r, p)
		return
	}
	if action == "import-preview" {
		var preview ImportPreview
		err := a.store.book(r.Context(), s.organization, false, func(b *Book) error {
			result, err := b.importRooms(period, r.Form.Get("csv"), false)
			preview = result
			if err != nil {
				return inputError{err}
			}
			return nil
		})
		if err != nil {
			a.problem(w, r, err)
			return
		}
		p.Screen, p.Import = "import", &preview
		a.page(w, r, p)
		return
	}
	err := a.store.book(r.Context(), s.organization, true, func(b *Book) error {
		value := func(key string) string { return strings.TrimSpace(r.Form.Get(key)) }
		id := value("id")
		switch action {
		case "site":
			capacity, err := number(r, "capacity", 1, 150)
			if err != nil {
				return err
			}
			if err := b.addSite(Site{Name: value("name"), Address: value("address"), Coordinates: value("coordinates"), Capacity: int(capacity)}); err != nil {
				return inputError{err}
			}
			next = "/?tab=sites&saved=1&period=" + period
		case "policy":
			rent, e1 := number(r, "rent", 0, 100000000)
			electric, e2 := number(r, "electricRate", 0, 100000)
			water, e3 := number(r, "waterRate", 0, 1000000)
			fee, e4 := number(r, "fee", 0, 100000000)
			if err := errors.Join(e1, e2, e3, e4); err != nil {
				return errInvalid
			}
			if err := b.addPolicy(RoomPolicy{Name: value("name"), Rent: rent, ElectricRate: electric, WaterRate: water, Fee: fee, FeeLabel: value("feeLabel")}); err != nil {
				return inputError{err}
			}
			next = "/?tab=policies&saved=1&period=" + period
		case "policy-apply":
			if err := b.applyPolicy(value("scope"), value("policy"), value("effective"), value("revision"), time.Now().In(vietnam).Format(time.DateOnly)); err != nil {
				return inputError{err}
			}
			next = "/?tab=policies&saved=policy-applied&period=" + period
		case "resident", "resident-out":
			room := b.room(id)
			if room == nil {
				return errMissing
			}
			var err error
			if action == "resident-out" {
				err = b.endResidency(id, value("resident"), value("leftOn"))
			} else {
				err = b.addResident(room, period, Resident{Name: value("name"), Phone: value("phone"), JoinedOn: value("joinedOn")})
			}
			if err != nil {
				return inputError{err}
			}
			next = "/rooms/" + id + "?saved=resident&period=" + period + "#room-residents"
		case "cash-entry":
			amount, err := number(r, "amount", 1, 100000000000)
			if err != nil {
				return err
			}
			entry := CashEntry{Key: value("key"), Kind: value("kind"), Date: value("date"),
				RoomID: value("room"), Category: value("category"), Reference: value("reference"),
				Note: value("note"), Amount: amount}
			if err := b.postCash(entry); err != nil {
				return inputError{err}
			}
			next = "/cash?saved=1&period=" + entry.Date[:7]
		case "room-fee":
			amount, err := number(r, "amount", 1, 100000000)
			kind := value("kind")
			if err != nil || (kind != "once" && kind != "monthly") {
				return errInvalid
			}
			if err := b.addRoomFee(RoomFee{ID: value("key"), RoomID: id, Label: value("label"),
				Start: value("start"), Monthly: kind == "monthly", Amount: amount}); err != nil {
				return inputError{err}
			}
			next = "/rooms/" + id + "?saved=fee&period=" + period + "#room-fees"
		case "room-fee-stop":
			if err := b.stopRoomFee(id, value("fee"), value("end")); err != nil {
				return inputError{err}
			}
			next = "/rooms/" + id + "?saved=fee&period=" + period + "#room-fees"
		case "billing-settings":
			fee, err := number(r, "commonFee", 0, 100000000)
			if err != nil {
				return err
			}
			if err := b.setBilling(period, BillingSettings{DueDate: value("dueDate"), CommonFeeLabel: value("commonFeeLabel"), CommonFee: fee}); err != nil {
				return inputError{err}
			}
		case "lease":
			terms := LeaseTerms{RoomID: id, Effective: value("effective"), Tenant: value("tenant"),
				Phone: value("phone"), SignedOn: value("signedOn"),
				Renewal: value("renewal") == "yes"}
			if value("useScheduledPrice") == "yes" {
				room := b.room(id)
				if room == nil {
					return errMissing
				}
				policy := b.policyAt(*room, terms.Effective)
				terms.Rent, terms.ElectricRate, terms.WaterRate = policy.Rent, policy.ElectricRate, policy.WaterRate
			} else {
				var e1, e2, e3 error
				terms.Rent, e1 = number(r, "rent", 0, 100000000)
				terms.ElectricRate, e2 = number(r, "electricRate", 0, 100000)
				terms.WaterRate, e3 = number(r, "waterRate", 0, 1000000)
				if e1 != nil || e2 != nil || e3 != nil {
					return errInvalid
				}
			}
			if value("months") != "" {
				months, err := number(r, "months", 1, 120)
				if err != nil {
					return err
				}
				terms.Months = int(months)
			}
			if err := b.scheduleLease(terms); err != nil {
				return inputError{err}
			}
			next = "/rooms/" + id + "?saved=lease&period=" + terms.Effective
		case "import":
			if _, err := b.importRooms(period, r.Form.Get("csv"), true); err != nil {
				return inputError{err}
			}
			next = "/?tab=rooms&saved=1&period=" + period
		case "room", "room-edit":
			next = "/?tab=rooms&saved=1&period=" + period
			room := Room{Name: value("name"), Tenant: value("tenant"), Start: period, SiteID: value("site"), Description: value("description")}
			if room.Name == "" || len(room.Name) > 60 || len(room.Tenant) > 150 || len(room.Description) > 1000 {
				return errInvalid
			}
			var lockOpening, lockRates bool
			policyMutable := true
			if action == "room-edit" {
				current := b.room(id)
				if current == nil {
					return errMissing
				}
				room.ID, room.Start = current.ID, current.Start
				room.Prepared = current.Prepared
				room.PolicyID, room.PolicyFee, room.PolicyFeeLabel = current.PolicyID, current.PolicyFee, current.PolicyFeeLabel
				if !r.Form.Has("site") {
					room.SiteID = current.SiteID
				}
				if !r.Form.Has("description") {
					room.Description = current.Description
				}
				if room.Prepared {
					room.Tenant = ""
				}
				room.LeaseMonths = current.LeaseMonths
				room.Rent, room.ElectricRate, room.WaterRate = current.Rent, current.ElectricRate, current.WaterRate
				room.Electric, room.Water = current.Electric, current.Water
				lockOpening, lockRates = b.roomLocks(id)
				policyMutable = current.Prepared && len(b.leaseHistory(*current)) == 0 && !lockRates
				if lockRates {
					room.Tenant = current.Tenant
				}
				if (lockRates && (r.Form.Has("rent") || r.Form.Has("electricRate") || r.Form.Has("waterRate"))) ||
					(lockOpening && (r.Form.Has("electric") || r.Form.Has("water"))) {
					return inputError{errors.New("Phòng đã có chỉ số hoặc bảng kê phát hành. Hãy mở lại phòng trước khi lưu.")}
				}
				next = "/rooms/" + id + "?saved=1&period=" + period
			} else if value("months") != "" {
				months, err := number(r, "months", 1, 120)
				if err != nil {
					return err
				}
				room.LeaseMonths = int(months)
			}
			if policyMutable && r.Form.Has("policy") {
				room.PolicyID, room.PolicyFee, room.PolicyFeeLabel = "", 0, ""
				if selected := value("policy"); selected != "" {
					index := slices.IndexFunc(b.Policies, func(policy RoomPolicy) bool { return policy.ID == selected })
					if index < 0 {
						return errMissing
					}
					policy := b.Policies[index]
					room.PolicyID, room.PolicyFee, room.PolicyFeeLabel = policy.ID, policy.Fee, policy.FeeLabel
					room.Rent, room.ElectricRate, room.WaterRate = policy.Rent, policy.ElectricRate, policy.WaterRate
				}
			} else if !policyMutable && r.Form.Has("policy") {
				return inputError{errors.New("Mẫu giá áp dụng khi chuẩn bị phòng. Với phòng đã có hợp đồng, hãy thêm điều khoản cho kỳ mới.")}
			}
			if !lockRates && (!policyMutable || !r.Form.Has("policy") || room.PolicyID == "") {
				var e1, e2, e3 error
				room.Rent, e1 = number(r, "rent", 0, 100000000)
				room.ElectricRate, e2 = number(r, "electricRate", 0, 100000)
				room.WaterRate, e3 = number(r, "waterRate", 0, 1000000)
				if e1 != nil || e2 != nil || e3 != nil {
					return errInvalid
				}
			}
			if !lockOpening {
				var e1, e2 error
				room.Electric, e1 = number(r, "electric", 0, 999999)
				room.Water, e2 = number(r, "water", 0, 999999)
				if e1 != nil || e2 != nil {
					return errInvalid
				}
			}
			if action == "room" && room.Tenant == "" {
				room.Prepared, room.LeaseMonths = true, 0
			}
			if action == "room" && room.SiteID != "" {
				next += "&site=" + url.QueryEscape(room.SiteID)
			}
			if err := b.saveRoom(room, value("revision")); err != nil {
				return inputError{err}
			}
		case "read", "read-special":
			next = "/?tab=meters&saved=1&period=" + period
			if value("filter") == "missing" {
				next += "&filter=missing"
			}
			e, e1 := number(r, "electric", 0, 999999)
			v, e2 := number(r, "water", 0, 999999)
			if e1 != nil || e2 != nil {
				return errInvalid
			}
			if action == "read-special" {
				adjustment, err := formMeterAdjustment(r)
				if err != nil {
					return err
				}
				if err := b.readWithAdjustment(id, period, e, v, adjustment); err != nil {
					return inputError{err}
				}
			} else if err := b.read(id, period, e, v); err != nil {
				return inputError{err}
			}
		case "draft":
			b.draft(period)
			if value("filter") == "draft" || value("filter") == "due" {
				next += "&filter=" + value("filter")
			}
		case "issue":
			i := b.invoice(id)
			if i == nil {
				return errMissing
			}
			if i.Revision() != value("revision") {
				return inputError{errors.New("Bảng kê đã thay đổi. Hãy mở lại bảng kê và kiểm tra trước khi phát hành.")}
			}
			if err := b.issue(id); err != nil {
				return inputError{err}
			}
			next = "/invoice/" + id
		case "post":
			amount, err := number(r, "amount", 1, 100000000000)
			if err != nil {
				return err
			}
			entry := Entry{Key: value("key"), Kind: value("kind"), Amount: amount, Note: value("note"),
				Method: value("method"), Reference: value("reference")}
			if err = b.post(id, entry); err != nil {
				return inputError{err}
			}
			next = "/invoice/" + id
		case "reverse":
			if err := b.reverse(id, value("key"), value("original"), value("note")); err != nil {
				return inputError{err}
			}
			next = "/invoice/" + id
		case "invite":
			i := b.invoice(id)
			if i == nil || !i.Issued {
				return errInvalid
			}
			for j := range b.Invites {
				if b.Invites[j].InvoiceID == id {
					b.Invites[j].Revoked = true
				}
			}
			if len(b.Invites) >= 5000 {
				return inputError{errors.New("Đã đạt giới hạn lời mời của bản thử.")}
			}
			p.Code = accessCode()
			v := Invite{ID: token(), InvoiceID: id, CodeHash: hash(p.Code), Expires: time.Now().Add(24 * time.Hour)}
			b.Invites = append(b.Invites, v)
			p.Invoice = i
			route, err := a.store.shareToken(s.organization, v.ID)
			if err != nil {
				return err
			}
			p.ShareURL = a.origin + "/p/" + route
		case "revoke":
			v := b.invite(id)
			if v == nil {
				return errMissing
			}
			v.Revoked = true
			next = "/invoice/" + v.InvoiceID
		default:
			return errMissing
		}
		return nil
	})
	if err != nil {
		a.problem(w, r, err)
		return
	}
	if p.Code != "" {
		a.page(w, r, p)
		return
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}
