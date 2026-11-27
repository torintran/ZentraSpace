import {
  ClipboardList,
  DoorOpen,
  Gauge,
  Landmark,
  LayoutDashboard,
  MapPin,
  Menu,
  Settings,
  Tags,
  Wallet,
} from 'lucide-react';
import type { PageData } from '../screens/view';

const navigation = [
  { tab: 'dashboard', label: 'Tổng quan', Icon: LayoutDashboard },
  { tab: 'sites', label: 'Dãy trọ', Icon: MapPin },
  { tab: 'rooms', label: 'Phòng và hợp đồng', Icon: DoorOpen },
  { tab: 'policies', label: 'Mẫu giá và phí', Icon: Tags },
  { tab: 'meters', label: 'Nhập điện nước', Icon: Gauge },
  { tab: 'billing', label: 'Chốt kỳ thu', Icon: ClipboardList },
  { tab: 'bank', label: 'Đối soát ngân hàng', Icon: Landmark },
  { tab: 'cash', label: 'Thu chi và cọc', Icon: Wallet },
  { tab: 'settings', label: 'Cấu hình', Icon: Settings },
] as const;

function WorkspaceLinks({ page }: { page: PageData }) {
  const period = page.Period || page.Invoice?.Period;
  const selected =
    page.Screen === 'home'
      ? page.Tab
      : (
          {
            'bank-review': 'bank',
            invoice: 'billing',
            invite: 'billing',
            room: 'rooms',
            import: 'rooms',
          } as Record<string, string>
        )[page.Screen] || page.Screen;
  return (
    <nav className="workspace-links" aria-label="Khu vực quản lý">
      {navigation.map(({ tab, label, Icon }) => {
        const params = new URLSearchParams(period ? { period } : {});
        if (tab !== 'bank' && tab !== 'cash') params.set('tab', tab);
        const query = params.size ? '?' + params.toString() : '';
        const href =
          tab === 'bank'
            ? '/bank'
            : tab === 'cash'
              ? '/cash' + query
              : '/' + query;
        return (
          <a
            key={tab}
            href={href}
            aria-current={selected === tab ? 'page' : undefined}
            title={label}
          >
            <Icon aria-hidden size={18} strokeWidth={1.9} />
            <span className="workspace-link-label">{label}</span>
          </a>
        );
      })}
    </nav>
  );
}

export function WorkspaceNav({ page }: { page: PageData }) {
  return (
    <aside className="workspace-sidebar">
      <div className="workspace-sidebar-desktop" id="workspace-desktop-links">
        <WorkspaceLinks page={page} />
      </div>
      <details className="workspace-sidebar-mobile">
        <summary>
          <Menu size={20} aria-hidden /> Menu quản lý
        </summary>
        <WorkspaceLinks page={page} />
      </details>
    </aside>
  );
}
