# ZentraSpace

Bản chạy thử cục bộ cho chủ trọ một tòa. Luồng chính gồm nhập phòng và người đại diện, lưu công tơ, tạo bảng kê nháp, phát hành, chia sẻ riêng, ghi thu hoặc điều chỉnh và xuất công nợ.

## Chạy trên Windows với Docker Desktop

1. Tại thư mục dự án, chạy `./scripts/init-local.ps1` để sinh mật khẩu và khóa. Nếu đã có cấu hình, script giữ lại tệp cũ.
2. Chạy `docker compose up --build -d --wait`.
3. Mở [ZentraSpace](http://localhost:8087/). Chọn **Chủ trọ** và dùng giá trị `ZS_OWNER_PASSWORD` trong tệp `.env`. Để thử vai trò **Nhân viên công tơ**, dùng `ZS_METER_PASSWORD`.

Chọn kỳ thu ở đầu màn hình. Vào **Phòng và hợp đồng** để nhập một phòng hoặc dán CSV theo mẫu. Xem kết quả kiểm CSV trước khi lưu. Sang **Nhập điện nước** để ghi chỉ số cuối kỳ, sau đó vào **Chốt kỳ thu** để tạo bản nháp. Mở từng bảng kê, kiểm tiền và phát hành. Thao tác thu thiếu, thu dư, hoàn và ghi giảm đều có lịch sử. CSV công nợ tải được từ màn chốt kỳ.

Để thử chia sẻ, tạo link trên bảng kê đã phát hành. Mở link bằng một phiên trình duyệt riêng trên cùng máy và nhập mã dùng một lần. Trong bản chạy thử, chủ trọ trao mã thủ công qua kênh đã xác nhận với người thuê. Lời nhắc được sao chép không chứa mã hoặc số tiền. Có thể thu hồi link ngay cả khi người thuê đang có phiên xem.

## VietQR và đối soát sao kê

Chủ trọ mở **Đối soát ngân hàng**, nhập mã BIN, số tài khoản và tên chủ tài khoản đã kiểm tra. VietQR được tạo ngay trên máy chủ ZentraSpace và chỉ hiện trên bảng kê đã xác minh còn số tiền phải trả. Sau khi ghi thu một phần, mã dùng số tiền còn lại và giữ nguyên nội dung chuyển khoản của bảng kê. Khi đổi BIN hoặc số tài khoản, ứng dụng thu hồi link cũ. Kiểm tên thụ hưởng trên ứng dụng ngân hàng trước khi xác nhận bất kỳ chuyển khoản thật nào. [Chuyển nhanh qua VietQR của NAPAS](https://napas.com.vn/dich-vu-chuyen-tien-nhanh-napas-247)

Để đối soát, dán các giao dịch tiền vào của tài khoản đang chọn theo mẫu `ma_giao_dich,ngay,so_tien,noi_dung`. Ngày theo YYYY-MM-DD, số tiền VND nguyên dương và mã lấy từ sao kê ngân hàng. Khi nội dung có một mã bảng kê xác định, hệ thống chỉ gợi ý phòng và kỳ. Chủ trọ kiểm rồi chọn ghi khoản thu mới hoặc ghép với khoản thu cùng số tiền đã ghi. Nhập lại và duyệt lại cùng giao dịch giữ nguyên công nợ. Sao kê, tài khoản và liên kết đối soát được lưu trong sổ mã hóa hiện có.

## Phạm vi của lần chạy thử

- Go render giao diện HTML và PostgreSQL giữ dữ liệu. Tôi chọn render phía máy chủ cho bản chạy thử để chạy được bằng một tiến trình web trong Docker khi npm trên máy không truy cập được registry. Blueprint mô tả Next.js cho kiến trúc mở rộng.
- Sổ thử được lưu thành tài liệu mã hóa AES-256-GCM theo phạm vi tổ chức. Transaction khóa hàng khi thay đổi để giữ bất biến tài chính. PostgreSQL RLS giới hạn theo ngữ cảnh giao dịch, tài khoản DB chạy web không sở hữu bảng. Service migration giữ tài khoản bootstrap riêng. Mật khẩu đăng nhập được xác minh bằng bcrypt.
- Bản này phục vụ một tòa với hai vai trò cục bộ. Phần lưu trữ hiện là sổ mã hóa theo tổ chức. Mô hình quan hệ đầy đủ cho timeline hợp đồng, phân quyền nhiều tòa, quản lý vòng đời khóa và audit triển khai thuộc các bước tiếp theo. Phiên đăng nhập và phiên xem link được tạo lại sau khi khởi động lại web.
- Điện/nước dùng đơn giá đã được người quản lý xác nhận, chưa tự chọn biểu giá pháp lý. VietQR thử nghiệm và đối soát từ CSV dùng dữ liệu do chủ trọ cung cấp. Webhook ngân hàng, gửi mã xác minh tự động và gửi Zalo tự động vẫn thuộc phạm vi sau.
- Compose mẫu chỉ mở trên loopback để thử bằng dữ liệu mẫu. Trước pilot có người thuê cần môi trường HTTPS, chọn kênh xác minh đã đăng ký, quản lý khóa và backup tách biệt, kiểm tra trên thiết bị thật và tổ chức hai kỳ đối chiếu. Tệp `.env` chứa riêng khóa và thông tin đăng nhập của máy này.

## Kiểm tra lại

Dockerfile chạy test Go khi build. Để chạy bộ kiểm thử HTTP trên bản Docker tách riêng, dùng PowerShell 7:

```powershell
$env:ZS_PORT = '8088'
$env:ZS_ORIGIN = 'http://localhost:8088'
docker compose -p zentraspace-check up --build -d --wait
./scripts/smoke.ps1
./scripts/bank-smoke.ps1
docker compose -p zentraspace-check down
Remove-Item Env:ZS_PORT
Remove-Item Env:ZS_ORIGIN
```

Bộ kiểm thử thêm dữ liệu tổng hợp vào project `zentraspace-check`. Các bài kiểm bao phủ phân quyền công tơ, đối chiếu số tiền, phát hành và thu lặp, nhập/xuất CSV, link chuyển tiếp, mã dùng một lần, thu hồi quyền xem, ảnh QR theo số dư và duyệt sao kê. Bài kiểm sao kê dùng tài khoản mẫu và chưa quét trên ứng dụng ngân hàng thật.

Xem [blueprint sản phẩm](docs/ZENTRASPACE-BLUEPRINT.md) và [backlog pilot](tasks/todo.md).

## Đóng góp và giấy phép

Đọc [Quy tắc ứng xử](CODE_OF_CONDUCT.md) trước khi tham gia. Báo vấn đề về dữ liệu hoặc quyền truy cập theo [chính sách bảo mật](SECURITY.md).

Copyright 2026 Torin Tran. ZentraSpace được phát hành theo [Apache License 2.0](LICENSE).
