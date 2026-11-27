package app

type RoomView struct {
	Room
	Done       bool
	Lease      LeaseTerms
	HasLease   bool
	Renewed    bool
	Handover   bool
	Adjustment MeterAdjustment
	People     int
	HasPhoto   bool
}
type Page struct {
	LoginProfile                                                              string
	Sites                                                                     []SiteView
	Policies                                                                  []RoomPolicy
	SelectedSite                                                              string
	RoomSite                                                                  *SiteView
	Residents                                                                 []Resident
	RoomTimeline                                                              []RoomEvent
	PolicyMutable, HasPhoto                                                   bool
	Screen, Role, Tab, Period, Error, Message, CSRF, InviteID, ShareURL, Code string
	Filter                                                                    string
	BillingSettings                                                           BillingSettings
	Cash                                                                      CashReport
	CashReceipts                                                              map[string]bool
	EntryDate                                                                 string
	Rooms                                                                     []RoomView
	RoomDetail                                                                *Room
	RoomFees                                                                  []RoomFee
	Lease                                                                     LeaseTerms
	LeaseHistory                                                              []LeaseTerms
	LeaseDocuments                                                            map[string]bool
	LeaseForm                                                                 bool
	RoomForm, LockOpening, LockRates                                          bool
	Dashboard                                                                 Dashboard
	Invoices                                                                  []Invoice
	Invoice                                                                   *Invoice
	ReconciledEntries                                                         map[string]bool
	Invites                                                                   []Invite
	Import                                                                    *ImportPreview
	Status                                                                    int
	Bank                                                                      *BankScreen
	BankReview                                                                *TransferReview
	PaymentBank                                                               BankAccount
	MockEvent                                                                 *mockTransfer
}
