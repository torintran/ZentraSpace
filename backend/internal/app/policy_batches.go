package app

import (
	"encoding/json"
	"errors"
	"slices"
	"time"
)

const unassignedPolicyScope = "unassigned"

type PolicyAssignment struct {
	RoomID, Scope, Effective, AppliedOn, Batch string
	Policy                                     RoomPolicy
}

type PolicyPreviewRoom struct {
	ID, Name, Blocked string
	Before            RoomPolicy
}

type PolicyPreview struct {
	Scope, ScopeName, Effective, Revision string
	Policy                                RoomPolicy
	Rooms                                 []PolicyPreviewRoom
	Eligible                              int
}

func (b *Book) policyAt(room Room, period string) RoomPolicy {
	var effective string
	policy := RoomPolicy{
		ID: room.PolicyID, Rent: room.Rent, ElectricRate: room.ElectricRate,
		WaterRate: room.WaterRate, FeeLabel: room.PolicyFeeLabel, Fee: room.PolicyFee,
	}
	for _, assignment := range b.PolicyAssignments {
		if assignment.RoomID == room.ID && assignment.Effective <= period && assignment.Effective > effective {
			policy, effective = assignment.Policy, assignment.Effective
		}
	}
	return policy
}

func (b *Book) previewPolicy(scope, policyID, effective string) (PolicyPreview, error) {
	if _, err := time.Parse("2006-01", effective); err != nil {
		return PolicyPreview{}, errors.New("Chọn kỳ hiệu lực hợp lệ.")
	}
	preview := PolicyPreview{Scope: scope, Effective: effective}
	if scope == unassignedPolicyScope {
		preview.ScopeName = "Phòng chưa gán dãy"
	} else {
		i := slices.IndexFunc(b.Sites, func(site Site) bool { return site.ID == scope })
		if i < 0 {
			return PolicyPreview{}, errMissing
		}
		preview.ScopeName = b.Sites[i].Name
	}
	i := slices.IndexFunc(b.Policies, func(policy RoomPolicy) bool { return policy.ID == policyID })
	if i < 0 {
		return PolicyPreview{}, errMissing
	}
	preview.Policy = b.Policies[i]
	for _, room := range b.Rooms {
		if room.SiteID != scope && !(scope == unassignedPolicyScope && room.SiteID == "") {
			continue
		}
		row := PolicyPreviewRoom{ID: room.ID, Name: room.Name, Before: b.policyAt(room, effective)}
		switch {
		case effective < room.Start:
			row.Blocked = "Kỳ hiệu lực trước kỳ bắt đầu của phòng"
		case !room.Prepared || len(b.leaseHistory(room)) > 0:
			row.Blocked = "Phòng đã có hồ sơ hợp đồng"
		case slices.ContainsFunc(b.Invoices, func(invoice Invoice) bool { return invoice.RoomID == room.ID && invoice.Period >= effective }):
			row.Blocked = "Phòng đã có bảng kê trong phạm vi này"
		case slices.ContainsFunc(b.PolicyAssignments, func(assignment PolicyAssignment) bool {
			return assignment.RoomID == room.ID && assignment.Effective == effective
		}):
			row.Blocked = "Phòng đã có mẫu giá cho kỳ này"
		default:
			preview.Eligible++
		}
		preview.Rooms = append(preview.Rooms, row)
	}
	raw, err := json.Marshal(preview)
	if err != nil {
		return PolicyPreview{}, err
	}
	preview.Revision = hash(string(raw))
	return preview, nil
}

func (b *Book) applyPolicy(scope, policyID, effective, revision, appliedOn string) error {
	if revision != "" && slices.ContainsFunc(b.PolicyAssignments, func(assignment PolicyAssignment) bool {
		return assignment.Batch == revision && assignment.Scope == scope &&
			assignment.Policy.ID == policyID && assignment.Effective == effective
	}) {
		return nil
	}
	preview, err := b.previewPolicy(scope, policyID, effective)
	if err != nil {
		return err
	}
	if revision != preview.Revision {
		return errors.New("Danh sách hoặc giá phòng đã thay đổi. Hãy xem trước lại.")
	}
	if preview.Eligible == 0 {
		return errors.New("Không có phòng đủ điều kiện áp dụng.")
	}
	for _, row := range preview.Rooms {
		if row.Blocked == "" {
			b.PolicyAssignments = append(b.PolicyAssignments, PolicyAssignment{
				RoomID: row.ID, Scope: scope, Effective: effective, AppliedOn: appliedOn, Batch: revision, Policy: preview.Policy,
			})
		}
	}
	return nil
}
