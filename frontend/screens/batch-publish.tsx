'use client';

import { useState } from 'react';
import { Action } from './components';
import { money } from './view';
import type { Invoice } from './view';

const batchLimit = 100;
const pageSize = 25;

export function BatchPublish({ drafts, csrf, period, site }: {
  drafts: Pick<Invoice, 'ID' | 'Revision' | 'RoomName' | 'Total'>[];
  csrf: string;
  period: string;
  site: string;
}) {
  const [page, setPage] = useState(0);
  const [selected, setSelected] = useState<Set<string>>(new Set());
  const [confirmed, setConfirmed] = useState(false);
  const [error, setError] = useState('');
  const pageCount = Math.ceil(drafts.length / pageSize);
  const visible = drafts.slice(page * pageSize, (page + 1) * pageSize);
  const allVisibleSelected = visible.every((draft) => selected.has(draft.ID));
  const chosen = drafts.filter((draft) => selected.has(draft.ID));

  function changeSelection(ids: string[], include: boolean) {
    const next = new Set(selected);
    for (const id of ids) {
      if (include) next.add(id);
      else next.delete(id);
    }
    if (next.size > batchLimit) {
      setError('Mỗi lô tối đa 100 bảng kê. Bỏ bớt lựa chọn rồi thử lại.');
      return;
    }
    setSelected(next);
    setConfirmed(false);
    setError('');
  }

  return (
    <form method="post" action="/action/issue-batch" className="form batch-publish"
      onSubmit={(event) => {
        if (!chosen.length) {
          event.preventDefault();
          setError('Chọn ít nhất một bảng kê để phát hành.');
        }
      }}>
      <input type="hidden" name="csrf" value={csrf} />
      <input type="hidden" name="period" value={period} />
      {site && <input type="hidden" name="site" value={site} />}
      {chosen.map((draft) => (
        <input key={draft.ID} type="hidden" name="selected" value={draft.ID + ':' + draft.Revision} />
      ))}
      <p className="subtle">Kiểm tra chi tiết từng phòng trước khi phát hành. Mỗi lô tối đa 100 bảng kê.</p>
      <p role="status">
        Đã chọn {chosen.length}/{batchLimit} · Tổng {money(chosen.reduce((sum, draft) => sum + draft.Total, 0))}
      </p>
      <Action type="button" className="secondary"
        onClick={() => changeSelection(visible.map((draft) => draft.ID), !allVisibleSelected)}>
        {allVisibleSelected ? 'Bỏ chọn trang này' : 'Chọn cả trang này'}
      </Action>
      {visible.map((draft) => (
        <label className="check" key={draft.ID}>
          <input type="checkbox" checked={selected.has(draft.ID)}
            onChange={(event) => changeSelection([draft.ID], event.currentTarget.checked)} />
          <span>{draft.RoomName} · {money(draft.Total)}</span>
        </label>
      ))}
      {pageCount > 1 && (
        <nav className="actions" aria-label="Trang bảng kê trong lô">
          <Action type="button" className="secondary" disabled={page === 0}
            onClick={() => setPage(page - 1)}>Trang trước</Action>
          <span>Trang {page + 1}/{pageCount}</span>
          <Action type="button" className="secondary" disabled={page + 1 >= pageCount}
            onClick={() => setPage(page + 1)}>Trang sau</Action>
        </nav>
      )}
      {error && <p role="alert">{error}</p>}
      <label className="check">
        <input type="checkbox" name="confirm" value="yes" required checked={confirmed}
          onChange={(event) => setConfirmed(event.currentTarget.checked)} />
        <span>Tôi đã kiểm tra các bảng kê đã chọn.</span>
      </label>
      <button>Phát hành bảng kê đã chọn</button>
    </form>
  );
}
