import { Action, FieldInput, FieldSelect, Panel } from './components';
import { clock, money, randomKey } from './view';
import type { PageData } from './view';

export function InvoiceCredit({ page }: { page: PageData }) {
  const bill = page.Invoice;
  if (!bill || (!page.AvailableCredit && !page.CreditTransfers?.length))
    return null;
  return (
    <Panel className="card history" id="invoice-credit">
      <h2>Tiền thu dư</h2>
      <p>
        Có thể chuyển <strong>{money(page.AvailableCredit)}</strong> sang kỳ sau
        của cùng hợp đồng.
      </p>
      {page.AvailableCredit > 0 &&
        (page.CreditTargets?.length ? (
          <form className="form" method="post" action="/action/credit-transfer">
            <input type="hidden" name="csrf" value={page.CSRF} />
            <input type="hidden" name="id" value={bill.ID} />
            <input type="hidden" name="key" value={randomKey()} />
            <label>
              Bảng kê nhận tiền dư
              <FieldSelect name="target" required>
                <option value="">Chọn kỳ nhận</option>
                {page.CreditTargets.map((target) => (
                  <option key={target.ID} value={target.ID}>
                    {target.Period} · Còn {money(target.Remaining)}
                  </option>
                ))}
              </FieldSelect>
            </label>
            <label>
              Số tiền chuyển (đ)
              <FieldInput
                type="number"
                name="amount"
                min={1}
                max={page.AvailableCredit}
                inputMode="numeric"
                required
              />
            </label>
            <label>
              Ghi chú (tùy chọn)
              <FieldInput name="note" maxLength={160} />
            </label>
            <Action>Chuyển tiền dư</Action>
          </form>
        ) : (
          <p className="subtle">
            Tiền dư được giữ lại cho đến khi có bảng kê kỳ sau còn nợ.
          </p>
        ))}
      {page.CreditTransfers?.map((transfer) => (
        <article className="list-row" key={transfer.Key}>
          <div>
            <a href={'/invoice/' + transfer.TargetID}>
              Chuyển sang kỳ {transfer.TargetPeriod}
            </a>
            <p>
              {money(transfer.Amount)} · {clock(transfer.Date)}
              {transfer.Reversed ? ' · Đã đảo' : ''}
            </p>
            {transfer.Note && <p>{transfer.Note}</p>}
          </div>
          {!transfer.Reversed && (
            <details>
              <summary>Đảo chuyển tiền dư</summary>
              <form
                className="form"
                method="post"
                action="/action/credit-reverse"
              >
                <input type="hidden" name="csrf" value={page.CSRF} />
                <input type="hidden" name="id" value={bill.ID} />
                <input type="hidden" name="key" value={randomKey()} />
                <input type="hidden" name="target" value={transfer.TargetID} />
                <input type="hidden" name="original" value={transfer.Key} />
                <input type="hidden" name="amount" value={transfer.Amount} />
                <label>
                  Lý do đảo bút toán
                  <FieldInput name="note" maxLength={160} required />
                </label>
                <Action className="secondary">Xác nhận đảo</Action>
              </form>
            </details>
          )}
        </article>
      ))}
    </Panel>
  );
}
