import { Trash2 } from 'lucide-react';
import { Action } from './components';
import type { PageData } from './view';

export function PropertyDelete({
  page,
  kind,
  id,
  name,
  reason,
  revision,
}: {
  page: PageData;
  kind: 'site' | 'room';
  id: string;
  name: string;
  reason?: string;
  revision?: string;
}) {
  const noun = kind === 'site' ? 'dãy trọ' : 'phòng';
  return (
    <details className="property-delete">
      <summary>
        <Trash2 size={16} aria-hidden /> Xóa {noun}
      </summary>
      {reason ? (
        <p className="subtle">{reason}</p>
      ) : (
        <form
          method="post"
          action={'/action/' + kind + '-delete'}
          className="form"
        >
          <input type="hidden" name="id" value={id} />
          <input type="hidden" name="csrf" value={page.CSRF} />
          <input type="hidden" name="period" value={page.Period} />
          {revision && <input type="hidden" name="revision" value={revision} />}
          <label className="check">
            <input type="checkbox" name="confirm" value="yes" required />
            <span>
              Tôi xác nhận xóa {noun} <strong>{name}</strong>
            </span>
          </label>
          <Action className="property-delete-submit">Xóa {noun}</Action>
        </form>
      )}
    </details>
  );
}
