package app

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

type MeterStaff struct {
	Username, Revision string
	SiteIDs            []string
	PasswordHash       []byte
	Disabled           bool
}

type MeterStaffView struct {
	Username, Revision string
	SiteIDs            []string
	Disabled           bool
}

type StaffChange struct {
	Username, Actor, Operation, At string
}

func (b *Book) meterStaff(username string) *MeterStaff {
	for i := range b.MeterStaff {
		if b.MeterStaff[i].Username == username {
			return &b.MeterStaff[i]
		}
	}
	return nil
}

func (s session) managedMeterStaff(b *Book) (*MeterStaff, bool) {
	if s.staffRevision == "" {
		return nil, true
	}
	account := b.meterStaff(s.username)
	return account, account != nil && !account.Disabled && account.Revision == s.staffRevision
}

var errMeterAccess = errors.New("Bạn không có quyền nhập chỉ số cho phòng này.")

func (b *Book) saveMeterStaff(staff MeterStaff, expected, actor string, now time.Time) error {
	if !accountIdentifier.MatchString(staff.Username) || len(staff.SiteIDs) == 0 || len(staff.SiteIDs) > len(b.Sites) {
		return inputError{errors.New("Nhập tên đăng nhập và chọn ít nhất một dãy thuộc tổ chức.")}
	}
	for _, id := range staff.SiteIDs {
		if !slices.ContainsFunc(b.Sites, func(site Site) bool { return site.ID == id }) {
			return inputError{errors.New("Dãy được chọn không còn trong tổ chức.")}
		}
	}
	slices.Sort(staff.SiteIDs)
	staff.SiteIDs = slices.Compact(staff.SiteIDs)
	prior := b.meterStaff(staff.Username)
	operation := "Tạo tài khoản"
	if prior == nil {
		if expected != "" || len(staff.PasswordHash) == 0 || len(b.MeterStaff) >= 100 {
			return inputError{errors.New("Cần mật khẩu mới để tạo tài khoản. Mỗi tổ chức hỗ trợ tối đa 100 tài khoản công tơ.")}
		}
	} else {
		if expected == "" || prior.Revision != expected {
			return inputError{errors.New("Tài khoản đã tồn tại hoặc đã thay đổi. Tải lại trang trước khi lưu.")}
		}
		operation = "Cập nhật quyền"
		if len(staff.PasswordHash) == 0 {
			if prior.Disabled == staff.Disabled && slices.Equal(prior.SiteIDs, staff.SiteIDs) {
				return nil
			}
			staff.PasswordHash = prior.PasswordHash
		} else {
			operation = "Đổi mật khẩu và cập nhật quyền"
		}
		if prior.Disabled != staff.Disabled {
			if staff.Disabled {
				operation = "Tắt tài khoản"
			} else {
				operation = "Bật tài khoản"
			}
		}
	}
	staff.Revision = token()
	if prior == nil {
		b.MeterStaff = append(b.MeterStaff, staff)
	} else {
		*prior = staff
	}
	b.StaffChanges = append(b.StaffChanges, StaffChange{Username: staff.Username, Actor: actor, Operation: operation, At: now.UTC().Format(time.RFC3339)})
	return nil
}

func (a *App) saveManagedStaff(w http.ResponseWriter, r *http.Request, s session) {
	username := strings.ToLower(strings.TrimSpace(r.Form.Get("username")))
	if !accountIdentifier.MatchString(username) || !a.allow("staff-save:"+s.organization, 12, time.Minute) {
		a.problem(w, r, inputError{errors.New("Kiểm tra tên đăng nhập hoặc thử lại sau một phút.")})
		return
	}
	staff := MeterStaff{Username: username, SiteIDs: slices.Clone(r.Form["site"]), Disabled: r.Form.Get("enabled") != "yes"}
	if password := r.Form.Get("password"); password != "" {
		if utf8.RuneCountInString(password) < 16 || len(password) > 72 {
			a.problem(w, r, inputError{errors.New("Mật khẩu cần ít nhất 16 ký tự và tối đa 72 byte.")})
			return
		}
		var err error
		staff.PasswordHash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			a.problem(w, r, err)
			return
		}
	}
	if err := a.store.book(r.Context(), s.organization, true, func(b *Book) error {
		return b.saveMeterStaff(staff, r.Form.Get("revision"), s.username, time.Now())
	}); err != nil {
		a.problem(w, r, err)
		return
	}
	http.Redirect(w, r, "/?tab=settings&saved=meter-staff#meter-staff", http.StatusSeeOther)
}
