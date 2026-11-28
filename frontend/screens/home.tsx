import { Action, FieldInput, FieldSelect } from './components';
import { has } from './view';
import type { PageData } from './view';

import { DashboardPanel } from './home-dashboard';
import { RoomsPanel } from './home-rooms';
import { MetersPanel } from './home-meters';
import { BillingPanel } from './home-billing';
import { PoliciesPanel, SitesPanel } from './property-panels';
import { SettingsPanel } from './settings';

const workspace = {
  dashboard: { label: 'Tổng quan kỳ thu', Panel: DashboardPanel },
  meters: { label: 'Nhập điện nước', Panel: MetersPanel },
  rooms: { label: 'Phòng và hợp đồng', Panel: RoomsPanel },
  sites: { label: 'Dãy trọ', Panel: SitesPanel },
  policies: { label: 'Mẫu giá và phí', Panel: PoliciesPanel },
  billing: { label: 'Chốt kỳ thu', Panel: BillingPanel },
  settings: { label: 'Cấu hình', Panel: SettingsPanel },
};

export function HomeScreen({ page }: { page: PageData }) {
  const { label, Panel } =
    workspace[page.Tab as keyof typeof workspace] ?? workspace.dashboard;
  const sites = page.Tab === 'meters' || page.Tab === 'billing' ? page.Sites ?? [] : [];
  return (
    <>
      <div className="page-title">
        <div>
          <p className="eyebrow">QUẢN LÝ NHÀ TRỌ</p>
          <h1>{label}</h1>
        </div>
        {page.Tab !== 'settings' && (
          <form method="get" action="/" className="period">
            <input type="hidden" name="tab" value={page.Tab} />
            {sites.length > 0 ? (
              <label>
                Dãy trọ
                <FieldSelect name="site" aria-label="Dãy trọ" defaultValue={page.SelectedSite} key={page.SelectedSite}>
                  <option value="">Tất cả dãy và phòng lẻ</option>
                  {sites.map((site) => <option key={site.ID} value={site.ID}>{site.Name}</option>)}
                </FieldSelect>
              </label>
            ) : page.SelectedSite && (
              <input type="hidden" name="site" value={page.SelectedSite} />
            )}
            {page.ShowSiteDetails && (
              <input type="hidden" name="view" value="details" />
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
        )}
      </div>
      <Panel page={page} />
    </>
  );
}
