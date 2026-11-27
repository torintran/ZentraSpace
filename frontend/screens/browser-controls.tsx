'use client';

import { PasswordInput, TextInput, Textarea } from '@mantine/core';
import { ClipboardCopy, Eye, EyeOff, Search } from 'lucide-react';
import { type ReactNode, useRef, useState } from 'react';
import { Action } from './components';

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
}: {
  rooms: { id: string; name: string; row: ReactNode }[];
}) {
  const [value, setValue] = useState('');
  const query = value.trim().toLocaleLowerCase('vi');
  const visible = rooms.filter((room) =>
    room.name.toLocaleLowerCase('vi').includes(query),
  );
  return (
    <>
      <TextInput
        className="room-search"
        type="search"
        label="Tìm theo mã phòng"
        leftSection={<Search aria-hidden size={18} />}
        value={value}
        onChange={(event) => setValue(event.currentTarget.value)}
      />
      <div className="room-cards">
        {visible.map((room) => (
          <div key={room.id}>{room.row}</div>
        ))}
      </div>
      {visible.length === 0 && (
        <p role="status" className="subtle">
          Không tìm thấy phòng.
        </p>
      )}
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
