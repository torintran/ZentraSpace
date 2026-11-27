import { Fragment } from 'react';
import { FieldSelect, Action, FieldInput, Panel, Status } from './components';
import {
  clock,
  has,
  itemKey,
  methodLabel,
  money,
  randomKey,
  reversed,
} from './view';
import type { PageData, Invoice } from './view';
import { InvoiceDetails } from './common';
import { InvoiceCredit } from './invoice-credit';

export function InvoiceScreen({ page }: { page: PageData }) {
  const bill = page.Invoice;
  if (!bill) return null;
  return (
    <>
      <a className="back" href={'/?period=' + bill.Period}>
        ← Về kỳ thu
      </a>
      <div className="page-title">
        <div>
          <p className="eyebrow">PHÒNG {bill.RoomName}</p>
          <h1>Bảng kê tiền thuê</h1>
          <p className="subtle">Kỳ {bill.Period}</p>
        </div>
        <Status className={'badge ' + (has(bill.Issued) ? 'good' : '')}>
          {has(bill.Issued) ? <>Đã phát hành</> : <>Bản nháp</>}
        </Status>
      </div>
      {has(bill.LeaseTenant) ? (
        <>
          <p className="subtle">Người đại diện {bill.LeaseTenant}</p>
        </>
      ) : null}
      <div className="columns">
        <Panel className="card" as="section">
          <InvoiceDetails invoice={bill} />
          {!!bill.ResidentShares?.length && (
            <section className="history">
              <h2>Chia theo số ngày ở</h2>
              <dl className="charges">
                {bill.ResidentShares.map((share, index) => (
                  <div key={index}><dt>{share.Name} · {share.Days} ngày</dt><dd>{money(share.Amount)}</dd></div>
                ))}
              </dl>
            </section>
          )}
        </Panel>
        <Panel className="card" as="section">
          {has(bill.Issued) ? (
            <>
              <h2>Ghi nhận thu tiền</h2>
              <form method="post" action="/action/post" className="form">
                <input type="hidden" name="csrf" value={page.CSRF} />
                <input type="hidden" name="id" value={bill.ID} />
                <input type="hidden" name="key" value={randomKey()} />
                <input type="hidden" name="kind" value="Thu" />
                <label>
                  Số tiền (đồng)
                  <FieldInput
                    type="number"
                    min="1"
                    max="100000000000"
                    inputMode="numeric"
                    name="amount"
                    required
                  />
                </label>
                <label>
                  Phương thức
                  <FieldSelect name="method" required>
                    <option value="">Chọn phương thức</option>
                    <option value="cash">Tiền mặt</option>
                    <option value="transfer">Chuyển khoản</option>
                  </FieldSelect>
                </label>
                <label>
                  Mã chứng từ (không bắt buộc)
                  <FieldInput name="reference" maxLength={40} />
                </label>
                <label>
                  Ghi chú (không bắt buộc)
                  <FieldInput name="note" maxLength={80} />
                </label>
                <Action>Ghi khoản thu</Action>
              </form>
              <details className="history">
                <summary>Ghi giảm hoặc hoàn tiền</summary>
                <form method="post" action="/action/post" className="form">
                  <input type="hidden" name="csrf" value={page.CSRF} />
                  <input type="hidden" name="id" value={bill.ID} />
                  <input type="hidden" name="key" value={randomKey()} />
                  <label>
                    Loại giao dịch
                    <FieldSelect name="kind">
                      <option>Giảm</option>
                      <option>Hoàn</option>
                    </FieldSelect>
                  </label>
                  <label>
                    Số tiền (đồng)
                    <FieldInput
                      type="number"
                      min="1"
                      max="100000000000"
                      inputMode="numeric"
                      name="amount"
                      required
                    />
                  </label>
                  <label>
                    Lý do
                    <FieldInput name="note" maxLength={80} required />
                  </label>
                  <Action className="secondary">Ghi điều chỉnh</Action>
                </form>
              </details>
              <hr />
              <h2>Gửi bảng kê riêng</h2>
              {has(bill.PaymentCode) ? (
                <>
                  <p className="subtle">
                    Mã đối soát <strong>{bill.PaymentCode}</strong>
                  </p>
                </>
              ) : null}
              <p className="subtle">
                Tạo link để gửi qua Zalo. {page.ManualInvite
                  ? 'Có thể trao mã riêng khi thử cục bộ.'
                  : 'Mã dùng một lần được gửi tới email đã đăng ký của người thuê.'}
              </p>
              {page.ManualInvite && <form method="post" action="/action/invite">
                <input type="hidden" name="csrf" value={page.CSRF} />
                <input type="hidden" name="id" value={bill.ID} />
                <Action className="secondary">Tạo link mới</Action>
              </form>}
              {page.VerificationEmail && (
                <form
                  method="post"
                  action="/action/invite-email"
                  className="form"
                >
                  <input type="hidden" name="csrf" value={page.CSRF} />
                  <input type="hidden" name="id" value={bill.ID} />
                  <p className="subtle">
                    Email nhận mã: {page.VerificationEmail}
                  </p>
                  <Action className="secondary">
                    Tạo link và gửi mã qua email
                  </Action>
                </form>
              )}
              {!page.SMTP.Configured ? (
                <p className="subtle">
                  Chưa cấu hình gửi mã qua email.{' '}
                  <a href="/?tab=settings#smtp">Xem cấu hình SMTP</a>
                </p>
              ) : (
                !page.VerificationEmail && (
                  <p className="subtle">
                    Chưa có email nhận mã.{' '}
                    <a
                      href={
                        '/rooms/' +
                        bill.RoomID +
                        '?period=' +
                        bill.Period +
                        (bill.LeaseEffective
                          ? '&history=1&kind=contracts&archivePeriod=' +
                            bill.LeaseEffective
                          : '') +
                        '#room-contracts'
                      }
                    >
                      Mở hồ sơ hợp đồng
                    </a>
                  </p>
                )
              )}
              {has(page.Invites)
                ? page.Invites.map((invite, index) => (
                    <Fragment key={itemKey(invite, index)}>
                      <form
                        method="post"
                        action="/action/revoke"
                        className="revoke"
                      >
                        <input type="hidden" name="csrf" value={page.CSRF} />
                        <input type="hidden" name="id" value={invite.ID} />
                        <Action className="secondary">
                          Thu hồi link hiện tại
                        </Action>{' '}
                        <small>Hết hạn {clock(invite.Expires)}</small>
                      </form>
                    </Fragment>
                  ))
                : null}
            </>
          ) : (
            <>
              <h2>Duyệt bảng kê</h2>
              <p className="subtle">
                Kiểm tra tiền thuê và cách tính điện nước. Sau khi phát hành,
                thay đổi được ghi thành giao dịch điều chỉnh.
              </p>
              {page.MeterReviewHeld ? (
                <p className="empty">
                  Dãy trọ còn công tơ tổng cần duyệt.{' '}
                  <a
                    href={'/?tab=meters&period=' + bill.Period + '#site-meters'}
                  >
                    Xem chênh lệch
                  </a>
                </p>
              ) : (
                <form method="post" action="/action/issue">
                  <input type="hidden" name="csrf" value={page.CSRF} />
                  <input type="hidden" name="id" value={bill.ID} />
                  <input type="hidden" name="revision" value={bill.Revision} />
                  <Action>Phát hành bảng kê</Action>
                </form>
              )}
            </>
          )}
        </Panel>
      </div>
      <InvoiceCredit page={page} />
      {has(bill.Entries) ? (
        <>
          <Panel className="card history" as="section">
            <h2>Lịch sử giao dịch</h2>
            {has(bill.Entries)
              ? bill.Entries.map((entry, index) => (
                  <Fragment key={itemKey(entry, index)}>
                    <div className="list-row">
                      <div>
                        <strong>
                          {entry.Kind}
                          {has(methodLabel(entry)) ? (
                            <> · {methodLabel(entry)}</>
                          ) : null}
                        </strong>
                        <p>
                          {clock(entry.Date)}{' '}
                          {has(entry.Note) ? <> · {entry.Note}</> : null}
                        </p>
                        {has(entry.Reference) ? (
                          <>
                            <p>
                              Mã chứng từ <code>{entry.Reference}</code>
                            </p>
                          </>
                        ) : null}
                        {has(entry.Reverses) ? (
                          <>
                            <p>
                              Chứng từ gốc <code>{entry.Reverses}</code>
                            </p>
                          </>
                        ) : null}
                        <p>
                          Mã bút toán <code>{entry.Key}</code>
                        </p>
                      </div>
                      <strong>{money(entry.Amount)}</strong>
                    </div>
                    {has(
                      has(entry.Kind === 'Thu' || entry.Kind === 'Giảm') &&
                        has(!has(reversed(bill, entry.Key))) &&
                        !has(page.ReconciledEntries?.[entry.Key]),
                    ) ? (
                      <>
                        <details>
                          <summary>Đảo bút toán</summary>
                          <form
                            method="post"
                            action="/action/reverse"
                            className="form"
                          >
                            <input
                              type="hidden"
                              name="csrf"
                              value={page.CSRF}
                            />
                            <input type="hidden" name="id" value={bill.ID} />
                            <input
                              type="hidden"
                              name="original"
                              value={entry.Key}
                            />
                            <input
                              type="hidden"
                              name="key"
                              value={randomKey()}
                            />
                            <label>
                              Lý do đảo bút toán
                              <FieldInput name="note" maxLength={80} required />
                            </label>
                            <Action className="secondary">
                              Xác nhận đảo {money(entry.Amount)}
                            </Action>
                          </form>
                        </details>
                      </>
                    ) : null}
                  </Fragment>
                ))
              : null}
          </Panel>
        </>
      ) : null}
    </>
  );
}
