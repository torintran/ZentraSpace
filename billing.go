package main

import "time"

type BillingSettings struct {
	DueDate, CommonFeeLabel string
	CommonFee               int64
}

func (b *Book) setBilling(period string, settings BillingSettings) error {
	if _, err := time.Parse("2006-01", period); err != nil {
		return errInvalid
	}
	if settings.DueDate != "" {
		if _, err := time.Parse("2006-01-02", settings.DueDate); err != nil || settings.DueDate < period+"-01" {
			return errInvalid
		}
	}
	if len(settings.CommonFeeLabel) > 160 || settings.CommonFee < 0 || settings.CommonFee > 100000000 ||
		(settings.CommonFee > 0 && settings.CommonFeeLabel == "") {
		return errInvalid
	}
	if settings.CommonFee == 0 {
		settings.CommonFeeLabel = ""
	}
	if b.Billing == nil {
		b.Billing = make(map[string]BillingSettings)
	}
	b.Billing[period] = settings
	for j := range b.Invoices {
		i := &b.Invoices[j]
		if i.Period == period && !i.Issued {
			if room := b.room(i.RoomID); room != nil {
				*i = b.calculateInvoice(*room, i.Reading, i.ID)
			}
		}
	}
	return nil
}
