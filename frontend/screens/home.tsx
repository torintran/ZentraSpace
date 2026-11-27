import { Action, FieldInput } from './components';
import { has } from './view';
import type { PageData, Invoice } from './view';

import { DashboardPanel } from './home-dashboard';
import { RoomsPanel } from './home-rooms';
import { MetersPanel } from './home-meters';
import { BillingPanel } from './home-billing';

export function HomeScreen({ page }: { page: PageData }) {
  return (
    <>
      <div className="page-title">
        <div>
          <p className="eyebrow">QUẢN LÝ NHÀ TRỌ</p>
          <h1>
            {page.Tab === 'dashboard' ? (
              <>Tổng quan kỳ thu</>
            ) : page.Tab === 'meters' ? (
              <>Nhập điện nước</>
            ) : page.Tab === 'rooms' ? (
              <>Phòng và hợp đồng</>
            ) : (
              <>Chốt kỳ thu</>
            )}
          </h1>
        </div>
        <form method="get" action="/" className="period">
          <input type="hidden" name="tab" value={page.Tab} />
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
      {page.Tab === 'dashboard' ? (
        <DashboardPanel page={page} />
      ) : page.Tab === 'rooms' ? (
        <RoomsPanel page={page} />
      ) : page.Tab === 'meters' ? (
        <MetersPanel page={page} />
      ) : (
        <BillingPanel page={page} />
      )}
    </>
  );
}
