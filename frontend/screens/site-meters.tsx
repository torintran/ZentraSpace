import { Action, FieldInput, Panel, Status } from './components';
import type { PageData } from './view';

export function SiteMeters({ page }: { page: PageData }) {
  if (!page.SiteMeterChecks?.length) return null;
  return (
    <Panel className="card history" id="site-meters">
      <h2>Đối chiếu công tơ tổng</h2>
      <p className="subtle">
        Chênh lệch được kiểm và ghi lý do trước khi chốt bảng kê. Hệ thống giữ
        tiền điện nước theo chỉ số của từng phòng.
      </p>
      {page.SiteMeterChecks.map((check) => (
        <details
          className="meter-special"
          key={check.SiteID}
          id={'site-meter-' + check.SiteID}
          open={page.SiteMeterFocus === check.SiteID || undefined}
        >
          <summary>
            {check.SiteName}{' '}
            <Status
              className={
                check.Recorded && !check.NeedsReview ? 'badge good' : 'badge'
              }
            >
              {!check.Recorded
                ? 'Chưa nhập công tơ tổng'
                : check.NeedsReview
                  ? 'Chờ duyệt'
                  : check.Approved
                    ? 'Đã duyệt'
                    : 'Khớp chỉ số'}
            </Status>
          </summary>
          <dl className="charges">
            <div>
              <dt>Tổng điện nước các phòng</dt>
              <dd>
                {check.RoomElectric} kWh · {check.RoomWater} m³
              </dd>
            </div>
            <div>
              <dt>Phòng thiếu chỉ số</dt>
              <dd>{check.Missing}</dd>
            </div>
            {check.Recorded && (
              <div>
                <dt>Công tơ tổng trừ tổng các phòng</dt>
                <dd>
                  {check.ElectricDelta} kWh · {check.WaterDelta} m³
                </dd>
              </div>
            )}
          </dl>
          {check.Approved && (
            <p>
              Đã duyệt ngày {check.ApprovedOn} bởi {check.ApprovedBy}.{' '}
              {check.ApprovalNote}
            </p>
          )}
          {check.NeedsReview && !check.Missing && (
            <form
              method="post"
              action="/action/site-meter-approve"
              className="form"
            >
              <input type="hidden" name="csrf" value={page.CSRF} />
              <input type="hidden" name="period" value={page.Period} />
              <input type="hidden" name="id" value={check.SiteID} />
              <input type="hidden" name="revision" value={check.Revision} />
              <label>
                Lý do chênh lệch
                <FieldInput name="note" maxLength={240} required />
              </label>
              <Action>Duyệt chênh lệch</Action>
            </form>
          )}
          <form method="post" action="/action/site-meter-save" className="form">
            <input type="hidden" name="csrf" value={page.CSRF} />
            <input type="hidden" name="period" value={page.Period} />
            <input type="hidden" name="id" value={check.SiteID} />
            <input type="hidden" name="revision" value={check.Revision} />
            {(['Electric', 'Water'] as const).map((kind) => {
              const label = kind === 'Electric' ? 'điện' : 'nước';
              const field = kind.toLowerCase();
              return (
                <fieldset className="form" key={kind}>
                  <legend>Công tơ tổng {label}</legend>
                  <label>
                    Mã công tơ {label}
                    <FieldInput
                      name={field + 'Meter'}
                      maxLength={48}
                      defaultValue={check[`${kind}Meter` as const]}
                      required
                    />
                  </label>
                  <div className="pair">
                    <label>
                      {label === 'điện' ? 'Điện đầu kỳ' : 'Nước đầu kỳ'}
                      <FieldInput
                        type="number"
                        inputMode="numeric"
                        name={field + 'Before'}
                        min={0}
                        max={999999}
                        defaultValue={
                          check.Recorded ? check[`${kind}Before` as const] : undefined
                        }
                        required
                      />
                    </label>
                    <label>
                      {label === 'điện' ? 'Điện cuối kỳ' : 'Nước cuối kỳ'}
                      <FieldInput
                        type="number"
                        inputMode="numeric"
                        name={field + 'After'}
                        min={0}
                        max={999999}
                        defaultValue={
                          check.Recorded ? check[`${kind}After` as const] : undefined
                        }
                        required
                      />
                    </label>
                  </div>
                </fieldset>
              );
            })}
            <Action className="secondary">Lưu công tơ tổng</Action>
          </form>
        </details>
      ))}
    </Panel>
  );
}
