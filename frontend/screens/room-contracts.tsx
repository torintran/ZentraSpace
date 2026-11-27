import { Download, FilePlus2, UserRoundPlus } from 'lucide-react';
import { Action, ActionLink, FieldInput, Panel } from './components';
import { activeAt, money, until } from './view';
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
        <h2>{page.RoomArchive ? 'Danh sách hợp đồng' : 'Hợp đồng trong kỳ'}</h2>
        {!page.RoomArchive && (
          <ActionLink
            className="button secondary"
            href={'/rooms/' + roomId + '?period=' + page.Period + '&contract=1'}
          >
            <UserRoundPlus size={18} aria-hidden />{' '}
            {page.LeaseCount ? 'Thêm hợp đồng' : 'Tiếp nhận người thuê'}
          </ActionLink>
        )}
      </div>
      {!page.RoomArchive &&
        page.Lease.Tenant &&
        !page.Lease.EndedOn &&
        activeAt(page.Lease, page.Period) && (
          <div className="actions">
            <ActionLink
              className="secondary"
              href={
                '/rooms/' + roomId + '?period=' + page.Period + '&checkout=1'
              }
            >
              Trả phòng và tất toán
            </ActionLink>
          </div>
        )}
      {history.length === 0 && (
        <p className="subtle">
          {page.RoomArchive
            ? 'Chưa có hợp đồng phù hợp.'
            : 'Kỳ này chưa có hợp đồng đang sử dụng.'}
        </p>
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
                {lease.EndedOn && (
                  <p>
                    Đã trả phòng ngày{' '}
                    <time dateTime={lease.EndedOn}>{lease.EndedOn}</time>
                  </p>
                )}
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
                  {page.RoomArchive && (
                    <input type="hidden" name="archive" value="1" />
                  )}
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
            {lease.Tenant && (
              <details>
                <summary>Email xác minh bảng kê</summary>
                <form
                  className="form"
                  method="post"
                  action="/action/lease-email"
                >
                  <input type="hidden" name="csrf" value={page.CSRF} />
                  {page.RoomArchive && (
                    <input type="hidden" name="archive" value="1" />
                  )}
                  <input type="hidden" name="period" value={page.Period} />
                  <input type="hidden" name="id" value={roomId} />
                  <input
                    type="hidden"
                    name="effective"
                    value={lease.Effective}
                  />
                  <label>
                    Email của người nhận
                    <FieldInput
                      type="email"
                      name="email"
                      maxLength={254}
                      defaultValue={page.LeaseEmails?.[lease.Effective] ?? ''}
                    />
                  </label>
                  <p className="subtle">
                    Để trống để bỏ email. Khi đổi email, link bảng kê của hợp
                    đồng này được thu hồi.
                  </p>
                  <Action className="secondary">Lưu email</Action>
                </form>
              </details>
            )}
          </article>
        );
      })}
      {!page.RoomArchive && page.LeaseCount > 0 && (
        <div className="actions">
          <ActionLink
            className="secondary"
            href={
              '/rooms/' +
              roomId +
              '?period=' +
              page.Period +
              '&history=1&kind=contracts'
            }
          >
            Xem tất cả hợp đồng
          </ActionLink>
        </div>
      )}
    </Panel>
  );
}
