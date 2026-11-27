package app

import "testing"

func TestRoomImportIsAtomicAndIdempotent(t *testing.T) {
	const header = "phong,nguoi_dai_dien,tien_thue,gia_dien,gia_nuoc,dien_dau,nuoc_dau\n"
	const row = "P101,Người thuê,2500000,3000,15000,10,20\n"
	b := Book{}
	if _, err := b.importRooms("2026-11", header+row+"broken,row", true); err == nil || len(b.Rooms) != 0 {
		t.Fatal("invalid import wrote data")
	}
	if p, err := b.importRooms("2026-11", header+row, false); err != nil || p.Add != 1 || len(b.Rooms) != 0 {
		t.Fatal(p, err)
	}
	if p, err := b.importRooms("2026-11", header+row, true); err != nil || p.Add != 1 || len(b.Rooms) != 1 {
		t.Fatal(p, err)
	}
	if p, err := b.importRooms("2026-11", header+row, true); err != nil || p.Skip != 1 || len(b.Rooms) != 1 {
		t.Fatal(p, err)
	}
}
