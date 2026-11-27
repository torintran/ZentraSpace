import { Action, ActionLink, FieldArea, Panel, Status } from './components';
import { has, until } from './view';
import type { PageData } from './view';
import { RoomForm } from './common';
import { RoomList } from './browser-controls';

export function RoomsPanel({ page }: { page: PageData }) {
  return (
    <>
      {has(page.RoomForm) ? (
        <>
          <a className="back" href={'/?tab=rooms&period=' + page.Period}>
            ← Về danh sách phòng
          </a>
          <Panel className="card room-form" as="section">
            <h2>Thêm phòng</h2>
            <p className="subtle">
              Chuẩn bị phòng từ kỳ <strong>{page.Period}</strong>. Bạn có thể
              lưu giá phòng ngay và để trống tên người thuê.
            </p>
            <RoomForm page={page} />
          </Panel>
        </>
      ) : (
        <>
          <Panel className="card" as="section">
            <div className="room-toolbar">
              <h2>Danh sách phòng</h2>
              <ActionLink
                className="button"
                href={'/?tab=rooms&period=' + page.Period + '&add=1'}
              >
                Thêm phòng
              </ActionLink>
            </div>
            {has(page.Rooms) ? (
              <RoomList
                rooms={page.Rooms.map((room) => ({
                  id: room.ID,
                  name: room.Name,
                  row: (
                    <article className="list-row">
                      <div>
                        <strong>{room.Name}</strong>
                        {!room.HasLease && (
                          <Status className="badge">Phòng trống</Status>
                        )}
                        {has(room.Renewed) ? (
                          <>
                            {' '}
                            <Status className="badge good">Đã ký lại</Status>
                          </>
                        ) : null}
                        {has(room.Tenant) ? (
                          <>
                            <p>{room.Tenant}</p>
                          </>
                        ) : null}
                        <p>
                          {has(room.HasLease) ? (
                            <>
                              {has(room.Lease.Months) ? (
                                <>
                                  Hợp đồng {room.Lease.Months} tháng · đến hết{' '}
                                  {until(room.Lease)}
                                </>
                              ) : (
                                <>Từ {room.Lease.Effective}</>
                              )}
                            </>
                          ) : (
                            <>Sẵn sàng tiếp nhận người thuê</>
                          )}
                        </p>
                      </div>
                      <ActionLink
                        className="button secondary"
                        href={'/rooms/' + room.ID + '?period=' + page.Period}
                        aria-label={'Xem chi tiết phòng ' + room.Name}
                      >
                        Xem chi tiết
                      </ActionLink>
                    </article>
                  ),
                }))}
              />
            ) : (
              <>
                <p className="subtle">
                  Chưa có phòng. Hãy thêm phòng đầu tiên.
                </p>
              </>
            )}
          </Panel>
          <details className="card room-import">
            <summary>Nhập danh sách CSV</summary>
            <p className="subtle">
              Điền thông tin vào mẫu, rồi dán nội dung CSV vào ô bên dưới. Kỳ
              đầu của các phòng được nhập là {page.Period}. Dòng bỏ trống người
              đại diện sẽ tạo phòng trống.
            </p>
            <p>
              <ActionLink
                className="button secondary"
                href="/rooms/template.csv"
              >
                Tải mẫu CSV
              </ActionLink>
            </p>
            <form
              method="post"
              action="/action/import-preview"
              className="form"
            >
              <input type="hidden" name="csrf" value={page.CSRF} />
              <input type="hidden" name="period" value={page.Period} />
              <label>
                Nội dung CSV
                <FieldArea
                  name="csv"
                  rows={4}
                  spellCheck={false}
                  required
                  defaultValue="phong,nguoi_dai_dien,tien_thue,gia_dien,gia_nuoc,dien_dau,nuoc_dau"
                ></FieldArea>
              </label>
              <Action className="secondary">Kiểm tra danh sách</Action>
            </form>
          </details>
        </>
      )}
    </>
  );
}
