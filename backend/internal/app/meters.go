package app

import (
	"errors"
	"strings"
	"time"
)

type MeterSwap struct {
	Replaced         bool
	OldEnd, NewStart int64
}

type MeterAdjustment struct {
	Handover                      bool
	OpeningElectric, OpeningWater int64
	Electric, Water               MeterSwap
	Reason                        string
}

type meterOpening struct {
	Electric, Water int64
	Handover        bool
}

func (b *Book) meterOpening(room Room, period string) (meterOpening, error) {
	start, err := time.Parse("2006-01", period)
	if err != nil || period < room.Start {
		return meterOpening{}, errInvalid
	}
	lease := b.leaseAt(room, period)
	if !lease.ActiveAt(period) {
		return meterOpening{}, errors.New("Phòng không có hợp đồng hiệu lực trong kỳ này.")
	}
	if period == room.Start {
		return meterOpening{Electric: room.Electric, Water: room.Water}, nil
	}
	previous := start.AddDate(0, -1, 0).Format("2006-01")
	if lease.Effective == period && !b.leaseAt(room, previous).ActiveAt(previous) {
		return meterOpening{Handover: true}, nil
	}
	for _, i := range b.Invoices {
		if i.RoomID == room.ID && i.Period == previous && i.Issued {
			return meterOpening{Electric: i.Reading.ElectricAfter, Water: i.Reading.WaterAfter}, nil
		}
	}
	return meterOpening{}, errors.New("Hãy phát hành bảng kê của kỳ trước để chốt chỉ số đầu kỳ.")
}

func (b *Book) read(id, period string, electric, water int64) error {
	var adjustment MeterAdjustment
	if r := b.reading(id, period); r != nil {
		adjustment = r.Adjustment
	}
	return b.readWithAdjustment(id, period, electric, water, adjustment)
}

func (b *Book) readWithAdjustment(id, period string, electric, water int64, adjustment MeterAdjustment) error {
	room := b.room(id)
	if room == nil {
		return errMissing
	}
	for _, i := range b.Invoices {
		if i.RoomID == id && i.Period == period && i.Issued {
			return errLocked
		}
	}
	opening, err := b.meterOpening(*room, period)
	if err != nil {
		return err
	}
	if opening.Handover != adjustment.Handover {
		if opening.Handover {
			return errors.New("Cần chỉ số bàn giao đầu kỳ cho hợp đồng sau thời gian phòng trống.")
		}
		return errors.New("Chỉ được nhập mốc bàn giao cho kỳ đầu của hợp đồng sau thời gian phòng trống.")
	}
	adjustment.Reason = strings.TrimSpace(adjustment.Reason)
	if adjustment.Handover || adjustment.Electric.Replaced || adjustment.Water.Replaced {
		if adjustment.Reason == "" || len(adjustment.Reason) > 240 {
			return errors.New("Nhập lý do hoặc tham chiếu biên bản công tơ.")
		}
	} else {
		adjustment.Reason = ""
	}
	if adjustment.Handover {
		opening.Electric, opening.Water = adjustment.OpeningElectric, adjustment.OpeningWater
	}
	if !validMeter(opening.Electric, electric, adjustment.Electric) || !validMeter(opening.Water, water, adjustment.Water) {
		return errors.New("Kiểm tra chỉ số 0–999.999. Số cuối phải từ số đầu trở lên trên từng công tơ.")
	}
	r := Reading{RoomID: id, Period: period, ElectricBefore: opening.Electric, WaterBefore: opening.Water,
		ElectricAfter: electric, WaterAfter: water, Adjustment: adjustment}
	if prior := b.reading(id, period); prior != nil {
		*prior = r
	} else {
		b.Readings = append(b.Readings, r)
	}
	for j := range b.Invoices {
		i := &b.Invoices[j]
		if i.RoomID == id && i.Period == period && !i.Issued {
			*i = b.calculateInvoice(*room, r, i.ID)
		}
	}
	return nil
}

func validMeter(before, after int64, swap MeterSwap) bool {
	if before < 0 || before > 999999 || after < 0 || after > 999999 {
		return false
	}
	if !swap.Replaced {
		return after >= before
	}
	return swap.OldEnd >= before && swap.OldEnd <= 999999 && swap.NewStart >= 0 && after >= swap.NewStart
}

func meterUsage(before, after int64, swap MeterSwap) int64 {
	if swap.Replaced {
		return swap.OldEnd - before + after - swap.NewStart
	}
	return after - before
}

func (r Reading) ElectricUsage() int64 {
	return meterUsage(r.ElectricBefore, r.ElectricAfter, r.Adjustment.Electric)
}
func (r Reading) WaterUsage() int64 {
	return meterUsage(r.WaterBefore, r.WaterAfter, r.Adjustment.Water)
}
