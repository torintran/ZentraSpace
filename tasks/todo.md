# Backlog triển khai ZentraSpace

Trạng thái ngày 27/11/2026. Cột **Sau** ghi phụ thuộc trực tiếp. Cột **Nghiệm thu và kiểm tra** mô tả bằng chứng cần có khi đóng việc. Tên thư mục và contract là dự kiến theo [kế hoạch](plan.md).

**Bản chạy thử cục bộ đã có.** Docker kiểm được luồng phòng, import CSV, công tơ, bảng kê, ghi thu, xuất công nợ và link xác minh. Đã kiểm quyền tài khoản công tơ, RLS ở PostgreSQL và khôi phục bản dump dữ liệu mã hóa vào DB khác. Các ô dưới đây theo dõi mức nghiệm thu pilot đầy đủ. Thử nghiệm với chủ trọ/người thuê thật, HTTPS và kênh gửi mã tự động vẫn nằm ngoài xác minh cục bộ. [Hướng dẫn chạy và kiểm thử](../README.md)

| Giai đoạn | Bằng chứng cục bộ | Điều kiện còn mở |
| --- | --- | --- |
| 0 | Có bản thử cho luồng chốt và xác minh người nhận. | Tuyển người tham gia, quan sát kỳ thu thật, thử trên thiết bị thật và ghi nhận cam kết trả phí. |
| 1A | Compose chạy migration trên DB mới. Chủ trọ tìm, xem và sửa phòng. Smoke test kiểm quyền sửa, khóa thông tin ảnh hưởng bảng kê đã phát hành, nhập CSV có xem trước và tải mẫu, công tơ, phát hành và link riêng. | Quyền ứng dụng nhiều tổ chức/tòa, hợp đồng và đơn giá theo thời gian, công nợ mở, ca công tơ ngoại lệ, hạn thanh toán và xác minh qua kênh đã đăng ký của pilot. |
| 1B | Smoke test kiểm phát hành và ghi thu đồng thời, phương thức thu và mã chứng từ, đảo bút toán, lọc công tơ/bảng kê và đối chiếu CSV công nợ. Bộ kiểm Docker diễn tập khôi phục DB và kiểm giới hạn quyền đọc. | Số dư chưa phân bổ tách sổ, thao tác lô theo quyền, HTTPS, kiểm với người dùng mục tiêu và hai kỳ đối chiếu thật. |
| 2 | Luồng nhập sao kê có người duyệt và VietQR theo số dư được kiểm với dữ liệu tổng hợp. | Quét bằng app ngân hàng, đo thời gian đối soát, chọn nhánh mở rộng theo pilot. Webhook cần provider và contract thử. |
| 3 | Compose hỗ trợ chạy thử trên một máy và kiểm khôi phục DB. | Xác nhận nhu cầu, dữ liệu và người duyệt theo từng nhánh Advanced. |

## Giai đoạn 0 — Xác thực nhu cầu

| Mã | Việc | Sau | Nghiệm thu và kiểm tra |
| --- | --- | --- | --- |
| G01 | [ ] Tuyển chủ trọ và quan sát ca chốt | — | Tiếp cận 5–8 chủ trọ, ghi công cụ đang dùng và đo từng bước. Xin mẫu biểu, tin nhắn, công nợ và công tơ đã ẩn danh. Ghi ca hỏi lại tiền điện, ca ghép tiền mơ hồ hoặc ca ngoài quy trình nếu có. |
| G02 | [ ] Chạy thử một kỳ có hỗ trợ | G01 | Chủ trọ duyệt rồi tự gửi Zalo, người thuê thử đọc nguồn số tiền. Đo thời gian trên cùng nhóm phòng, số lần hỏi lại theo lý do, khoản lệch và phút hỗ trợ. Chốt đầu mối và khung hỗ trợ ngày chốt. |
| G05 | [ ] Thử mẫu thao tác và xác minh người nhận | G01 | So bản mẫu chốt kỳ với cách hiện tại. Ghi lỗi, lần được hỗ trợ và thời gian của 5 quản lý từ 55 tuổi và ít nhất 3 người dùng quen công nghệ. Thử xác minh trên máy thực, chọn một kênh liên hệ đã đăng ký và chốt chi phí. |
| G03 | [ ] Quyết định phạm vi bản đầu | G02, G05 | Tổng hợp ít nhất 3 người cùng vấn đề, 2 ca hoàn thành, 1 chấp nhận đề nghị phí cụ thể để qua cổng. Chốt tòa, người duyệt, ngày kỳ tiếp theo và bước ưu tiên. Ghi quyết định lặp lại pilot nếu chưa đạt. |
| G04 | [ ] Chuẩn bị trang giới thiệu pilot nếu cần tuyển thêm | G01 | Có bảng kê mẫu, lời mời chạy thử và đầu mối liên hệ. Áp dụng Torin Kit trên trang kênh Torin, kiểm nội dung/đọc trên điện thoại và chỉ đưa số liệu đã đo. |

## Giai đoạn 1A — Bảng kê và Zalo

| Mã | Việc | Sau | Nghiệm thu và kiểm tra |
| --- | --- | --- | --- |
| F01 | [x] Tạo môi trường phát triển tối thiểu | G03 | Docker dựng cổng Go, frontend Next.js và PostgreSQL từ DB trống. CI chạy Go test, build, migration và smoke test HTTP với dữ liệu tổng hợp. |
| F02 | [ ] Đăng nhập và cô lập tổ chức | F01 | Có cô lập nhiều tổ chức, cấp tài khoản công tơ theo dãy, thu hồi phiên khi đổi quyền và nhật ký sửa quyền. Còn vai trò quản lý tài chính theo dãy và audit truy cập hỗ trợ. |
| F03 | [x] Quản lý dãy và danh sách phòng | F02 | Tạo và sửa dãy, ghim bản đồ, tìm và lọc phòng, kiểm mã phòng theo dãy, giới hạn sức chứa, xóa phòng chưa dùng và kiểm quyền theo tổ chức. |
| F04 | [ ] Ghi hợp đồng và đại diện | F03 | Có ngày vào/ra giữa tháng, tính tiền thuê theo ngày và ghi chỉ số điện nước lúc bàn giao. Còn hỗ trợ hai hợp đồng nối tiếp trong cùng tháng và phân đoạn công tơ tương ứng. |
| F05 | [ ] Chuyển dữ liệu đầu kỳ có hỗ trợ | F04 | Nhập bảng do pilot cung cấp, báo lỗi từng dòng, đối chiếu số phòng/hợp đồng/chỉ số/số dư mở. Chạy lại cùng lần nhập giữ số bản ghi; lưu biên bản đối chiếu. |
| F06 | [ ] Nhập và xác nhận điện nước | F04 | Có mốc bàn giao, thay công tơ, mã công tơ tổng, đối chiếu với tổng điện nước phòng và duyệt có lý do. Còn danh mục định danh công tơ từng phòng, tem QR và hồ sơ ca có nhiều lần thay trong một kỳ. |
| F07 | [x] Tính và tạo lô bảng kê nháp | F04, F06 | Tính kỳ đủ và kỳ lẻ bằng VND nguyên, có phí chung và phí riêng. Hàng chờ chênh lệch công tơ giữ bảng kê tại dãy bị ảnh hưởng cho đến khi chủ trọ duyệt. |
| F08 | [x] Phát hành bất biến | F07 | Bản xem trước được kiểm phiên bản trong transaction phát hành. Kiểm phát hành đồng thời, bỏ qua phòng thiếu chỉ số và khóa chỉ số của bảng kê đã phát hành. |
| F14 | [ ] Xem bảng kê riêng sau xác minh | F08, G05 | HTTPS yêu cầu email đã đăng ký cho cả xác minh và truy cập bảng kê/QR. Có kiểm tra hết hạn, chuyển tiếp và thu hồi phiên. Còn nghiệm thu gửi mã qua SMTP thực tế và đối chiếu trên thiết bị người thuê. |
| F09 | [ ] Gửi lời nhắc qua Zalo | F14 | Nội dung copy và preview không có số tiền, tên hoặc số phòng. Người nhận mở link để xác minh. Đối chiếu màn đã xác minh với bảng kê chốt trên điện thoại Android/iPhone. |

**Cổng 1A.** F05, F06–F08, F14 và F09 chạy liền với dữ liệu thử. Chủ trọ đối chiếu bảng kê theo từng phòng; kiểm luồng xác minh và thử link chuyển tiếp trước khi gửi thật.

## Giai đoạn 1B — Ghi thu, công nợ và pilot

| Mã | Việc | Sau | Nghiệm thu và kiểm tra |
| --- | --- | --- | --- |
| F10 | [x] Ghi nhận thu thủ công | F08 | Ghi thu thiếu, đủ, dư và chống trùng. Có sổ chuyển tiền dư sang bảng kê kỳ sau của cùng hợp đồng và đảo chuyển. Khoản tiền mặt chưa có bảng kê đích được ghi riêng vào hàng chờ và giữ nguyên phương thức khi phân bổ. |
| F11 | [x] Điều chỉnh có chứng từ tham chiếu | F10 | Đảo khoản thu hoặc ghi giảm bằng bút toán mới có lý do và mã chứng từ gốc. Kiểm số dư sau thu một phần, gửi đồng thời, đảo lặp và trường hợp đã đối soát ngân hàng. |
| F12 | [x] Xuất công nợ theo kỳ | F10, F11 | Có CSV công nợ và CSV khoản chưa phân bổ cho cả tiền mặt và chuyển khoản, lọc theo tháng giao dịch. Khoản đã phân bổ rời khỏi file chờ. |
| F13 | [ ] Chuẩn bị vận hành pilot | F02, F08, F10, F14 | Docker đã diễn tập chuyển sang DB phục hồi rồi quay về DB gốc, so CSV công nợ và tệp chứng từ qua ứng dụng. README có hướng dẫn giữ khóa riêng và quay lại sổ cũ. Còn nghiệm thu trên môi trường pilot, kiểm log/analytics/preview của hạ tầng triển khai và đối chiếu sổ cũ với người vận hành. |
| F15 | [ ] Hoàn thiện luồng chốt kỳ dễ thao tác | G05, F06, F09, F10 | Có tìm phòng, lọc bảng kê, che số liệu tổng và phát hành lô theo quyền chủ trọ. Playwright kiểm thao tác bàn phím, focus, tương phản và nút tối thiểu 44 px cho form phát hành lô, không cuộn ngang ở 640/320 px. Còn nghiệm thu zoom 200% toàn luồng và thao tác với người dùng mục tiêu. |
| P01 | [ ] Chạy song song hai kỳ thật | F05, F09, F12, F13, F15 | Mục tiêu 4/5 quản lý từ 55 tuổi tự hoàn thành luồng. Kiểm người thuê đọc đúng bảng kê sau xác minh. Đối chiếu từng đồng, đo thời gian khách, hỏi lại, điều chỉnh, thời gian hỗ trợ và lượt quay lại. |
| P02 | [ ] Quyết định mở bản trả phí | P01 | Đối chiếu đủ hai kỳ và giải quyết chênh lệch. Đo tiến độ so với mục tiêu 30%, ghi khách thực hiện cam kết phí và chọn vấn đề dẫn tới nhánh mở rộng đầu tiên. |

## Giai đoạn 2 — Mở rộng theo vấn đề đã đo

VietQR riêng và duyệt sao kê thủ công đã qua kiểm thử cục bộ với dữ liệu tổng hợp. Sandbox Docker có webhook mock ký HMAC, chống ghi trùng với CSV và tùy chọn tự ghép khi mã bảng kê và công nợ khớp. E01 và E04–E07 còn nghiệm thu với ứng dụng ngân hàng, contract của nhà cung cấp và sao kê thật.

| Mã | Việc | Sau | Nghiệm thu và kiểm tra |
| --- | --- | --- | --- |
| E01 | [ ] VietQR trong bảng kê riêng | P02, F14 | Chỉ thấy QR sau xác minh. Kiểm tiền còn nợ và mã đối soát sau trả một phần trên app ngân hàng pilot. |
| E02 | [x] Import CSV tự phục vụ và XLSX | P02 nếu chuyển dữ liệu lặp lại | Tải mẫu CSV hoặc XLSX, chọn dãy nhận dữ liệu, xem trước và chạy lại không nhân bản phòng. Bộ kiểm XLSX chặn công thức và kiểm giới hạn phòng của dãy. |
| E03 | [x] Sổ chi và cọc cơ bản | P02 | Có danh mục chi, chứng từ mã hóa theo tổ chức, nhận/hoàn cọc, chặn hoàn vượt số đang giữ và CSV tách tiền vào/ra. Đã đối chiếu tổng sổ, CSV, quyền chứng từ và khôi phục DB trong Docker/Playwright. |
| E07 | [ ] Thử ghép sao kê có người duyệt | P02 nếu đối soát là vấn đề ưu tiên | Thử trên mẫu ẩn danh, ghi số giao dịch ghép được, mơ hồ và đã nhập tay. So thời gian với cách cũ. Gợi ý chỉ sau khi có căn cứ; trạng thái đã thu chỉ đổi khi ghi thu có duyệt. |
| E04 | [ ] Chọn adapter và ghi giao dịch ngân hàng | E07 nếu vẫn có nhu cầu tự động | Chọn một provider cho ngân hàng pilot, có payload/test mode và định danh giao dịch. Contract test body/signature, sai tài khoản, retry và ack sau commit. Dùng Postman CLI cho collection khi môi trường thử đã sẵn sàng. |
| E05 | [ ] Duyệt ghép giao dịch và backfill | E04, F10 | Kiểm thiếu/sai mã, trả thiếu/dư, đã nhập tay, chạy đồng thời và webhook thất lạc. Lặp backfill không phát sinh giao dịch, tiền và công nợ khớp sao kê. |
| E06 | [ ] Bật auto-match theo phạm vi pilot | E05 | Chạy đối chiếu ở chế độ quan sát, chủ trọ ký xác nhận kết quả. Chỉ bật cho ca ghép xác định được và có cách tắt để quay về duyệt tay. |

## Giai đoạn 3 — Advanced và vận hành theo nhu cầu

| Mã | Việc | Sau | Nghiệm thu và kiểm tra |
| --- | --- | --- | --- |
| A01 | [ ] Timeline người ở và chia theo người | P02, có ca ở ghép | Chủ trọ chọn chia theo số ngày ở từ một kỳ cụ thể. Bảng kê đã phát hành giữ nguyên phần chia. Có kiểm thử ngày ở, số âm và làm tròn về đúng tổng. Còn theo dõi và thu theo từng người. |
| A02 | [ ] Thuê theo slot và pool điện nước | A01, có ca ký túc xá | Test slot trống, giao thời, sub-meter và phần chủ nhà chịu. Tổng phân bổ cộng phần chủ nhà bằng pool. |
| A03 | [ ] Giữ chỗ, trả trước và tất toán | E03, có nhu cầu; A01 nếu tất toán từng người | Chia đợt giữ chỗ/trả trước và tất toán. Test cấn cọc, hoàn, rời giữa kỳ và không thu lại phần đã tính. |
| A04 | [ ] Portal tenant và nghiệp vụ bàn giao | P02, có người thuê tham gia thử | Đã có màn bảng kê và thanh toán trên điện thoại sau xác minh. Còn tài khoản người thuê, lịch sử nhiều kỳ theo hợp đồng, thu hồi quyền khi rời hợp đồng, bàn giao và bảo trì. |
| A05 | [ ] Gói tuân thủ Việt Nam | P02, có người duyệt chuyên môn | Có version, ngày hiệu lực, phạm vi và nguồn. Kiểm sai nhóm, kỳ giao thời điện/VAT và rule chưa duyệt. Nhắc cư trú/PCCC, xuất số liệu thuế trước khi xem xét tự động hóa. |
| A06 | [ ] AI gợi ý chỉ số | F06, có bộ ảnh và mục đích xử lý được duyệt | Đo lỗi trên ảnh thực, xóa metadata, kiểm đúng công tơ, người duyệt xác nhận trước khi tính tiền. Đo sai số ảnh hưởng tiền trước khi đề xuất tự duyệt. |
| A07 | [ ] Chính sách phí nâng cao | A01 hoặc A02 theo ca sử dụng | Resolver theo scope, mô phỏng và publish không đổi bảng kê cũ. Kiểm fixture policy/rule giao thời. CEL chỉ thêm khi calculator có sẵn không đủ. |
| A08 | [ ] Vận hành mở rộng | F13, có nhu cầu tự cài, job hoặc tải đã đo | Release self-host có license chốt, tài liệu cài/restore; tách worker khi cần retry nền. Thử notification, adapter bổ sung, multi-node hoặc site tài liệu theo từng nhu cầu riêng. |

## Tiếp theo sau luồng chuẩn bị phòng

| Ưu tiên | Tính năng | Tiêu chí nghiệm thu |
| --- | --- | --- |
| 1 | Mẫu giá và phí dùng chung | Có xem trước và gán hàng loạt theo dãy hoặc nhóm chưa gán dãy, theo kỳ hiệu lực. Phòng đang chuẩn bị nhận bản sao giá, lịch sử phòng lưu lần áp dụng. |
| 2 | Bàn giao và trả phòng | Có xem trước và xác nhận ngày trả, chỉ số cuối, khoản cọc cấn hoặc hoàn. Phòng chuyển sang trống ngay khi xác nhận và có thể lập hợp đồng mới từ tháng sau. Còn bàn giao ngày vào giữa kỳ và hồ sơ ảnh hoặc biên bản. |
| 3 | Quản lý hợp đồng sắp hết hạn | Có danh sách 60 ngày và luồng ký lại điền sẵn kỳ bắt đầu, giữ lịch sử hợp đồng. |
| 4 | Hồ sơ nhiều người ở và nhiều tệp | Quản lý người đại diện, người ở cùng, hợp đồng chính và phụ lục có tên. Chủ trọ tải đúng tệp của đúng kỳ theo quyền tổ chức. |
