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
  Effective: string;
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
};
type RoomView = Room & {
  Done: boolean;
  Lease: Lease;
  HasLease: boolean;
  Renewed: boolean;
  Handover: boolean;
  Adjustment: Adjustment;
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
};
export type Invoice = {
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
  Held: number;
  Rows: CashRow[] | null;
  Rooms: { ID: string; Name: string; Held: number }[] | null;
};
type BankAccount = { BIN: string; Account: string; Holder: string };
type BankTransfer = {
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
  LeaseForm: boolean;
  RoomForm: boolean;
  LockOpening: boolean;
  LockRates: boolean;
  Dashboard: Dashboard;
  Invoices: Invoice[];
  Invoice: Invoice | null;
  ReconciledEntries: Record<string, boolean> | null;
  Invites: { ID: string; Expires: string }[] | null;
  Import: { Period: string; Add: number; Skip: number; Data: string } | null;
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
  if (!lease.Effective || lease.Months === 0) return '';
  const start = new Date(lease.Effective + '-01T00:00:00Z');
  start.setUTCMonth(start.getUTCMonth() + lease.Months - 1);
  return start.toISOString().slice(0, 7);
}
export function activeAt(lease: Lease, period: string) {
  const end = until(lease);
  return period >= lease.Effective && (!end || period <= end);
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
    { expense: 'Chi phí', deposit_in: 'Nhận cọc', deposit_out: 'Hoàn cọc' }[
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
