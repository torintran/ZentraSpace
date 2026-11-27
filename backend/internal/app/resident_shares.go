package app

import (
	"errors"
	"sort"
	"strings"
	"time"
)

type ResidentShare struct {
	ResidentID string `json:",omitempty"`
	Name       string
	Days       int
	Amount     int64
}

func (i Invoice) residentCollected(share int) int64 {
	var total int64
	for _, e := range i.Entries {
		if e.ResidentShare == share {
			collected, _ := e.balanceEffects()
			total += collected
		}
	}
	return total
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
			shares = append(shares, ResidentShare{ResidentID: resident.ID, Name: resident.Name, Days: count})
		}
	}
	allocateResidentShares(i.Total, shares)
	i.ResidentPolicyVersion, i.ResidentShares = version, shares
}

type PersonalBill struct {
	Room, Period, DueDate, Name string
	Days                        int
	Amount, Collected, Reduced  int64
	Remaining, Credit           int64
}

func (i Invoice) personalBill(share int) *PersonalBill {
	if !i.Issued || share < 1 || share > len(i.ResidentShares) {
		return nil
	}
	s := i.ResidentShares[share-1]
	p := &PersonalBill{Room: i.RoomName, Period: i.Period, DueDate: i.DueDate, Name: s.Name, Days: s.Days, Amount: s.Amount}
	for _, entry := range i.Entries {
		if entry.ResidentShare == share {
			collected, reduced := entry.balanceEffects()
			p.Collected += collected
			p.Reduced += reduced
		}
	}
	balance := p.Amount - p.Collected - p.Reduced
	p.Remaining, p.Credit = max(0, balance), max(0, -balance)
	return p
}

func (b *Book) verificationAddress(i *Invoice, share int) string {
	if share == 0 || (share == 1 && len(i.ResidentShares) > 0) {
		return b.LeaseEmails[i.RoomID+"/"+i.LeaseEffective]
	}
	if share < 1 || share > len(i.ResidentShares) {
		return ""
	}
	id := i.ResidentShares[share-1].ResidentID
	for _, r := range b.Residents {
		if id != "" && r.ID == id && r.RoomID == i.RoomID && r.Effective == i.LeaseEffective {
			return r.Email
		}
	}
	return ""
}

func (b *Book) setResidentEmail(roomID, residentID, email string) error {
	if residentID == "" || (email != "" && !validEmail(email)) {
		return errInvalid
	}
	for j := range b.Residents {
		r := &b.Residents[j]
		if r.ID != residentID || r.RoomID != roomID {
			continue
		}
		if r.Email == email {
			return nil
		}
		r.Email = email
		for k := range b.Invites {
			v := &b.Invites[k]
			if v.Revoked || v.ResidentShare < 2 {
				continue
			}
			i := b.invoice(v.InvoiceID)
			if i != nil && i.RoomID == roomID && v.ResidentShare <= len(i.ResidentShares) &&
				i.ResidentShares[v.ResidentShare-1].ResidentID == residentID {
				v.Revoked = true
			}
		}
		return nil
	}
	return errMissing
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
