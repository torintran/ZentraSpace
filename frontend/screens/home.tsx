import { Action, FieldInput } from './components';
import { has } from './view';
import type { PageData, Invoice } from './view';

import { DashboardPanel } from './home-dashboard';
import { RoomsPanel } from './home-rooms';
import { MetersPanel } from './home-meters';
import { BillingPanel } from './home-billing';
import { PoliciesPanel, SitesPanel } from './property-panels';

const workspace = {
  dashboard: { label: 'Tổng quan kỳ thu', Panel: DashboardPanel },
  meters: { label: 'Nhập điện nước', Panel: MetersPanel },
  rooms: { label: 'Phòng và hợp đồng', Panel: RoomsPanel },
  sites: { label: 'Dãy trọ', Panel: SitesPanel },
  policies: { label: 'Mẫu giá và phí', Panel: PoliciesPanel },
  billing: { label: 'Chốt kỳ thu', Panel: BillingPanel },
};

export function HomeScreen({ page }: { page: PageData }) {
  const { label, Panel } =
    workspace[page.Tab as keyof typeof workspace] ?? workspace.dashboard;
  return (
    <>
      <div className="page-title">
        <div>
          <p className="eyebrow">QUẢN LÝ NHÀ TRỌ</p>
          <h1>{label}</h1>
        </div>
        <form method="get" action="/" className="period">
          <input type="hidden" name="tab" value={page.Tab} />
          {page.SelectedSite && (
            <input type="hidden" name="site" value={page.SelectedSite} />
          )}
          {has(page.Filter) ? (
            <>
              <input type="hidden" name="filter" value={page.Filter} />
            </>
          ) : null}
          <label>
            Kỳ thu
            <FieldInput
              type="month"
              name="period"
              required
              defaultValue={page.Period}
            />
          </label>
          <Action className="secondary">Xem kỳ</Action>
        </form>
      </div>
      <Panel page={page} />
    </>
  );
}
