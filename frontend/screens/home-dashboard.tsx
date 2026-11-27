import { Fragment } from 'react';
import { Panel, Status } from './components';
import { has, itemKey, money } from './view';
import type { PageData, Invoice } from './view';

export function DashboardPanel({ page }: { page: PageData }) {
  return (
    <>
      <Panel className="card dashboard-overview" as="section">
        <h2>Việc cần làm</h2>
        <div className="dashboard-tasks">
          <a
            className="status-missing"
            href={'/?tab=meters&period=' + page.Period + '&filter=missing'}
          >
            <strong>{page.Dashboard.Missing}</strong>
            <span>Phòng thiếu chỉ số</span>
          </a>
          <a
            className="status-draft"
            href={'/?tab=billing&period=' + page.Period + '&filter=draft'}
          >
            <strong>{page.Dashboard.Drafts}</strong>
            <span>Bảng kê chờ phát hành</span>
          </a>
          <a
            className="status-due"
            href={'/?tab=billing&period=' + page.Period + '&filter=due'}
          >
            <strong>{page.Dashboard.Outstanding}</strong>
            <span>Bảng kê còn phải thu</span>
          </a>
        </div>
      </Panel>
      <div className="dashboard-workspace">
        <Panel className="card" aria-labelledby="room-map-heading" as="section">
          <h2 id="room-map-heading">Sơ đồ trạng thái phòng</h2>
          <p className="subtle">
            {page.Dashboard.Rooms} phòng trong kỳ {page.Period}
          </p>
          {has(page.Dashboard.RoomStatus) ? (
            <>
              <ul className="room-map">
                {has(page.Dashboard.RoomStatus)
                  ? page.Dashboard.RoomStatus.map((room, index) => (
                      <Fragment key={itemKey(room, index)}>
                        <li>
                          <a
                            className={'room-tile status-' + room.Status}
                            href={
                              has(room.InvoiceID)
                                ? '/invoice/' + room.InvoiceID
                                : room.Status === 'missing'
                                  ? '/?tab=meters&period=' +
                                    page.Period +
                                    '&filter=missing'
                                  : '/?tab=billing&period=' + page.Period
                            }
                          >
                            <strong>{room.Name}</strong>
                            <span>{room.Label}</span>
                          </a>
                        </li>
                      </Fragment>
                    ))
                  : null}
              </ul>
            </>
          ) : (
            <>
              <p className="empty">Chưa có phòng trong kỳ này.</p>
            </>
          )}
        </Panel>
        <Panel className="card" aria-labelledby="progress-heading" as="section">
          <h2 id="progress-heading">Tiến độ chốt kỳ</h2>
          {has(page.Dashboard.Rooms) ? (
            <>
              <div className="stage-progress">
                <div>
                  <label htmlFor="reading-progress">
                    Đã nhập chỉ số{' '}
                    <strong>
                      {page.Dashboard.Rooms - page.Dashboard.Missing} /{' '}
                      {page.Dashboard.Rooms}
                    </strong>
                  </label>
                  <progress
                    id="reading-progress"
                    value={page.Dashboard.Rooms - page.Dashboard.Missing}
                    max={page.Dashboard.Rooms}
                  ></progress>
                </div>
                <div>
                  <label htmlFor="issue-progress">
                    Đã phát hành{' '}
                    <strong>
                      {page.Dashboard.Issued} / {page.Dashboard.Rooms}
                    </strong>
                  </label>
                  <progress
                    id="issue-progress"
                    value={page.Dashboard.Issued}
                    max={page.Dashboard.Rooms}
                  ></progress>
                </div>
                <div>
                  <label htmlFor="settled-progress">
                    Phòng đã thu đủ{' '}
                    <strong>
                      {page.Dashboard.Settled} / {page.Dashboard.Rooms}
                    </strong>
                  </label>
                  <progress
                    id="settled-progress"
                    value={page.Dashboard.Settled}
                    max={page.Dashboard.Rooms}
                  ></progress>
                </div>
              </div>
            </>
          ) : (
            <>
              <p className="empty">Tiến độ sẽ hiện khi có phòng trong kỳ.</p>
            </>
          )}
        </Panel>
      </div>
      <details className="card dashboard-finances">
        <summary>
          <span>Số liệu thu tiền</span>
          <span className="show-finances">Hiện số liệu</span>
          <span className="hide-finances">Ẩn số liệu</span>
        </summary>
        <dl className="finance-metrics">
          <div>
            <dt>Phải thu sau điều chỉnh</dt>
            <dd>{money(page.Dashboard.Billed)}</dd>
          </div>
          <div>
            <dt>Đã thu ròng</dt>
            <dd>{money(page.Dashboard.Collected)}</dd>
          </div>
          <div>
            <dt>Còn phải thu</dt>
            <dd>{money(page.Dashboard.Due)}</dd>
          </div>
          <div>
            <dt>Thu dư</dt>
            <dd>{money(page.Dashboard.Credit)}</dd>
          </div>
        </dl>
        <section className="collection-trend" aria-labelledby="trend-heading">
          <h2 id="trend-heading">Đã thu và còn phải thu trong 6 kỳ</h2>
          <p className="subtle">
            Tính theo kỳ ghi trên bảng kê đã phát hành, cập nhật theo số dư hiện
            tại.
          </p>
          {has(page.Dashboard.TrendMax) ? (
            <>
              <div className="trend-legend">
                <span className="legend-collected">Đã thu ròng</span>
                <span className="legend-due">Còn phải thu</span>
              </div>
              <ol className="trend-chart">
                {has(page.Dashboard.Trend)
                  ? page.Dashboard.Trend.map((month, index) => (
                      <Fragment key={itemKey(month, index)}>
                        <li
                          className={
                            'trend-month' +
                            (month.Period === page.Period
                              ? ' trend-current'
                              : '')
                          }
                        >
                          <time dateTime={month.Period}>{month.Label}</time>
                          {has(month.Issued) ? (
                            <>
                              <div className="trend-series">
                                <div>
                                  <progress
                                    className="bar-collected"
                                    aria-label={
                                      'Đã thu ròng kỳ ' + month.Label
                                    }
                                    value={month.Collected}
                                    max={page.Dashboard.TrendMax}
                                  ></progress>
                                  <span>{money(month.Collected)}</span>
                                </div>
                                <div>
                                  <progress
                                    className="bar-due"
                                    aria-label={
                                      'Còn phải thu kỳ ' + month.Label
                                    }
                                    value={month.Due}
                                    max={page.Dashboard.TrendMax}
                                  ></progress>
                                  <span>{money(month.Due)}</span>
                                </div>
                              </div>
                            </>
                          ) : (
                            <>
                              <span className="subtle">Chưa phát hành</span>
                            </>
                          )}
                        </li>
                      </Fragment>
                    ))
                  : null}
              </ol>
            </>
          ) : (
            <>
              <p className="empty">
                Chưa có số tiền phải thu hoặc đã thu để vẽ biểu đồ trong 6 kỳ
                này.
              </p>
            </>
          )}
        </section>
      </details>
    </>
  );
}
