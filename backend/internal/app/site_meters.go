package app

import (
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type SiteMeterReading struct {
	SiteID, Period, ElectricMeter, WaterMeter              string
	ElectricBefore, ElectricAfter, WaterBefore, WaterAfter int64
	ApprovalRevision, ApprovedBy, ApprovalNote, ApprovedOn string
}

type SiteMeterCheck struct {
	SiteMeterReading
	SiteName, Revision                                 string
	RoomElectric, RoomWater, ElectricDelta, WaterDelta int64
	Missing                                            int
	Recorded, NeedsReview, Approved                    bool
}

type SiteMeterApproval struct{ SiteID, Period, Revision, Reviewer, Note, RecordedOn string }

func (b *Book) saveSiteReading(reading SiteMeterReading, revision string) error {
	if !slices.ContainsFunc(b.Sites, func(site Site) bool { return site.ID == reading.SiteID }) {
		return errMissing
	}
	if _, err := time.Parse("2006-01", reading.Period); err != nil {
		return errInvalid
	}
	for _, meter := range []string{reading.ElectricMeter, reading.WaterMeter} {
		if meter == "" || utf8.RuneCountInString(meter) > 48 || strings.ContainsFunc(meter, unicode.IsControl) {
			return errInvalid
		}
	}
	if !validMeter(reading.ElectricBefore, reading.ElectricAfter, MeterSwap{}) || !validMeter(reading.WaterBefore, reading.WaterAfter, MeterSwap{}) {
		return errors.New("Chỉ số tổng phải từ 0 đến 999.999, số cuối từ số đầu trở lên.")
	}
	for i := range b.SiteMeterReadings {
		saved := &b.SiteMeterReadings[i]
		if saved.SiteID == reading.SiteID && saved.Period == reading.Period {
			current := *saved
			current.ApprovalRevision, current.ApprovedBy, current.ApprovalNote, current.ApprovedOn = "", "", "", ""
			if current == reading {
				return nil
			}
			if revision != b.checkSiteMeters(Site{ID: reading.SiteID}, reading.Period).Revision {
				return errors.New("Chỉ số đã thay đổi. Hãy mở lại dãy để kiểm tra trước khi lưu.")
			}
			*saved = reading
			return nil
		}
	}
	if revision != "" {
		return errInvalid
	}
	b.SiteMeterReadings = append(b.SiteMeterReadings, reading)
	return nil
}

func (b *Book) checkSiteMeters(site Site, period string) SiteMeterCheck {
	check := SiteMeterCheck{SiteMeterReading: SiteMeterReading{SiteID: site.ID, Period: period}, SiteName: site.Name}
	for _, reading := range b.SiteMeterReadings {
		if reading.SiteID == site.ID && reading.Period == period {
			check.SiteMeterReading, check.Recorded = reading, true
			break
		}
	}
	// Include reading identities and values so an approval cannot survive offsetting edits.
	var inputs []Reading
	for _, room := range b.Rooms {
		if room.SiteID != site.ID || !b.leaseAt(room, period).ActiveAt(period) {
			continue
		}
		reading := b.reading(room.ID, period)
		if reading == nil {
			check.Missing++
			continue
		}
		check.RoomElectric += reading.ElectricUsage()
		check.RoomWater += reading.WaterUsage()
		inputs = append(inputs, *reading)
	}
	if !check.Recorded {
		return check
	}
	check.ElectricDelta = check.ElectricAfter - check.ElectricBefore - check.RoomElectric
	check.WaterDelta = check.WaterAfter - check.WaterBefore - check.RoomWater
	reading := check.SiteMeterReading
	reading.ApprovalRevision, reading.ApprovedBy, reading.ApprovalNote, reading.ApprovedOn = "", "", "", ""
	data, _ := json.Marshal(struct {
		Main    SiteMeterReading
		Missing int
		Rooms   []Reading
	}{reading, check.Missing, inputs})
	check.Revision = hash(string(data))
	check.Approved = check.ApprovalRevision != "" && check.ApprovalRevision == check.Revision
	check.NeedsReview = !check.Approved && (check.Missing > 0 || check.ElectricDelta != 0 || check.WaterDelta != 0)
	return check
}

func (b *Book) approveSiteMeters(siteID, period, revision, reviewer, note string, now time.Time) error {
	if note == "" || utf8.RuneCountInString(note) > 240 || strings.ContainsFunc(note, unicode.IsControl) {
		return errInvalid
	}
	for _, site := range b.Sites {
		if site.ID != siteID {
			continue
		}
		check := b.checkSiteMeters(site, period)
		if !check.Recorded || check.Missing > 0 || revision != check.Revision {
			return errors.New("Chỉ duyệt sau khi đã nhập đủ chỉ số và xem lại chênh lệch hiện tại.")
		}
		if !check.NeedsReview {
			return nil
		}
		for i := range b.SiteMeterReadings {
			r := &b.SiteMeterReadings[i]
			if r.SiteID == siteID && r.Period == period {
				r.ApprovalRevision, r.ApprovedBy, r.ApprovalNote, r.ApprovedOn = revision, reviewer, note, now.In(vietnam).Format(time.DateOnly)
				b.SiteMeterApprovals = append(b.SiteMeterApprovals, SiteMeterApproval{SiteID: siteID, Period: period, Revision: revision, Reviewer: reviewer, Note: note, RecordedOn: now.UTC().Format(time.RFC3339)})
				return nil
			}
		}
	}
	return errMissing
}

func (b *Book) siteMeterHold(siteID, period string) bool {
	if siteID == "" || !slices.ContainsFunc(b.SiteMeterReadings, func(reading SiteMeterReading) bool {
		return reading.SiteID == siteID && reading.Period == period
	}) {
		return false
	}
	for _, site := range b.Sites {
		if site.ID == siteID {
			return b.checkSiteMeters(site, period).NeedsReview
		}
	}
	return false
}
