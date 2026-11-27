import { ActionLink, Panel } from './components';

import type { PageData, Invoice } from './view';

export function ErrorScreen({ page }: { page: PageData }) {
  return (
    <>
      <Panel className="card" as="div">
        <ActionLink className="button secondary" href="/">
          Về trang quản lý
        </ActionLink>
      </Panel>
    </>
  );
}
