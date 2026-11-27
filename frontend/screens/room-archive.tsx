import {
  Action,
  ActionLink,
  FieldInput,
  FieldSelect,
  Panel,
} from './components';
import { RoomContracts } from './room-contracts';
import { RoomTimeline } from './room-records';
import type { PageData } from './view';

export function RoomArchive({ page }: { page: PageData }) {
  const room = page.RoomDetail;
  if (!room) return null;
  const path = '/rooms/' + room.ID;
  const summary = path + '?period=' + page.Period;
  const query = new URLSearchParams({ period: page.Period, history: '1' });
  if (page.ArchiveKind) query.set('kind', page.ArchiveKind);
  if (page.ArchivePeriod) query.set('archivePeriod', page.ArchivePeriod);
  const destination = (number: number) =>
    path + '?' + new URLSearchParams([...query, ['page', String(number)]]);
  return (
    <>
      <a className="back" href={summary}>
        ← Về hồ sơ phòng
      </a>
      <div className="page-title">
        <h1>Lịch sử phòng · {room.Name}</h1>
      </div>
      <Panel className="card history">
        <form method="get" action={path} className="period">
          <input type="hidden" name="history" value="1" />
          <input type="hidden" name="period" value={page.Period} />
          <label>
            Nội dung
            <FieldSelect name="kind" defaultValue={page.ArchiveKind}>
              <option value="">Toàn bộ hoạt động</option>
              <option value="contracts">Hồ sơ hợp đồng</option>
              <option value="lease">Bắt đầu hợp đồng</option>
              <option value="bill">Phát hành bảng kê</option>
              <option value="payment">Thanh toán và điều chỉnh</option>
              <option value="deposit">Tiền cọc</option>
              <option value="resident">Người ở</option>
              <option value="policy">Mẫu giá</option>
              <option value="checkout">Trả phòng</option>
            </FieldSelect>
          </label>
          <label>
            Tháng (tùy chọn)
            <FieldInput
              name="archivePeriod"
              type="month"
              defaultValue={page.ArchivePeriod}
            />
          </label>
          <Action className="secondary">Lọc lịch sử</Action>
        </form>
      </Panel>
      {page.ArchiveKind === 'contracts' ? (
        <RoomContracts page={page} roomId={room.ID} />
      ) : (
        <RoomTimeline page={page} />
      )}
      {page.ArchivePages > 1 && (
        <nav className="actions" aria-label="Phân trang lịch sử phòng">
          {page.ArchivePage > 1 && (
            <ActionLink
              className="secondary"
              href={destination(page.ArchivePage - 1)}
            >
              Trang trước
            </ActionLink>
          )}
          <span>
            Trang {page.ArchivePage}/{page.ArchivePages}
          </span>
          {page.ArchivePage < page.ArchivePages && (
            <ActionLink
              className="secondary"
              href={destination(page.ArchivePage + 1)}
            >
              Trang sau
            </ActionLink>
          )}
        </nav>
      )}
    </>
  );
}
