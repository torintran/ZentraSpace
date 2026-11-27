import { ActionLink, Panel } from './components';
import { configured, has, money } from './view';
import type { PageData, Invoice } from './view';
import { InvoiceDetails } from './common';

export function SharedScreen({ page }: { page: PageData }) {
  const bill = page.Invoice;
  if (!bill) return null;
  return (
    <>
      <header className="tenant-summary">
        <p className="eyebrow">
          PHÒNG {bill.RoomName} · KỲ {bill.Period}
        </p>
        <h1>Bảng kê tiền thuê</h1>
        <p>{has(bill.Remaining) ? <>Còn phải trả</> : <>Đã thanh toán đủ</>}</p>
        <strong className="tenant-balance">{money(bill.Remaining)}</strong>
        {has(bill.DueDate) ? (
          <>
            <p>
              Hạn thanh toán <time dateTime={bill.DueDate}>{bill.DueDate}</time>
            </p>
          </>
        ) : null}
        {has(configured(page.PaymentBank)) ? (
          <>
            <ActionLink className="button" href="#payment">
              Xem mã thanh toán
            </ActionLink>
          </>
        ) : null}
      </header>
      <div
        className={
          'tenant-columns' +
          (has(!has(configured(page.PaymentBank)))
            ? ' tenant-columns-single'
            : '')
        }
      >
        <Panel className="card" as="section">
          <InvoiceDetails invoice={bill} />
        </Panel>
        {has(configured(page.PaymentBank)) ? (
          <>
            <Panel className="card" id="payment" as="section">
              <h2>Chuyển khoản bằng VietQR</h2>
              <p>
                Kiểm tra tên người nhận và số tiền trên ứng dụng ngân hàng trước
                khi chuyển.
              </p>
              {page.PaymentBank.Account === '0000000000000' ? (
                <>
                  <p className="notice">
                    Đây là tài khoản mô phỏng. Dùng dịch vụ ngân hàng mock để
                    tạo giao dịch thử.
                  </p>
                </>
              ) : null}
              <img
                className="payment-qr"
                src={'/p/' + page.InviteID + '/qr'}
                alt="Mã VietQR cho số tiền còn phải trả"
                width="280"
                height="280"
              />
              <dl className="charges">
                <div>
                  <dt>Chủ tài khoản</dt>
                  <dd>{page.PaymentBank.Holder}</dd>
                </div>
                <div>
                  <dt>Số tài khoản</dt>
                  <dd>{page.PaymentBank.Account}</dd>
                </div>
                <div>
                  <dt>Nội dung chuyển khoản</dt>
                  <dd>{bill.PaymentCode}</dd>
                </div>
              </dl>
            </Panel>
          </>
        ) : null}
      </div>
    </>
  );
}
