package app

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSharedFrontendViewKeepsBalanceWithoutPrivateNotes(t *testing.T) {
	invoice := &Invoice{
		LeaseTenant: "private tenant label",
		Total:       150,
		Entries:     []Entry{{Kind: "Thu", Amount: 50, Note: "private payment note"}},
		Reading:     Reading{Adjustment: MeterAdjustment{Reason: "private meter note"}},
	}
	view := frontendView(Page{Screen: "shared", Invoice: invoice})
	raw, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "private") || view.Invoice.Remaining != 100 || view.Invoice.NetCollected != 50 {
		t.Fatal("shared view leaked private notes or changed the displayed balance")
	}
	if invoice.Entries[0].Note == "" || invoice.Reading.Adjustment.Reason == "" {
		t.Fatal("shared view modified the source invoice")
	}
}
