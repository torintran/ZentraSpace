import { Fragment } from 'react';
import { RoomList } from './browser-controls';
import { BatchPublish } from './batch-publish';
import { Action, ActionLink, FieldInput, Panel, Status } from './components';
import { has, itemKey, money } from './view';
import type { PageData, Invoice } from './view';

export function BillingPanel({ page }: { page: PageData }) {
  const siteQuery = page.SelectedSite ? '&site=' + encodeURIComponent(page.SelectedSite) : '';
  const drafts = page.Filter === 'draft' ? (page.Invoices ?? []).filter((invoice) => !invoice.Issued) : [];
  return (
    <>
      <div className="columns">
        <Panel className="card" as="section">
          <h2>Bảng kê kỳ {page.Period}</h2>
          <p className="subtle">Tạo bản nháp cho phòng đã chốt điện nước.</p>
          {!!page.SiteMeterChecks?.length && (
            <p className="empty">
              Chờ duyệt công tơ tổng của{' '}
              {page.SiteMeterChecks.map((check) => check.SiteName).join(', ')}.{' '}
              <a href={'/?tab=meters&period=' + page.Period + siteQuery + '#site-meters'}>
                Xem chênh lệch
              </a>
            </p>
          )}
          <div className="actions">
            <form method="post" action="/action/draft">
              <input type="hidden" name="csrf" value={page.CSRF} />
              <input type="hidden" name="period" value={page.Period} />
              {page.SelectedSite && <input type="hidden" name="site" value={page.SelectedSite} />}
              {has(page.Filter) ? (
                <>
                  <input type="hidden" name="filter" value={page.Filter} />
                </>
              ) : null}
              <Action>Tạo bảng kê nháp</Action>
            </form>
            <ActionLink
              className="button secondary"
              href={'/export?period=' + page.Period + siteQuery}
            >
              Xuất công nợ
            </ActionLink>
            <ActionLink className="button secondary" href={'/export?view=pending&period=' + page.Period}>
              {page.SelectedSite ? 'Xuất khoản chờ đối soát (tất cả dãy)' : 'Xuất khoản nhận chờ đối soát'}
            </ActionLink>
          </div>
          <nav className="tabs" aria-label="Lọc bảng kê">
            <a
              href={'/?tab=billing&period=' + page.Period + siteQuery}
              className={!has(page.Filter) ? 'active' : undefined}
              aria-current={!has(page.Filter) ? 'page' : undefined}
            >
              Tất cả
            </a>
            <a
              href={'/?tab=billing&period=' + page.Period + siteQuery + '&filter=draft'}
              className={page.Filter === 'draft' ? 'active' : undefined}
              aria-current={page.Filter === 'draft' ? 'page' : undefined}
            >
              Chờ phát hành
            </a>
            <a
              href={'/?tab=billing&period=' + page.Period + siteQuery + '&filter=due'}
              className={page.Filter === 'due' ? 'active' : undefined}
              aria-current={page.Filter === 'due' ? 'page' : undefined}
            >
              Còn phải thu
            </a>
            <a
              href={'/?tab=billing&period=' + page.Period + siteQuery + '&filter=overdue'}
              className={page.Filter === 'overdue' ? 'active' : undefined}
              aria-current={page.Filter === 'overdue' ? 'page' : undefined}
            >
              Quá hạn
            </a>
            <a
              href={'/?tab=billing&period=' + page.Period + siteQuery + '&filter=credit'}
              className={page.Filter === 'credit' ? 'active' : undefined}
              aria-current={page.Filter === 'credit' ? 'page' : undefined}
            >
              Có tiền dư
            </a>
          </nav>
          {drafts.length > 0 && (
            <details>
              <summary>Phát hành nhiều bảng kê</summary>
              <BatchPublish drafts={drafts.map(({ID, Revision, RoomName, Total}) => ({ID, Revision, RoomName, Total}))}
                csrf={page.CSRF} period={page.Period} site={page.SelectedSite} />
            </details>
          )}
          {has(page.Invoices) ? (
            <RoomList
              label="Tìm theo mã phòng hoặc tên người thuê"
              empty="Không tìm thấy bảng kê."
              listClass="billing-results"
              rooms={page.Invoices.map((invoice) => ({
                id: invoice.ID,
                name: [invoice.RoomName, invoice.LeaseTenant].filter(Boolean).join(' '),
                row: (
                  <article className="list-row">
                    <div>
                      <strong>{invoice.RoomName}</strong>
                      {invoice.LeaseTenant && <p>{invoice.LeaseTenant}</p>}
                      <p>
                        {!has(invoice.Issued) ? (
                          <>Bản nháp</>
                        ) : has(invoice.Credit) ? (
                          <><span>Có tiền thu dư</span> · {money(invoice.Credit)}</>
                        ) : has(invoice.Remaining) ? (
                          <>
                            <span>{invoice.DueDate && invoice.DueDate < page.Today ? 'Quá hạn' : 'Còn phải thu'}</span> · {money(invoice.Remaining)}
                          </>
                        ) : (
                          <>Đã thu đủ</>
                        )}
                      </p>
                      {!invoice.Issued && <p>Tổng bảng kê {money(invoice.Total)}</p>}
                      {invoice.DueDate && <p>Hạn thu <time dateTime={invoice.DueDate}>{invoice.DueDate.split('-').reverse().join('/')}</time></p>}
                    </div>
                    <ActionLink
                      className="button secondary"
                      href={'/invoice/' + invoice.ID}
                    >
                      Xem bảng kê
                    </ActionLink>
                  </article>
                ),
              }))}
            />
          ) : (
            <>
              <div className="empty">
                {has(page.Filter) ? (
                  <>Không có bảng kê khớp bộ lọc.</>
                ) : (
                  <>Chưa có bảng kê cho kỳ này.</>
                )}
              </div>
            </>
          )}
        </Panel>
        <Panel className="card" as="section">
          <h2>Chuẩn bị kỳ thu</h2>
          {has(page.Rooms) ? (
            page.Rooms.map((room, index) => (
              <Fragment key={itemKey(room, index)}>
                <div className="list-row">
                  <strong>{room.Name}</strong>
                  <Status className={'badge ' + (has(room.Done) ? 'good' : '')}>
                    {has(room.Done) ? <>Có chỉ số</> : <>Thiếu chỉ số</>}
                  </Status>
                </div>
              </Fragment>
            ))
          ) : (
            <>
              <p className="subtle">Thêm một phòng để bắt đầu kỳ thu.</p>
              <ActionLink
                className="button secondary"
                href={'/?tab=rooms&period=' + page.Period + siteQuery}
              >
                Thêm phòng
              </ActionLink>
            </>
          )}
          <hr />
          <details>
            <summary>Hạn thanh toán và phí chung</summary>
            <p className="subtle">
              Áp dụng toàn bộ dãy cho bảng kê nháp và bảng kê tạo sau trong kỳ {page.Period}
              .
            </p>
            <form
              method="post"
              action="/action/billing-settings"
              className="form"
            >
              <input type="hidden" name="csrf" value={page.CSRF} />
              <input type="hidden" name="period" value={page.Period} />
              {page.SelectedSite && <input type="hidden" name="site" value={page.SelectedSite} />}
              <label>
                Hạn thanh toán
                <FieldInput
                  name="dueDate"
                  type="date"
                  min={page.Period + '-01'}
                  defaultValue={page.BillingSettings.DueDate}
                />
              </label>
              <label>
                Tên khoản phí
                <FieldInput
                  name="commonFeeLabel"
                  maxLength={40}
                  defaultValue={page.BillingSettings.CommonFeeLabel}
                />
              </label>
              <label>
                Phí chung mỗi phòng (đồng)
                <FieldInput
                  name="commonFee"
                  type="number"
                  inputMode="numeric"
                  min="0"
                  max="100000000"
                  required
                  defaultValue={page.BillingSettings.CommonFee}
                />
              </label>
              <Action className="secondary">Lưu thiết lập kỳ</Action>
            </form>
          </details>
        </Panel>
      </div>
    </>
  );
}
