import { Download, FilePlus2, UserRoundPlus } from 'lucide-react';
import { Action, ActionLink, Panel } from './components';
import { money, until } from './view';
import type { PageData } from './view';

export function RoomContracts({
  page,
  roomId,
}: {
  page: PageData;
  roomId: string;
}) {
  const history = page.LeaseHistory ?? [];
  return (
    <Panel className="card history" id="room-contracts">
      <div className="room-toolbar">
        <h2>Hợp đồng và tệp đính kèm</h2>
        <ActionLink
          className="button secondary"
          href={'/rooms/' + roomId + '?period=' + page.Period + '&contract=1'}
        >
          <UserRoundPlus size={18} aria-hidden />{' '}
          {history.length ? 'Thêm hợp đồng' : 'Tiếp nhận người thuê'}
        </ActionLink>
      </div>
      {history.length === 0 && (
        <p className="subtle">Chưa có hồ sơ hợp đồng cho phòng này.</p>
      )}
      {history.map((lease) => {
        const document =
          '/rooms/' + roomId + '/contracts/' + lease.Effective + '/file';
        const end = until(lease);
        return (
          <article className="contract-row" key={lease.Effective}>
            <div className="list-row">
              <div>
                <strong>
                  {'Từ ' + lease.Effective + (end ? ' đến hết ' + end : '')}
                </strong>
                {lease.Tenant && (
                  <p>{lease.Tenant + (lease.Renewal ? ' · Ký lại' : '')}</p>
                )}
                {lease.Phone && <p>Điện thoại {lease.Phone}</p>}
                {lease.SignedOn && (
                  <p>
                    Ngày ký{' '}
                    <time dateTime={lease.SignedOn}>{lease.SignedOn}</time>
                  </p>
                )}
              </div>
              <span>{money(lease.Rent)}/tháng</span>
            </div>
            {page.LeaseDocuments?.[lease.Effective] ? (
              <ActionLink className="button secondary" href={document}>
                <Download size={18} aria-hidden /> Tải hợp đồng
              </ActionLink>
            ) : (
              <details>
                <summary>
                  <FilePlus2 size={18} aria-hidden /> Đính kèm tệp hợp đồng
                </summary>
                <form
                  method="post"
                  action={document}
                  encType="multipart/form-data"
                  className="form"
                >
                  <input type="hidden" name="csrf" value={page.CSRF} />
                  <label>
                    Tệp hợp đồng (PDF, JPG hoặc PNG, tối đa 8 MB)
                    <input
                      type="file"
                      name="document"
                      accept=".pdf,.jpg,.jpeg,.png"
                      required
                    />
                  </label>
                  <Action>Lưu tệp hợp đồng</Action>
                </form>
              </details>
            )}
          </article>
        );
      })}
    </Panel>
  );
}
