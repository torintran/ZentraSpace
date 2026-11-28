import { Fragment } from 'react';
import {
  FieldSelect,
  Action,
  ActionLink,
  FieldArea,
  FieldInput,
  Panel,
} from './components';
import { configured, has, itemKey, money, randomKey, suffix } from './view';
import type { PageData } from './view';

export function BankScreen({ page }: { page: PageData }) {
  const bank = page.Bank;
  if (!bank) return null;
  function listURL(reviewed: boolean, offset = 0, search = bank?.Search || '') {
    const query = new URLSearchParams();
    if (reviewed) query.set('status', 'reviewed');
    if (offset) query.set('offset', String(offset));
    if (search) query.set('q', search);
    return '/bank' + (query.size ? '?' + query.toString() : '');
  }
  return (
    <>
      <a className="back" href="/">
        ← Về kỳ thu
      </a>
      <div className="page-title">
        <div>
          <p className="eyebrow">THU TIỀN</p>
          <h1>Đối soát ngân hàng</h1>
        </div>
      </div>
      <div className="columns">
        <Panel className="card" as="section">
          <h2>Giao dịch tiền vào</h2>
          <nav className="tabs" aria-label="Trạng thái đối soát">
            <a className={!has(bank.Reviewed) ? 'active' : ''} href={listURL(false)}>
              Chờ duyệt
            </a>
            <a
              className={has(bank.Reviewed) ? 'active' : ''}
              href={listURL(true)}
            >
              Đã đối soát
            </a>
          </nav>
          <form method="get" action="/bank" className="form">
            {bank.Reviewed && <input type="hidden" name="status" value="reviewed" />}
            <label>
              Tìm giao dịch
              <FieldInput type="search" name="q" maxLength={120} defaultValue={bank.Search}
                placeholder="Mã, nội dung, ngày hoặc số tiền" />
            </label>
            <div className="actions">
              <Action className="secondary">Tìm giao dịch</Action>
              {bank.Search && <ActionLink className="secondary" href={listURL(bank.Reviewed, 0, '')}>Xóa tìm kiếm</ActionLink>}
            </div>
          </form>
          {bank.Matches > 0 && <p className="subtle">{bank.Matches} giao dịch</p>}
          <details>
            <summary>Ghi tiền mặt chưa phân bổ</summary>
            <form method="post" action="/bank/cash-receipt" className="form">
              <input type="hidden" name="csrf" value={page.CSRF} />
              <input type="hidden" name="key" value={randomKey()} />
              <label>Ngày nhận<FieldInput name="date" type="date" required /></label>
              <label>Số tiền (đồng)<FieldInput name="amount" type="number" min={1} max={100000000000} required /></label>
              <label>Mã chứng từ<FieldInput name="reference" maxLength={80} /></label>
              <label>Ghi chú<FieldInput name="note" maxLength={80} /></label>
              <Action>Lưu khoản nhận</Action>
            </form>
          </details>
          {has(bank.Transfers) ? (
            bank.Transfers.map((transfer, index) => (
              <Fragment key={itemKey(transfer, index)}>
                <article className="list-row">
                  <div>
                    <strong>{money(transfer.Amount)}</strong>
                    <p>
                      {transfer.Date} · {transfer.Reference} · {transfer.ManualCash ? 'Tiền mặt' : 'TK …' + transfer.AccountSuffix}
                    </p>
                    {has(transfer.Memo) ? (
                      <>
                        <p>{transfer.Memo}</p>
                      </>
                    ) : null}
                  </div>
                  <ActionLink
                    className="button secondary"
                    href={'/bank/review/' + transfer.ID}
                  >
                    {has(transfer.InvoiceID) ? (
                      <>Xem đối soát</>
                    ) : (
                      <>Duyệt giao dịch</>
                    )}
                  </ActionLink>
                </article>
              </Fragment>
            ))
          ) : (
            <>
              <p className="empty">{bank.Search ? 'Không tìm thấy giao dịch khớp.' : 'Chưa có giao dịch trong mục này.'}</p>
            </>
          )}
          {has(has(bank.HasPrevious) || has(bank.HasNext)) ? (
            <>
              <nav className="actions" aria-label="Trang giao dịch">
                {has(bank.HasPrevious) ? (
                  <>
                    <ActionLink
                      className="button secondary"
                      href={listURL(bank.Reviewed, bank.Previous)}
                    >
                      Trang trước
                    </ActionLink>
                  </>
                ) : null}
                {has(bank.HasNext) ? (
                  <>
                    <ActionLink
                      className="button secondary"
                      href={listURL(bank.Reviewed, bank.Next)}
                    >
                      Trang sau
                    </ActionLink>
                  </>
                ) : null}
              </nav>
            </>
          ) : null}
        </Panel>
        <Panel className="card" as="section">
          <h2>Tài khoản nhận tiền</h2>
          {has(configured(bank.Account)) ? (
            <>
              <p className="notice">
                {bank.BankName || 'Tài khoản nhận tiền'} · ••••
                {suffix(bank.Account)}
              </p>
            </>
          ) : (
            <>
              <p className="subtle">
                Lưu tài khoản để người thuê có thể xem VietQR sau khi xác minh
                bảng kê.
              </p>
            </>
          )}
          <details
            open={has(!has(configured(bank.Account))) ? true : undefined}
          >
            <summary>
              {has(configured(bank.Account)) ? (
                <>Đổi tài khoản</>
              ) : (
                <>Nhập thông tin tài khoản</>
              )}
            </summary>
            <form method="post" action="/bank/account" className="form">
              <input type="hidden" name="csrf" value={page.CSRF} />
              <label>
                Ngân hàng
                <FieldSelect
                  name="bank"
                  required
                  defaultValue={bank.BankName ? bank.Account.BIN : ''}
                >
                  <option value="">Chọn ngân hàng</option>
                  {bank.Options.map((bank) => (
                    <option key={bank.BIN} value={bank.BIN}>
                      {bank.Name}
                    </option>
                  ))}
                </FieldSelect>
              </label>
              <label>
                Số tài khoản nhận
                <FieldInput
                  type="text"
                  name="account"
                  inputMode="numeric"
                  pattern="[0-9]{4,19}"
                  maxLength={19}
                  required
                />
              </label>
              <label>
                Tên chủ tài khoản
                <FieldInput name="holder" maxLength={80} required />
              </label>
              {has(configured(bank.Account)) ? (
                <>
                  <p className="subtle">
                    Đổi ngân hàng hoặc số tài khoản sẽ thu hồi các link xem cũ.
                    Hãy gửi link mới sau khi kiểm tra tài khoản.
                  </p>
                </>
              ) : null}
              <label className="check">
                <input type="checkbox" name="confirm" value="yes" required />{' '}
                Tôi đã kiểm tra tài khoản nhận.
              </label>
              <Action>Lưu tài khoản</Action>
            </form>
          </details>
          {has(bank.MockAvailable) ? (
            <>
              <hr />
              <h2>Đối soát mô phỏng</h2>
              <p className="subtle">
                {has(bank.MockAutoMatch) ? (
                  <>Tự đối soát đang bật.</>
                ) : (
                  <>Giao dịch mô phỏng đang chờ chủ trọ duyệt.</>
                )}{' '}
                Chỉ tự ghép khi có đúng một mã bảng kê, số tiền khớp công nợ còn
                lại và không có khoản ghi tay mơ hồ.
              </p>
              <form method="post" action="/bank/mock-settings">
                <input type="hidden" name="csrf" value={page.CSRF} />
                <Action
                  className="secondary"
                  name="mode"
                  value={has(bank.MockAutoMatch) ? 'manual' : 'auto'}
                >
                  {has(bank.MockAutoMatch) ? (
                    <>Chuyển về duyệt tay</>
                  ) : (
                    <>Bật tự đối soát mô phỏng</>
                  )}
                </Action>
              </form>
            </>
          ) : null}
          {has(configured(bank.Account)) ? (
            <>
              <hr />
              <h2>Nhập sao kê CSV</h2>
              <p className="subtle">
                Dùng giao dịch tiền vào của tài khoản đang chọn. Ngày dạng
                YYYY-MM-DD, số tiền nguyên dương. Mỗi dòng cần mã giao dịch ngân
                hàng.
              </p>
              <form method="post" action="/bank/import" className="form">
                <input type="hidden" name="csrf" value={page.CSRF} />
                <label>
                  Nội dung CSV
                  <FieldArea
                    name="csv"
                    rows={4}
                    spellCheck={false}
                    required
                    defaultValue="ma_giao_dich,ngay,so_tien,noi_dung"
                  ></FieldArea>
                </label>
                <Action className="secondary">Thêm vào hàng chờ</Action>
              </form>
            </>
          ) : null}
        </Panel>
      </div>
    </>
  );
}
