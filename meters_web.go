package main

import "net/http"

func formMeterAdjustment(r *http.Request) (MeterAdjustment, error) {
	form := r.Form
	a := MeterAdjustment{Handover: form.Get("handover") == "yes", Reason: form.Get("reason")}
	if a.Handover {
		electric, e1 := number(r, "openingElectric", 0, 999999)
		water, e2 := number(r, "openingWater", 0, 999999)
		if e1 != nil || e2 != nil {
			return a, errInvalid
		}
		a.OpeningElectric, a.OpeningWater = electric, water
	}
	for _, item := range []struct {
		prefix string
		swap   *MeterSwap
	}{{"electric", &a.Electric}, {"water", &a.Water}} {
		if form.Get(item.prefix+"Replaced") != "yes" {
			continue
		}
		old, e1 := number(r, item.prefix+"OldEnd", 0, 999999)
		start, e2 := number(r, item.prefix+"NewStart", 0, 999999)
		if e1 != nil || e2 != nil {
			return a, errInvalid
		}
		*item.swap = MeterSwap{Replaced: true, OldEnd: old, NewStart: start}
	}
	return a, nil
}
