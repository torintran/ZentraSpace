import { TextInput } from '@mantine/core';
import { ShieldCheck } from 'lucide-react';
import { Action, Panel } from './components';
import type { PageData } from './view';

export function VerifyScreen({ page }: { page: PageData }) {
  return (
    <Panel className="card auth" as="section">
      <h1>Xác minh để xem bảng kê</h1>
      <p className="subtle">
        Nhập mã 8 chữ số người quản lý đã cung cấp riêng.
      </p>
      <form method="post" action={'/p/' + page.InviteID} className="form">
        <TextInput
          label="Mã xác minh"
          name="code"
          inputMode="numeric"
          autoComplete="one-time-code"
          pattern="[0-9]{8}"
          maxLength={8}
          size="md"
          required
        />
        <Action leftSection={<ShieldCheck aria-hidden size={18} />}>
          Xem bảng kê
        </Action>
      </form>
    </Panel>
  );
}
