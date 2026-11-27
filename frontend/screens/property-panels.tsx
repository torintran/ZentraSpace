import { ArrowUpRight, MapPin } from 'lucide-react';
import { Action, ActionLink, FieldInput, Panel } from './components';
import { money } from './view';
import type { PageData } from './view';
import { PolicyBatchForm, PolicyBatchPreview } from './policy-batch';

export function SitesPanel({ page }: { page: PageData }) {
  return (
    <div className="columns">
      <Panel className="card">
        <h2>Danh sách dãy trọ</h2>
        {(page.Sites ?? []).map((site) => (
          <article className="list-row" key={site.ID}>
            <div>
              <strong>{site.Name}</strong>
              {site.Address && <p>{site.Address}</p>}
              <p>
                {site.Rooms}/{site.Capacity} phòng đã tạo ·{' '}
                {site.Rooms - site.Occupied} phòng trống
              </p>
            </div>
            <div className="actions">
              {site.MapURL && (
                <a
                  href={site.MapURL}
                  target="_blank"
                  rel="noreferrer"
                  className="button secondary"
                >
                  <MapPin size={17} aria-hidden /> Mở bản đồ{' '}
                  <ArrowUpRight size={15} aria-hidden />
                </a>
              )}
              <ActionLink
                className="button secondary"
                href={'/?tab=rooms&period=' + page.Period + '&site=' + site.ID}
              >
                Xem phòng
              </ActionLink>
            </div>
          </article>
        ))}
        {!page.Sites?.length && (
          <p className="subtle">
            Chưa có dãy trọ. Thêm dãy đầu tiên để nhóm các phòng theo địa chỉ.
          </p>
        )}
      </Panel>
      <Panel className="card">
        <h2>Thêm dãy trọ</h2>
        <form className="form" method="post" action="/action/site">
          <input type="hidden" name="csrf" value={page.CSRF} />
          <input type="hidden" name="period" value={page.Period} />
          <label>
            Tên dãy trọ
            <FieldInput name="name" maxLength={120} required />
          </label>
          <label>
            Địa chỉ
            <FieldInput name="address" maxLength={360} />
          </label>
          <label>
            Số phòng tối đa
            <FieldInput
              name="capacity"
              type="number"
              min={1}
              max={150}
              required
            />
          </label>
          <details>
            <summary>Ghim vị trí chính xác (tuỳ chọn)</summary>
            <label>
              Tọa độ
              <FieldInput
                name="coordinates"
                placeholder="Ví dụ 10.7769, 106.7009"
                maxLength={65}
              />
            </label>
          </details>
          <Action>Lưu dãy trọ</Action>
        </form>
      </Panel>
    </div>
  );
}

export function PoliciesPanel({ page }: { page: PageData }) {
  if (page.PolicyPreview) {
    return <PolicyBatchPreview page={page} preview={page.PolicyPreview} />;
  }
  return (
    <>
      <div className="columns">
        <Panel className="card">
          <h2>Mẫu đã lưu</h2>
          {(page.Policies ?? []).map((policy) => (
            <article className="list-row" key={policy.ID}>
              <div>
                <strong>{policy.Name}</strong>
                <p>
                  {money(policy.Rent)}/tháng · Điện {money(policy.ElectricRate)}
                  /kWh · Nước {money(policy.WaterRate)}/m³
                </p>
                {policy.Fee > 0 && (
                  <p>
                    {policy.FeeLabel} · {money(policy.Fee)}/tháng
                  </p>
                )}
              </div>
            </article>
          ))}
          {!page.Policies?.length && (
            <p className="subtle">
              Chưa có mẫu nào. Lưu một mẫu để dùng lại khi tạo phòng.
            </p>
          )}
        </Panel>
        <Panel className="card">
          <h2>Tạo mẫu giá và phí</h2>
          <form className="form" method="post" action="/action/policy">
            <input type="hidden" name="csrf" value={page.CSRF} />
            <input type="hidden" name="period" value={page.Period} />
            <label>
              Tên mẫu
              <FieldInput
                name="name"
                placeholder="Ví dụ Phòng tiêu chuẩn"
                maxLength={120}
                required
              />
            </label>
            <label>
              Tiền thuê mỗi tháng (đồng)
              <FieldInput
                name="rent"
                type="number"
                min={0}
                max={100000000}
                required
              />
            </label>
            <div className="pair">
              <label>
                Điện (đồng/kWh)
                <FieldInput
                  name="electricRate"
                  type="number"
                  min={0}
                  max={100000}
                  required
                />
              </label>
              <label>
                Nước (đồng/m³)
                <FieldInput
                  name="waterRate"
                  type="number"
                  min={0}
                  max={1000000}
                  required
                />
              </label>
            </div>
            <label>
              Tên khoản phí định kỳ
              <FieldInput
                name="feeLabel"
                placeholder="Ví dụ Phí Wi-Fi"
                maxLength={120}
              />
            </label>
            <label>
              Phí mỗi tháng (đồng)
              <FieldInput
                name="fee"
                type="number"
                min={0}
                max={100000000}
                defaultValue={0}
                required
              />
            </label>
            <Action>Lưu mẫu giá và phí</Action>
          </form>
        </Panel>
      </div>
      <PolicyBatchForm page={page} />
    </>
  );
}
