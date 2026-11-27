import {
  Action,
  ActionLink,
  FieldArea,
  FieldInput,
  Notice,
  Panel,
} from './components';
import { money } from './view';
import type { PageData } from './view';

export function RoomCheckout({
  page,
  roomID,
}: {
  page: PageData;
  roomID: string;
}) {
  const roomURL = '/rooms/' + roomID;
  const quote = page.CheckoutQuote;
  return (
    <Panel className="card room-form">
      <h2>Trả phòng và tất toán</h2>
      <form method="get" action={roomURL} className="form">
        <input type="hidden" name="period" value={page.Period} />
        <input type="hidden" name="checkout" value="1" />
        <label>
          Ngày trả phòng
          <FieldInput
            type="date"
            name="date"
            defaultValue={page.CheckoutDate}
            max={page.Today}
            required
          />
        </label>
        <div className="actions">
          <Action>Xem trước tất toán</Action>
          <ActionLink
            className="secondary"
            href={roomURL + '?period=' + page.Period}
          >
            Hủy
          </ActionLink>
        </div>
      </form>
      {page.CheckoutError && <Notice role="alert">{page.CheckoutError}</Notice>}
      {quote && (
        <>
          <p>
            {quote.Invoice.Issued
              ? 'Bảng kê cuối đã phát hành. Số tiền được giữ nguyên.'
              : 'Xác nhận sẽ phát hành bảng kê cuối. Tiền thuê tính đến hết ngày trả và các khoản phí khác theo cấu hình kỳ.'}
          </p>
          <dl className="charges">
            <div>
              <dt>Chỉ số điện khi bàn giao</dt>
              <dd>{quote.Invoice.Reading.ElectricAfter}</dd>
            </div>
            <div>
              <dt>Chỉ số nước khi bàn giao</dt>
              <dd>{quote.Invoice.Reading.WaterAfter}</dd>
            </div>
            <div>
              <dt>Tiền thuê kỳ cuối</dt>
              <dd>{money(quote.Invoice.Rent)}</dd>
            </div>
            <div>
              <dt>Còn phải thu kỳ cuối</dt>
              <dd>{money(quote.Remaining)}</dd>
            </div>
            <div>
              <dt>Công nợ các kỳ trước</dt>
              <dd>{money(quote.EarlierDue)}</dd>
            </div>
            <div>
              <dt>Cọc đang giữ</dt>
              <dd>{money(quote.Held)}</dd>
            </div>
          </dl>
          <p className="subtle">
            Phòng chuyển sang trạng thái trống ngay sau xác nhận. Hợp đồng mới
            có thể bắt đầu từ kỳ sau. Công nợ chưa thanh toán và cọc còn dư vẫn
            được giữ trong sổ.
          </p>
          <form className="form" method="post" action="/action/checkout">
            <input type="hidden" name="csrf" value={page.CSRF} />
            <input type="hidden" name="period" value={page.Period} />
            <input type="hidden" name="id" value={roomID} />
            <input type="hidden" name="date" value={quote.Date} />
            <input type="hidden" name="effective" value={quote.Effective} />
            <input type="hidden" name="revision" value={quote.Revision} />
            <label>
              Cấn cọc vào tiền kỳ cuối (đ)
              <FieldInput
                name="applied"
                type="number"
                inputMode="numeric"
                min={0}
                max={Math.min(quote.Held, quote.Remaining)}
                defaultValue={0}
                required
              />
            </label>
            <label>
              Hoàn cọc thực tế (đ)
              <FieldInput
                name="refunded"
                type="number"
                inputMode="numeric"
                min={0}
                max={quote.Held}
                defaultValue={0}
                required
              />
            </label>
            <label>
              Ghi chú bàn giao
              <FieldArea name="note" maxLength={160} required />
            </label>
            <label className="check">
              <input type="checkbox" name="confirm" value="yes" required /> Tôi
              đã kiểm tra chỉ số và các khoản tiền.
            </label>
            <Action>Xác nhận trả phòng</Action>
          </form>
        </>
      )}
    </Panel>
  );
}
