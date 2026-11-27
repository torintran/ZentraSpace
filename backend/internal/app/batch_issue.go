package app

import (
	"fmt"
	"strings"
	"time"
)

func (b *Book) issueBatch(period string, selections []string) error {
	if len(selections) == 0 || len(selections) > 100 {
		return fmt.Errorf("Chọn từ 1 đến 100 bảng kê.")
	}
	pending := make([]*Invoice, 0, len(selections))
	seen := make(map[string]bool, len(selections))
	for _, selection := range selections {
		id, revision, ok := strings.Cut(selection, ":")
		i := b.invoice(id)
		if !ok || seen[id] || i == nil || i.Period != period {
			return errInvalid
		}
		seen[id] = true
		if i.Revision() != revision {
			return fmt.Errorf("Bảng kê phòng %s đã thay đổi. Mở lại danh sách để kiểm tra.", i.RoomName)
		}
		if !i.Issued {
			if err := b.canIssue(i); err != nil {
				return fmt.Errorf("Phòng %s: %w", i.RoomName, err)
			}
			pending = append(pending, i)
		}
	}
	issuedOn := time.Now().In(vietnam).Format(time.DateOnly)
	for _, i := range pending {
		b.finalizeInvoice(i, issuedOn)
	}
	return nil
}
