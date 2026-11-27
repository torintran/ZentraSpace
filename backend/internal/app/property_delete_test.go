package app

import "testing"

func TestDeleteUnusedRoomAndEmptySite(t *testing.T) {
	room := Room{ID: "room", SiteID: "site", Prepared: true}
	book := Book{
		Rooms: []Room{room}, Sites: []Site{{ID: "site"}},
		RoomFees:          []RoomFee{{RoomID: room.ID}},
		PolicyAssignments: []PolicyAssignment{{RoomID: room.ID}},
	}
	if err := book.deleteSite("site"); err == nil {
		t.Fatal("a site containing a room must not be deleted")
	}
	if _, err := book.deleteUnusedRoom(room.ID, "stale"); err == nil {
		t.Fatal("stale confirmation must not delete a room")
	}
	if siteID, err := book.deleteUnusedRoom(room.ID, room.Revision()); err != nil || siteID != "site" {
		t.Fatalf("delete unused room: site %q, error %v", siteID, err)
	}
	if len(book.Rooms) != 0 || len(book.RoomFees) != 0 || len(book.PolicyAssignments) != 0 {
		t.Fatal("unused room setup should be removed with the room")
	}
	if err := book.deleteSite("site"); err != nil || len(book.Sites) != 0 {
		t.Fatalf("delete empty site: %v", err)
	}
}

func TestDeleteRoomPreservesBusinessHistory(t *testing.T) {
	for name, book := range map[string]Book{
		"leases":    {Leases: []LeaseTerms{{RoomID: "room"}}},
		"invoices":  {Invoices: []Invoice{{RoomID: "room"}}},
		"readings":  {Readings: []Reading{{RoomID: "room"}}},
		"cash":      {CashEntries: []CashEntry{{RoomID: "room"}}},
		"residents": {Residents: []Resident{{RoomID: "room"}}},
	} {
		t.Run(name, func(t *testing.T) {
			room := Room{ID: "room", Prepared: true}
			book.Rooms = []Room{room}
			if _, err := book.deleteUnusedRoom(room.ID, room.Revision()); err == nil || len(book.Rooms) != 1 {
				t.Fatal("the room must be retained while its history exists")
			}
		})
	}
}
