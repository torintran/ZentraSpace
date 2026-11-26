# ZentraSpace — Blueprint sản phẩm và kiến trúc

**v3.4 đề xuất · Vietnam-first · 26/11/2026**

Hợp nhất kiến trúc v2.1 và v3.0, lấy nhu cầu của chủ trọ làm căn cứ quyết định phạm vi xây dựng.

## 1. Khách hàng và bài toán đầu tiên

**Bắt đầu bằng một kỳ thu thật cùng chủ trọ.** Phân khúc thử nghiệm đề xuất là chủ trọ tự thu tiền hoặc có một người phụ trách, khoảng **10–30 phòng**, đang dùng Excel/sổ tay và gửi tiền phòng qua Zalo. Chủ trọ là người quyết định trả tiền. Người quản lý cần biết phòng nào đã trả và còn thiếu bao nhiêu. Người thuê cần một bảng kê dễ hiểu trên điện thoại.

**Lời hứa để thử với khách.** “Chốt tiền trọ trên điện thoại. Người thuê tự kiểm được tiền phòng mình. Số liệu cả tòa chỉ hiện với người có quyền.” Bắt đầu từ danh sách phòng khách đã có. Người quản lý nhập chỉ số, duyệt bảng kê, gửi lời nhắc qua Zalo và ghi tiền đã nhận. Bảng kê là khoản cần trả, phiếu thu là chứng từ khoản đã nhận.

### Điểm khác biệt cần chứng minh

EasyTro đã công bố giao diện chữ lớn, nút lớn và nhập chỉ số bằng giọng nói. ProHome đã công bố giao diện dễ dùng và gửi QR qua Zalo. ZentraSpace sẽ cạnh tranh bằng kết quả của cả một kỳ thu, đo trên cùng nhóm phòng và cùng bộ tác vụ với công cụ khách đang dùng. [EasyTro](https://easytro.net/) · [ProHome](https://prohome.com.vn/chu-tro)

| Lợi thế ZentraSpace đặt mục tiêu | Tiêu chí nghiệm thu pilot |
| --- | --- |
| **Người quản lý lớn tuổi tự chốt được kỳ.** Màn hình chính ưu tiên Nhập điện nước, Gửi bảng kê, Ghi đã thu. Hiện phòng cần xử lý, dùng tiếng Việt quen thuộc, nút lớn có nhãn và xác nhận rõ trước khi phát hành. | Mời tối thiểu 5 người quản lý từ 55 tuổi trở lên. Ít nhất 4/5 tự hoàn thành nhập chỉ số, gửi bảng kê và ghi thu trên điện thoại sau phần giới thiệu ngắn, không cần người thử thao tác hộ. Mục tiêu giảm 30% thời gian xử lý trên cùng nhóm phòng so với cách hiện tại và không ghi sai tiền. |
| **Giao diện hiện đại cho người dùng quen công nghệ.** Cùng một giao diện responsive, có trạng thái bằng chữ, lưu nháp, và mở thêm tìm phòng, bộ lọc, nhập liên tiếp hoặc thao tác hàng loạt ở đúng màn hình. Người thuê xem trên web di động sau xác minh, thấy cách tính và số còn phải trả, không cần cài app. | Thử với ít nhất 3 người dùng quen công nghệ, gồm người quản lý và người thuê. Kiểm luồng tìm phòng, xử lý nhiều phòng ở vai trò quản lý và khả năng đọc đúng số tiền, giải thích khoản điện/nước ở vai trò người thuê. |
| **Riêng tư có thể kiểm chứng.** Mặc định tin nhắn Zalo chỉ mang lời nhắc và link, chế độ Che số tiền che chỉ số tổng trên màn hình và người ghi công tơ không có quyền xem doanh thu. | Mọi ca thử sai tổ chức, sai tòa, sai phòng, link chuyển tiếp, link đã thu hồi và bản xem trước Zalo đều không trả số liệu ngoài quyền. Người có quyền vẫn thực hiện được việc của họ. |

Mục tiêu giao diện là WCAG 2.2 AA cho các luồng chính. Nút thao tác chính tối thiểu 44×44 CSS px theo chuẩn nội bộ, chữ nội dung từ 16 px và hỗ trợ phóng to 200%. Kiểm tương phản, bàn phím, focus rõ, thông báo lỗi cạnh trường nhập và trạng thái đọc được bằng chữ. Kiểm trên Android và iPhone phổ thông. [WCAG 2.2](https://www.w3.org/TR/WCAG22/)

### Tín hiệu từ VOZ và Reddit

Các thảo luận công khai là nguồn để chọn ca thử. Người đăng tự chọn kể trải nghiệm, một số bình luận giới thiệu sản phẩm, và nhóm quốc tế có quy trình thu khác Việt Nam. Chưa có dữ liệu đủ để suy ra tỷ lệ gặp vấn đề, mức giá khách sẽ trả hay kết luận pháp lý.

| Tín hiệu quan sát | Thay đổi cho ZentraSpace |
| --- | --- |
| Một chủ đề VOZ hỏi tìm app quản lý trọ còn hoạt động và được hỗ trợ, sau khi liên hệ hotline/Zalo ở vài ứng dụng chưa có phản hồi. Một người nêu ngân sách khoảng 200.000 đồng/tháng. | Pilot có người phụ trách và khung hỗ trợ vào ngày chốt. Đo thời gian hỗ trợ trên mỗi tòa và thử đề nghị phí riêng với khách. Mức 200.000 đồng là một trường hợp tự khai. [VOZ · Tư vấn phần mềm quản lý phòng trọ](https://voz.vn/t/tu-van-phan-mem-quan-ly-phong-tro.1064105/) |
| Người đăng trên VOZ mô tả chênh lệch giữa chỉ số công tơ phòng và hóa đơn. Người thuê trên r/VietNam hỏi cách xin chứng từ điện; r/vozforums có trao đổi bất đồng về điện phòng và chi phí chung. | Đưa khả năng kiểm lại số tiền vào bản đầu. Bảng kê ghi kỳ, mốc đầu/cuối, sản lượng, căn cứ tính và dòng phí chung riêng nếu phát sinh hợp lệ. Nội bộ phân biệt meter tổng/meter phòng, giữ khoản có chênh lệch chưa giải thích ở nháp. Có thể bổ sung ảnh hoặc chứng từ khi có nhu cầu, qua quyền truy cập phù hợp. [VOZ · Tranh luận tiền điện](https://voz.vn/t/cai-lon-vs-nguoi-thue-vu-tien-dien.1266635/) · [r/VietNam · Electricity bills](https://www.reddit.com/r/VietNam/comments/1ldpw71/electricity_bills/) · [r/vozforums · Điện và nước nhà trọ](https://www.reddit.com/r/vozforums/comments/1tr1utl/) |
| Trong các chủ đề r/Landlord, vài người thấy bảng tính đủ dùng; người khác gặp khó với việc chép giao dịch ngân hàng, dữ liệu phân tán hoặc công cụ quá phức tạp. Cũng có người hài lòng với ứng dụng họ đang dùng. | So với Excel/sổ và công cụ hiện tại. Giữ xuất CSV và chuyển dữ liệu có hỗ trợ. Nếu ghép tiền tốn nhiều thời gian, thử so khớp sao kê có người duyệt trước khi đầu tư webhook. [r/Landlord · Quay lại bảng tính](https://www.reddit.com/r/Landlord/comments/1u4x1yk/landlord_usny_anyone_else_ditch_the_property/) · [r/Landlord · Công cụ đang dùng](https://www.reddit.com/r/Landlord/comments/1spks9j/general_uspa_what_softwaretools_do_you_actually/) |
| EasyTro công bố quản lý cơ bản miễn phí tối đa 20 phòng và gửi bảng kê qua Zalo. ProHome công bố gói miễn phí một khu, chuyển từ Excel và gửi bảng kê kèm QR qua Zalo. Đây là tính năng do nhà cung cấp mô tả. | Dùng Zalo để phù hợp thói quen hiện tại. Thử giá trị khác biệt bằng thời gian chốt và giải thích khoản thu, số lần người thuê phải hỏi lại, khả năng xuất dữ liệu và mức hỗ trợ thực tế. [EasyTro](https://app.easytro.vn/) · [ProHome](https://prohome.com.vn/chu-tro) |

### Cách tiếp cận và kiểm chứng

1. **Xem khách chốt tiền một kỳ.** Mời 5–8 chủ trọ thuộc cùng phân khúc qua giới thiệu hoặc nhóm chủ trọ, bằng lời mời tìm hiểu cách họ làm hiện nay. Xin xem bảng kê, tin nhắn và ghi chép đã ẩn danh. Đo thời gian theo từng việc, chỗ phải nhập lại, lỗi và số lần người thuê hỏi lại. Ghi nguyên văn vấn đề họ gặp.
2. **Mời chạy thử có người hỗ trợ.** Trên cùng nhóm phòng, so cách đang dùng với bảng mẫu và bảng kê copy sang Zalo. Cho người thuê kiểm thử xem có đọc được vì sao phải trả số tiền đó. Chủ trọ duyệt mọi khoản trước khi gửi. Thử ca meter phòng lệch meter tổng, ca trả thiếu hoặc chưa rõ nội dung chuyển khoản nếu có. Đo riêng thời gian khách thao tác và thời gian hỗ trợ. Hỏi khách có muốn dùng kỳ sau và đồng ý với đề nghị phí cụ thể hay không.
3. **Xây phần việc đã có bằng chứng.** Mục tiêu đề xuất để bắt đầu xây là ít nhất 3 chủ trọ có cùng vướng mắc, 2 người hoàn thành một kỳ có hỗ trợ và ít nhất 1 người đồng ý với một đề nghị trả phí cụ thể. Mục tiêu sau khi có bản đầu là giảm 30% thời gian cho khâu đã đo, đối chiếu từng đồng và có khách thực sự quay lại kỳ tiếp theo. Ghi riêng lời khen, cam kết và hành vi thực tế. Nếu khách ngại chuyển dữ liệu hoặc vấn đề chính là đối soát chuyển khoản, thử điều chỉnh luồng trước khi mở rộng.

### Phạm vi bản đầu và điều kiện mở rộng

| Bản đầu | Khi nào thêm |
| --- | --- |
| Một tòa, nhập danh sách phòng từ bảng khách có, tiền thuê và điện nước theo căn cứ đã xác nhận | Thêm import tự phục vụ khi việc chuyển dữ liệu lặp lại ở nhiều khách. |
| Một bảng kê nháp mỗi phòng/kỳ, có cách tính điện nước và duyệt trước phát hành. Gửi lời nhắc Zalo kèm link xem riêng có xác minh. | Thử VietQR nếu khó trả đúng tiền hoặc đúng nội dung chuyển. Mở portal có lịch sử nhiều kỳ khi có nhu cầu. |
| Ghi tiền đã nhận, còn thiếu, lưu lịch sử điều chỉnh và xuất danh sách công nợ | Nếu ghép giao dịch là bước tốn thời gian lặp lại, thử sao kê có người duyệt. Thêm webhook khi lượng việc còn lại và điều kiện kết nối chứng minh nhu cầu. |
| Luồng có người hướng dẫn, dùng trên điện thoại và có hỗ trợ khi chốt kỳ | Thêm chi phí, cọc phức tạp, ở ghép, portal tenant, AI và tùy chọn self-host theo yêu cầu lặp lại đã có ví dụ thực tế. |

### Cách giới thiệu theo Torin Kit

Torin Kit là cẩm nang thương hiệu cá nhân. Đề xuất giới thiệu **ZentraSpace là một dự án của Torin Tran** trên trang dự án của Torin. Dùng giọng gần gũi, rõ và cụ thể theo kit. Trang giới thiệu kể việc chủ trọ đang làm, cho xem một bảng kê mẫu và mời **“Đăng ký chạy thử một kỳ thu”**. Chỉ công bố kết quả tiết kiệm thời gian sau khi đã đo, ghi rõ quy mô và phạm vi thử.

> Mình đang thử ZentraSpace cùng các chủ trọ để chốt tiền phòng và theo dõi khoản còn thiếu. Nếu anh/chị đang gửi bảng kê qua Zalo, mình có thể cùng anh/chị thử trên vài phòng trong kỳ tới. Anh/chị xem và duyệt số tiền trước khi gửi cho người thuê.

Nếu làm trang giới thiệu, có thể tham chiếu Be Vietnam Pro, xanh ngọc #103D35, trắng ngà #F7F3E8, vàng nhấn #C9A35F và chữ #14242B. Logo và favicon TT dùng cho phần nhận diện Torin trên kênh của Torin. Nhận diện sản phẩm ZentraSpace sẽ được chốt riêng khi cần. Torin Brand Kit v1.0 là tài liệu nội bộ.

### Thiết kế kỹ thuật theo nhu cầu

Các mục kiến trúc bên dưới mô tả hướng mở rộng khi khách đã dùng. Với bản đầu, ưu tiên một ứng dụng triển khai gọn, DB có sao lưu, bảo vệ dữ liệu từng khách, lịch sử tiền kiểm được và xuất dữ liệu. Tách worker khi tác vụ nền thực sự cần. Workspace chưa có runtime để kiểm chứng. Phân khúc, mức phí, cổng tiếp tục và lịch ở đây là giả thuyết hoặc mục tiêu đề xuất.

| Quyết định đề xuất | Lý do và hệ quả |
| --- | --- |
| Lấy Standard của v3.0 làm mặc định, giữ mô hình nhiều người và policy phiên bản hóa của v2.1 | Tạo lease với một PRIMARY, một share và mẫu phí sẵn. Advanced mở thêm nghiệp vụ trên cùng mô hình dữ liệu. |
| Modular monolith Go, có thể tách API và worker khi cần | Giữ ghi chứng từ và sổ trong transaction PostgreSQL. Khi có queue, enqueue cùng transaction. |
| Thu thủ công trước, webhook theo bằng chứng | Pilot thử trọn chu kỳ. Chọn adapter theo ngân hàng của khách khi chi phí đối soát tay đủ lớn. |
| Sổ phụ append-only | Chứng từ và bút toán có nguồn gốc. Số tiền đã trả và số dư là dữ liệu dẫn xuất có thể tái tính. |
| Kho tệp riêng qua adapter | Volume riêng có thể dùng cho pilot một node nếu đã thử khôi phục. S3-compatible là đường mở rộng. MinIO Community đã được lưu trữ và chuyển sang phát hành mã nguồn, nên bỏ khỏi Compose mặc định. [MinIO Community](https://github.com/minio/minio) |
| Rule Việt Nam có ngày hiệu lực và người duyệt | Giữ khả năng tái tạo theo kỳ. Quy tắc chưa xác minh đưa khoản thu liên quan vào hàng chờ duyệt. |
| Apache-2.0 | Giấy phép phát hành đã được chủ dự án chọn để hỗ trợ tích hợp và tái sử dụng. [Apache License 2.0](../LICENSE) |

## 2. Kiến trúc đích khi mở rộng

~~~mermaid
flowchart LR
    U["Chủ trọ và người thuê"] --> W["Next.js web"]
    W --> A["Go API"]
    B["Ngân hàng hoặc provider"] --> A
    A --> D[("PostgreSQL + River")]
    D --> J["Go worker"]
    A --> S["Kho tệp riêng"]
    J --> S
    J --> N["Email, Telegram, webhook hoặc in-app"]
~~~

- Go, chi, pgx/v5, sqlc và goose. Ghim Go 1.27.x cho môi trường xây dựng tại thời điểm thiết kế, kiểm tra thêm Go 1.26.x trong CI nếu dependency tương thích. PostgreSQL 16+ dùng btree_gist, UUIDv7 do ứng dụng sinh để tương thích PostgreSQL 16. Mốc thời gian lưu UTC, ranh giới ngày theo timezone property, mặc định Asia/Ho_Chi_Minh. [Go Release History](https://go.dev/doc/devel/release) · [PostgreSQL btree_gist](https://www.postgresql.org/docs/16/btree-gist.html)
- Next.js App Router standalone, Tailwind, next-intl vi/en. Ghim Next.js và Node.js thuộc các nhánh LTS tương thích. Ưu tiên thao tác trên điện thoại và nhập chỉ số theo danh sách phòng. Ảnh công tơ chụp bằng camera sau khi thiết bị hỗ trợ, resize cạnh dài ≤ 1280 px. PWA triển khai sau luồng chính, chỉ cache tài nguyên giao diện công khai. [Next.js Support Policy](https://nextjs.org/support-policy) · [Node.js Releases](https://nodejs.org/en/about/previous-releases)
- River InsertTx enqueue cùng transaction nghiệp vụ. Worker chịu được retry và xử lý trùng. Hiệu ứng ngoài hệ thống dùng khóa chống gửi lặp nếu adapter hỗ trợ. Redis/Valkey là profile tùy chọn khi cần rate limit hoặc cache phân tán. [River transactional enqueueing](https://riverqueue.com/docs/transactional-enqueueing)
- Compose một node gồm reverse proxy TLS, API, worker, web, PostgreSQL và kho tệp theo cấu hình. OIDC là cấu hình cài đặt, có thể cung cấp profile IdP cho self-host. Multi-node cần kho tệp dùng chung, session nhất quán, cấu hình cache Next.js và ngân sách kết nối DB. Kubernetes và apps/site triển khai khi có nhu cầu.
- Monorepo gồm apps/api, apps/worker, apps/web, thư mục Go internal dùng chung cho API/worker, packages/contracts, packages/policy-presets, packages/jurisdictions/vn, migrations, deploy, docs và .github. REST có OpenAPI; cấu hình calculator/split/import có JSON Schema. CI kiểm migration trên DB sạch, tài chính, ma trận quyền, contract adapter và chuỗi cung ứng.

## 3. Mô hình nghiệp vụ và dữ liệu

| Mô hình | Cấu hình | Luồng thu |
| --- | --- | --- |
| Thuê cả phòng, một người đại diện | WHOLE_UNIT + SINGLE_PAYER | Một lease, một PRIMARY, một share. Đây là mặc định. |
| Đồng thuê cả phòng | WHOLE_UNIT + SPLIT_SHARES | Một lease, nhiều party và share. Trách nhiệm JOINT_AND_SEVERAL hoặc SEVERAL lấy từ thỏa thuận; SEVERAL cần kiểm tổng tỷ lệ bằng 10.000 bps. |
| Thuê theo giường hoặc chỗ | PER_SLOT | Mỗi slot có lease và bảng kê riêng, tiện ích chung qua pool cấp unit. |

**Person** lưu thông tin cá nhân tối thiểu và có thể chưa có tài khoản. **LeaseParty** là người có vai trò hợp đồng/thanh toán, gồm PRIMARY, CO_TENANT, GUARANTOR và PAYER_ONLY. **Occupancy** biểu diễn người thực tế ở trong khoảng ngày [from, to), là nguồn headcount và person-days. **StayNotice** quản lý khách lưu trú, tách khỏi phép tính phí. Chuyển giữa WHOLE_UNIT và PER_SLOT yêu cầu xử lý các hợp đồng đang hiệu lực và đã xếp lịch, xác lập ngày chuyển rồi kiểm tra giao thời.

| Miền | Các bảng chính |
| --- | --- |
| Tổ chức, người dùng và cơ sở | organizations, users, persons, memberships, properties, units, unit_slots, unit_assets |
| Thuê và người ở | leases, lease_parties, occupancies, stay_notices |
| Tiện ích và policy | meters, meter_readings, fee_policies, split_policies, jurisdiction_rule_sets, cost_pools, cost_pool_allocations |
| Bảng kê, sổ phụ và thanh toán | billing_runs, invoices, invoice_line_items, invoice_shares, invoice_access_invites, ledger_events, ledger_entries, payment_connections, payment_transactions, payment_allocations |
| Vận hành | property_expenses, import_runs, audit_logs, consents, compliance_items, maintenance_tickets, notification_preferences, notifications |

### Ràng buộc bắt buộc trước migration

1. **Tenant isolation.** FK ghép theo organization_id và id trên các liên kết nghiệp vụ, gồm bảng allocation. Kiểm tính hợp lệ của scope đa hình. Áp RLS đọc và ghi trên bảng thuộc tổ chức, vai trò runtime không sở hữu bảng và không có BYPASSRLS, đặt ngữ cảnh org trong từng transaction. Kiểm thêm quyền theo property/lease/share và quyền worker. Trang bảng kê riêng chỉ đọc dữ liệu trong phạm vi phiên đã xác minh. DDL mẫu hai bản chỉ bật RLS trên invoices. [PostgreSQL Row Security](https://www.postgresql.org/docs/16/ddl-rowsecurity.html)
2. **Timeline và scope.** Hai partial unique index ACTIVE chưa xử lý hợp đồng xếp lịch chồng ngày hoặc thuê cả unit đồng thời với lease theo slot. Dùng exclusion theo khoảng ngày và kiểm chéo nhất quán trong transaction có khóa phù hợp. Kiểm slot thuộc đúng unit, PRIMARY theo timeline, dải ngày và chuyển trạng thái. Bổ sung scope PROPERTY cho công tơ tổng cấp tòa nhà vì schema nguồn yêu cầu meter có unit.
3. **Bất biến sau khi phát hành.** Chặn sửa hoặc xóa số tiền, dòng phí, share và snapshot đã khóa. Trạng thái thu tiền tính từ sổ hoặc cập nhật dưới dạng projection. Bỏ cascade delete trên chuỗi tài chính; giới hạn UPDATE, DELETE và TRUNCATE trên ledger. Audit chỉ lưu dữ liệu cần thiết cho mục đích truy vết.
4. **Khóa nghiệp vụ.** Bảng kê định kỳ duy nhất theo lease và kỳ/lịch áp dụng. Tất toán dùng định danh sự kiện kết thúc lease hoặc rời party để xử lý nhiều sự kiện hợp lệ trong một tháng. Credit note có khóa thao tác và tham chiếu chứng từ gốc. Giao dịch provider dùng định danh theo phạm vi provider cam kết và có khóa ghi sổ theo sự kiện. Tài khoản dùng chung nhiều property cần định tuyến share hoặc đưa vào hàng chờ.
5. **Phiên bản lịch sử.** Policy và rule set có định danh, phiên bản, ngày hiệu lực, checksum và cấu hình đã validate. Publish khóa khoảng hiệu lực chồng lấn và giữ bản từng có hiệu lực. Snapshot lưu thêm phiên bản engine và đầu vào chuẩn hóa. Các DDL nguồn dùng làm tham chiếu để viết migration theo giai đoạn.

## 4. Tính phí, phát hành và sổ phụ

**Kỳ và batch.** Standard dùng tháng lịch cho tiện ích, ngày chạy batch và hạn thu là cấu hình riêng. Chọn mọi lease giao với kỳ, kể cả lease kết thúc giữa tháng. Tính tiền thuê kỳ lẻ theo điều khoản đã chốt. Một khoản thuê dùng một nguồn giá hiệu lực, snapshot ngay khi tính. Batch báo phòng thiếu chỉ số hoặc đang NEEDS_REVIEW, vẫn tạo nháp cho phòng đủ dữ liệu. Phát hành từng bảng kê trong transaction riêng để chạy lại lô và theo dõi phát hành từng phần.

**Policy.** Resolver xét override_key từ LEASE > UNIT > PROPERTY > ORG. ADD cộng, REPLACE áp bản hiện tại rồi dừng, DISABLE dừng kế thừa. Standard dùng tiền thuê cố định/theo ngày, phí cố định và điện/nước theo biểu đã duyệt. Advanced thêm PER_CAPITA, HEADCOUNT_TABLE, METER_TIERED, ELECTRIC_RESIDENTIAL, WATER_LOCAL_TIERED, PERCENT_OF, LATE_FEE theo điều khoản, ONE_TIME_ADJUSTMENT và FORMULA. DRAFT → SIMULATED → PUBLISHED → SUPERSEDED, thử trên kỳ hiện tại và tối đa hai kỳ lịch sử có dữ liệu trước khi publish. Kiểm phí trùng, vòng lặp và sai tổng. CEL chỉ đọc dữ liệu đầu vào đã duyệt, giới hạn độ dài, cost và thời gian, không có I/O. Preset Standard, ở ghép, gia đình và ký túc xá tạo bản nháp.

**Compute(input) → output.** Chốt snapshot lease/party, occupancy, meter segments, policy và rule hiệu lực. Tách đoạn khi đổi công tơ, người ở, giá hoặc hợp đồng theo căn cứ áp dụng. Tính phí thô bằng số hữu tỉ hoặc fixed-point; dùng số nguyên VND khi ghi sổ, kiểm tràn và biểu diễn số tiền 64-bit bằng chuỗi thập phân trong JSON. Tiền phí làm tròn một lần tại dòng hoặc tổng pool. Đề xuất mặc định 1 đồng, cấu hình bước 1.000 đồng khi điều khoản và giới hạn thu cho phép. Sau phát hành, giữ các đầu vào và dấu vết làm tròn để đối chiếu.

**Chia phí.** Pool → lease rồi từng dòng → party. Standard dồn một share cho PRIMARY. Advanced có EQUAL, PERSON_DAYS, PERCENT, WEIGHT, FIXED_THEN_REMAINDER, PRIMARY_ALL và BY_SUBMETER cho pool. Person-days là tổng số ngày giao kỳ nhân trọng số trên occupancy đủ điều kiện; có thể chọn snapshot/max/min nếu policy quy định. Sub-meter tính riêng và phần dư phân bổ theo policy, đưa vào ngoại lệ nếu tổng sub-meter vượt meter tổng. Slot trống mặc định phần chủ nhà chịu, có thể phân bổ lại theo policy. Largest remainder với khóa ổn định để xử lý hòa. Khoản âm phân bổ trị tuyệt đối rồi đảo dấu. Tổng share bằng invoice, tổng phần chi tiết bằng line, tổng phân bổ cộng phần chủ nhà chịu bằng pool. Ví dụ 301.500 đồng làm tròn HALF_UP theo bước 1.000 đồng thành 302.000. Chia theo 30/30/20 person-days cho kết quả 113.000/113.000/76.000 đồng.

| Sự kiện | Hạch toán |
| --- | --- |
| Phát hành | Ghi CHARGE tăng công nợ theo share, khóa chứng từ và enqueue thông báo cùng transaction. |
| Nhận tiền | Ghi giao dịch thu một lần. Phân bổ tối đa bằng phần còn nợ và giảm RECEIVABLE. Trả thiếu là PARTIALLY_PAID. Thu dư hoặc chưa ghép vào số dư riêng để cấn trừ hoặc hoàn. |
| Cọc và giữ chỗ | DEPOSIT_IN tăng cọc đang giữ, DEPOSIT_OUT/DEPOSIT_APPLIED giảm cọc. Giữ chỗ trước ACTIVE có phân loại riêng. Chuyển giữ chỗ thành cọc hoặc chuyển cọc giữa party ghi cặp bút toán phi tiền mặt; giữ cọc vì vi phạm cần căn cứ, lý do và audit. |
| Trả trước và thu nhiều tháng | Tách lịch thu tiền thuê 1/3/6/12 tháng khỏi lịch tiện ích. Có thể nhận tiền thuê trả trước vào số dư riêng rồi cấn kỳ, chỉ ghi dòng tiền ở lần nhận thực. |
| Sửa sai và tất toán | Giảm bằng CREDIT_NOTE tham chiếu bản gốc, tăng bằng bảng kê bổ sung. Đảo bút toán có tham chiếu khi sửa giao dịch. Tất toán toàn lease hoặc một party tới ngày rời, loại phần đã tính, ghi hoàn/cấn cọc theo biên bản. REFUND_DUE khi có khoản phải hoàn. |
| Báo cáo | Tách công nợ, thực thu, cọc, trả trước, hoàn tiền và chi theo property/kỳ. Hạch toán nội bộ, chuyển cọc hoặc cấn trừ không sinh dòng tiền mới. |

Số dư cọc và paid_minor, nếu lưu cache, tái tính từ ledger. Dương trên RECEIVABLE tăng nợ và dương trên DEPOSIT tăng cọc đang giữ. OVERDUE suy ra từ số còn nợ và hạn thu. Advanced hỗ trợ ảnh/bằng chứng bàn giao, xác nhận khấu trừ, tranh chấp và chuyển PRIMARY trước khi người đại diện rời. Pilot hỗ trợ hoàn cọc cơ bản và điều chỉnh có phê duyệt.

## 5. Thanh toán, dữ liệu và công tơ

- **Gửi bảng kê riêng qua Zalo.** Lời nhắc và bản xem trước chỉ mang nội dung chung. Link mời ngẫu nhiên, lưu hash, có hạn và thu hồi được. Link chỉ khởi tạo xác minh qua kênh liên hệ đã đăng ký, chẳng hạn mã dùng một lần gửi riêng. Sau xác minh, phiên ngắn hạn chỉ được xem bảng kê hoặc phần thanh toán đã cấp. Kiểm phạm vi và trạng thái thu hồi ở mỗi lượt đọc. Dùng no-store, noindex, Referrer-Policy strict-origin để giữ nguồn gửi form và ẩn đường dẫn lời mời và loại token khỏi log/analytics. VietQR nếu bật chỉ hiện sau xác minh, gồm tài khoản nhận, số còn nợ và mã đối soát ngẫu nhiên duy nhất.
- **Webhook đúng contract.** Thử SePay trước nếu ngân hàng, gói và tài khoản pilot phù hợp. Giới hạn body, tìm cấu hình kết nối, xác thực trước xử lý tài chính, kiểm tài khoản nhận và tiền vào. SePay xác thực theo loại webhook đã cấu hình. payOS kiểm chữ ký trên dữ liệu đã canonicalize theo tài liệu, nên yêu cầu “xác thực raw body trước khi parse” không áp dụng chung. Casso theo phiên bản API đã chọn. Ack chỉ sau commit và theo contract, kể cả callback kiểm tra hoặc retry. [SePay Webhooks](https://docs.sepay.vn/tich-hop-webhooks.html) · [payOS signature](https://payos.vn/docs/tich-hop-webhook/kiem-tra-du-lieu-voi-signature/) · [Casso Flow API](https://flow.casso.vn/docs/)
- **Đối soát.** Một transaction ghi giao dịch gốc, khóa số dư liên quan, ghép mã share hoặc VA nếu xác định được, ghi allocations/ledger. Giao dịch thiếu mã, xung đột hoặc phần dư nằm trong hàng chờ trên giao dịch gốc. So khớp theo số tiền, thời gian và người chuyển chỉ hỗ trợ người duyệt khi mơ hồ. Ghi tay và webhook đối chiếu cùng mã ngân hàng nếu có. Backfill lịch sử dùng cùng luồng idempotent và checkpoint. Kiểm cả retry, trả hộ, trả gộp, trả dư và giao dịch đã ghi tay. Thử fixture và test mode trước auto-match. [SePay Test mode](https://developer.sepay.vn/en/tien-ich-khac/test-mode/tao-webhook)
- **Chi và import.** Chi vận hành theo property, loại chi, ngày, chứng từ riêng, báo cáo thực thu trừ thực chi. CSV có mẫu, dry-run từng dòng, chuẩn hóa ngày/số tiền, khóa import chống lặp và đối chiếu số dư mở khi nhập phòng, lease, người đại diện, công tơ và cọc. Thêm XLSX khi cần, giới hạn kích thước và xử lý công thức an toàn.
- **Công tơ.** Theo dõi meter tổng property/unit và sub-meter; tem QR gắn đúng danh tính. Reading PERIODIC, MOVE_IN, MOVE_OUT, REPLACEMENT và CORRECTION. Sửa bằng bản ghi mới liên kết bản cũ, giữ nguyên dữ liệu đo gốc. Thay đồng hồ tính chênh lệch trên meter cũ cộng chênh lệch trên meter mới, đưa vào NEEDS_REVIEW. Chỉ dùng reading VERIFIED để tính. Điều chỉnh sau phát hành đi qua chứng từ.
- **AI.** MeterVisionProvider gồm none mặc định, Ollama nội bộ, Gemini/OpenAI theo cấu hình. Crop và re-encode bỏ EXIF/GPS, kiểm ảnh trước khi gửi, output dạng JSON chỉ đề xuất số đo. Khởi đầu duyệt tay toàn bộ. Chỉ thử tự duyệt sau khi đo lỗi trên ảnh thực, có QR, đối chiếu hai lần, kiểm xu hướng và ngưỡng. Ảnh tenant, thay meter và số bất thường luôn qua duyệt. Confidence do model tự báo là tín hiệu phụ.

## 6. Bảo mật và vận hành

- OIDC Authorization Code + PKCE S256, kiểm issuer/audience/state/nonce/chữ ký và ánh xạ (issuer, sub). Setup owner qua secret một lần. Session thu hồi được, cookie Secure/HttpOnly, chống CSRF cho thao tác ghi. Magic-link/OTP tenant có TTL ngắn, dùng một lần, lưu hash, giới hạn tần suất, hết quyền khi rời lease. Tenant thấy share của mình, tổng nợ chung khi thỏa thuận yêu cầu, và chỉ thấy chi tiết người khác theo chính sách được phép.
- Envelope AES-256-GCM với DEK bọc bởi KEK, nonce, key/version và AAD ràng buộc org/bản ghi/trường. KMS/Vault hoặc khóa cài đặt. Dùng blind index khóa riêng nếu cần tra cứu giấy tờ. Kế hoạch rewrap, backup và khôi phục khóa. Thu thập PII tối thiểu ngay từ Standard, chỉ bổ sung dữ liệu cư trú khi có mục đích và căn cứ.
- **Phạm vi mã hóa và quyền xem.** TLS khi truyền. Mã hóa PostgreSQL/volume, kho tệp và bản sao lưu khi lưu trữ. Mã hóa cấp trường hoặc tệp với giấy tờ, thông tin ngân hàng nhạy cảm, secret và chứng từ riêng. Khóa được quản lý tách khỏi dữ liệu, giới hạn quyền giải mã. Backend được cấp quyền xử lý số tiền để tính phí, đối soát và báo cáo, nên kiến trúc pilot có máy chủ tin cậy. Chủ sở hữu xem số liệu tổ chức mình; quản lý chỉ thấy số liệu trong phạm vi tòa được cấp; nhân viên công tơ chỉ thấy phòng/chỉ số cần nhập; người thuê chỉ thấy bảng kê hoặc phần thanh toán được cấp. Che số tiền là bảo vệ khi nhìn màn hình; phân quyền API và DB quyết định dữ liệu nào được trả về.
- **Người vận hành và xuất dữ liệu.** Tài khoản hỗ trợ mặc định không có quyền xem số liệu khách. Truy cập hỗ trợ ngoại lệ theo phạm vi, thời hạn và lý do, được chủ sở hữu duyệt và ghi nhật ký. Kiểm quyền riêng cho CSV và chứng từ; giữ luồng tải riêng qua phiên đã xác thực. Bản self-host cho phép chủ trọ tự giữ máy chủ và khóa khi cần tự kiểm soát hạ tầng.
- Tệp private, giới hạn mục đích/kích thước, kiểm nội dung và quarantine khi cần; re-encode ảnh trước khi phục vụ. URL ký ngắn hạn nếu adapter hỗ trợ; API kiểm quyền khi cấp/đọc file local. Hạn chế egress cho URL do người dùng cấu hình, kiểm DNS và redirect. Endpoint Ollama nội bộ qua danh sách do admin cấu hình.
- Audit ghi tác nhân, thao tác, thời điểm và lý do cho phát hành, ghép tiền, thay giá và truy cập định danh. Redaction PII/secret khỏi log. Có retention, xóa/ẩn danh và legal hold theo loại dữ liệu. Giám sát lỗi job, tuổi hàng chờ, khoản chưa khớp và backup. Diễn tập restore DB, tệp và khóa. CI quét dependency, secret, image, tạo SBOM và ký image phát hành.

## 7. Gói quy tắc Việt Nam tại 26/11/2026

Rule có văn bản/điều khoản, phạm vi, version, khoảng hiệu lực, người duyệt, checksum và ca mẫu. VERIFIED mới được dùng tự động. Quy tắc thiếu căn cứ hoặc thiếu thông tin đối tượng chuyển sang duyệt thủ công. Duyệt đơn giá và căn cứ trước pilot dù nhập tay. Tách biểu giá khỏi quy tắc chọn biểu giá. Giá nước phụ thuộc địa phương, nhà cung cấp và kỳ.

| Lĩnh vực | Đối chiếu và hàm ý thiết kế |
| --- | --- |
| Điện sinh hoạt | **QĐ 2892/QĐ-BCT áp dụng 01/11/2026**, biểu sinh hoạt chuyển từ 6 sang **5 bậc**, thay mốc giá 1279 trong bản nguồn. Theo Thông tư 60/2025/TT-BCT, trường hợp thuê dưới 12 tháng và chủ nhà kê khai thiếu số người thuộc điều khoản áp **bậc 2** cho toàn bộ sản lượng trong biểu mới. Giữ nhánh theo đối tượng thuê, hồ sơ cư trú và loại hợp đồng mua điện, tách số người đã khai khỏi headcount chia phí. Kiểm kỳ giao thời, VAT, thay đổi số người và hồ sơ còn thiếu trước tự động hóa. NĐ 133/2026 là nguồn về xử phạt lĩnh vực điện. [QĐ 2892 và phụ lục](https://www.evn.com.vn/userfile/User/tcdl/files/2026/11/1_QD-BCT_2026_2892-20261104131123509.pdf) · [Thông tư 60](https://vbpl.moj.gov.vn/botuphap/Pages/vbpq-toanvan.aspx?ItemID=184102) · [NĐ 133](https://vanban.chinhphu.vn/?classid=1&docid=217612&orggroupid=2&pageid=27160) |
| Thuế và hóa đơn | **NĐ 141/2026** sửa NĐ 68, nâng mốc **500 triệu lên 1 tỷ đồng/năm** tại các điều được sửa, gồm quy định GTGT/TNCN liên quan hộ, cá nhân kinh doanh và cá nhân cho thuê bất động sản. Tách rule kê khai/thông báo khỏi rule hóa đơn điện tử, có loại chủ thể, điều kiện theo năm và dữ liệu ngoài hệ thống. Sổ doanh thu hỗ trợ đối chiếu, xuất cho kế toán. [NĐ 141, Điều 1 và 3](https://datafiles.chinhphu.vn/cpp/files/vbpq/2026/4/141-ndcp.signed.pdf) · [Bộ Tài chính về cá nhân có nhiều bất động sản cho thuê](https://qlg.mof.gov.vn/hoidapcstc/home/cthoidap/162069) · [NĐ 252 về quản lý thuế](https://vanban.chinhphu.vn/?classid=1&docid=218690&orggroupid=2&pageid=27160) |
| Cư trú | Phân biệt đăng ký tạm trú với thông báo lưu trú. Mốc 30 ngày tại Điều 27 có điều kiện về nơi cư trú, vì vậy tránh một deadline chung cho mọi occupancy. Lưu giờ đến để tính nhắc thông báo lưu trú trước 23 giờ ngày đến, hoặc trước 08 giờ hôm sau nếu đến sau 23 giờ, theo đối tượng áp dụng. Hỗ trợ xuất dữ liệu và ghi xác nhận thao tác bên ngoài. [Luật Cư trú hợp nhất, Điều 27 và 30](https://chupah.gialai.gov.vn/upload/105657/20260309/flatten-06_VBHN-VPQH_Luat_cu_tru_99519.pdf) · [Cổng Bộ Công an](https://tbltkbtt.bocongan.gov.vn/auth/register) |
| PCCC | Checklist theo nhóm cơ sở, địa phương, kỳ kiểm tra và hồ sơ/bằng chứng. Đối chiếu NĐ 105/2025 với văn bản sửa đổi trước khi cung cấp checklist mặc định. Người phụ trách xác nhận phân loại cơ sở và điều kiện bảo hiểm. [NĐ 105/2025](https://vanban.chinhphu.vn/?classid=1&docid=213702&orggroupid=2&pageid=27160) · [NĐ 217/2026](https://datafiles.chinhphu.vn/cpp/files/vbpq/2026/6/217-ndcp.signed.pdf) · [NĐ 347/2026](https://chinhphu.vn/?docid=219411&pageid=27160) |
| Dữ liệu cá nhân | Luật 91/2025/QH15 và NĐ 356/2025/NĐ-CP định hướng phân vai xử lý, căn cứ/mục đích, quyền của người thuê, xử lý sự cố và đánh giá chuyển dữ liệu xuyên biên giới theo trường hợp. Rà luồng dữ liệu và hồ sơ trước khi bật AI cloud. Gắn điều kiện và thời hạn thông báo sự cố với quy định áp dụng. [Luật 91/2025/QH15](https://mps.gov.vn/chinh-sach-phap-luat/co-so-du-lieu-van-ban/luat-bao-ve-du-lieu-ca-nhan-1753688803) · [NĐ 356/2025](https://vbpl.moj.gov.vn/bocongan/Pages/vbpq-thuoctinh.aspx?ItemID=187276) |

## 8. Khả thi, lộ trình và nghiệm thu

**Khả thi kỹ thuật của luồng Standard được đánh giá 4/5. Nhu cầu và khả năng bán chưa được kiểm chứng.** Khách có thể hoàn thành một kỳ thu với phương án có hỗ trợ trước khi đầu tư vào phần mềm. Quyết định xây dựa trên việc họ dùng lại, số phút thực sự tiết kiệm và phản hồi trước một đề nghị trả phí cụ thể.

Điểm định tính trên thang 5, xét năng lực công nghệ, độ phức tạp nghiệp vụ, dữ liệu pilot và phụ thuộc bên ngoài. Điểm 5 biểu thị đường triển khai rõ và ít phụ thuộc; 3 cần chứng minh qua thử nghiệm; 1–2 cần giải quyết phụ thuộc lớn hoặc xác minh chuyên môn.

| Hạng mục | Điểm | Điều kiện chính |
| --- | --- | --- |
| Standard, batch, thu tay, VietQR và Zalo | **4/5** | Stack phù hợp. Cần dữ liệu đầu kỳ sạch, phát hành có kiểm soát và đối chiếu từng đồng. |
| Ledger, cọc, sửa sai và import | **3,5/5** | Cần bộ ca công nợ mở, thu thiếu/dư, đổi người và đảo bút toán; sai số có chi phí xử lý trực tiếp. |
| Tự đối soát một provider | **3/5** | Phụ thuộc ngân hàng, gói merchant, HTTPS, mã giao dịch, chất lượng memo và khả năng backfill. |
| Advanced chia người, slot, tất toán | **3/5** | Khó ở timeline, dữ liệu chuyển phòng/giường, thỏa thuận pool và chống thu trùng khi tất toán. |
| Rule Việt Nam nhắc việc và xuất dữ liệu | **3/5** | Duyệt phạm vi và phiên bản định kỳ. Tự động xác định nghĩa vụ hoặc áp điện đại trà khoảng **2/5** ở hiện trạng thiết kế. |
| AI công tơ | **3/5** khi gợi ý, **2/5** khi tự xác nhận | Cần ảnh thực và số đo lỗi theo loại công tơ, ánh sáng, góc chụp và mốc tiêu thụ. |
| Self-host một node | **4/5** | Cần người vận hành, TLS, sao lưu, khôi phục và cấu hình kho tệp. Báo giá thực tế tùy nơi chạy, lưu lượng và nhà cung cấp. |

| Giai đoạn | Phạm vi | Ước lượng đề xuất |
| --- | --- | --- |
| 0. Quan sát và chạy thử có hỗ trợ | Phỏng vấn và quan sát ca thật, lần lượt thử bảng kê, gửi Zalo, theo dõi thu. Đo thời gian, ghi nhu cầu lặp lại và thử đề nghị trả phí. | 1–2 tuần chuẩn bị, sau đó theo lịch thu của khách |
| 1. Bản trả phí đầu tiên | Tự động hóa phần việc đã có bằng chứng. Gợi ý phạm vi là nhập phòng có hỗ trợ, chốt và gửi bảng kê, ghi nhận thu và xuất công nợ. | Ước lượng sau kỳ thử, theo phạm vi khách cam kết dùng |
| 2. Đối soát và mở rộng | Bổ sung từng nhu cầu đã lặp lại. Nếu ưu tiên là đối soát, thử sao kê có người duyệt. Khi cần adapter, kiểm retry, backfill, giao dịch chưa khớp và ghi tay trùng webhook. | Ước lượng sau khi có dữ liệu giao dịch và kết nối thử |

Tính cả thời gian thiết lập, hỗ trợ và sửa ngoại lệ khi đánh giá khả năng phục vụ khách. Dùng chi phí đó, tần suất khách quay lại và đề nghị trả phí đã thử để quyết định có tự động hóa thêm hay đổi phân khúc.

### Cổng nghiệm thu

- Trước khi dùng bảng kê thật, khách duyệt tiền từng phòng, đối chiếu tổng đã thu và công nợ, giải thích mọi chênh lệch và giữ bản xuất để quay lại sổ cũ nếu cần. Đo thời gian của chính chủ trọ trên cùng một nhóm phòng.
- Khi nhiều khách dùng ứng dụng, kiểm quyền giữa ít nhất hai tổ chức. Kiểm thu hồi link nếu có, khôi phục dữ liệu và chống ghi trùng ở các thao tác đã triển khai.
- Đo lượt quay lại ở kỳ thứ hai, thời gian hỗ trợ và số khách thực hiện cam kết trả phí. Trước khi chuyển sang sổ chính, đối chiếu đủ hai kỳ và xử lý các chênh lệch.
- Bật auto-match khi đã qua contract, retry, đối chiếu sổ ngân hàng, thu tay trùng webhook và backfill. Bật calculator điện sau khi đã duyệt nhóm áp dụng, version giá, kỳ giao thời, VAT và ca kiểm tra. Xem xét AI tự duyệt sau khi đo được lỗi đọc có ảnh hưởng tiền.

## 9. Đối chiếu hai bản nguồn

| Chủ đề | Quyết định trong bản hợp nhất |
| --- | --- |
| Thế mạnh v3.0 | Giữ luồng đơn giản với bảng kê, Zalo và thu tay. QR, sổ chi, import tự phục vụ và webhook theo nhu cầu đã quan sát. |
| Thế mạnh v2.1 | Giữ mô hình nhiều người, policy có phiên bản, split hai cấp, công tơ theo sự kiện, đối soát, tất toán, bảo mật và jurisdiction pack. |
| Điều kiện hiện tại | Cập nhật điện, thuế, PCCC và MinIO. Xác thực webhook theo từng contract. |
| DDL | Đưa về schema logic, bổ sung ràng buộc tenant, thời gian, bất biến và chống giao dịch trùng để viết migration theo giai đoạn. |
