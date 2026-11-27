import { Fragment } from 'react';
import { Action, ActionLink, FieldInput, Panel } from './components';
import { activeAt, has, itemKey, money, randomKey, until } from './view';
import type { PageData, Invoice } from './view';
import { RoomForm } from './common';

export function RoomScreen({ page }: { page: PageData }) {
  const room = page.RoomDetail;
  if (!room) return null;
  return (
    <>
      <a className="back" href={'/?tab=rooms&period=' + page.Period}>
        ← Về danh sách phòng
      </a>
      <div className="page-title">
        <h1>
          {has(page.RoomForm) ? (
            <>Sửa phòng </>
          ) : has(page.LeaseForm) ? (
            <>Điều khoản mới · </>
          ) : null}
          {room.Name}
        </h1>
        {has(!has(page.RoomForm) && !has(page.LeaseForm)) ? (
          <>
            <ActionLink
              className="button"
              href={'/rooms/' + room.ID + '?period=' + page.Period + '&edit=1'}
            >
              Chỉnh sửa
            </ActionLink>
          </>
        ) : null}
      </div>
      {has(page.RoomForm) ? (
        <>
          <Panel className="card room-form" as="section">
            {has(page.LockRates) ? (
              <>
                <p className="subtle">
                  Để đổi người đại diện hoặc đơn giá, thêm điều khoản mới từ màn
                  chi tiết phòng.
                </p>
              </>
            ) : null}
            {has(page.LockOpening) ? (
              <>
                <p className="subtle">
                  Chỉ số gốc được khóa sau lần ghi công tơ đầu tiên.
                </p>
              </>
            ) : null}
            <RoomForm page={page} />
          </Panel>
        </>
      ) : has(page.LeaseForm) ? (
        <>
          <Panel className="card room-form" as="section">
            <p className="subtle">
              Điều khoản bắt đầu vào ngày đầu tháng. Chọn kỳ sau bảng kê đã phát
              hành gần nhất.
            </p>
            <form method="post" action="/action/lease" className="form">
              <input type="hidden" name="csrf" value={page.CSRF} />
              <input type="hidden" name="id" value={room.ID} />
              <label>
                Kỳ bắt đầu hiệu lực
                <FieldInput name="effective" type="month" required />
              </label>
              <label>
                Thời hạn (tháng, để trống nếu chưa xác định)
                <FieldInput
                  name="months"
                  type="number"
                  inputMode="numeric"
                  min="1"
                  max="120"
                />
              </label>
              <label className="check">
                <input type="checkbox" name="renewal" value="yes" /> Ký lại hợp
                đồng hiện có
              </label>
              <label>
                Người đại diện
                <FieldInput
                  name="tenant"
                  maxLength={150}
                  required
                  defaultValue={page.Lease.Tenant}
                />
              </label>
              <label>
                Tiền thuê mỗi tháng (đồng)
                <FieldInput
                  name="rent"
                  type="number"
                  min="0"
                  max="100000000"
                  required
                  defaultValue={page.Lease.Rent}
                />
              </label>
              <div className="pair">
                <label>
                  Giá điện (đồng/kWh)
                  <FieldInput
                    name="electricRate"
                    type="number"
                    min="0"
                    max="100000"
                    required
                    defaultValue={page.Lease.ElectricRate}
                  />
                </label>
                <label>
                  Giá nước (đồng/m³)
                  <FieldInput
                    name="waterRate"
                    type="number"
                    min="0"
                    max="1000000"
                    required
                    defaultValue={page.Lease.WaterRate}
                  />
                </label>
              </div>
              <div className="room-toolbar">
                <Action>Lưu điều khoản</Action>
                <ActionLink
                  className="button secondary"
                  href={'/rooms/' + room.ID + '?period=' + page.Period}
                >
                  Hủy
                </ActionLink>
              </div>
            </form>
          </Panel>
        </>
      ) : (
        <>
          <div className="columns">
            <Panel className="card" as="section">
              <h2>Thông tin phòng</h2>
              <dl className="charges">
                <div>
                  <dt>Người đại diện kỳ {page.Period}</dt>
                  <dd>
                    {has(
                      has(activeAt(page.Lease, page.Period)) &&
                        has(page.Lease.Tenant),
                    ) ? (
                      <>{page.Lease.Tenant}</>
                    ) : (
                      <>Chưa có</>
                    )}
                  </dd>
                </div>
                <div>
                  <dt>Kỳ đầu tiên</dt>
                  <dd>{room.Start}</dd>
                </div>
                {has(page.Lease.Months) ? (
                  <>
                    <div>
                      <dt>Thời hạn hợp đồng</dt>
                      <dd>
                        {page.Lease.Months} tháng · đến hết {until(page.Lease)}
                      </dd>
                    </div>
                  </>
                ) : null}
                <div>
                  <dt>Tiền thuê mỗi tháng</dt>
                  <dd>{money(page.Lease.Rent)}</dd>
                </div>
                <div>
                  <dt>Đơn giá điện</dt>
                  <dd>{money(page.Lease.ElectricRate)}/kWh</dd>
                </div>
                <div>
                  <dt>Đơn giá nước</dt>
                  <dd>{money(page.Lease.WaterRate)}/m³</dd>
                </div>
                <div>
                  <dt>Chỉ số điện gốc</dt>
                  <dd>{room.Electric} kWh</dd>
                </div>
                <div>
                  <dt>Chỉ số nước gốc</dt>
                  <dd>{room.Water} m³</dd>
                </div>
              </dl>
            </Panel>
            <Panel className="card" as="section">
              <h2>Bảng kê kỳ {page.Period}</h2>
              {has(page.Invoice) ? (
                <>
                  <p>
                    {!has(page.Invoice.Issued) ? (
                      <>Bản nháp</>
                    ) : has(page.Invoice.Remaining) ? (
                      <>Còn phải thu</>
                    ) : (
                      <>Đã thu đủ</>
                    )}
                  </p>
                  <ActionLink
                    className="button secondary"
                    href={'/invoice/' + page.Invoice.ID}
                  >
                    Xem bảng kê
                  </ActionLink>
                </>
              ) : (
                <>
                  <p className="subtle">Chưa có bảng kê kỳ này.</p>
                  <ActionLink
                    className="button secondary"
                    href={'/?tab=billing&period=' + page.Period}
                  >
                    Mở kỳ thu
                  </ActionLink>
                </>
              )}
            </Panel>
          </div>
          <Panel className="card history" id="room-fees" as="section">
            <h2>Phí riêng của phòng</h2>
            {has(page.RoomFees) ? (
              page.RoomFees.map((fee, index) => (
                <Fragment key={itemKey(fee, index)}>
                  <div className="list-row">
                    <div>
                      <strong>{fee.Label}</strong>
                      <p>
                        {has(fee.Monthly) ? (
                          <>Hằng tháng từ</>
                        ) : (
                          <>Một lần vào kỳ</>
                        )}{' '}
                        {fee.Start}
                        {has(fee.End) ? (
                          <> · Dừng từ {fee.End}</>
                        ) : null}
                      </p>
                    </div>
                    <strong>{money(fee.Amount)}</strong>
                  </div>
                  {!has(fee.End) ? (
                    <>
                      <details>
                        <summary>
                          {has(fee.Monthly) ? (
                            <>Dừng phí {fee.Label}</>
                          ) : (
                            <>Hủy phí {fee.Label}</>
                          )}
                        </summary>
                        <form
                          method="post"
                          action="/action/room-fee-stop"
                          className="form room-form"
                        >
                          <input type="hidden" name="csrf" value={page.CSRF} />
                          <input type="hidden" name="id" value={room.ID} />
                          <input
                            type="hidden"
                            name="period"
                            value={page.Period}
                          />
                          <input type="hidden" name="fee" value={fee.ID} />
                          {has(fee.Monthly) ? (
                            <>
                              <label>
                                Dừng từ kỳ
                                <FieldInput
                                  type="month"
                                  name="end"
                                  min={fee.Start}
                                  required
                                />
                              </label>
                            </>
                          ) : null}
                          <Action className="secondary">
                            {has(fee.Monthly) ? (
                              <>Xác nhận dừng</>
                            ) : (
                              <>Xác nhận hủy</>
                            )}
                          </Action>
                        </form>
                      </details>
                    </>
                  ) : null}
                </Fragment>
              ))
            ) : (
              <>
                <p className="subtle">Phòng chưa có khoản phí riêng.</p>
              </>
            )}
            <hr />
            <details>
              <summary>Thêm khoản phí</summary>
              <form
                method="post"
                action="/action/room-fee"
                className="form room-form"
              >
                <input type="hidden" name="csrf" value={page.CSRF} />
                <input type="hidden" name="id" value={room.ID} />
                <input type="hidden" name="period" value={page.Period} />
                <input type="hidden" name="key" value={randomKey()} />
                <label>
                  Tên khoản phí
                  <FieldInput
                    name="label"
                    placeholder="VD: Wi-Fi"
                    maxLength={40}
                    required
                  />
                </label>
                <label>
                  Số tiền (đồng)
                  <FieldInput
                    name="amount"
                    type="number"
                    inputMode="numeric"
                    min="1"
                    max="100000000"
                    required
                  />
                </label>
                <div className="pair">
                  <label>
                    Chu kỳ thu
                    <select name="kind">
                      <option value="monthly">Hằng tháng</option>
                      <option value="once">Một lần</option>
                    </select>
                  </label>
                  <label>
                    Kỳ bắt đầu
                    <FieldInput
                      name="start"
                      type="month"
                      min={room.Start}
                      required
                      defaultValue={page.Period}
                    />
                  </label>
                </div>
                <Action>Lưu khoản phí</Action>
              </form>
            </details>
          </Panel>
          <Panel className="card history" as="section">
            <div className="room-toolbar">
              <h2>Điều khoản theo kỳ</h2>
              <ActionLink
                className="button secondary"
                href={
                  '/rooms/' + room.ID + '?period=' + page.Period + '&contract=1'
                }
              >
                Thêm điều khoản
              </ActionLink>
            </div>
            {has(page.LeaseHistory)
              ? page.LeaseHistory.map((lease, index) => (
                  <Fragment key={itemKey(lease, index)}>
                    <div className="list-row">
                      <div>
                        <strong>
                          Từ {lease.Effective}
                          {has(lease.Months) ? (
                            <> đến hết {until(lease)}</>
                          ) : null}
                        </strong>
                        {has(lease.Tenant) ? (
                          <>
                            <p>
                              {lease.Tenant}
                              {has(lease.Renewal) ? <> · Ký lại</> : null}
                            </p>
                          </>
                        ) : null}
                      </div>
                      <span>{money(lease.Rent)}/tháng</span>
                    </div>
                  </Fragment>
                ))
              : null}
          </Panel>
        </>
      )}
    </>
  );
}
