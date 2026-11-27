import { TextInput } from '@mantine/core';
import { LogIn } from 'lucide-react';
import { LoginPassword } from './browser-controls';
import { Action, Panel } from './components';
import type { PageData } from './view';

export function LoginScreen(_: { page: PageData }) {
  return (
    <Panel className="card auth" as="section">
      <h1>Đăng nhập</h1>
      <p className="subtle">Quản lý phòng và các kỳ thu của bạn.</p>
      <form method="post" action="/login" className="form">
        <TextInput
          label="Tài khoản"
          name="account"
          list="local-accounts"
          maxLength={40}
          autoComplete="username"
          defaultValue="owner"
          size="md"
          required
        />
        <datalist id="local-accounts">
          <option value="owner" label="Chủ trọ" />
          <option value="meter" label="Nhân viên công tơ" />
        </datalist>
        <LoginPassword />
        <Action leftSection={<LogIn aria-hidden size={18} />}>Đăng nhập</Action>
      </form>
    </Panel>
  );
}
