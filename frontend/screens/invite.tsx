import { ShareMessage } from './browser-controls';

import { Panel } from './components';

import type { PageData, Invoice } from './view';

export function InviteScreen({ page }: { page: PageData }) {
  const bill = page.Invoice;
  if (!bill) return null;
  return (
    <>
      <a className="back" href={'/invoice/' + bill.ID}>
        ← Về bảng kê
      </a>
      <Panel className="card auth" as="section">
        <h1>Gửi bảng kê riêng</h1>
        <ShareMessage link={page.ShareURL} />
        <div className="code">
          <span>Mã dùng một lần</span>
          <strong>{page.Code}</strong>
        </div>
        <p className="subtle">
          Trao mã qua một kênh riêng, chẳng hạn cuộc gọi tới số đã xác nhận.
          Link và mã này hết hạn sau 24 giờ.
        </p>
      </Panel>
    </>
  );
}
