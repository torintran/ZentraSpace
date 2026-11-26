# Kế hoạch triển khai feature ZentraSpace

**26/11/2026 · Căn cứ Blueprint v3.4**

ZentraSpace bắt đầu với chủ trọ 10–30 phòng đang dùng Excel hoặc sổ tay và gửi bảng kê qua Zalo. Đơn vị triển khai đầu tiên là một tòa, một người đại diện trả tiền mỗi phòng. Kết quả cần đạt là chủ trọ có thể tự chốt bảng kê, gửi cho người thuê, ghi nhận thu và biết phòng nào còn nợ.

Đọc cùng [Blueprint v3.4](../docs/ZENTRASPACE-BLUEPRINT.md) và [backlog triển khai](todo.md). Lịch giả định hai kỹ sư toàn thời gian và một người phụ trách pilot. Thời gian chờ lịch thu, khách và nhà cung cấp tính riêng.

## Lộ trình và quyết định chuyển giai đoạn

| Giai đoạn | Feature và kết quả cần bàn giao | Ước lượng sơ bộ | Cổng quyết định |
| --- | --- | --- | --- |
| 0. Xác thực nhu cầu | So với cách khách đang dùng trên cùng nhóm phòng, thử người thuê tự kiểm bảng kê, chạy một kỳ có hỗ trợ và thử đề nghị phí | 1–2 tuần chuẩn bị, sau đó theo lịch thu thật | Ít nhất 3 chủ trọ gặp cùng vướng mắc, 2 hoàn thành một kỳ thử, 1 đồng ý đề nghị trả phí cụ thể. Xác định bước gây tốn thời gian và chi phí hỗ trợ. |
| 1A. Từ dữ liệu đầu kỳ đến bảng kê | Đăng nhập, nhập một tòa, tiền thuê/điện nước có căn cứ, lô nháp, duyệt và gửi bảng kê riêng qua Zalo | Ước lượng lại sau thử luồng xác minh và phân quyền theo vai trò | Diễn giải khoản điện nước đến từng đồng, giữ chênh lệch chưa rõ ở nháp, phát hành lại không sinh khoản thu trùng; bản xem trước Zalo và link chuyển tiếp không lộ số tiền. |
| 1B. Thu tiền và chuyển sang bản trả phí | Thu tay, số còn nợ, điều chỉnh có lịch sử, xuất công nợ, bàn giao vận hành và chạy song song | 2–3 tuần xây dựng sau 1A, cộng hai kỳ thu thật để quyết định chuyển sổ | Đối chiếu đủ hai kỳ, xử lý mọi chênh lệch, có khách quay lại và ghi nhận thực hiện cam kết trả phí. Mục tiêu đề xuất giảm 30% thời gian trên khâu đã đo. |
| 2. Đối soát và mở rộng có bằng chứng | Nếu đối soát là ưu tiên, thử ghép sao kê có người duyệt, đo phần còn tốn công rồi quyết định webhook. VietQR có thể thử riêng trong trang bảng kê đã xác minh. | 2–4 tuần riêng cho một adapter nếu được chọn, sau khi có tài khoản thử, giao dịch mẫu và contract | Số liệu khớp sao kê, nhập tay và quy trình thử không ghi trùng. Đo thời gian ghép trước và sau, xác định nhu cầu theo từng nhánh. |
| 3. Advanced và mở rộng vận hành | Ở ghép, slot, trả trước, tất toán, tenant portal, tuân thủ, AI và self-host theo nhu cầu | Ước lượng từng nhánh sau dữ liệu thực tế | Mỗi nhánh có khách hoặc người vận hành xác nhận nhu cầu, ca nghiệm thu và người phụ trách. |

Nếu cổng 0 cho thấy phần tốn thời gian nhất khác luồng dự kiến, cập nhật backlog và kiểm chứng lại bằng ca thật. Trong hai kỳ chạy song song, giữ nguồn đối chiếu của chủ trọ và xuất dữ liệu sau mỗi kỳ. Chuyển sổ chính khi người phụ trách xác nhận số tiền, dữ liệu và quy trình khôi phục.

Khi đo pilot, ghi riêng thời gian thao tác của chủ trọ, số lần người thuê hỏi lại theo nhóm lý do, số phòng phải điều chỉnh, thời gian hỗ trợ của nhóm và số tiền đối chiếu chênh. Cùng một nhóm phòng và khâu làm việc là cơ sở so trước/sau. VOZ và Reddit gợi ý ca cần thử; dữ liệu pilot quyết định phạm vi và giá. [Tổng hợp nguồn và quyết định trong blueprint](../docs/ZENTRASPACE-BLUEPRINT.md#tín-hiệu-từ-voz-và-reddit)

## Thiết kế cho bản đầu

- Một luồng web tiếng Việt dùng được trên điện thoại. Dùng Go/PostgreSQL cho nghiệp vụ và Next.js cho web theo blueprint. Giữ một môi trường pilot với DB có sao lưu. Chỉ thêm worker và queue khi công việc chạy nền cần retry hoặc tách khỏi request.
- Có ranh giới org từ migration đầu và kiểm quyền với hai bộ dữ liệu trước khi cấp quyền dùng thật. Thực hiện từng lát dọc qua dữ liệu, dịch vụ và UI. Bảng kê đã phát hành bất biến, thu và điều chỉnh ghi vào sổ phụ; số dư tính lại được từ bút toán.
- Mẫu thử màn hình chính, màn nhập chỉ số và màn người thuê thuộc giai đoạn 0. Đo với người từ 55 tuổi và người dùng quen công nghệ. Mặc định ưu tiên việc cần làm và che số liệu tổng. Thao tác hàng loạt mở theo vai trò tại màn phù hợp. Mục tiêu 4/5 người quản lý lớn tuổi tự hoàn thành luồng và giảm 30% thời gian với cùng dữ liệu. [Tiêu chí sản phẩm](../docs/ZENTRASPACE-BLUEPRINT.md#điểm-khác-biệt-cần-chứng-minh)
- Mã hóa khi truyền và lưu, quản lý khóa tách biệt, kiểm quyền ở API/DB. Gửi lời nhắc Zalo không kèm số tiền, người nhận xác minh qua kênh đã đăng ký trước khi đọc. Thử xác minh trên điện thoại pilot và giá nhà cung cấp trước khi chốt lại lịch.
- Định nghĩa giao diện nghiệp vụ trước khi chia việc. Dữ liệu đầu kỳ cung cấp phòng, hợp đồng, đại diện và chỉ số gốc. Tính phí nhận kỳ, hợp đồng, chỉ số và cấu hình đã duyệt; trả dòng phí, tổng và ngoại lệ. Phát hành nhận revision nháp và khóa thao tác; trả bảng kê đã khóa. Ghi thu/điều chỉnh nhận khóa thao tác và chứng từ gốc; trả giao dịch cùng số dư.
- Torin Brand Kit là tài liệu tham chiếu cho trang giới thiệu trên kênh Torin. Nếu cần trang mời pilot, dùng cách viết cụ thể, font Be Vietnam Pro và màu theo kit, với lời mời “Đăng ký chạy thử một kỳ thu”. Các SVG Torin và TT phục vụ nhận diện Torin; thương hiệu sản phẩm ZentraSpace sẽ chốt riêng. Torin Brand Kit v1.0 là tài liệu nội bộ.

## Thứ tự và cách chia việc

Sau G01, chạy kỳ có hỗ trợ G02 và thử mẫu/xác minh G05 trước quyết định G03. Từ F08, hai nhánh có thể chạy song song: **F14 → F09** cho bảng kê riêng và **F10 → F11 → F12** cho thu và xuất công nợ. Nhập dữ liệu đầu kỳ, kiểm luồng dễ thao tác và chuẩn bị vận hành đều hoàn tất trước P01. Một người tích hợp migration và thứ tự khóa sổ theo lượt để tránh xung đột.

| Chặng | Phần việc nên nhận trước | Kết quả kiểm tại cuối chặng |
| --- | --- | --- |
| Nền tảng | Ca mẫu thật đã ẩn danh, môi trường chạy, tổ chức/phòng/hợp đồng | Setup từ máy sạch, hai org tách quyền, một tòa nhập đủ dữ liệu. |
| Bảng kê | Công tơ, tính phí, batch nháp, phát hành, gửi lời nhắc Zalo | Từ đầu kỳ tới bản ghi đã khóa; người nhận xác minh trước khi thấy số liệu. |
| Thu và pilot | Thu tay, điều chỉnh, sổ nợ, backup/export | Thu thiếu/đủ/dư và đảo chứng từ khớp sổ; chạy song song hai kỳ. |
| Mở rộng | Nhánh phù hợp với dữ liệu pilot | Qua bộ ca của nhánh trước khi bật cho người dùng thật. |

## Rủi ro cần xử lý trong kế hoạch

| Rủi ro | Cách kiểm và xử lý |
| --- | --- |
| Xây nhầm bước chủ trọ ít cần | Đo từng bước nhập, chốt, gửi và đối soát. Ghi số lần lặp, thời gian hỗ trợ và cam kết trả phí. |
| Dữ liệu đầu kỳ thiếu hoặc có công nợ cũ | Báo từng dòng lỗi, xác nhận chỉ số gốc và công nợ mở với chủ trọ. Chặn phát hành dòng chưa đủ dữ liệu. |
| Người thuê hỏi lại về điện hoặc chi phí chung | Cho đọc thử bảng kê trên điện thoại, ghi lý do hỏi lại. Đối chiếu chỉ số, mốc đọc, căn cứ tính và dòng phí chung trước khi phát hành. |
| Lệch tiền hoặc ghi trùng | Fixture từng đồng, test retry/đồng thời, nhật ký điều chỉnh, đối chiếu độc lập trước khi chuyển sổ chính. |
| Lộ dữ liệu phòng hoặc người thuê | Ma trận quyền, dữ liệu tối thiểu, kiểm link chia sẻ nếu triển khai, kiểm khôi phục trước khi chạy thật. |
| Chậm tích hợp ngân hàng hoặc quy tắc thay đổi | Giữ luồng thủ công có đối chiếu; adapter và rule chỉ bật sau khi kiểm contract, căn cứ áp dụng và vận hành thực. |

Các quyết định còn mở gồm tòa thí điểm, thời điểm kỳ thu thật, mức phí thử, người duyệt số tiền và máy chủ chạy pilot. Chúng thuộc công việc giai đoạn 0–1, với người phụ trách và cách xác nhận ghi trong backlog.
