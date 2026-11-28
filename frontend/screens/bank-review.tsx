'use client';

import { Fragment, useState } from 'react';
import { FieldSelect, Action, ActionLink, Panel } from './components';
import { InvoicePicker } from './browser-controls';
import { clock, has, itemKey, money } from './view';
import type { PageData } from './view';

export function BankReviewScreen({ page }: { page: PageData }) {
  const review = page.BankReview;
  const [choice, setChoice] = useState(review?.Selected?.ID || '');
  if (!review) return null;
  const balanceAfterReceipt = review.Selected
    ? review.Selected.Remaining - review.Selected.Credit - review.Transfer.Amount
    : 0;
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
              <InvoicePicker invoices={review.Invoices ?? []} choice={choice} onChoice={setChoice} />
              <Action className="secondary">Kiểm tra bảng kê</Action>
            </form>
            {choice !== (review.Selected?.ID || '') && (
              <p role="status">Bấm Kiểm tra bảng kê để xem lại số tiền trước khi đối soát.</p>
            )}
            {review.Selected && review.Selected.ID === choice ? (
              <>
                <hr />
                <p>
                  Phòng <strong>{review.Selected.RoomName}</strong> · Kỳ{' '}
                  {review.Selected.Period} · Còn{' '}
                  {money(review.Selected.Remaining)}
                </p>
                {review.Selected.LeaseTenant && <p>Người thuê {review.Selected.LeaseTenant}</p>}
                {review.Suggested && review.Suggested.ID !== review.Selected.ID && (
                  <p className="notice">Bảng kê đang chọn khác phòng hoặc kỳ gợi ý từ nội dung chuyển khoản.</p>
                )}
                <p className={balanceAfterReceipt < 0 ? 'notice' : 'subtle'}>
                  Nếu ghi khoản thu mới, {balanceAfterReceipt < 0 ? 'tổng tiền dư' : 'còn phải thu'} {money(Math.abs(balanceAfterReceipt))}.
                </p>
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
                    <FieldSelect name="existing" required>
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
                    </FieldSelect>
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
