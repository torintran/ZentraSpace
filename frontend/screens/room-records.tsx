import { CalendarClock, UserRound } from 'lucide-react';
import { Action, ActionLink, FieldInput, Panel } from './components';
import { money } from './view';
import type { PageData } from './view';

export function PersonIcons({ count }: { count: number }) {
  return (
    <span className="person-count" role="img" aria-label={count + ' người ở'}>
      {Array.from({ length: Math.min(6, count) }, (_, index) => (
        <span className="person-avatar" key={index}>
          <UserRound size={15} aria-hidden />
        </span>
      ))}
      {count > 6 && <span aria-hidden>+{count - 6}</span>}
    </span>
  );
}

export function RoomPhoto({
  page,
  roomID,
}: {
  page: PageData;
  roomID: string;
}) {
  const url = '/rooms/' + roomID + '/photo';
  return (
    <>
      {page.HasPhoto && (
        <img
          className="room-profile-photo"
          src={url}
          alt={'Ảnh phòng ' + page.RoomDetail?.Name}
          width={600}
          height={360}
        />
      )}
      <details>
        <summary>{page.HasPhoto ? 'Đổi ảnh phòng' : 'Thêm ảnh phòng'}</summary>
        <form
          className="form"
          method="post"
          action={url}
          encType="multipart/form-data"
        >
          <input type="hidden" name="csrf" value={page.CSRF} />
          <input type="hidden" name="period" value={page.Period} />
          <label>
            Ảnh JPG hoặc PNG, tối đa 2 MB
            <input
              type="file"
              name="photo"
              accept="image/jpeg,image/png"
              required
            />
          </label>
          <Action>Lưu ảnh phòng</Action>
        </form>
      </details>
    </>
  );
}

export function RoomResidents({
  page,
  roomID,
  occupied,
}: {
  page: PageData;
  roomID: string;
  occupied: boolean;
}) {
  const residents = page.Residents ?? [];
  return (
    <Panel className="card history" id="room-residents">
      <div className="room-toolbar">
        <h2>Người ở trong kỳ</h2>
        {residents.length > 0 && <PersonIcons count={residents.length} />}
      </div>
      {!occupied && (
        <p className="subtle">
          Khi có hợp đồng hiệu lực, bạn có thể thêm người ở cùng.
        </p>
      )}
      {residents.map((person) => (
        <article className="list-row" key={person.ID || 'representative'}>
          <div>
            <strong>{person.Name}</strong>
            <p>
              {person.ID
                ? 'Vào ở từ ' + person.JoinedOn
                : 'Người đại diện hợp đồng'}
            </p>
            {person.Phone && <p>{person.Phone}</p>}
            {person.LeftOn && <p>Rời phòng {person.LeftOn}</p>}
          </div>
          {person.ID && !person.LeftOn && (
            <details>
              <summary>Ghi ngày rời phòng</summary>
              <form
                className="form"
                method="post"
                action="/action/resident-out"
              >
                <input type="hidden" name="id" value={roomID} />
                <input type="hidden" name="resident" value={person.ID} />
                <input type="hidden" name="csrf" value={page.CSRF} />
                <input type="hidden" name="period" value={page.Period} />
                <label>
                  Ngày rời phòng
                  <FieldInput
                    name="leftOn"
                    type="date"
                    min={person.JoinedOn}
                    defaultValue={
                      page.PeriodDate >= person.JoinedOn
                        ? page.PeriodDate
                        : person.JoinedOn
                    }
                    required
                  />
                </label>
                <Action className="secondary">Lưu ngày rời</Action>
              </form>
            </details>
          )}
        </article>
      ))}
      {occupied && (
        <details>
          <summary>Thêm người ở cùng</summary>
          <form className="form" method="post" action="/action/resident">
            <input type="hidden" name="id" value={roomID} />
            <input type="hidden" name="csrf" value={page.CSRF} />
            <input type="hidden" name="period" value={page.Period} />
            <label>
              Họ và tên
              <FieldInput name="name" maxLength={150} required />
            </label>
            <label>
              Điện thoại (tuỳ chọn)
              <FieldInput name="phone" type="tel" maxLength={30} />
            </label>
            <label>
              Ngày vào ở
              <FieldInput
                name="joinedOn"
                type="date"
                min={page.Period + '-01'}
                max={page.PeriodEnd}
                defaultValue={page.PeriodDate}
                required
              />
            </label>
            <Action>Lưu người ở</Action>
          </form>
        </details>
      )}
    </Panel>
  );
}

export function RoomTimeline({ page }: { page: PageData }) {
  const events = page.RoomTimeline ?? [];
  return (
    <Panel className="card history" id="room-timeline">
      <h2>
        <CalendarClock size={20} aria-hidden />{' '}
        {page.RoomArchive ? 'Nhật ký phòng' : 'Hoạt động trong kỳ'}
      </h2>
      {!events.length && <p className="subtle">Chưa có hoạt động phù hợp.</p>}
      <ol className="room-timeline">
        {events.map((event, index) => (
          <li key={index}>
            <time dateTime={event.Date}>{event.Date}</time>
            <div>
              <strong>{event.Title}</strong>
              {event.Detail && <p>{event.Detail}</p>}
              {!!event.Amount && <span>{money(event.Amount)}</span>}
              {event.InvoiceID && (
                <p>
                  <a href={'/invoice/' + event.InvoiceID}>Xem bảng kê</a>
                </p>
              )}
            </div>
          </li>
        ))}
      </ol>
      {!page.RoomArchive && (
        <ActionLink
          className="secondary"
          href={
            '/rooms/' +
            page.RoomDetail?.ID +
            '?period=' +
            page.Period +
            '&history=1'
          }
        >
          Mở lịch sử phòng
        </ActionLink>
      )}
    </Panel>
  );
}
