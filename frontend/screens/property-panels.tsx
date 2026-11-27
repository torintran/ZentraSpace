import { ArrowUpRight, DoorOpen, MapPin, Plus } from 'lucide-react';
import { Action, ActionLink, FieldInput, Panel } from './components';
import { money } from './view';
import type { PageData } from './view';
import { PolicyBatchForm, PolicyBatchPreview } from './policy-batch';
import { SavedSiteMap, SiteLocationField } from './site-location';
import { PropertyDelete } from './property-delete';

type RentalSite = NonNullable<PageData['Sites']>[number];

function SiteLocationPanel({
  page,
  site,
  href,
}: {
  page: PageData;
  site: RentalSite;
  href: string;
}) {
  return (
    <Panel className="card" id="site-map-panel">
      <div className="room-toolbar">
        <h2>{page.EditSiteLocation ? 'Sửa vị trí' : 'Vị trí dãy trọ'}</h2>
        <ActionLink
          className="secondary"
          href={href + '&view=details#site-details-panel'}
        >
          Chi tiết dãy trọ
        </ActionLink>
      </div>
      <p>
        <strong>{site.Name}</strong>
      </p>
      {page.EditSiteLocation ? (
        <form className="form" method="post" action="/action/site-location">
          <input type="hidden" name="csrf" value={page.CSRF} />
          <input type="hidden" name="period" value={page.Period} />
          <input type="hidden" name="id" value={site.ID} />
          <label>
            Địa chỉ
            <FieldInput
              name="address"
              maxLength={360}
              defaultValue={site.Address}
            />
          </label>
          <SiteLocationField initialCoordinates={site.Coordinates} expanded />
          <div className="actions">
            <Action>Lưu vị trí</Action>
            <ActionLink className="secondary" href={href + '#site-map-panel'}>
              Hủy
            </ActionLink>
          </div>
        </form>
      ) : (
        <>
          {site.Address && <p>{site.Address}</p>}
          {site.Coordinates ? (
            <SavedSiteMap coordinates={site.Coordinates} />
          ) : (
            <p className="empty">
              Chưa ghim vị trí. Mở phần sửa vị trí để chọn điểm trên bản đồ.
            </p>
          )}
          <div className="actions">
            <ActionLink href={href + '&location=edit#site-map-panel'}>
              Sửa vị trí
            </ActionLink>
            {site.MapURL && (
              <ActionLink
                href={site.MapURL}
                target="_blank"
                rel="noreferrer"
                className="secondary"
                leftSection={<ArrowUpRight size={17} aria-hidden />}
              >
                Mở OpenStreetMap
              </ActionLink>
            )}
          </div>
        </>
      )}
    </Panel>
  );
}

function SiteDetailsPanel({
  page,
  site,
  href,
}: {
  page: PageData;
  site: RentalSite;
  href: string;
}) {
  const roomURL =
    '/?tab=rooms&period=' +
    page.Period +
    '&site=' +
    encodeURIComponent(site.ID);
  const detailsURL = href + '&view=details';
  return (
    <Panel
      className="card"
      id="site-details-panel"
      aria-labelledby="site-details-heading"
    >
      <h2 id="site-details-heading">Chi tiết dãy trọ</h2>
      {page.EditSiteDetails ? (
        <form className="form" method="post" action="/action/site-edit">
          <input type="hidden" name="csrf" value={page.CSRF} />
          <input type="hidden" name="period" value={page.Period} />
          <input type="hidden" name="id" value={site.ID} />
          <input type="hidden" name="revision" value={page.SiteRevision} />
          <label>
            Tên dãy trọ
            <FieldInput
              name="name"
              maxLength={120}
              defaultValue={site.Name}
              required
            />
          </label>
          <label>
            Số phòng tối đa
            <FieldInput
              name="capacity"
              type="number"
              min={Math.max(1, site.Rooms)}
              max={150}
              defaultValue={site.Capacity}
              required
            />
          </label>
          <div className="actions">
            <Action>Lưu thông tin</Action>
            <ActionLink
              className="secondary"
              href={detailsURL + '#site-details-panel'}
            >
              Hủy
            </ActionLink>
          </div>
        </form>
      ) : (
        <>
          <p className="site-details-name">{site.Name}</p>
          <p className="subtle">{site.Address || 'Chưa có địa chỉ'}</p>
          <dl className="site-details-stats">
            <div>
              <dt>Phòng đã tạo</dt>
              <dd>
                {site.Rooms}/{site.Capacity}
              </dd>
            </div>
            <div>
              <dt>Đang thuê</dt>
              <dd>{site.Occupied}</dd>
            </div>
            <div>
              <dt>Phòng trống</dt>
              <dd>{site.Rooms - site.Occupied}</dd>
            </div>
          </dl>
          <div className="actions">
            <ActionLink
              className="secondary"
              href={detailsURL + '&edit=1#site-details-panel'}
            >
              Sửa thông tin
            </ActionLink>
            <ActionLink
              href={roomURL}
              leftSection={<DoorOpen size={17} aria-hidden />}
            >
              Xem phòng
            </ActionLink>
            <ActionLink
              className="secondary"
              href={href + '#site-map-panel'}
              leftSection={<MapPin size={17} aria-hidden />}
            >
              Xem vị trí
            </ActionLink>
          </div>
        </>
      )}
      <div className="site-details-management">
        <PropertyDelete
          page={page}
          kind="site"
          id={site.ID}
          name={site.Name}
          reason={
            site.Rooms > 0
              ? 'Dãy trọ vẫn còn phòng. Hãy chuyển phòng hoặc xóa các phòng đủ điều kiện trước.'
              : site.HasMeterHistory
                ? 'Dãy trọ đã có lịch sử công tơ tổng.'
                : undefined
          }
        />
      </div>
    </Panel>
  );
}

export function SitesPanel({ page }: { page: PageData }) {
  const sites = page.Sites ?? [];
  const current = sites.find((site) => site.ID === page.SelectedSite);
  const baseURL = '/?tab=sites&period=' + page.Period;
  const siteURL = (id: string) => baseURL + '&site=' + encodeURIComponent(id);
  return (
    <div className="columns site-workspace">
      <Panel className="card">
        <div className="room-toolbar">
          <h2>Danh sách dãy trọ</h2>
          {current && (
            <ActionLink
              className="secondary"
              href={baseURL + '#site-create-panel'}
              leftSection={<Plus size={17} aria-hidden />}
            >
              Thêm dãy trọ
            </ActionLink>
          )}
        </div>
        {sites.map((site) => (
          <article
            className="list-row site-list-item"
            key={site.ID}
            id={'site-' + site.ID}
          >
            <div>
              <a
                className="site-name"
                href={siteURL(site.ID) + '&view=details#site-details-panel'}
              >
                {site.Name}
              </a>
              {site.Address && <p>{site.Address}</p>}
              <p>
                {site.Rooms}/{site.Capacity} phòng đã tạo ·{' '}
                {site.Rooms - site.Occupied} phòng trống
              </p>
            </div>
            <div className="actions">
              <ActionLink
                className="secondary"
                aria-current={
                  current?.ID === site.ID && page.ShowSiteDetails
                    ? 'location'
                    : undefined
                }
                href={siteURL(site.ID) + '&view=details#site-details-panel'}
              >
                Chi tiết
              </ActionLink>
              <ActionLink
                className="secondary"
                aria-current={
                  current?.ID === site.ID && !page.ShowSiteDetails
                    ? 'location'
                    : undefined
                }
                href={siteURL(site.ID) + '#site-map-panel'}
                leftSection={<MapPin size={17} aria-hidden />}
              >
                Xem vị trí
              </ActionLink>
              <ActionLink
                className="secondary"
                href={
                  '/?tab=rooms&period=' +
                  page.Period +
                  '&site=' +
                  encodeURIComponent(site.ID)
                }
              >
                Xem phòng
              </ActionLink>
            </div>
          </article>
        ))}
        {!sites.length && (
          <p className="subtle">
            Chưa có dãy trọ. Thêm dãy đầu tiên để nhóm các phòng theo địa chỉ.
          </p>
        )}
      </Panel>
      {current && page.ShowSiteDetails ? (
        <SiteDetailsPanel
          page={page}
          site={current}
          href={siteURL(current.ID)}
        />
      ) : current ? (
        <SiteLocationPanel
          page={page}
          site={current}
          href={siteURL(current.ID)}
        />
      ) : (
        <Panel className="card" id="site-create-panel">
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
            <SiteLocationField />
            <Action>Lưu dãy trọ</Action>
          </form>
        </Panel>
      )}
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
