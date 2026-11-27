package app

import (
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestRoomWorkbookRoundTripAndFormulaGuard(t *testing.T) {
	template, err := roomImportWorkbook()
	if err != nil {
		t.Fatal(err)
	}
	wb, err := excelize.OpenReader(template)
	if err != nil {
		t.Fatal(err)
	}
	defer wb.Close()
	sheet := wb.GetSheetName(0)
	row := []interface{}{"P101", "Người thuê", 2500000, 3000, 15000, 10, 20}
	if err := wb.SetSheetRow(sheet, "A2", &row); err != nil {
		t.Fatal(err)
	}
	data, err := wb.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	csv, err := roomWorkbookCSV(data)
	if err != nil || !strings.Contains(csv, "P101,Người thuê,2500000") {
		t.Fatalf("CSV %q: %v", csv, err)
	}
	b := Book{Sites: []Site{{ID: "a", Name: "Dãy A", Capacity: 1}}}
	if report, err := b.importRooms("2026-11", "a", csv, true); err != nil || report.Add != 1 || b.Rooms[0].SiteID != "a" {
		t.Fatalf("import: %+v %v", report, err)
	}
	if report, err := b.importRooms("2026-11", "a", csv, true); err != nil || report.Skip != 1 {
		t.Fatalf("replay: %+v %v", report, err)
	}
	if _, err := b.importRooms("2026-11", "a", strings.Replace(csv, "P101", "P102", 1), true); err == nil || len(b.Rooms) != 1 {
		t.Fatal("import exceeded site capacity")
	}
	if err := wb.SetCellFormula(sheet, "C2", "=1+1"); err != nil {
		t.Fatal(err)
	}
	data, err = wb.WriteToBuffer()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := roomWorkbookCSV(data); err == nil || !strings.Contains(err.Error(), "C2") {
		t.Fatalf("formula validation: %v", err)
	}
}
