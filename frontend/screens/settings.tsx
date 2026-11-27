import { BadgeCheck } from 'lucide-react';
import { Action, FieldInput, FieldSelect, Panel } from './components';
import type { PageData } from './view';
import { SettingsStaff } from './settings-staff';

export function SettingsPanel({ page }: { page: PageData }) {
  const smtp = page.SMTP;
  return (
    <>
      <Panel className="card room-form" id="smtp" aria-labelledby="smtp-title">
        <div className="room-toolbar">
          <h2 id="smtp-title">Email xác minh qua SMTP</h2>
          <span className={smtp.Configured ? 'badge good' : 'badge'}>
            {smtp.Configured ? 'Đã bật' : 'Chưa bật'}
          </span>
        </div>
        {smtp.CanEdit ? (
          <>
            <p className="subtle">
              Cấu hình dùng chung cho mọi tổ chức trong ZentraSpace.
            </p>
            <form method="post" action="/action/smtp-save" className="form">
              <input type="hidden" name="csrf" value={page.CSRF} />
              <input type="hidden" name="revision" value={smtp.Revision} />
              <div className="pair">
                <label>
                  Máy chủ SMTP
                  <FieldInput
                    name="host"
                    placeholder="smtp.example.com"
                    defaultValue={smtp.Host}
                    maxLength={253}
                    required
                  />
                </label>
                <label>
                  Cổng
                  <FieldInput
                    name="port"
                    type="number"
                    inputMode="numeric"
                    min={1}
                    max={65535}
                    defaultValue={smtp.Port}
                    required
                  />
                </label>
              </div>
              <label>
                Mã hóa kết nối
                <FieldSelect name="mode" defaultValue={smtp.Mode}>
                  <option value="tls">
                    TLS trực tiếp · thường dùng cổng 465
                  </option>
                  <option value="starttls">
                    STARTTLS · thường dùng cổng 587
                  </option>
                </FieldSelect>
              </label>
              <label>
                Địa chỉ gửi
                <FieldInput
                  name="sender"
                  type="email"
                  defaultValue={smtp.Sender}
                  maxLength={254}
                  required
                />
              </label>
              <label>
                Tên đăng nhập SMTP (tùy chọn)
                <FieldInput
                  name="username"
                  defaultValue={smtp.Username}
                  maxLength={254}
                  autoComplete="off"
                />
              </label>
              <label>
                Mật khẩu SMTP
                <FieldInput
                  name="password"
                  type="password"
                  maxLength={1024}
                  autoComplete="new-password"
                  placeholder={smtp.Authenticated ? 'Đã lưu mật khẩu' : ''}
                />
              </label>
              <p className="subtle">
                Để trống mật khẩu để giữ mật khẩu đã lưu khi máy chủ, cổng, kiểu
                mã hóa và tên đăng nhập giữ nguyên.
              </p>
              <Action>Lưu và bật SMTP</Action>
            </form>
            {smtp.Configured && (
              <div className="actions">
                <form method="post" action="/action/smtp-test">
                  <input type="hidden" name="csrf" value={page.CSRF} />
                  <Action className="secondary">Kiểm tra kết nối đã lưu</Action>
                </form>
                <form method="post" action="/action/smtp-disable">
                  <input type="hidden" name="csrf" value={page.CSRF} />
                  <input type="hidden" name="revision" value={smtp.Revision} />
                  <Action className="secondary">Tắt SMTP</Action>
                </form>
              </div>
            )}
          </>
        ) : (
          <p className="subtle">
            Quản trị hệ thống quản lý cấu hình gửi email dùng chung.
          </p>
        )}
        <details>
          <summary>
            <BadgeCheck size={17} aria-hidden /> Gửi mã cho người thuê
          </summary>
          <p>
            Lưu email ở mục <strong>Email xác minh bảng kê</strong> trong hợp
            đồng. Trên bảng kê đã phát hành, chọn{' '}
            <strong>Tạo link và gửi mã qua email</strong>.
          </p>
        </details>
      </Panel>
      <SettingsStaff page={page} />
    </>
  );
}
