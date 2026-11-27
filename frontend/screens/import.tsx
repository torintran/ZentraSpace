import { Action, FieldArea, Panel } from './components';

import type { PageData, Invoice } from './view';

export function ImportScreen({ page }: { page: PageData }) {
  const preview = page.Import;
  if (!preview) return null;
  return (
    <>
      <Panel className="card auth" as="section">
        <h1>Kiểm tra danh sách</h1>
        <p>
          Kỳ đầu {preview.Period}. Thêm {preview.Add} phòng mới. Bỏ qua{' '}
          {preview.Skip} dòng đã có.
        </p>
        <p>Dãy trọ {preview.SiteName || 'chưa gán'}</p>
        <form method="post" action="/action/import" className="form">
          <input type="hidden" name="csrf" value={page.CSRF} />
          <input type="hidden" name="period" value={preview.Period} />
          <input type="hidden" name="site" value={preview.SiteID} />
          <label>
            Danh sách đã kiểm tra
            <FieldArea
              name="csv"
              rows={6}
              readOnly
              defaultValue={preview.Data}
            ></FieldArea>
          </label>
          <Action>Lưu danh sách phòng</Action>
        </form>
        <a className="back" href={'/?tab=rooms&period=' + preview.Period}>
          Về danh sách phòng
        </a>
      </Panel>
    </>
  );
}
