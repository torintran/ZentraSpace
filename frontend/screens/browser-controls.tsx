'use client';

import { PasswordInput, TextInput, Textarea } from '@mantine/core';
import { ClipboardCopy, Eye, EyeOff, Search } from 'lucide-react';
import { type ReactNode, useRef, useState } from 'react';
import { Action, FieldSelect } from './components';
import { money } from './view';

function normalizeSearch(text: string) {
  return text.toLocaleLowerCase('vi').normalize('NFD').replace(/\p{M}/gu, '').replace(/đ/g, 'd');
}

export function LoginPassword() {
  return (
    <PasswordInput
      label="Mật khẩu"
      name="password"
      autoComplete="current-password"
      maxLength={72}
      visibilityToggleIcon={({ reveal }) =>
        reveal ? (
          <EyeOff aria-hidden size={18} />
        ) : (
          <Eye aria-hidden size={18} />
        )
      }
      visibilityToggleButtonProps={{ 'aria-label': 'Hiện hoặc ẩn mật khẩu' }}
      visibilityToggleFocusable
      size="md"
      required
    />
  );
}

export function RoomList({
  rooms,
  label = 'Tìm theo mã phòng',
  empty = 'Không tìm thấy phòng.',
  listClass = 'room-cards',
  preserveInputs = false,
}: {
  rooms: { id: string; name: string; row: ReactNode }[];
  label?: string;
  empty?: string;
  listClass?: string;
  preserveInputs?: boolean;
}) {
  const [value, setValue] = useState('');
  const query = normalizeSearch(value.trim());
  const visible = rooms.filter((room) =>
    normalizeSearch(room.name).includes(query),
  );
  return (
    <>
      <TextInput
        className="room-search"
        type="search"
        label={label}
        leftSection={<Search aria-hidden size={18} />}
        value={value}
        onChange={(event) => setValue(event.currentTarget.value)}
      />
      <div className={listClass}>
        {(preserveInputs ? rooms : visible).map((room) => (
          <div key={room.id} hidden={preserveInputs && !visible.includes(room)}>{room.row}</div>
        ))}
      </div>
      {visible.length === 0 && (
        <p role="status" className="subtle">
          {empty}
        </p>
      )}
    </>
  );
}

export function InvoicePicker({ invoices, choice, onChoice }: {
  invoices: { ID: string; RoomName: string; LeaseTenant: string; Period: string; Remaining: number }[];
  choice: string;
  onChoice: (id: string) => void;
}) {
  const [query, setQuery] = useState('');
  const words = normalizeSearch(query.trim()).split(/\s+/);
  const matches = invoices.filter((invoice) => {
    const text = normalizeSearch([invoice.RoomName, invoice.LeaseTenant, invoice.Period].join(' '));
    return words.every((word) => text.includes(word));
  });
  const retained = choice && invoices.find((invoice) => invoice.ID === choice && !matches.includes(invoice));
  return (
    <>
      <TextInput type="search" label="Tìm mã phòng, người thuê hoặc kỳ"
        leftSection={<Search aria-hidden size={18} />} value={query}
        onChange={(event) => setQuery(event.currentTarget.value)} />
      <label>
        Bảng kê đã phát hành
        <FieldSelect name="invoice" required value={choice} onChange={(event) => onChoice(event.currentTarget.value)}>
          <option value="">Chọn phòng và kỳ</option>
          {(retained ? [retained, ...matches] : matches).map((invoice) => (
            <option value={invoice.ID} key={invoice.ID}>
              {[invoice.RoomName, invoice.LeaseTenant, invoice.Period, 'Còn ' + money(invoice.Remaining)].filter(Boolean).join(' · ')}
            </option>
          ))}
        </FieldSelect>
      </label>
      {matches.length === 0 && <p role="status" className="subtle">Không tìm thấy bảng kê khớp từ khóa.</p>}
    </>
  );
}

export function ShareMessage({ link }: { link: string }) {
  const messageField = useRef<HTMLTextAreaElement>(null);
  const [status, setStatus] = useState('');
  const message =
    'Bạn có bảng kê mới. Mở đường dẫn và xác minh để xem: ' + link;
  async function copy() {
    try {
      await navigator.clipboard.writeText(message);
      setStatus('Đã sao chép lời nhắc.');
    } catch {
      messageField.current?.focus();
      messageField.current?.select();
      setStatus('Đã chọn lời nhắc. Hãy dùng thao tác sao chép trên thiết bị.');
    }
  }
  return (
    <>
      <Textarea
        ref={messageField}
        id="share-message"
        label="Lời nhắc để gửi qua Zalo"
        readOnly
        rows={3}
        value={message}
      />
      <Action
        type="button"
        className="secondary"
        onClick={copy}
        leftSection={<ClipboardCopy aria-hidden size={17} />}
      >
        Sao chép lời nhắc
      </Action>
      <p id="copy-status" role="status" className="subtle">
        {status}
      </p>
    </>
  );
}
