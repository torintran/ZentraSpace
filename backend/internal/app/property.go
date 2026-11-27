package app

import (
	"errors"
	"math"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"
)

type Site struct {
	ID, Name, Address, Coordinates string
	Capacity                       int
}

type SiteView struct {
	Site
	Rooms, Occupied int
	MapURL          string
}

func (b *Book) roomLabel(room Room) string {
	for _, site := range b.Sites {
		if site.ID == room.SiteID {
			return room.Name + " · " + site.Name
		}
	}
	return room.Name
}

func (s Site) mapURL() string {
	if parts := strings.Split(s.Coordinates, ","); len(parts) == 2 {
		lat, lng := parts[0], parts[1]
		return "https://www.openstreetmap.org/?mlat=" + url.QueryEscape(lat) + "&mlon=" + url.QueryEscape(lng) + "#map=18/" + lat + "/" + lng
	}
	if s.Address == "" {
		return ""
	}
	return "https://www.openstreetmap.org/search?query=" + url.QueryEscape(s.Address)
}

func (b *Book) addSite(site Site) (string, error) {
	if site.Name == "" || len(site.Name) > 120 || len(site.Address) > 360 || site.Capacity < 1 || site.Capacity > 150 || len(b.Sites) >= 100 {
		return "", errInvalid
	}
	coordinates, err := normalizeCoordinates(site.Coordinates)
	if err != nil {
		return "", err
	}
	if slices.ContainsFunc(b.Sites, func(existing Site) bool { return strings.EqualFold(existing.Name, site.Name) }) {
		return "", errors.New("Tên dãy trọ đã tồn tại.")
	}
	site.ID = token()
	site.Coordinates = coordinates
	b.Sites = append(b.Sites, site)
	return site.ID, nil
}

func normalizeCoordinates(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return "", errors.New("Nhập tọa độ theo dạng vĩ độ, kinh độ.")
	}
	lat, errLat := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	lng, errLng := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if errLat != nil || errLng != nil || math.IsNaN(lat) || math.IsNaN(lng) || math.Abs(lat) > 90 || math.Abs(lng) > 180 {
		return "", errors.New("Tọa độ chưa hợp lệ.")
	}
	return strconv.FormatFloat(lat, 'f', 6, 64) + "," + strconv.FormatFloat(lng, 'f', 6, 64), nil
}

func (b *Book) updateSiteLocation(id, address, value string) error {
	index := slices.IndexFunc(b.Sites, func(site Site) bool { return site.ID == id })
	if index < 0 {
		return errMissing
	}
	if len(address) > 360 {
		return errInvalid
	}
	coordinates, err := normalizeCoordinates(value)
	if err != nil {
		return err
	}
	b.Sites[index].Address, b.Sites[index].Coordinates = address, coordinates
	return nil
}

func (s Site) Revision() string {
	return hash(s.Name + "\x00" + strconv.Itoa(s.Capacity))
}

func (b *Book) updateSite(id, revision, name string, capacity int) error {
	index := slices.IndexFunc(b.Sites, func(site Site) bool { return site.ID == id })
	if index < 0 {
		return errMissing
	}
	site := &b.Sites[index]
	if site.Revision() != revision {
		return errors.New("Thông tin dãy đã thay đổi. Hãy mở lại trước khi lưu.")
	}
	if name == "" || len(name) > 120 || capacity < 1 || capacity > 150 {
		return errInvalid
	}
	if slices.ContainsFunc(b.Sites, func(other Site) bool { return other.ID != id && strings.EqualFold(other.Name, name) }) {
		return errors.New("Tên dãy trọ đã tồn tại.")
	}
	count := 0
	for _, room := range b.Rooms {
		if room.SiteID == id {
			count++
		}
	}
	if capacity < count {
		return errors.New("Số phòng tối đa không thể nhỏ hơn số phòng đã tạo.")
	}
	site.Name, site.Capacity = name, capacity
	return nil
}

func (b *Book) checkSiteCapacity(room Room) error {
	if room.SiteID == "" {
		return nil
	}
	index := slices.IndexFunc(b.Sites, func(site Site) bool { return site.ID == room.SiteID })
	if index < 0 {
		return errMissing
	}
	count := 0
	for _, existing := range b.Rooms {
		if existing.SiteID == room.SiteID && existing.ID != room.ID {
			count++
		}
	}
	if count >= b.Sites[index].Capacity {
		return errors.New("Dãy trọ đã đủ số phòng đăng ký.")
	}
	return nil
}

func (b *Book) siteViews(period string) []SiteView {
	views := make([]SiteView, 0, len(b.Sites))
	for _, site := range b.Sites {
		view := SiteView{Site: site, MapURL: site.mapURL()}
		for _, room := range b.Rooms {
			if room.SiteID == site.ID {
				view.Rooms++
				if b.leaseAt(room, period).OccupiedIn(period) {
					view.Occupied++
				}
			}
		}
		views = append(views, view)
	}
	return views
}

type RoomPolicy struct {
	ID, Name, FeeLabel                 string
	Rent, ElectricRate, WaterRate, Fee int64
}

func (b *Book) addPolicy(policy RoomPolicy) error {
	if policy.Name == "" || len(policy.Name) > 120 || len(b.Policies) >= 100 ||
		(policy.Fee > 0 && policy.FeeLabel == "") || len(policy.FeeLabel) > 120 {
		return errInvalid
	}
	if slices.ContainsFunc(b.Policies, func(existing RoomPolicy) bool { return strings.EqualFold(existing.Name, policy.Name) }) {
		return errors.New("Tên mẫu giá đã tồn tại.")
	}
	policy.ID = token()
	b.Policies = append(b.Policies, policy)
	return nil
}

type Resident struct {
	ID, RoomID, Effective, Name, Phone, JoinedOn, LeftOn string
}

func (b *Book) addResident(room *Room, ownerPeriod string, resident Resident) error {
	terms := b.leaseAt(*room, ownerPeriod)
	if !terms.OccupiedIn(ownerPeriod) || resident.Name == "" || len(resident.Name) > 150 || len(resident.Phone) > 30 {
		return errInvalid
	}
	joined, err := time.Parse(time.DateOnly, resident.JoinedOn)
	if err != nil || resident.JoinedOn[:7] != ownerPeriod {
		return errors.New("Ngày vào ở phải thuộc kỳ đang xem.")
	}
	if resident.JoinedOn < terms.startDate() {
		return errors.New("Ngày vào ở phải từ ngày bắt đầu hợp đồng trở đi.")
	}
	if len(b.roomResidents(*room, ownerPeriod)) >= 20 {
		return errors.New("Hồ sơ hỗ trợ tối đa 20 người đang ở mỗi phòng.")
	}
	for _, existing := range b.Residents {
		if existing.RoomID == room.ID && existing.Effective == terms.Effective &&
			strings.EqualFold(existing.Name, resident.Name) && existing.JoinedOn == joined.Format(time.DateOnly) {
			return errors.New("Người ở này đã có trong hồ sơ.")
		}
	}
	resident.ID, resident.RoomID, resident.Effective = token(), room.ID, terms.Effective
	b.Residents = append(b.Residents, resident)
	return nil
}

func (b *Book) roomResidents(room Room, period string) []Resident {
	lease := b.leaseAt(room, period)
	if !lease.OccupiedIn(period) {
		return nil
	}
	var residents []Resident
	if lease.Tenant != "" {
		residents = append(residents, Resident{Name: lease.Tenant, Phone: lease.Phone})
	}
	first := period + "-01"
	last := period + "-31"
	for _, resident := range b.Residents {
		if resident.RoomID == room.ID && resident.Effective == lease.Effective && resident.JoinedOn <= last &&
			(resident.LeftOn == "" || resident.LeftOn > first) {
			residents = append(residents, resident)
		}
	}
	return residents
}

func (b *Book) endResidency(roomID, residentID, end string) error {
	if _, err := time.Parse(time.DateOnly, end); err != nil {
		return errInvalid
	}
	for i := range b.Residents {
		resident := &b.Residents[i]
		if resident.RoomID == roomID && resident.ID == residentID {
			if resident.LeftOn != "" || end < resident.JoinedOn {
				return errInvalid
			}
			resident.LeftOn = end
			return nil
		}
	}
	return errMissing
}
