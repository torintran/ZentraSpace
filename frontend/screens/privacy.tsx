import { Panel } from './components';

import type { PageData, Invoice } from './view';

export function PrivacyScreen({ page }: { page: PageData }) {
  return (
    <>
      <Panel className="card privacy" as="article">
        <h1>Quyền riêng tư</h1>
        <p>
          Thông tin này áp dụng cho bản thử nghiệm ZentraSpace. Người vận hành
          tòa nhà quản lý dữ liệu nhập vào hệ thống.
        </p>
        <h2>Dữ liệu được xử lý</h2>
        <p>
          ZentraSpace lưu thông tin phòng, người đại diện, chỉ số điện nước,
          bảng kê, các khoản thu và điều chỉnh. Khi chủ trọ dùng đối soát, hệ
          thống lưu thêm tài khoản nhận tiền và giao dịch sao kê được nhập.
        </p>
        <h2>Quyền truy cập</h2>
        <p>
          Chủ trọ quản lý bảng kê và công nợ. Nhân viên công tơ truy cập danh
          sách phòng và chỉ số. Người nhận bảng kê cần mã dùng một lần để xem
          link riêng. Chủ trọ có thể thu hồi link.
        </p>
        <h2>Phiên đăng nhập và lưu trữ</h2>
        <p>
          Ứng dụng dùng cookie để duy trì phiên đăng nhập và phiên xem bảng kê.
          Sổ nghiệp vụ được mã hóa trước khi lưu vào cơ sở dữ liệu. Người vận
          hành quản lý bản sao lưu và thời gian lưu giữ.
        </p>
        <h2>Yêu cầu về dữ liệu</h2>
        <p>
          Liên hệ chủ trọ hoặc đơn vị vận hành đã gửi bảng kê để yêu cầu kiểm
          tra, sửa hoặc xóa dữ liệu, hoặc để báo link bị chia sẻ nhầm.
        </p>
      </Panel>
    </>
  );
}
