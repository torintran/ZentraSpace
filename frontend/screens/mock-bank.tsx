import { TextInput } from '@mantine/core';
import { Banknote } from 'lucide-react';
import { Action, Panel } from './components';
import type { PageData } from './view';

export function MockBankScreen({ page }: { page: PageData }) {
  const event = page.MockEvent;
  if (!event) return null;
  return (
    <Panel className="card auth" as="section">
      <h1>Ngân hàng mô phỏng</h1>
      <p>
        Gửi giao dịch giả lập vào tài khoản thử 0000000000000. Giữ nguyên thông
        tin giao dịch để kiểm tra gửi lại.
      </p>
      <form method="post" action="/transfer" className="form">
        <TextInput
          label="Mã giao dịch"
          name="reference"
          defaultValue={event.transaction_id}
          maxLength={80}
          required
        />
        <TextInput
          label="Ngày giao dịch"
          name="date"
          type="date"
          defaultValue={event.date}
          required
        />
        <TextInput
          label="Số tiền (đồng)"
          name="amount"
          type="number"
          inputMode="numeric"
          defaultValue={event.amount_vnd}
          min={1}
          max={100000000000}
          required
        />
        <TextInput
          label="Nội dung chuyển khoản"
          name="memo"
          defaultValue={event.memo}
          maxLength={100}
        />
        <Action leftSection={<Banknote aria-hidden size={18} />}>
          Gửi giao dịch
        </Action>
      </form>
    </Panel>
  );
}
