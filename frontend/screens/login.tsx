import { TextInput } from '@mantine/core';
import { LogIn } from 'lucide-react';
import { LoginPassword } from './browser-controls';
import { FieldSelect, Action, Panel } from './components';
import type { PageData } from './view';

export function LoginScreen({ page }: { page: PageData }) {
  return (
    <Panel className="card auth" as="section">
      <h1>Đăng nhập</h1>
      <p className="subtle">Quản lý phòng và các kỳ thu của bạn.</p>
      <form method="post" action="/login" className="form login-form">
        <label>
          Tài khoản
          <FieldSelect
            name="profile"
            defaultValue={page.LoginProfile || 'owner'}
          >
            <option value="owner">Chủ trọ</option>
            <option value="meter">Nhân viên công tơ</option>
            <option value="other">Tài khoản khác</option>
          </FieldSelect>
        </label>
        <TextInput
          className="login-account-name"
          label="Tên đăng nhập"
          name="account"
          maxLength={40}
          autoComplete="username"
          size="md"
        />
        <LoginPassword />
        <Action leftSection={<LogIn aria-hidden size={18} />}>Đăng nhập</Action>
      </form>
    </Panel>
  );
}
