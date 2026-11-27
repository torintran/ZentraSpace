package app

import (
	"sort"
	"time"
)

type ExpiringLease struct {
	RoomID, RoomName, Tenant, EndsOn, NextPeriod string
}

func (b *Book) expiringLeases(today time.Time, site string) []ExpiringLease {
	today = today.In(vietnam)
	start, horizon := today.Format(time.DateOnly), today.AddDate(0, 0, 60).Format(time.DateOnly)
	period := today.Format("2006-01")
	var results []ExpiringLease
	for _, room := range b.Rooms {
		if site != "" && site != room.SiteID {
			continue
		}
		lease := b.leaseAt(room, period)
		last, err := time.Parse("2006-01", lease.Until())
		if err != nil || lease.EndedOn != "" || lease.Tenant == "" || !lease.ActiveAt(period) {
			continue
		}
		next := last.AddDate(0, 1, 0)
		endsOn, nextPeriod := next.AddDate(0, 0, -1).Format(time.DateOnly), next.Format("2006-01")
		if endsOn < start || endsOn > horizon || b.leaseAt(room, nextPeriod).ActiveAt(nextPeriod) {
			continue
		}
		results = append(results, ExpiringLease{RoomID: room.ID, RoomName: b.roomLabel(room), Tenant: lease.Tenant, EndsOn: endsOn, NextPeriod: nextPeriod})
	}
	sort.SliceStable(results, func(i, j int) bool { return results[i].EndsOn < results[j].EndsOn })
	return results
}
