# Backlog triển khai ZentraSpace

Trạng thái khởi tạo · 26/11/2026. Cột **Sau** ghi phụ thuộc trực tiếp. Cột **Nghiệm thu và kiểm tra** mô tả bằng chứng cần có khi đóng việc. Tên thư mục và contract là dự kiến theo [kế hoạch](plan.md). Các bài kiểm là công việc triển khai tương lai; F01 tạo lệnh test, build và kiểm migration cho workspace greenfield.

**Bản chạy thử cục bộ đã có.** Docker kiểm được luồng phòng, import CSV, công tơ, bảng kê, ghi thu, xuất công nợ và link xác minh. Đã kiểm quyền tài khoản công tơ, RLS ở PostgreSQL và khôi phục bản dump dữ liệu mã hóa vào DB khác. Các ô dưới đây theo dõi mức nghiệm thu pilot đầy đủ. Thử nghiệm với chủ trọ/người thuê thật, HTTPS và kênh gửi mã tự động vẫn nằm ngoài xác minh cục bộ. [Hướng dẫn chạy và kiểm thử](../README.md)

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
| F01 | [ ] Tạo môi trường phát triển tối thiểu | G03 | Khởi động API, web và PostgreSQL từ máy sạch. Ghim dependency, có lệnh kiểm migration, test backend/frontend và build chạy trong CI. Ghi cách tạo môi trường mẫu không chứa PII. |
| F02 | [ ] Đăng nhập và cô lập tổ chức | F01 | Test API và DB với owner hai tổ chức, quản lý chỉ được cấp một tòa và nhân viên chỉ đọc/ghi công tơ. Kiểm route tổng doanh thu, CSV, ghi thu và audit các lần truy cập hỗ trợ. |
| F03 | [ ] Quản lý tòa và danh sách phòng | F02 | Tạo một tòa, lưu mã phòng duy nhất trong tòa, hiển thị và tìm phòng trên điện thoại. Test retry/validation và truy cập chéo org. |
| F04 | [ ] Ghi hợp đồng và đại diện | F03 | Tạo lease WHOLE_UNIT + SINGLE_PAYER, chỉ định PRIMARY, kỳ hiệu lực và giá đã chốt. Test phòng trống, kỳ lẻ và ca hợp đồng gối ngày bằng hai yêu cầu đồng thời. |
| F05 | [ ] Chuyển dữ liệu đầu kỳ có hỗ trợ | F04 | Nhập bảng do pilot cung cấp, báo lỗi từng dòng, đối chiếu số phòng/hợp đồng/chỉ số/số dư mở. Chạy lại cùng lần nhập giữ số bản ghi; lưu biên bản đối chiếu. |
| F06 | [ ] Nhập và xác nhận điện nước | F04 | Ghi mốc đọc, số trước/sau theo phòng, người duyệt, căn cứ đơn giá và mã công tơ. Thử số lùi, thay công tơ và chênh giữa meter tổng/meter phòng. Gắn lý do khi giữ khoản cần đối chiếu. |
| F07 | [ ] Tính và tạo lô bảng kê nháp | F04, F06 | Kiểm fixture kỳ đủ/kỳ lẻ, tiền thuê, điện nước và phí chung hợp lệ ở dòng riêng. Tính VND nguyên, tổng khớp từng đồng. Chỉ số thiếu hoặc chênh lệch chưa giải thích giữ nháp; chạy lại không sinh nháp trùng. |
| F08 | [ ] Phát hành bất biến | F07 | Chủ trọ xem trước, chốt từng bảng kê và ghi công nợ cùng transaction. Test phát hành lại/đồng thời, phòng thiếu đầu vào, và thử sửa dòng hoặc tổng sau khóa. |
| F14 | [ ] Xem bảng kê riêng sau xác minh | F08, G05 | Link có hạn/thu hồi và cần xác minh qua kênh đã đăng ký. Test chuyển tiếp, đọc sai phòng, hết hạn và thu hồi khi đã mở phiên. Người thuê đã xác minh đọc được mốc điện nước, cách tính, tổng và hạn trên web di động. |
| F09 | [ ] Gửi lời nhắc qua Zalo | F14 | Nội dung copy và preview không có số tiền, tên hoặc số phòng. Người nhận mở link để xác minh. Đối chiếu màn đã xác minh với bảng kê chốt trên điện thoại Android/iPhone. |

**Cổng 1A.** F05, F06–F08, F14 và F09 chạy liền với dữ liệu thử. Chủ trọ đối chiếu bảng kê theo từng phòng; kiểm luồng xác minh và thử link chuyển tiếp trước khi gửi thật.

## Giai đoạn 1B — Ghi thu, công nợ và pilot

| Mã | Việc | Sau | Nghiệm thu và kiểm tra |
| --- | --- | --- | --- |
| F10 | [ ] Ghi nhận thu thủ công | F08 | Nhập tiền mặt/chuyển khoản và mã chứng từ khi có. Test thiếu, đủ, dư, nhấn lặp và hai lượt thu đồng thời. Số dư chưa phân bổ truy lại được. |
| F11 | [ ] Điều chỉnh có chứng từ tham chiếu | F10 | Sửa sai bằng chứng từ giảm/bổ sung hoặc bút toán đảo, có lý do và audit. Test sau phát hành, sau thu một phần và khi đảo lại. |
| F12 | [ ] Xuất công nợ theo kỳ | F10, F11 | Hiện từng phòng đã thu/còn thiếu và khoản chưa phân bổ. So CSV với sổ phụ và UI trên cùng kỳ, test dữ liệu có ký tự tiếng Việt. |
| F13 | [ ] Chuẩn bị vận hành pilot | F02, F08, F10, F14 | Diễn tập restore DB/tệp đã mã hóa và khóa cất riêng. Kiểm truy cập sai quyền, log/analytics/cache/preview không chứa số liệu chia sẻ, và rollback về sổ cũ. |
| F15 | [ ] Hoàn thiện luồng chốt kỳ dễ thao tác | G05, F06, F09, F10 | Ba thao tác chính dễ thấy, nút có nhãn ≥44×44, zoom 200%, kiểm tương phản và focus. Thử tìm phòng, lọc và xử lý lô theo quyền. Che số liệu tổng dùng được khi mở màn hình chung. |
| P01 | [ ] Chạy song song hai kỳ thật | F05, F09, F12, F13, F15 | Mục tiêu 4/5 quản lý từ 55 tuổi tự hoàn thành luồng. Kiểm người thuê đọc đúng bảng kê sau xác minh. Đối chiếu từng đồng, đo thời gian khách, hỏi lại, điều chỉnh, thời gian hỗ trợ và lượt quay lại. |
| P02 | [ ] Quyết định mở bản trả phí | P01 | Đối chiếu đủ hai kỳ và giải quyết chênh lệch. Đo tiến độ so với mục tiêu 30%, ghi khách thực hiện cam kết phí và chọn vấn đề dẫn tới nhánh mở rộng đầu tiên. |

## Giai đoạn 2 — Mở rộng theo vấn đề đã đo

VietQR riêng và duyệt sao kê thủ công đã qua kiểm thử cục bộ với dữ liệu tổng hợp. Giữ E01 và E07 ở trạng thái chờ nghiệm thu thực tế cho tới khi quét bằng ứng dụng ngân hàng pilot và đo thời gian đối soát. Các adapter nhận webhook vẫn thuộc bước sau.

| Mã | Việc | Sau | Nghiệm thu và kiểm tra |
| --- | --- | --- | --- |
| E01 | [ ] VietQR trong bảng kê riêng | P02, F14 | Chỉ thấy QR sau xác minh. Kiểm tiền còn nợ và mã đối soát sau trả một phần trên app ngân hàng pilot. |
| E02 | [ ] Import CSV tự phục vụ và XLSX khi cần | P02 nếu chuyển dữ liệu lặp lại | Người quản lý tải mẫu, dry-run và xem lỗi từng dòng, commit lại không trùng. XLSX có bài kiểm giới hạn và định dạng. |
| E03 | [ ] Sổ chi và cọc cơ bản | P02 nếu có yêu cầu lặp lại | Có loại chi, chứng từ private, tiền cọc nhận/hoàn và báo cáo tách thu, chi, cọc. Kiểm đối chiếu giao dịch tiền với bút toán và file xuất. Có thể phát hành sổ chi và cọc độc lập. |
| E07 | [ ] Thử ghép sao kê có người duyệt | P02 nếu đối soát là vấn đề ưu tiên | Thử trên mẫu ẩn danh, ghi số giao dịch ghép được, mơ hồ và đã nhập tay. So thời gian với cách cũ. Gợi ý chỉ sau khi có căn cứ; trạng thái đã thu chỉ đổi khi ghi thu có duyệt. |
| E04 | [ ] Chọn adapter và ghi giao dịch ngân hàng | E07 nếu vẫn có nhu cầu tự động | Chọn một provider cho ngân hàng pilot, có payload/test mode và định danh giao dịch. Contract test body/signature, sai tài khoản, retry và ack sau commit. Dùng Postman CLI cho collection khi môi trường thử đã sẵn sàng. |
| E05 | [ ] Duyệt ghép giao dịch và backfill | E04, F10 | Kiểm thiếu/sai mã, trả thiếu/dư, đã nhập tay, chạy đồng thời và webhook thất lạc. Lặp backfill không phát sinh giao dịch, tiền và công nợ khớp sao kê. |
| E06 | [ ] Bật auto-match theo phạm vi pilot | E05 | Chạy đối chiếu ở chế độ quan sát, chủ trọ ký xác nhận kết quả. Chỉ bật cho ca ghép xác định được và có cách tắt để quay về duyệt tay. |

## Giai đoạn 3 — Advanced và vận hành theo nhu cầu

| Mã | Việc | Sau | Nghiệm thu và kiểm tra |
| --- | --- | --- | --- |
| A01 | [ ] Timeline người ở và chia theo người | P02, có ca ở ghép | Version policy và snapshot occupancy theo ngày. Golden test người vào/ra, phần âm, làm tròn; tổng share bằng bảng kê. |
| A02 | [ ] Thuê theo slot và pool điện nước | A01, có ca ký túc xá | Test slot trống, giao thời, sub-meter và phần chủ nhà chịu. Tổng phân bổ cộng phần chủ nhà bằng pool. |
| A03 | [ ] Giữ chỗ, trả trước và tất toán | E03, có nhu cầu; A01 nếu tất toán từng người | Chia đợt giữ chỗ/trả trước và tất toán. Test cấn cọc, hoàn, rời giữa kỳ và không thu lại phần đã tính. |
| A04 | [ ] Portal tenant và nghiệp vụ bàn giao | P02, có người thuê tham gia thử | Test quyền share, mã đăng nhập một lần, rời lease mất quyền, biên bản/bằng chứng và luồng yêu cầu bảo trì. Triển khai từng phần theo nhu cầu. |
| A05 | [ ] Gói tuân thủ Việt Nam | P02, có người duyệt chuyên môn | Có version, ngày hiệu lực, phạm vi và nguồn. Kiểm sai nhóm, kỳ giao thời điện/VAT và rule chưa duyệt. Nhắc cư trú/PCCC, xuất số liệu thuế trước khi xem xét tự động hóa. |
| A06 | [ ] AI gợi ý chỉ số | F06, có bộ ảnh và mục đích xử lý được duyệt | Đo lỗi trên ảnh thực, xóa metadata, kiểm đúng công tơ, người duyệt xác nhận trước khi tính tiền. Đo sai số ảnh hưởng tiền trước khi đề xuất tự duyệt. |
| A07 | [ ] Chính sách phí nâng cao | A01 hoặc A02 theo ca sử dụng | Resolver theo scope, mô phỏng và publish không đổi bảng kê cũ. Kiểm fixture policy/rule giao thời. CEL chỉ thêm khi calculator có sẵn không đủ. |
| A08 | [ ] Vận hành mở rộng | F13, có nhu cầu tự cài, job hoặc tải đã đo | Release self-host có license chốt, tài liệu cài/restore; tách worker khi cần retry nền. Thử notification, adapter bổ sung, multi-node hoặc site tài liệu theo từng nhu cầu riêng. |
