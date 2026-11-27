# ADR 001 — Next.js, Mantine và Lucide cho frontend

## Trạng thái

Đã chốt ngày 27/11/2026.

## Bối cảnh

ZentraSpace cần một giao diện hiện đại, dễ đọc và dễ thao tác trên điện thoại cho người quản lý lớn tuổi, đồng thời hỗ trợ thao tác nhanh trên desktop. Các màn phòng, công tơ, bảng kê và thu tiền có nhiều mẫu form, danh sách và phản hồi lặp lại. Bộ component dùng chung cần giảm HTML, CSS và mã tương tác phải tự duy trì.

## Quyết định

- Dùng Next.js App Router, TypeScript, Mantine và Lucide React cho frontend. Ghim Mantine 9.8.2 khi bắt đầu và đồng bộ phiên bản các gói Mantine trong lockfile. [Hướng dẫn Mantine với Next.js](https://mantine.dev/guides/next/) · [Mantine 9.8.2](https://github.com/mantinedev/mantine/releases/tag/9.8.2) · [Lucide React](https://lucide.dev/guide/react)
- Dùng component Mantine làm mặc định cho form, bố cục và thông báo. Theme chung xác định màu, chữ, khoảng cách và vùng bấm. Các nút nghiệp vụ chính có nhãn chữ cùng icon Lucide.
- Giữ Go và PostgreSQL làm nguồn nghiệp vụ và dữ liệu. Go chịu trách nhiệm xác thực, phân quyền, kiểm đầu vào, tính phí và ghi sổ.

## Triển khai và kiểm chứng

Docker Compose chạy Next.js trên mạng nội bộ cùng Go. Go quyết định màn hình và trả mã trạng thái, chuyển hướng, cookie và tệp tải xuống cho trình duyệt. Tiến trình Node nhận dữ liệu của từng yêu cầu, dựng màn hình qua App Router và trả HTML. Go chỉ chuyển tiếp tài nguyên công khai dưới `/_next/static/`. Trình duyệt có thể gửi form nghiệp vụ qua Go ngay cả khi JavaScript bị tắt.

Kiểm hồi quy Docker bao phủ các luồng HTTP, kiểm thử qua trình duyệt, cô lập dữ liệu theo tổ chức và khôi phục bản sao lưu. Mốc hiệu năng trên thiết bị của người dùng thuộc đợt thử nghiệm thực địa.
