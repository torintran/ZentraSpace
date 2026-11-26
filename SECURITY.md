# Bảo mật ZentraSpace

## Phạm vi hỗ trợ

Nhánh `main` là bản đang phát triển và kiểm thử cục bộ. Khi báo lỗi, ghi commit hoặc phiên bản đang chạy, môi trường Docker và cấu hình liên quan đã loại bỏ bí mật.

## Báo lỗi riêng tư

Dùng [báo cáo lỗ hổng riêng tư trên GitHub](https://github.com/torintran/ZentraSpace/security/advisories/new) nếu nút báo cáo có sẵn. Nếu kênh đó chưa bật, yêu cầu người duy trì cung cấp đầu mối riêng tư bằng một issue chung. Giữ chi tiết lỗi và dữ liệu nhạy cảm trong kênh riêng.

Một báo cáo nên có phạm vi ảnh hưởng, mô tả hành vi, kết quả mong đợi và dữ liệu tổng hợp đủ để người duy trì xác nhận. Che toàn bộ mật khẩu, khóa mã hóa, token chia sẻ, dữ liệu sao kê và định danh người thuê. Dùng môi trường tự dựng hoặc dữ liệu được phép kiểm thử.

Người duy trì sẽ đánh giá ảnh hưởng, xác định hướng khắc phục và trao đổi thời điểm thông báo. Trong sự cố liên quan đến khóa hoặc thông tin đăng nhập, người vận hành nên thu hồi quyền đang dùng, thay bí mật đã lộ và kiểm tra nhật ký phù hợp với phạm vi sự cố.

## Cài đặt cục bộ

- Sinh khóa và mật khẩu bằng `scripts/init-local.ps1`. Giữ `.env` ngoài Git và giữ khóa khôi phục tách khỏi bản sao dữ liệu.
- Compose mẫu chỉ mở trên loopback. Hoàn tất HTTPS và kiểm thử phân quyền, sao lưu, khôi phục trước khi dùng dữ liệu thật.
- Link bảng kê yêu cầu xác minh qua kênh riêng, có hạn và thu hồi được. Cấu hình tài khoản nhận tiền chỉ dành cho Chủ trọ.
