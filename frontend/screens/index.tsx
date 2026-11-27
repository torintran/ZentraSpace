import { LogOut } from 'lucide-react';
import { SiteHeader } from '../layout/site-header';
import { WorkspaceNav } from '../layout/workspace-nav';
import { WorkspaceShell } from '../layout/workspace-shell';
import { Action, Notice } from './components';
import type { PageData, Screen } from './view';
import { LoginScreen } from './login';
import { PrivacyScreen } from './privacy';
import { ErrorScreen } from './error';
import { HomeScreen } from './home';
import { RoomScreen } from './room';
import { CashScreen } from './cash';
import { BankScreen } from './bank';
import { BankReviewScreen } from './bank-review';
import { InvoiceScreen } from './invoice';
import { ImportScreen } from './import';
import { InviteScreen } from './invite';
import { VerifyScreen } from './verify';
import { SharedScreen } from './shared';
import { MockBankScreen } from './mock-bank';

const screens = {
  login: LoginScreen,
  privacy: PrivacyScreen,
  error: ErrorScreen,
  home: HomeScreen,
  room: RoomScreen,
  cash: CashScreen,
  bank: BankScreen,
  'bank-review': BankReviewScreen,
  invoice: InvoiceScreen,
  import: ImportScreen,
  invite: InviteScreen,
  verify: VerifyScreen,
  shared: SharedScreen,
  'mock-bank': MockBankScreen,
};

const wide = new Set<Screen>([
  'home',
  'invoice',
  'bank',
  'room',
  'shared',
  'cash',
]);
export function Screens({ page }: { page: PageData }) {
  const ScreenComponent = screens[page.Screen];
  const brand = (
    <a className="brand" href="/" aria-label="ZentraSpace">
      <img src="/assets/icon.svg" alt="" width={37} height={37} />
      <span>
        Zentra<span className="accent">Space</span>
      </span>
    </a>
  );
  const account = page.Role && (
    <form action="/action/logout" method="post">
      <input type="hidden" name="csrf" value={page.CSRF} />
      <Action
        className="secondary small"
        leftSection={<LogOut aria-hidden size={17} />}
      >
        Đăng xuất
      </Action>
    </form>
  );
  const content = (
    <main
      id="main"
      className={wide.has(page.Screen) ? 'layout' : 'layout narrow'}
    >
      {page.Error && (
        <Notice className="notice error" role="alert">
          {page.Error}
        </Notice>
      )}
      {page.Message && (
        <Notice className="notice" role="status">
          {page.Message}
        </Notice>
      )}
      <ScreenComponent page={page} />
    </main>
  );
  return (
    <>
      <a className="skip" href="#main">
        Đến nội dung
      </a>
      {page.Role === 'owner' ? (
        <WorkspaceShell
          brand={brand}
          account={account}
          sidebar={<WorkspaceNav page={page} />}
        >
          {content}
        </WorkspaceShell>
      ) : (
        <>
          <SiteHeader brand={brand} account={account} />
          <div className="workspace-frame">{content}</div>
        </>
      )}
      <footer className="site-footer">
        <div className="layout site-footer-inner">
          <div className="site-footer-identity">
            <span className="site-footer-brand">
              <img src="/assets/icon.svg" alt="" width={24} height={24} />
              ZentraSpace
            </span>
            <span className="site-footer-copyright">© 2026 Torin Tran</span>
          </div>
          {page.Screen !== 'mock-bank' && (
            <a
              className="site-footer-link"
              href="/privacy"
              aria-current={page.Screen === 'privacy' ? 'page' : undefined}
            >
              Quyền riêng tư
            </a>
          )}
        </div>
      </footer>
    </>
  );
}
