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
	MeterStaffNamePrefix                                                      string
	MeterStaff                                                                []MeterStaffView
	StaffChanges                                                              []StaffChange
	AvailableCredit                                                           int64
	CreditTargets                                                             []CreditTarget
	CreditTransfers                                                           []CreditTransferView
	SiteMeterChecks                                                           []SiteMeterCheck
	SiteMeterFocus                                                            string
	MeterReviewHeld                                                           bool
	EditSiteLocation                                                          bool
	ShowSiteDetails                                                           bool
	EditSiteDetails                                                           bool
	SiteRevision                                                              string
	CheckoutForm                                                              bool
	CheckoutDate, CheckoutError                                               string
	Today, PeriodDate, PeriodEnd, NextLeasePeriod                             string
	SMTP                                                                      SMTPSettings
	RoomArchive                                                               bool
	ArchiveKind, ArchivePeriod                                                string
	ArchivePage, ArchivePages, LeaseCount                                     int
	CheckoutQuote                                                             *CheckoutQuote
	ExpiringLeases                                                            []ExpiringLease
	RenewPeriod                                                               string
	LeaseEmails                                                               map[string]string
	VerificationEmail, CodeDelivery                                           string
	ResidentEmailReady                                                        []bool        `json:",omitempty"`
	PersonalBill                                                              *PersonalBill `json:",omitempty"`
	ManualInvite                                                              bool
	RoomDeleteReason                                                          string
	PolicyPreview                                                             *PolicyPreview
	RoomPolicyFee                                                             FeeLine
	RoomHasScheduledPolicy                                                    bool
	LoginProfile                                                              string
	Sites                                                                     []SiteView
	Policies                                                                  []RoomPolicy
	SelectedSite                                                              string
	RoomSite                                                                  *SiteView
	Residents                                                                 []Resident
	ResidentSplitMode                                                         string
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
	LeaseDocuments                                                            map[string][]LeaseAttachment
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
