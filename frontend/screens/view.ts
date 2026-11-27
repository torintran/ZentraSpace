import { randomBytes } from 'node:crypto';

export type Screen =
  | 'login'
  | 'privacy'
  | 'error'
  | 'home'
  | 'room'
  | 'cash'
  | 'bank'
  | 'bank-review'
  | 'invoice'
  | 'import'
  | 'invite'
  | 'verify'
  | 'shared'
  | 'mock-bank';

type MeterSwap = { Replaced: boolean; OldEnd: number; NewStart: number };
type Adjustment = {
  Handover: boolean;
  OpeningElectric: number;
  OpeningWater: number;
  Electric: MeterSwap;
  Water: MeterSwap;
  Reason: string;
};
type Reading = {
  ElectricBefore: number;
  ElectricAfter: number;
  WaterBefore: number;
  WaterAfter: number;
  Adjustment: Adjustment;
};
type Lease = {
  Tenant: string;
  Phone: string;
  SignedOn: string;
  StartOn: string;
  Effective: string;
  EndedOn: string;
  Rent: number;
  ElectricRate: number;
  WaterRate: number;
  Months: number;
  Renewal: boolean;
};
type Room = {
  ID: string;
  Name: string;
  Tenant: string;
  Start: string;
  Rent: number;
  ElectricRate: number;
  WaterRate: number;
  Electric: number;
  Water: number;
  LeaseMonths: number;
  Prepared?: boolean;
  SiteID: string;
  PolicyID: string;
  Description: string;
  PolicyFeeLabel: string;
  PolicyFee: number;
};
type RoomView = Room & {
  Done: boolean;
  Lease: Lease;
  HasLease: boolean;
  Renewed: boolean;
  Handover: boolean;
  Adjustment: Adjustment;
  People: number;
  HasPhoto: boolean;
};
type Site = {
  HasMeterHistory: boolean;
  ID: string;
  Name: string;
  Address: string;
  Coordinates: string;
  Capacity: number;
  Rooms: number;
  Occupied: number;
  MapURL: string;
};
type RoomPolicy = {
  ID: string;
  Name: string;
  Rent: number;
  ElectricRate: number;
  WaterRate: number;
  Fee: number;
  FeeLabel: string;
};
type Resident = {
  ID: string;
  Name: string;
  Phone: string;
  Email?: string;
  JoinedOn: string;
  LeftOn: string;
};
type RoomEvent = {
  Date: string;
  Kind: string;
  Title: string;
  Detail: string;
  Amount: number;
  InvoiceID: string;
};
type FeeLine = { Label: string; Amount: number };
type RoomFee = FeeLine & {
  ID: string;
  Start: string;
  End: string;
  Monthly: boolean;
};
type Entry = {
  Key: string;
  Kind: string;
  Date: string;
  Note: string;
  Amount: number;
  Reverses?: string;
  Method?: string;
  Reference?: string;
  ResidentShare?: number;
};
export type Invoice = {
  RentDays?: number;
  RentMonthDays?: number;
  ID: string;
  RoomID: string;
  RoomName: string;
  Period: string;
  LeaseEffective: string;
  LeaseTenant: string;
  DueDate: string;
  CommonFeeLabel: string;
  CommonFee: number;
  RoomFees?: FeeLine[];
  ResidentShares?: {Name: string; Days: number; Amount: number}[];
  ResidentBalances?: number[];
  ResidentUnallocated?: number;
  Rent: number;
  ElectricRate: number;
  WaterRate: number;
  ElectricAmount: number;
  WaterAmount: number;
  Total: number;
  Reading: Reading;
  Issued: boolean;
  Entries: Entry[] | null;
  PaymentCode: string;
  Revision: string;
  Remaining: number;
  Credit: number;
  NetCollected: number;
  NetReduction: number;
};
type CashRow = {
  Key: string;
  Kind: string;
  Date: string;
  RoomID: string;
  RoomName: string;
  Category: string;
  Reference: string;
  Note: string;
  Amount: number;
};
type Cash = {
  RentNet: number;
  Expense: number;
  Received: number;
  Returned: number;
  Applied: number;
  Held: number;
  Rows: CashRow[] | null;
  Rooms: { ID: string; Name: string; Held: number }[] | null;
};
type BankAccount = { BIN: string; Account: string; Holder: string };
type BankTransfer = {
  ManualCash?: boolean;
  ID: string;
  AccountSuffix: string;
  Reference: string;
  Date: string;
  Memo: string;
  Amount: number;
  InvoiceID: string;
  EntryKey: string;
};
type Bank = {
  Account: BankAccount;
  Options: { BIN: string; Code: string; Name: string }[];
  BankName: string;
  Transfers: BankTransfer[] | null;
  Reviewed: boolean;
  Next: number;
  Previous: number;
  HasNext: boolean;
  HasPrevious: boolean;
  MockAvailable: boolean;
  MockAutoMatch: boolean;
};
type BankReview = {
  Transfer: BankTransfer;
  Invoices: Invoice[];
  Selected: Invoice | null;
  Suggested: Invoice | null;
  Existing: Entry[] | null;
};
type Dashboard = {
  Rooms: number;
  Missing: number;
  Drafts: number;
  Outstanding: number;
  Issued: number;
  Settled: number;
  Billed: number;
  Collected: number;
  Due: number;
  Credit: number;
  TrendMax: number;
  RoomStatus:
    { Name: string; Status: string; Label: string; InvoiceID: string }[] | null;
  Trend:
    | {
        Period: string;
        Label: string;
        Issued: number;
        Collected: number;
        Due: number;
      }[]
    | null;
};
export type PageData = {
  ResidentSplitMode?: string;
  MeterStaffNamePrefix: string;
  MeterStaff: { Username: string; Revision: string; SiteIDs: string[]; Disabled: boolean }[] | null;
  StaffChanges: { Username: string; Actor: string; Operation: string; At: string }[] | null;
  AvailableCredit: number;
  CreditTargets: { ID: string; Period: string; Remaining: number }[] | null;
  CreditTransfers: { Key: string; TargetID: string; TargetPeriod: string; Amount: number; Date: string; Note: string; Reversed: boolean }[] | null;
  SiteMeterFocus: string;
  SiteMeterChecks: {SiteID: string; SiteName: string; Period: string; Revision: string; ElectricMeter: string; WaterMeter: string; ElectricBefore: number; ElectricAfter: number; WaterBefore: number; WaterAfter: number; RoomElectric: number; RoomWater: number; ElectricDelta: number; WaterDelta: number; Missing: number; Recorded: boolean; NeedsReview: boolean; Approved: boolean; ApprovedBy: string; ApprovedOn: string; ApprovalNote: string}[] | null;
  MeterReviewHeld: boolean;
  RoomArchive: boolean;
  ArchiveKind: string;
  ArchivePeriod: string;
  ArchivePage: number;
  ArchivePages: number;
  LeaseCount: number;
  LeaseEmails: Record<string, string> | null;
  VerificationEmail: string;
  ResidentEmailReady?: boolean[];
  PersonalBill?: {
    Room: string;
    Period: string;
    DueDate: string;
    Name: string;
    Days: number;
    Amount: number;
    Collected: number;
    Reduced: number;
    Remaining: number;
    Credit: number;
  };
  ManualInvite?: boolean;
  CodeDelivery: string;
  ExpiringLeases: {RoomID: string; RoomName: string; Tenant: string; EndsOn: string; NextPeriod: string}[] | null;
  RenewPeriod: string;
  CheckoutForm: boolean;
  CheckoutDate: string;
  Today: string;
  PeriodDate: string;
  PeriodEnd: string;
  NextLeasePeriod: string;
  SMTP: {
    Configured: boolean; Authenticated: boolean; CanEdit: boolean;
    Address: string; Host: string; Port: string; Sender: string;
    Mode: string; Username: string; Revision: string;
  };
  CheckoutError: string;
  CheckoutQuote: {Date: string; Effective: string; Revision: string; Invoice: Invoice; Held: number; EarlierDue: number; Remaining: number} | null;
  EditSiteLocation: boolean;
  ShowSiteDetails: boolean;
  EditSiteDetails: boolean;
  SiteRevision: string;
  RoomDeleteReason: string;
  PolicyPreview: {
    Scope: string;
    ScopeName: string;
    Effective: string;
    Revision: string;
    Policy: RoomPolicy;
    Rooms:
      | { ID: string; Name: string; Blocked: string; Before: RoomPolicy }[]
      | null;
    Eligible: number;
  } | null;
  RoomPolicyFee: FeeLine;
  RoomHasScheduledPolicy: boolean;
  LoginProfile: string;
  Sites: Site[] | null;
  Policies: RoomPolicy[] | null;
  SelectedSite: string;
  RoomSite: Site | null;
  Residents: Resident[] | null;
  RoomTimeline: RoomEvent[] | null;
  PolicyMutable: boolean;
  HasPhoto: boolean;
  Screen: Screen;
  Role: string;
  Tab: string;
  Period: string;
  Error: string;
  Message: string;
  CSRF: string;
  InviteID: string;
  ShareURL: string;
  Code: string;
  Filter: string;
  EntryDate: string;
  BillingSettings: {
    DueDate: string;
    CommonFeeLabel: string;
    CommonFee: number;
  };
  Cash: Cash;
  CashReceipts: Record<string, boolean> | null;
  Rooms: RoomView[] | null;
  RoomDetail: (Room & { Revision: string }) | null;
  RoomFees: RoomFee[] | null;
  Lease: Lease;
  LeaseHistory: Lease[] | null;
  LeaseDocuments: Record<string, { ID: string; Label: string }[]> | null;
  LeaseForm: boolean;
  RoomForm: boolean;
  LockOpening: boolean;
  LockRates: boolean;
  Dashboard: Dashboard;
  Invoices: Invoice[];
  Invoice: Invoice | null;
  ReconciledEntries: Record<string, boolean> | null;
  Invites: { ID: string; Expires: string; ResidentShare?: number }[] | null;
  Import: { Period: string; Add: number; Skip: number; Data: string; SiteID: string; SiteName: string } | null;
  Bank: Bank | null;
  BankReview: BankReview | null;
  PaymentBank: BankAccount;
  MockEvent: {
    transaction_id: string;
    date: string;
    amount_vnd: number;
    memo: string;
  } | null;
};

const amountFormat = new Intl.NumberFormat('vi-VN');
export function money(amount: number) {
  return amountFormat.format(amount) + ' ₫';
}

const clockFormat = new Intl.DateTimeFormat('vi-VN', {
  timeZone: 'Asia/Ho_Chi_Minh',
  day: '2-digit',
  month: '2-digit',
  year: 'numeric',
  hour: '2-digit',
  minute: '2-digit',
  hourCycle: 'h23',
});
export function clock(value: string) {
  const time = Object.fromEntries(
    clockFormat
      .formatToParts(new Date(value))
      .map(({ type, value }) => [type, value]),
  );
  return (
    time.hour +
    ':' +
    time.minute +
    ' · ' +
    time.day +
    '/' +
    time.month +
    '/' +
    time.year
  );
}
export function has<T>(
  value: T,
): value is Exclude<T, null | undefined | false | '' | 0> {
  return Array.isArray(value) ? value.length > 0 : Boolean(value);
}
export function itemKey(item: object, index: number) {
  for (const property of ['ID', 'Key', 'Period', 'Name']) {
    const value = Reflect.get(item, property);
    if (typeof value === 'string' && value) return value;
  }
  return index;
}
export function randomKey() {
  return randomBytes(24).toString('hex');
}
export function until(lease: Lease) {
  if (lease.EndedOn) return lease.EndedOn.slice(0, 7);
  if (!lease.Effective || lease.Months === 0) return '';
  const start = new Date(lease.Effective + '-01T00:00:00Z');
  start.setUTCMonth(start.getUTCMonth() + lease.Months - 1);
  return start.toISOString().slice(0, 7);
}
export function activeAt(lease: Lease, period: string) {
  const end = until(lease);
  return (
    !!lease.Effective && period >= lease.Effective && (!end || period <= end)
  );
}
export function reversed(invoice: Invoice, key: string) {
  return invoice.Entries?.some((entry) => entry.Reverses === key) ?? false;
}
export function usage(reading: Reading, kind: 'Electric' | 'Water') {
  const meter = reading.Adjustment[kind];
  return (
    reading[(kind + 'After') as 'ElectricAfter' | 'WaterAfter'] -
    reading[(kind + 'Before') as 'ElectricBefore' | 'WaterBefore'] +
    (meter.Replaced ? meter.OldEnd - meter.NewStart : 0)
  );
}
export function configured(bank: BankAccount) {
  return (
    /^\d{6}$/.test(bank.BIN) && /^\d{4,19}$/.test(bank.Account) && !!bank.Holder
  );
}
export function suffix(bank: BankAccount) {
  return bank.Account.slice(-4);
}
export function methodLabel(entry: Entry) {
  if (entry.Method === 'cash') return 'Tiền mặt';
  if (entry.Method === 'transfer') return 'Chuyển khoản';
  return '';
}
export function kindLabel(row: CashRow) {
  return (
    { expense: 'Chi phí', deposit_in: 'Nhận cọc', deposit_out: 'Hoàn cọc', deposit_applied: 'Cấn cọc vào bảng kê' }[
      row.Kind
    ] ?? row.Kind
  );
}
export function income(row: CashRow) {
  return row.Kind === 'deposit_in' || row.Kind === 'Thu';
}
export function canAttach(row: CashRow) {
  return ['expense', 'deposit_in', 'deposit_out'].includes(row.Kind);
}
