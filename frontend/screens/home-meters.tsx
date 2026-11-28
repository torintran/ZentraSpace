import { Action, FieldInput, Panel, Status } from './components';
import { RoomList } from './browser-controls';
import { has } from './view';
import type { PageData } from './view';

import { Check, Gauge } from 'lucide-react';
import { SiteMeters } from './site-meters';
export function MetersPanel({ page }: { page: PageData }) {
  const siteQuery = page.SelectedSite ? '&site=' + encodeURIComponent(page.SelectedSite) : '';
  return (
    <>
      <Panel className="card" as="section">
        <h2>Chỉ số cuối kỳ {page.Period}</h2>
        <p className="subtle">
          Kiểm tra số trên công tơ, rồi lưu theo từng phòng.
        </p>
        {!!page.SiteMeterChecks?.length && (
          <p>
            <a href="#site-meters">Đối chiếu công tơ tổng</a>
          </p>
        )}
        <nav className="tabs" aria-label="Lọc theo chỉ số">
          <a
            href={'/?tab=meters&period=' + page.Period + siteQuery}
            className={!has(page.Filter) ? 'active' : undefined}
            aria-current={!has(page.Filter) ? 'page' : undefined}
          >
            Tất cả
          </a>
          <a
            href={'/?tab=meters&period=' + page.Period + siteQuery + '&filter=missing'}
            className={page.Filter === 'missing' ? 'active' : undefined}
            aria-current={page.Filter === 'missing' ? 'page' : undefined}
          >
            Thiếu chỉ số
          </a>
        </nav>
        {has(page.Rooms) ? (
          <RoomList listClass="meter-results" preserveInputs rooms={page.Rooms.map((room) => ({
            id: room.ID,
            name: room.Name,
            row: <>
              {room.OpeningUnavailable ? (
                <div className="list-row">
                  <div>
                    <strong>{room.Name}</strong>
                    <p>Chờ phát hành bảng kê kỳ trước để có chỉ số đầu kỳ.</p>
                  </div>
                </div>
              ) : has(!has(room.Handover) || has(room.Done)) ? (
                <>
                  <form
                    className="meter-row"
                    method="post"
                    action="/action/read"
                  >
                    <input type="hidden" name="csrf" value={page.CSRF} />
                    <input type="hidden" name="period" value={page.Period} />
                    {page.SelectedSite && <input type="hidden" name="site" value={page.SelectedSite} />}
                    <input type="hidden" name="id" value={room.ID} />
                    {has(page.Filter) ? (
                      <>
                        <input
                          type="hidden"
                          name="filter"
                          value={page.Filter}
                        />
                      </>
                    ) : null}
                    <div>
                      <strong>{room.Name}</strong>{' '}
                      {room.MeterLocked ? (
                        <Status className="badge good">Đã chốt</Status>
                      ) : has(room.Done) ? (
                        <>
                          <Status className="badge good">Đã lưu</Status>
                        </>
                      ) : (
                        <>
                          <Status className="badge">Chưa nhập</Status>
                        </>
                      )}
                    </div>
                    <label>
                      Điện cuối kỳ (kWh)
                      <FieldInput
                        type="number"
                        name="electric"
                        inputMode="numeric"
                        min="0"
                        max="999999"
                        readOnly={room.MeterLocked}
                        required
                        defaultValue={has(room.Done) ? room.Electric : ''}
                      />
                      <small>Đầu kỳ {room.ElectricOpening.toLocaleString('vi-VN')} kWh</small>
                    </label>
                    <label>
                      Nước cuối kỳ (m³)
                      <FieldInput
                        type="number"
                        name="water"
                        inputMode="numeric"
                        min="0"
                        max="999999"
                        readOnly={room.MeterLocked}
                        required
                        defaultValue={has(room.Done) ? room.Water : ''}
                      />
                      <small>Đầu kỳ {room.WaterOpening.toLocaleString('vi-VN')} m³</small>
                    </label>
                    {!room.MeterLocked && (
                      <Action leftSection={<Check size={18} aria-hidden />}>Lưu chỉ số</Action>
                    )}
                  </form>
                </>
              ) : (
                <>
                  <div className="list-row">
                    <strong>{room.Name}</strong>
                    <Status className="badge">
                      Cần chỉ số bàn giao đầu kỳ
                    </Status>
                  </div>
                </>
              )}
              {page.Role === 'owner' && !room.MeterLocked && !room.OpeningUnavailable ? (
                <>
                  <details
                    className="meter-special"
                    open={
                      has(has(room.Handover) && !has(room.Done))
                        ? true
                        : undefined
                    }
                  >
                    <summary>
                      {has(room.Handover) ? (
                        <>Chỉ số bàn giao và cuối kỳ · {room.Name}</>
                      ) : (
                        <>Thay công tơ trong kỳ · {room.Name}</>
                      )}
                    </summary>
                    <form
                      method="post"
                      action="/action/read-special"
                      className="form"
                    >
                      <input type="hidden" name="csrf" value={page.CSRF} />
                      <input type="hidden" name="period" value={page.Period} />
                      {page.SelectedSite && <input type="hidden" name="site" value={page.SelectedSite} />}
                      <input type="hidden" name="id" value={room.ID} />
                      {has(page.Filter) ? (
                        <>
                          <input
                            type="hidden"
                            name="filter"
                            value={page.Filter}
                          />
                        </>
                      ) : null}
                      {has(room.Handover) ? (
                        <>
                          <input type="hidden" name="handover" value="yes" />
                          <p className="subtle">
                            Nhập chỉ số tại lúc bàn giao. Điện nước phát sinh
                            khi phòng trống không cộng vào kỳ thuê này.
                          </p>
                          <div className="pair">
                            <label>
                              Điện lúc bàn giao
                              <FieldInput
                                name="openingElectric"
                                type="number"
                                min="0"
                                max="999999"
                                inputMode="numeric"
                                required
                                defaultValue={
                                  has(room.Done)
                                    ? room.Adjustment.OpeningElectric
                                    : ''
                                }
                              />
                            </label>
                            <label>
                              Nước lúc bàn giao
                              <FieldInput
                                name="openingWater"
                                type="number"
                                min="0"
                                max="999999"
                                inputMode="numeric"
                                required
                                defaultValue={
                                  has(room.Done)
                                    ? room.Adjustment.OpeningWater
                                    : ''
                                }
                              />
                            </label>
                          </div>
                        </>
                      ) : null}
                      <fieldset className="form">
                        <legend>Công tơ điện</legend>
                        <label className="check">
                          <input
                            type="checkbox"
                            name="electricReplaced"
                            value="yes"
                            defaultChecked={
                              has(room.Adjustment.Electric.Replaced)
                                ? true
                                : undefined
                            }
                          />{' '}
                          Có thay công tơ điện
                        </label>
                        <div className="pair">
                          <label>
                            Số cuối công tơ điện cũ
                            <FieldInput
                              name="electricOldEnd"
                              type="number"
                              min="0"
                              max="999999"
                              inputMode="numeric"
                              defaultValue={
                                has(room.Adjustment.Electric.Replaced)
                                  ? room.Adjustment.Electric.OldEnd
                                  : ''
                              }
                            />
                          </label>
                          <label>
                            Số đầu công tơ điện mới
                            <FieldInput
                              name="electricNewStart"
                              type="number"
                              min="0"
                              max="999999"
                              inputMode="numeric"
                              defaultValue={
                                has(room.Adjustment.Electric.Replaced)
                                  ? room.Adjustment.Electric.NewStart
                                  : ''
                              }
                            />
                          </label>
                        </div>
                      </fieldset>
                      <fieldset className="form">
                        <legend>Công tơ nước</legend>
                        <label className="check">
                          <input
                            type="checkbox"
                            name="waterReplaced"
                            value="yes"
                            defaultChecked={
                              has(room.Adjustment.Water.Replaced)
                                ? true
                                : undefined
                            }
                          />{' '}
                          Có thay công tơ nước
                        </label>
                        <div className="pair">
                          <label>
                            Số cuối công tơ nước cũ
                            <FieldInput
                              name="waterOldEnd"
                              type="number"
                              min="0"
                              max="999999"
                              inputMode="numeric"
                              defaultValue={
                                has(room.Adjustment.Water.Replaced)
                                  ? room.Adjustment.Water.OldEnd
                                  : ''
                              }
                            />
                          </label>
                          <label>
                            Số đầu công tơ nước mới
                            <FieldInput
                              name="waterNewStart"
                              type="number"
                              min="0"
                              max="999999"
                              inputMode="numeric"
                              defaultValue={
                                has(room.Adjustment.Water.Replaced)
                                  ? room.Adjustment.Water.NewStart
                                  : ''
                              }
                            />
                          </label>
                        </div>
                      </fieldset>
                      <div className="pair">
                        <label>
                          Số điện cuối kỳ
                          <FieldInput
                            name="electric"
                            type="number"
                            min="0"
                            max="999999"
                            inputMode="numeric"
                            required
                            defaultValue={has(room.Done) ? room.Electric : ''}
                          />
                        </label>
                        <label>
                          Số nước cuối kỳ
                          <FieldInput
                            name="water"
                            type="number"
                            min="0"
                            max="999999"
                            inputMode="numeric"
                            required
                            defaultValue={has(room.Done) ? room.Water : ''}
                          />
                        </label>
                      </div>
                      <label>
                        Lý do hoặc tham chiếu biên bản
                        <FieldInput
                          name="reason"
                          maxLength={240}
                          required
                          defaultValue={room.Adjustment.Reason}
                        />
                      </label>
                      <Action leftSection={<Gauge size={18} aria-hidden />}>
                        Lưu chỉ số và biên bản
                      </Action>
                    </form>
                  </details>
                </>
              ) : null}
            </>,
          }))} />
        ) : (
          <>
            <p className="subtle">
              {page.Filter === 'missing' ? (
                <>Không có phòng thiếu chỉ số trong kỳ này.</>
              ) : (
                <>Chưa có phòng để nhập chỉ số.</>
              )}
            </p>
          </>
        )}
      </Panel>
      <SiteMeters page={page} />
    </>
  );
}
