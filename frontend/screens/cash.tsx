import { Fragment } from 'react';
import { Action, ActionLink, FieldInput, Panel } from './components';
import {
  canAttach,
  has,
  income,
  itemKey,
  kindLabel,
  money,
  randomKey,
} from './view';
import type { PageData, Invoice } from './view';
import { CashFields } from './common';

export function CashScreen({ page }: { page: PageData }) {
  return (
    <>
      <a className="back" href={'/?period=' + page.Period}>
        ← Về tổng quan
      </a>
      <div className="page-title">
        <h1>Thu chi và cọc</h1>
        <form method="get" action="/cash" className="period">
          <label>
            Kỳ báo cáo
            <FieldInput
              type="month"
              name="period"
              required
              defaultValue={page.Period}
            />
          </label>
          <Action className="secondary">Xem kỳ</Action>
        </form>
      </div>
      <Panel className="card" as="section">
        <h2>Tổng hợp kỳ {page.Period}</h2>
        <dl className="finance-metrics">
          <div>
            <dt>Thu bảng kê ròng</dt>
            <dd>{money(page.Cash.RentNet)}</dd>
          </div>
          <div>
            <dt>Chi phí vận hành</dt>
            <dd>{money(page.Cash.Expense)}</dd>
          </div>
          <div>
            <dt>Cọc đã nhận</dt>
            <dd>{money(page.Cash.Received)}</dd>
          </div>
          <div>
            <dt>Cọc đã hoàn</dt>
            <dd>{money(page.Cash.Returned)}</dd>
          </div>
          <div>
            <dt>Cọc đang giữ cuối kỳ</dt>
            <dd>{money(page.Cash.Held)}</dd>
          </div>
        </dl>
      </Panel>
      <div className="columns cash-actions">
        <Panel className="card" as="section">
          <h2>Ghi chi phí</h2>
          <form
            method="post"
            action="/action/cash-entry"
            className="form"
            aria-label="Ghi chi phí"
          >
            <input type="hidden" name="csrf" value={page.CSRF} />
            <input type="hidden" name="key" value={randomKey()} />
            <input type="hidden" name="kind" value="expense" />
            <label>
              Danh mục
              <select name="category" required>
                <option>Bảo trì và sửa chữa</option>
                <option>Điện nước chung</option>
                <option>Vệ sinh</option>
                <option>Vật tư</option>
                <option>Khác</option>
              </select>
            </label>
            <CashFields page={page} />
            <Action>Lưu khoản chi</Action>
          </form>
        </Panel>
        <Panel className="card" as="section">
          <h2>Nhận hoặc hoàn cọc</h2>
          {has(page.Cash.Rooms) ? (
            <>
              <form
                method="post"
                action="/action/cash-entry"
                className="form"
                aria-label="Nhận hoặc hoàn cọc"
              >
                <input type="hidden" name="csrf" value={page.CSRF} />
                <input type="hidden" name="key" value={randomKey()} />
                <label>
                  Loại giao dịch
                  <select name="kind">
                    <option value="deposit_in">Nhận cọc</option>
                    <option value="deposit_out">Hoàn cọc</option>
                  </select>
                </label>
                <label>
                  Phòng
                  <select name="room" required>
                    <option value="">Chọn phòng</option>
                    {has(page.Cash.Rooms)
                      ? page.Cash.Rooms.map((room, index) => (
                          <Fragment key={itemKey(room, index)}>
                            <option value={room.ID}>
                              {room.Name} · đang giữ cuối kỳ{' '}
                              {money(room.Held)}
                            </option>
                          </Fragment>
                        ))
                      : null}
                  </select>
                </label>
                <CashFields page={page} />
                <Action>Lưu giao dịch cọc</Action>
              </form>
            </>
          ) : (
            <>
              <p className="empty">Cần có phòng để ghi tiền cọc.</p>
            </>
          )}
        </Panel>
      </div>
      <Panel className="card history" as="section">
        <div className="room-toolbar">
          <h2>Sổ giao dịch</h2>
          <ActionLink
            href={'/cash.csv?period=' + page.Period}
            className="button secondary"
          >
            Xuất sổ CSV
          </ActionLink>
        </div>
        {has(page.Cash.Rows) ? (
          page.Cash.Rows.map((entry, index) => (
            <Fragment key={itemKey(entry, index)}>
              <article className="list-row">
                <div>
                  <strong>
                    {kindLabel(entry)}
                    {has(entry.RoomName) ? <> · {entry.RoomName}</> : null}
                  </strong>
                  <p>
                    {entry.Date}
                    {has(entry.Category) ? <> · {entry.Category}</> : null}
                    {has(entry.Reference) ? (
                      <> · {entry.Reference}</>
                    ) : null}
                  </p>
                  {has(entry.Note) ? (
                    <>
                      <p>{entry.Note}</p>
                    </>
                  ) : null}
                </div>
                <strong>
                  {has(income(entry)) ? <>+</> : <>−</>}
                  {money(entry.Amount)}
                </strong>
                {has(canAttach(entry)) ? (
                  <>
                    <div className="cash-receipt">
                      {has(page.CashReceipts?.[entry.Key]) ? (
                        <>
                          <ActionLink
                            className="button secondary"
                            href={'/cash/receipt/' + entry.Key}
                          >
                            Tải chứng từ
                          </ActionLink>
                        </>
                      ) : (
                        <>
                          <details>
                            <summary>Đính kèm chứng từ</summary>
                            <form
                              method="post"
                              action={'/cash/receipt/' + entry.Key}
                              encType="multipart/form-data"
                              className="form"
                            >
                              <input
                                type="hidden"
                                name="csrf"
                                value={page.CSRF}
                              />
                              <label>
                                PDF, JPG hoặc PNG, tối đa 2 MB
                                <input
                                  type="file"
                                  name="receipt"
                                  accept=".pdf,.jpg,.jpeg,.png"
                                  required
                                />
                              </label>
                              <Action className="secondary">
                                Lưu chứng từ
                              </Action>
                            </form>
                          </details>
                        </>
                      )}
                    </div>
                  </>
                ) : null}
              </article>
            </Fragment>
          ))
        ) : (
          <>
            <p className="empty">Chưa có giao dịch trong kỳ.</p>
          </>
        )}
      </Panel>
    </>
  );
}
