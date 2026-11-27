import {
  ClipboardList,
  DoorOpen,
  Gauge,
  Landmark,
  LayoutDashboard,
  Wallet,
} from 'lucide-react';
import type { PageData } from '../screens/view';

const navigation = [
  { tab: 'dashboard', label: 'Tổng quan', Icon: LayoutDashboard },
  { tab: 'rooms', label: 'Phòng và hợp đồng', Icon: DoorOpen },
  { tab: 'meters', label: 'Nhập điện nước', Icon: Gauge },
  { tab: 'billing', label: 'Chốt kỳ thu', Icon: ClipboardList },
  { tab: 'bank', label: 'Đối soát ngân hàng', Icon: Landmark },
  { tab: 'cash', label: 'Thu chi và cọc', Icon: Wallet },
] as const;

export function WorkspaceNav({ page }: { page: PageData }) {
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
    <nav className="workspace-nav" aria-label="Khu vực quản lý">
      <div className="layout workspace-nav-inner">
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
            >
              <Icon aria-hidden size={18} strokeWidth={1.9} />
              <span>{label}</span>
            </a>
          );
        })}
      </div>
    </nav>
  );
}
