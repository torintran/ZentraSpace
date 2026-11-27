import { Action, FieldInput, Panel } from './components';
import { clock } from './view';
import type { PageData } from './view';

type Account = NonNullable<PageData['MeterStaff']>[number];

function MeterStaffForm({
  page,
  account,
}: {
  page: PageData;
  account?: Account;
}) {
  return (
    <form method="post" action="/action/staff-save" className="form">
      <input type="hidden" name="csrf" value={page.CSRF} />
      {account ? (
        <>
          <input type="hidden" name="revision" value={account.Revision} />
          <input type="hidden" name="username" value={account.Username} />
        </>
      ) : (
        <label>
          Tên tài khoản
          <FieldInput
            name="username"
            autoCapitalize="none"
            autoComplete="off"
            pattern="[a-z0-9][a-z0-9_\-]{0,39}"
            maxLength={40}
            required
          />
        </label>
      )}
      <label>
        {account
          ? 'Mật khẩu mới (chỉ nhập khi cần đổi)'
          : 'Mật khẩu (tối thiểu 16 ký tự)'}
        <FieldInput
          name="password"
          type="password"
          autoComplete="new-password"
          minLength={16}
          maxLength={72}
          required={!account}
        />
      </label>
      <fieldset className="form">
        <legend>Dãy được nhập chỉ số</legend>
        {page.Sites?.map((site) => (
          <label className="check" key={site.ID}>
            <input
              type="checkbox"
              name="site"
              value={site.ID}
              defaultChecked={
                account?.SiteIDs.includes(site.ID) ?? page.Sites?.length === 1
              }
            />
            {site.Name}
          </label>
        ))}
      </fieldset>
      <label className="check">
        <input
          type="checkbox"
          name="enabled"
          value="yes"
          defaultChecked={!account?.Disabled}
        />{' '}
        Cho phép đăng nhập
      </label>
      <Action>{account ? 'Lưu quyền công tơ' : 'Tạo tài khoản công tơ'}</Action>
    </form>
  );
}

export function SettingsStaff({ page }: { page: PageData }) {
  return (
    <Panel
      className="card history room-form"
      id="meter-staff"
      aria-labelledby="meter-staff-title"
    >
      <h2 id="meter-staff-title">Nhân viên công tơ</h2>
      <p>
        Nhân viên dùng <strong>Tài khoản khác</strong> để đăng nhập. Họ chỉ xem
        phòng và ghi chỉ số cho những dãy được giao.
      </p>
      {!page.Sites?.length ? (
        <p>Thêm dãy trọ trước để cấp quyền công tơ.</p>
      ) : (
        <>
          {page.MeterStaff?.map((account) => (
            <details key={account.Username}>
              <summary>
                {page.MeterStaffNamePrefix}
                {account.Username} · {account.Disabled ? 'Đã tắt' : 'Đang bật'}
              </summary>
              <MeterStaffForm page={page} account={account} />
            </details>
          ))}
          <details>
            <summary>Thêm nhân viên công tơ</summary>
            <p className="subtle">
              Tên đăng nhập đầy đủ sẽ bắt đầu bằng{' '}
              <code>{page.MeterStaffNamePrefix}</code>.
            </p>
            <MeterStaffForm page={page} />
          </details>
        </>
      )}
      {!!page.StaffChanges?.length && (
        <details>
          <summary>Thay đổi tài khoản gần đây</summary>
          <ul>
            {page.StaffChanges.map((change, index) => (
              <li key={index}>
                {clock(change.At)} · {change.Actor} · {change.Operation} ·{' '}
                {change.Username}
              </li>
            ))}
          </ul>
        </details>
      )}
    </Panel>
  );
}
