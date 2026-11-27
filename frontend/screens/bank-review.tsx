import { Fragment } from 'react';
import { Action, ActionLink, Panel } from './components';
import { clock, has, itemKey, money } from './view';
import type { PageData, Invoice } from './view';

export function BankReviewScreen({ page }: { page: PageData }) {
  const review = page.BankReview;
  if (!review) return null;
  return (
    <>
      <a className="back" href="/bank">
        ← Về đối soát
      </a>
      <Panel className="card auth" as="section">
        <h1>Duyệt giao dịch</h1>
        <p>
          <strong>{money(review.Transfer.Amount)}</strong> ·{' '}
          {review.Transfer.Date} · {review.Transfer.Reference}
        </p>
        {has(review.Transfer.Memo) ? (
          <>
            <p className="subtle">{review.Transfer.Memo}</p>
          </>
        ) : null}
        {has(review.Transfer.InvoiceID) ? (
          <>
            <p className="notice">Giao dịch đã được đối soát.</p>
            <ActionLink
              className="button secondary"
              href={'/invoice/' + review.Transfer.InvoiceID}
            >
              Xem bảng kê
            </ActionLink>
          </>
        ) : (
          <>
            {has(review.Suggested) ? (
              <>
                <p className="notice">
                  Mã chuyển khoản gợi ý phòng {review.Suggested.RoomName} · kỳ{' '}
                  {review.Suggested.Period}.
                </p>
              </>
            ) : (
              <>
                <p className="subtle">
                  Chưa có mã đủ rõ. Hãy kiểm tra sao kê rồi chọn bảng kê.
                </p>
              </>
            )}
            <form
              method="get"
              action={'/bank/review/' + review.Transfer.ID}
              className="form"
            >
              <label>
                Bảng kê đã phát hành
                <select
                  name="invoice"
                  required
                  defaultValue={review.Selected?.ID || ''}
                >
                  <option value="">Chọn phòng và kỳ</option>
                  {has(review.Invoices)
                    ? review.Invoices.map((invoice, index) => (
                        <Fragment key={itemKey(invoice, index)}>
                          <option value={invoice.ID}>
                            {invoice.RoomName} · {invoice.Period} · còn{' '}
                            {money(invoice.Remaining)}
                          </option>
                        </Fragment>
                      ))
                    : null}
                </select>
              </label>
              <Action className="secondary">Kiểm tra bảng kê</Action>
            </form>
            {has(review.Selected) ? (
              <>
                <hr />
                <p>
                  Phòng <strong>{review.Selected.RoomName}</strong> · Kỳ{' '}
                  {review.Selected.Period} · Còn{' '}
                  {money(review.Selected.Remaining)}
                </p>
                {review.Transfer.Amount > review.Selected.Remaining ? (
                  <>
                    <p className="notice">
                      Nếu ghi khoản thu mới, phần vượt công nợ sẽ được ghi nhận
                      là thu dư.
                    </p>
                  </>
                ) : null}
                <form
                  method="post"
                  action={'/bank/approve/' + review.Transfer.ID}
                  className="form"
                >
                  <input type="hidden" name="csrf" value={page.CSRF} />
                  <input
                    type="hidden"
                    name="invoice"
                    value={review.Selected.ID}
                  />
                  <label>
                    Cách ghi nhận
                    <select name="existing" required>
                      <option value="">Chọn cách ghi</option>
                      <option value="new">Ghi một khoản thu mới</option>
                      {has(review.Existing)
                        ? review.Existing.map((entry, index) => (
                            <Fragment key={itemKey(entry, index)}>
                              <option value={entry.Key}>
                                Ghép khoản đã ghi · {clock(entry.Date)} ·{' '}
                                {entry.Note}
                              </option>
                            </Fragment>
                          ))
                        : null}
                    </select>
                  </label>
                  <label className="check">
                    <input
                      type="checkbox"
                      name="confirm"
                      value="yes"
                      required
                    />{' '}
                    Tôi đã kiểm tra bảng kê và khoản thu.
                  </label>
                  <Action>Xác nhận đối soát</Action>
                </form>
              </>
            ) : null}
          </>
        )}
      </Panel>
    </>
  );
}
