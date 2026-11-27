import { Fragment } from 'react';
import {
  FieldSelect,
  Action,
  ActionLink,
  FieldArea,
  FieldInput,
} from './components';
import { RoomPricing } from './room-pricing';
import { has, itemKey, money, usage } from './view';
import type { PageData, Invoice } from './view';

export function CashFields({ page }: { page: PageData }) {
  return (
    <>
      <div className="pair">
        <label>
          Ngày giao dịch
          <FieldInput
            name="date"
            type="date"
            required
            defaultValue={page.EntryDate}
          />
        </label>
        <label>
          Số tiền (đồng)
          <FieldInput
            name="amount"
            type="number"
            inputMode="numeric"
            min="1"
            max="100000000000"
            required
          />
        </label>
      </div>
      <label>
        Mã chứng từ (nếu có)
        <FieldInput name="reference" maxLength={40} />
      </label>
      <label>
        Ghi chú
        <FieldInput name="note" maxLength={80} />
      </label>
    </>
  );
}

export function RoomForm({ page }: { page: PageData }) {
  return (
    <>
      <form
        method="post"
        action={has(page.RoomDetail) ? '/action/room-edit' : '/action/room'}
        className="form"
      >
        <input type="hidden" name="csrf" value={page.CSRF} />
        <input type="hidden" name="period" value={page.Period} />
        {has(page.RoomDetail) ? (
          <>
            <input type="hidden" name="id" value={page.RoomDetail.ID} />
            <input
              type="hidden"
              name="revision"
              value={page.RoomDetail.Revision}
            />
          </>
        ) : null}
        <label>
          Mã phòng
          <FieldInput
            name="name"
            placeholder="VD: P101"
            maxLength={60}
            required
            defaultValue={has(page.RoomDetail) ? page.RoomDetail.Name : ''}
          />
        </label>
        {!!page.Sites?.length && (
          <label>
            Dãy trọ
            <FieldSelect
              name="site"
              defaultValue={page.RoomDetail?.SiteID || page.SelectedSite || ''}
            >
              <option value="">Chưa gán dãy</option>
              {page.Sites.map((site) => (
                <option key={site.ID} value={site.ID}>
                  {site.Name}
                </option>
              ))}
            </FieldSelect>
          </label>
        )}
        <label>
          Mô tả phòng
          <FieldArea
            name="description"
            rows={2}
            maxLength={1000}
            defaultValue={page.RoomDetail?.Description}
          />
        </label>
        {!has(page.RoomDetail) ? (
          <>
            <label>
              Thời hạn hợp đồng (tháng, nếu đã có người thuê)
              <FieldInput
                name="months"
                type="number"
                inputMode="numeric"
                min="1"
                max="120"
              />
            </label>
          </>
        ) : null}
        {!has(page.LockRates) ? (
          <>
            {!page.RoomDetail?.Prepared && (
              <label>
                Người đại diện <span className="subtle">(không bắt buộc)</span>
                <FieldInput
                  name="tenant"
                  maxLength={150}
                  defaultValue={
                    has(page.RoomDetail) ? page.RoomDetail.Tenant : ''
                  }
                />
              </label>
            )}
            <RoomPricing
              room={page.RoomDetail}
              policies={
                !page.RoomDetail || page.PolicyMutable ? page.Policies : null
              }
            />
          </>
        ) : null}
        {!has(page.LockOpening) ? (
          <>
            <div className="pair">
              <label>
                Chỉ số điện đầu kỳ
                <FieldInput
                  name="electric"
                  type="number"
                  inputMode="numeric"
                  min="0"
                  max="999999"
                  required
                  defaultValue={
                    has(page.RoomDetail) ? page.RoomDetail.Electric : '0'
                  }
                />
              </label>
              <label>
                Chỉ số nước đầu kỳ
                <FieldInput
                  name="water"
                  type="number"
                  inputMode="numeric"
                  min="0"
                  max="999999"
                  required
                  defaultValue={
                    has(page.RoomDetail) ? page.RoomDetail.Water : '0'
                  }
                />
              </label>
            </div>
          </>
        ) : null}
        <div className="room-toolbar">
          <Action>
            {has(page.RoomDetail) ? <>Lưu thay đổi</> : <>Lưu phòng</>}
          </Action>
          <ActionLink
            className="button secondary"
            href={
              has(page.RoomDetail)
                ? '/rooms/' + page.RoomDetail.ID + '?period=' + page.Period
                : '/?tab=rooms&period=' + page.Period
            }
          >
            Hủy
          </ActionLink>
        </div>
      </form>
    </>
  );
}

export function InvoiceDetails({ invoice }: { invoice: Invoice }) {
  return (
    <>
      <h2>Chi tiết khoản thu</h2>
      <dl className="charges">
        <div>
          <dt>Tiền thuê</dt>
          <dd>{money(invoice.Rent)}</dd>
        </div>
        <div>
          <dt>
            Điện{' '}
            <small>
              {has(invoice.Reading.Adjustment.Electric.Replaced) ? (
                <>
                  Cũ {invoice.Reading.ElectricBefore} →{' '}
                  {invoice.Reading.Adjustment.Electric.OldEnd} · Mới{' '}
                  {invoice.Reading.Adjustment.Electric.NewStart} →{' '}
                  {invoice.Reading.ElectricAfter}
                </>
              ) : (
                <>
                  {has(invoice.Reading.Adjustment.Handover) ? (
                    <>Bàn giao </>
                  ) : null}
                  {invoice.Reading.ElectricBefore} →{' '}
                  {invoice.Reading.ElectricAfter}
                </>
              )}{' '}
              · {usage(invoice.Reading, 'Electric')} kWh ×{' '}
              {money(invoice.ElectricRate)}/kWh
            </small>
          </dt>
          <dd>{money(invoice.ElectricAmount)}</dd>
        </div>
        <div>
          <dt>
            Nước{' '}
            <small>
              {has(invoice.Reading.Adjustment.Water.Replaced) ? (
                <>
                  Cũ {invoice.Reading.WaterBefore} →{' '}
                  {invoice.Reading.Adjustment.Water.OldEnd} · Mới{' '}
                  {invoice.Reading.Adjustment.Water.NewStart} →{' '}
                  {invoice.Reading.WaterAfter}
                </>
              ) : (
                <>
                  {has(invoice.Reading.Adjustment.Handover) ? (
                    <>Bàn giao </>
                  ) : null}
                  {invoice.Reading.WaterBefore} → {invoice.Reading.WaterAfter}
                </>
              )}{' '}
              · {usage(invoice.Reading, 'Water')} m³ ×{' '}
              {money(invoice.WaterRate)}/m³
            </small>
          </dt>
          <dd>{money(invoice.WaterAmount)}</dd>
        </div>
        {has(invoice.CommonFee) ? (
          <>
            <div>
              <dt>{invoice.CommonFeeLabel}</dt>
              <dd>{money(invoice.CommonFee)}</dd>
            </div>
          </>
        ) : null}
        {has(invoice.RoomFees)
          ? invoice.RoomFees.map((fee, index) => (
              <Fragment key={itemKey(fee, index)}>
                <div>
                  <dt>{fee.Label}</dt>
                  <dd>{money(fee.Amount)}</dd>
                </div>
              </Fragment>
            ))
          : null}
        <div>
          <dt>Tổng bảng kê</dt>
          <dd>{money(invoice.Total)}</dd>
        </div>
        {has(invoice.NetReduction) ? (
          <>
            <div>
              <dt>Giảm trừ</dt>
              <dd>{money(invoice.NetReduction)}</dd>
            </div>
          </>
        ) : null}
        {has(invoice.NetCollected) ? (
          <>
            <div>
              <dt>Đã thu ròng</dt>
              <dd>{money(invoice.NetCollected)}</dd>
            </div>
          </>
        ) : null}
        {has(invoice.DueDate) ? (
          <>
            <div>
              <dt>Hạn thanh toán</dt>
              <dd>
                <time dateTime={invoice.DueDate}>{invoice.DueDate}</time>
              </dd>
            </div>
          </>
        ) : null}
        {has(invoice.Issued) ? (
          <>
            <div className="total">
              <dt>Còn phải trả</dt>
              <dd>{money(invoice.Remaining)}</dd>
            </div>
            {has(invoice.Credit) ? (
              <>
                <div>
                  <dt>Thu dư</dt>
                  <dd>{money(invoice.Credit)}</dd>
                </div>
              </>
            ) : null}
          </>
        ) : null}
      </dl>
    </>
  );
}
