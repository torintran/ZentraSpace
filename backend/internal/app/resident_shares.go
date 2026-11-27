package app

import (
	"errors"
	"sort"
	"strings"
	"time"
)

type ResidentShare struct {
	Name   string
	Days   int
	Amount int64
}

func (b *Book) residentPolicy(roomID, lease, period string) (string, string) {
	prefix := roomID + "/" + lease + "/"
	var version, mode string
	for key, value := range b.ResidentPolicies {
		if from, ok := strings.CutPrefix(key, prefix); ok && from <= period && from > version {
			version, mode = from, value
		}
	}
	return version, mode
}

func (b *Book) setResidentPolicy(room Room, period, mode string) error {
	lease := b.leaseAt(room, period)
	if !lease.ActiveAt(period) || (mode != "days" && mode != "representative") {
		return errInvalid
	}
	key := room.ID + "/" + lease.Effective + "/" + period
	if prior, ok := b.ResidentPolicies[key]; ok {
		if prior == mode {
			return nil
		}
		return errors.New("Kỳ này đã có phiên bản chia phí. Chọn một kỳ khác.")
	}
	for _, invoice := range b.Invoices {
		if invoice.RoomID == room.ID && invoice.LeaseEffective == lease.Effective && invoice.Period >= period && invoice.Issued {
			return errors.New("Chọn kỳ sau bảng kê đã phát hành.")
		}
	}
	if b.ResidentPolicies == nil {
		b.ResidentPolicies = make(map[string]string)
	}
	b.ResidentPolicies[key] = mode
	b.refreshRoomFeeDrafts(room)
	return nil
}

func (b *Book) snapshotResidentShares(i *Invoice, lease LeaseTerms) {
	version, mode := b.residentPolicy(i.RoomID, lease.Effective, i.Period)
	if mode != "days" || lease.Tenant == "" {
		return
	}
	first, err := time.Parse("2006-01", i.Period)
	if err != nil {
		return
	}
	start := max(first.Format(time.DateOnly), lease.startDate())
	end := first.AddDate(0, 1, 0).Format(time.DateOnly)
	if last, err := time.Parse(time.DateOnly, lease.EndedOn); err == nil {
		end = min(end, last.AddDate(0, 0, 1).Format(time.DateOnly))
	}
	days := func(from, to string) int {
		a, e1 := time.Parse(time.DateOnly, from)
		z, e2 := time.Parse(time.DateOnly, to)
		if e1 != nil || e2 != nil || !z.After(a) {
			return 0
		}
		return int(z.Sub(a) / (24 * time.Hour))
	}
	representativeDays := days(start, end)
	if representativeDays == 0 {
		return
	}
	shares := []ResidentShare{{Name: lease.Tenant, Days: representativeDays}}
	for _, resident := range b.Residents {
		if resident.RoomID != i.RoomID || resident.Effective != lease.Effective {
			continue
		}
		left := end
		if resident.LeftOn != "" {
			left = min(left, resident.LeftOn)
		}
		if count := days(max(start, resident.JoinedOn), left); count > 0 {
			shares = append(shares, ResidentShare{Name: resident.Name, Days: count})
		}
	}
	allocateResidentShares(i.Total, shares)
	i.ResidentPolicyVersion, i.ResidentShares = version, shares
}

func allocateResidentShares(total int64, shares []ResidentShare) {
	var weight, assigned int64
	for _, s := range shares {
		weight += int64(s.Days)
	}
	if weight == 0 {
		return
	}
	type residual struct {
		index  int
		weight int64
	}
	order := make([]residual, 0, len(shares))
	for j := range shares {
		day := int64(shares[j].Days)
		part := (total % weight) * day
		shares[j].Amount = (total/weight)*day + part/weight
		assigned += shares[j].Amount
		rem := part % weight
		order = append(order, residual{j, max(rem, -rem)})
	}
	sort.SliceStable(order, func(i, j int) bool { return order[i].weight > order[j].weight })
	delta := total - assigned
	sign := int64(1)
	if delta < 0 {
		sign = -1
	}
	for j := 0; j < int(max(delta, -delta)); j++ {
		shares[order[j].index].Amount += sign
	}
}
