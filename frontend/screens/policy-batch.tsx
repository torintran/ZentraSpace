import { Action, ActionLink, FieldInput, Panel } from './components';
import { money } from './view';
import type { PageData } from './view';

type Preview = NonNullable<PageData['PolicyPreview']>;

function Prices({ policy }: { policy: Preview['Policy'] }) {
  return (
    <dl className="charges">
      <div>
        <dt>Tiền thuê</dt>
        <dd>{money(policy.Rent)}/tháng</dd>
      </div>
      <div>
        <dt>Điện</dt>
        <dd>{money(policy.ElectricRate)}/kWh</dd>
      </div>
      <div>
        <dt>Nước</dt>
        <dd>{money(policy.WaterRate)}/m³</dd>
      </div>
      <div>
        <dt>{policy.FeeLabel || 'Phí định kỳ'}</dt>
        <dd>{money(policy.Fee)}/tháng</dd>
      </div>
    </dl>
  );
}

export function PolicyBatchForm({ page }: { page: PageData }) {
  if (!page.Policies?.length) return null;
  return (
    <Panel className="card history">
      <h2>Áp dụng mẫu cho nhiều phòng</h2>
      <p>
        Chọn nhóm phòng và kỳ bắt đầu. Bạn sẽ thấy các thay đổi trước khi xác
        nhận.
      </p>
      <form className="form" method="post" action="/action/policy-preview">
        <input type="hidden" name="csrf" value={page.CSRF} />
        <input type="hidden" name="period" value={page.Period} />
        <label>
          Mẫu giá và phí
          <select name="policy" required defaultValue="">
            <option value="" disabled>
              Chọn mẫu giá
            </option>
            {page.Policies.map((policy) => (
              <option key={policy.ID} value={policy.ID}>
                {policy.Name}
              </option>
            ))}
          </select>
        </label>
        <label>
          Nhóm phòng
          <select name="scope" required defaultValue="">
            <option value="" disabled>
              Chọn nhóm phòng
            </option>
            <option value="unassigned">Phòng chưa gán dãy</option>
            {(page.Sites ?? []).map((site) => (
              <option key={site.ID} value={site.ID}>
                {site.Name}
              </option>
            ))}
          </select>
        </label>
        <label>
          Kỳ hiệu lực
          <FieldInput
            name="effective"
            type="month"
            defaultValue={page.Period}
            required
          />
        </label>
        <Action>Xem trước thay đổi</Action>
      </form>
    </Panel>
  );
}

export function PolicyBatchPreview({
  page,
  preview,
}: {
  page: PageData;
  preview: Preview;
}) {
  const back = '/?tab=policies&period=' + page.Period;
  return (
    <div className="columns">
      <Panel className="card">
        <h2>Xem trước áp dụng mẫu</h2>
        <p>
          {preview.ScopeName} · Từ kỳ {preview.Effective}
        </p>
        <p>
          <strong>{preview.Eligible} phòng đủ điều kiện</strong>
        </p>
        {(preview.Rooms ?? []).map((room) => (
          <article className="list-row" key={room.ID}>
            <div>
              <strong>{room.Name}</strong>
              {room.Blocked ? (
                <p>Bỏ qua · {room.Blocked}</p>
              ) : (
                <details>
                  <summary>Giá hiện tại của phòng</summary>
                  <Prices policy={room.Before} />
                </details>
              )}
            </div>
          </article>
        ))}
      </Panel>
      <Panel className="card">
        <h2>{preview.Policy.Name}</h2>
        <Prices policy={preview.Policy} />
        <div className="actions">
          {preview.Eligible > 0 && (
            <form method="post" action="/action/policy-apply">
              <input type="hidden" name="csrf" value={page.CSRF} />
              <input type="hidden" name="period" value={page.Period} />
              <input type="hidden" name="effective" value={preview.Effective} />
              <input type="hidden" name="policy" value={preview.Policy.ID} />
              <input type="hidden" name="scope" value={preview.Scope} />
              <input type="hidden" name="revision" value={preview.Revision} />
              <Action>Áp dụng cho {preview.Eligible} phòng</Action>
            </form>
          )}
          <ActionLink className="secondary" href={back}>
            Quay lại
          </ActionLink>
        </div>
      </Panel>
    </div>
  );
}
