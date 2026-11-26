param(
    [string]$Base = 'http://localhost:8088',
    [string]$EnvFile = (Join-Path (Split-Path $PSScriptRoot -Parent) '.env'),
    [switch]$CheckDemo,
    [string]$SecondOwnerPassword
)
$ErrorActionPreference = 'Stop'
$config = ConvertFrom-StringData (Get-Content -Raw -LiteralPath $EnvFile)
function Assert-That([bool]$Check, [string]$Label) {
    if (-not $Check) { throw "Failed: $Label" }
    Write-Output "PASS $Label"
}
function Open-Page($Client, [string]$Path, [hashtable]$Data, [string]$Origin = $Base) {
    $options = @{ Uri = "$Base$Path"; WebSession = $Client; SkipHttpErrorCheck = $true; TimeoutSec = 10 }
    if ($null -ne $Data) {
        $options.Method = 'Post'
        $options.Body = $Data
        $options.Headers = @{Origin = $Origin}
    }
    Invoke-WebRequest @options
}
function Get-Match([string]$Body, [string]$Pattern) {
    $match = [regex]::Match($Body, $Pattern)
    if (-not $match.Success) { throw "Expected form or link missing: $Pattern" }
    $match.Groups[1].Value
}
function Submit-Twice($Session, [string]$Path, [hashtable]$Form) {
    $handler = [Net.Http.HttpClientHandler]::new()
    $handler.CookieContainer = $Session.Cookies
    $handler.AllowAutoRedirect = $false
    $client = [Net.Http.HttpClient]::new($handler)
    $client.Timeout = [TimeSpan]::FromSeconds(10)
    [void]$client.DefaultRequestHeaders.TryAddWithoutValidation('Origin', $Base)
    $fields = [Collections.Generic.Dictionary[string,string]]::new()
    foreach ($name in $Form.Keys) { $fields.Add($name, [string]$Form[$name]) }
    try {
        $first = $client.PostAsync("$Base$Path", [Net.Http.FormUrlEncodedContent]::new($fields))
        $second = $client.PostAsync("$Base$Path", [Net.Http.FormUrlEncodedContent]::new($fields))
        foreach ($request in @($first, $second)) {
            $response = $request.GetAwaiter().GetResult()
            try { [int]$response.StatusCode } finally { $response.Dispose() }
        }
    }
    finally { $client.Dispose() }
}
$owner = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$meter = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$tenant = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$outsider = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$period = '2026-11'
$room = "Test-" + [Guid]::NewGuid().ToString('N').Substring(0,8)
$request = Open-Page $outsider '/healthz'
Assert-That ($request.StatusCode -eq 200) 'health endpoint'
$request = Open-Page $outsider '/privacy'
Assert-That ($request.StatusCode -eq 200 -and $request.Content.Contains('<h1>Quyền riêng tư</h1>') -and $request.Content.Contains('2026 Torin Tran')) 'public privacy page and copyright'
$request = Open-Page $owner '/login' @{account='owner';password=$config.ZS_OWNER_PASSWORD}
Assert-That ($request.StatusCode -eq 200 -and $request.Content.Contains('<h1>Tổng quan kỳ thu</h1>')) 'owner sign-in opens dashboard'
$csrf = Get-Match $request.Content 'name="csrf" value="([a-f0-9]+)"'
if ($CheckDemo) {
    $request = Open-Page $owner "/?tab=dashboard&period=$period"
    Assert-That ($request.Content.Contains("4 phòng trong kỳ $period") -and $request.Content.Contains('4.430.000') -and $request.Content.Contains('3.065.000') -and $request.Content.Contains('1.365.000')) 'seed replay preserves demo room count and dashboard totals'
    $request = Open-Page $owner '/?tab=dashboard&period=2026-10'
    Assert-That ($request.Content.Contains('0 phòng trong kỳ 2026-10') -and !$request.Content.Contains('4.430.000')) 'dashboard filters the selected period'
    $request = Open-Page $owner "/?tab=meters&period=$period&filter=missing"
    Assert-That ($request.Content.Contains('DEMO-2026-11-P04') -and !$request.Content.Contains('DEMO-2026-11-P01')) 'missing-meter filter excludes recorded rooms'
    foreach ($item in @(@{Filter='draft';Room='P03'}, @{Filter='due';Room='P02'})) {
        $request = Open-Page $owner "/?tab=billing&period=$period&filter=$($item.Filter)"
        $bills = [regex]::Matches($request.Content, '(?s)<article class="list-row">.*?</article>')
        Assert-That ($bills.Count -eq 1 -and $bills[0].Value.Contains("DEMO-2026-11-$($item.Room)")) "billing filter $($item.Filter) selects the correct bill"
    }
}
$request = Open-Page $owner '/action/room' @{csrf=$csrf;period=$period;name=$room;tenant='Người thuê thử';rent='2500000';electricRate='3000';waterRate='15000';electric='10';water='20'}
Assert-That ($request.Content.Contains($room)) 'room and contract'
$importRoom = "$room-CSV"
$template = Open-Page $owner '/rooms/template.csv'
Assert-That ($template.StatusCode -eq 200 -and $template.Headers['Content-Disposition'] -match 'mau-nhap-phong.csv') 'owner downloads CSV template'
$csv = "$($template.Content.Trim())`n$importRoom,,2000000,3000,15000,0,0"
$request = Open-Page $owner '/action/import-preview' @{csrf=$csrf;period=$period;csv=$csv}
Assert-That ($request.Content.Contains('Thêm 1 phòng mới')) 'CSV preview validates without writing'
$request = Open-Page $owner "/?tab=rooms&period=$period"
Assert-That (-not $request.Content.Contains($importRoom)) 'CSV preview leaves rooms unchanged'
$request = Open-Page $owner '/action/import' @{csrf=$csrf;period=$period;csv=$csv}
Assert-That ($request.Content.Contains($importRoom)) 'CSV commit persists rooms'
$request = Open-Page $owner '/action/import-preview' @{csrf=$csrf;period=$period;csv=$csv}
Assert-That ($request.Content.Contains('Bỏ qua 1 dòng')) 'CSV re-import skips existing room'
$request = Open-Page $owner "/?tab=meters&period=$period"
$form = [regex]::Matches($request.Content, '(?s)<form class="meter-row".*?</form>') | Where-Object { $_.Value.Contains("<strong>$room</strong>") } | Select-Object -First 1
$roomId = Get-Match $form.Value 'name="id" value="([a-f0-9]+)"'
$request = Open-Page $owner "/rooms/${roomId}?period=$period"
Assert-That ($request.Content.Contains('Người thuê thử') -and $request.Content.Contains('2.500.000') -and $request.Content.Contains('Chỉnh sửa')) 'owner sees room details'
$request = Open-Page $owner "/rooms/${roomId}?period=$period&edit=1"
$roomRevision = Get-Match $request.Content 'name="revision" value="([a-f0-9]+)"'
$edit = @{csrf=$csrf;id=$roomId;period=$period;revision=$roomRevision;name=$room;tenant='Người thuê thử đã sửa';rent='2500000';electricRate='3000';waterRate='15000';electric='10';water='20'}
$request = Open-Page $owner '/action/room-edit' $edit
Assert-That ($request.Content.Contains('Đã lưu thông tin phòng') -and $request.Content.Contains('Người thuê thử đã sửa')) 'owner updates room details'
$request = Open-Page $owner '/action/room-edit' $edit
Assert-That ($request.StatusCode -eq 422 -and $request.Content.Contains('Thông tin phòng đã thay đổi')) 'stale room edit cannot overwrite a saved edit'
$request = Open-Page $meter '/login' @{account='meter';password=$config.ZS_METER_PASSWORD}
$meterCsrf = Get-Match $request.Content 'name="csrf" value="([a-f0-9]+)"'
Assert-That ($request.Content.Contains($room) -and -not $request.Content.Contains('Người thuê thử')) 'meter role only receives meter fields'
$request = Open-Page $meter "/?tab=dashboard&period=$period"
Assert-That ($request.Content.Contains('<h1>Nhập điện nước</h1>') -and !$request.Content.Contains('Đã thu ròng')) 'meter role cannot access dashboard totals'
$request = Open-Page $meter "/rooms/$roomId"
Assert-That ($request.StatusCode -eq 403) 'meter role cannot read room details'
$request = Open-Page $meter '/action/room-edit' @{csrf=$meterCsrf;id=$roomId}
Assert-That ($request.StatusCode -eq 403) 'meter role cannot edit rooms'
$request = Open-Page $meter '/export'
Assert-That ($request.StatusCode -eq 403) 'meter role cannot export finances'
$request = Open-Page $meter '/rooms/template.csv'
Assert-That ($request.StatusCode -eq 403) 'meter role cannot access room CSV template'
$reading = @{csrf=$meterCsrf;id=$roomId;period=$period;electric='33';water='26'}
$request = Open-Page $meter '/action/read' $reading 'http://invalid.example'
Assert-That ($request.StatusCode -eq 403) 'cross-site form rejected'
$request = Open-Page $meter '/action/read' $reading
Assert-That ($request.Content.Contains('Đã lưu')) 'meter reading persisted'
$request = Open-Page $owner '/action/draft' @{csrf=$csrf;period=$period}
$article = [regex]::Matches($request.Content,'(?s)<article class="list-row">.*?</article>') | Where-Object Value -Match $room | Select-Object -First 1
$invoice = Get-Match $article.Value 'href="/invoice/([a-f0-9]+)"'
$request = Open-Page $owner "/invoice/$invoice"
Assert-That ($request.Content.Contains('2.659.000') -and $request.Content.Contains('Bản nháp')) 'exact draft amount'
$revision = Get-Match $request.Content 'name="revision" value="([a-f0-9]+)"'
$request = Open-Page $meter "/invoice/$invoice"
Assert-That ($request.StatusCode -eq 403) 'meter role cannot read invoice'
$request = Open-Page $meter '/action/read' (@{csrf=$meterCsrf;id=$roomId;period=$period;electric='35';water='26'})
$request = Open-Page $owner '/action/issue' @{csrf=$csrf;id=$invoice;revision=$revision}
Assert-That ($request.StatusCode -eq 422 -and $request.Content.Contains('Bảng kê đã thay đổi')) 'stale preview cannot issue a changed draft'
$request = Open-Page $meter '/action/read' $reading
$request = Open-Page $owner "/invoice/$invoice"
$revision = Get-Match $request.Content 'name="revision" value="([a-f0-9]+)"'
Assert-That (((Submit-Twice $owner '/action/issue' @{csrf=$csrf;id=$invoice;revision=$revision}) -join ',') -eq '303,303') 'concurrent issuance succeeds for both requests'
$request = Open-Page $owner "/invoice/$invoice"
Assert-That ($request.Content.Contains('Đã phát hành')) 'invoice issuance'
$request = Open-Page $owner "/rooms/${roomId}?period=$period&edit=1"
Assert-That (!$request.Content.Contains('name="rent"') -and !$request.Content.Contains('name="electric"')) 'issued room locks financial and opening fields in its edit form'
$roomRevision = Get-Match $request.Content 'name="revision" value="([a-f0-9]+)"'
$request = Open-Page $owner '/action/room-edit' @{csrf=$csrf;id=$roomId;period=$period;revision=$roomRevision;name=$room;tenant='Người thuê thử đã sửa';rent='1'}
Assert-That ($request.StatusCode -eq 422) 'issued room rejects attempted rate changes on the server'
$terms = @{csrf=$csrf;id=$roomId;effective='2026-12';tenant='Người thuê kỳ sau';rent='2700000';electricRate='4000';waterRate='20000'}
Assert-That (((Submit-Twice $owner '/action/lease' $terms) -join ',') -eq '303,303') 'concurrent identical lease scheduling is idempotent'
$request = Open-Page $owner "/rooms/${roomId}?period=2026-12"
Assert-That ($request.Content.Contains('Người thuê kỳ sau') -and $request.Content.Contains('2.700.000')) 'room shows terms for the selected period'
$request = Open-Page $owner "/invoice/$invoice"
Assert-That ($request.Content.Contains('2.659.000') -and !$request.Content.Contains('Người thuê kỳ sau')) 'lease change preserves issued invoice and representative'
$request = Open-Page $meter '/action/read' @{csrf=$meterCsrf;id=$roomId;period='2026-12';electric='35';water='27'}
Assert-That ($request.Content.Contains('Đã lưu')) 'next-period reading uses prior issued close'
$request = Open-Page $owner '/action/draft' @{csrf=$csrf;period='2026-12'}
$nextArticle = [regex]::Matches($request.Content,'(?s)<article class="list-row">.*?</article>') | Where-Object Value -Match $room | Select-Object -First 1
$nextInvoice = Get-Match $nextArticle.Value 'href="/invoice/([a-f0-9]+)"'
$request = Open-Page $owner "/invoice/$nextInvoice"
Assert-That ($request.Content.Contains('2.728.000') -and $request.Content.Contains('Người thuê kỳ sau')) 'next-period draft uses new lease terms'
$beforeFee = Get-Match $request.Content 'name="revision" value="([a-f0-9]+)"'
$request = Open-Page $owner '/action/billing-settings' @{csrf=$csrf;period='2026-12';dueDate='2027-01-05';commonFeeLabel='Vệ sinh';commonFee='20000'}
Assert-That ($request.Content.Contains('Đã lưu')) 'owner configures a due date and common fee'
$request = Open-Page $owner "/invoice/$nextInvoice"
Assert-That ($request.Content.Contains('2.748.000') -and $request.Content.Contains('Vệ sinh') -and $request.Content.Contains('2027-01-05')) 'draft incorporates fee as a separate line and shows due date'
$request = Open-Page $owner '/action/issue' @{csrf=$csrf;id=$nextInvoice;revision=$beforeFee}
Assert-That ($request.StatusCode -eq 422) 'billing settings invalidate an older approval preview'
$request = Open-Page $meter '/action/read' (@{csrf=$meterCsrf;id=$roomId;period=$period;electric='35';water='26'})
Assert-That ($request.Content.Contains('Kỳ này đã phát hành')) 'issued readings protected'
$payment = @{csrf=$csrf;id=$invoice;kind='Thu';amount='500000';key=[Guid]::NewGuid().ToString('N');method='cash';reference='PT-THU-01'}
Assert-That (((Submit-Twice $owner '/action/post' $payment) -join ',') -eq '303,303') 'concurrent payment replay succeeds for both requests'
$request = Open-Page $owner "/invoice/$invoice"
Assert-That ($request.Content.Contains('2.159.000') -and -not $request.Content.Contains('1.659.000')) 'cash payment is idempotent'
Assert-That ($request.Content.Contains('Thu · Tiền mặt') -and $request.Content.Contains('PT-THU-01')) 'payment history displays the method and receipt reference'
$payment.reference = 'PT-THU-02'
$request = Open-Page $owner '/action/post' $payment
Assert-That ($request.StatusCode -eq 422) 'same payment key cannot change the receipt reference'
$payment.reference = 'PT-THU-01'
$request = Open-Page $owner "/export?period=$period"
Assert-That ($request.Content.Contains("$room,$period,2659000,500000,2159000,0,0,2659000")) 'CSV matches outstanding and adjusted balances'
$reversal = @{csrf=$csrf;id=$invoice;original=$payment.key;key=[Guid]::NewGuid().ToString('N');note='Nhập nhầm chứng từ'}
Assert-That (((Submit-Twice $owner '/action/reverse' $reversal) -join ',') -eq '303,303') 'concurrent reversal replay succeeds for both requests'
$request = Open-Page $owner "/export?period=$period"
Assert-That ($request.Content.Contains("$room,$period,2659000,0,2659000,0")) 'reversal restores exact balance'
$reversal.key = [Guid]::NewGuid().ToString('N')
$request = Open-Page $owner '/action/reverse' $reversal
Assert-That ($request.StatusCode -eq 422 -and $request.Content.Contains('Bút toán này đã được đảo')) 'a new request cannot reverse the same entry twice'
$payment.key = [Guid]::NewGuid().ToString('N')
$request = Open-Page $owner '/action/post' $payment
$request = Open-Page $owner '/action/invite' @{csrf=$csrf;id=$invoice}
$invite = Get-Match $request.Content '/p/([a-f0-9]{48})'
$code = Get-Match $request.Content '<strong>([0-9]{8})</strong>'
$request = Open-Page $outsider "/p/$invite"
Assert-That ($request.Content.Contains('Xác minh') -and -not $request.Content.Contains($room) -and -not $request.Content.Contains('2.159.000')) 'forwarded link reveals no room or amount'
$request = Open-Page $outsider "/p/$invite" @{code='INVALID'}
Assert-That ($request.Content.Contains('Mã không hợp lệ') -and -not $request.Content.Contains('2.159.000')) 'wrong code rejected'
$request = Open-Page $tenant "/p/$invite" @{code=$code}
Assert-That ($request.Content.Contains('2.159.000') -and $request.Content.Contains($room)) 'authorized tenant sees exact bill'
$request = Open-Page $outsider "/p/$invite" @{code=$code}
Assert-That ($request.Content.Contains('Mã không hợp lệ')) 'single-use code cannot be replayed'
$request = Open-Page $owner '/action/revoke' @{csrf=$csrf;id=$invite}
$request = Open-Page $tenant "/p/$invite"
Assert-That (-not $request.Content.Contains('2.159.000')) 'revocation invalidates active viewing session'

if ($SecondOwnerPassword) {
    $second = New-Object Microsoft.PowerShell.Commands.WebRequestSession
    $secondReader = New-Object Microsoft.PowerShell.Commands.WebRequestSession
    $request = Open-Page $second '/login' @{account='owner-secondary';password=$SecondOwnerPassword}
    $secondCsrf = Get-Match $request.Content 'name="csrf" value="([a-f0-9]+)"'
    $request = Open-Page $second "/rooms/$roomId"
    Assert-That ($request.StatusCode -eq 404) 'second owner cannot read first organization room'
    $request = Open-Page $second "/invoice/$invoice"
    Assert-That ($request.StatusCode -eq 404) 'second owner cannot read first organization invoice'
    $request = Open-Page $second "/export?period=$period"
    Assert-That (!$request.Content.Contains($room)) 'second owner export excludes first organization balances'
    $request = Open-Page $second '/action/post' @{csrf=$secondCsrf;id=$invoice;kind='Thu';amount='1';key=[Guid]::NewGuid().ToString('N');method='cash'}
    Assert-That ($request.StatusCode -eq 422 -or $request.StatusCode -eq 404) 'second owner cannot post to first organization invoice'
    $request = Open-Page $second '/action/room' @{csrf=$secondCsrf;period=$period;name=$room;tenant='Đại diện tổ chức hai';rent='1230000';electricRate='0';waterRate='0';electric='0';water='0'}
    Assert-That ($request.Content.Contains('Đại diện tổ chức hai')) 'same room code is allowed in a different organization'
    $link = Get-Match $request.Content '/rooms/([a-f0-9]{48})'
    $request = Open-Page $owner "/rooms/$link"
    Assert-That ($request.StatusCode -eq 404) 'first owner cannot read second organization room'
    $request = Open-Page $second '/bank/account' @{csrf=$secondCsrf;bin='970407';account='9999000007788';holder='TAI KHOAN MAU HAI';confirm='yes'}
    Assert-That ($request.Content.Contains('7788')) 'second owner sets a separate bank account'
    $request = Open-Page $owner '/bank'
    Assert-That (!$request.Content.Contains('7788')) 'first organization cannot see second organization bank account'
    $null = Open-Page $second '/action/read' @{csrf=$secondCsrf;id=$link;period=$period;electric='0';water='0'}
    $request = Open-Page $second '/action/draft' @{csrf=$secondCsrf;period=$period}
    $secondInvoice = Get-Match $request.Content 'href="/invoice/([a-f0-9]{48})"'
    $request = Open-Page $second "/invoice/$secondInvoice"
    $revision = Get-Match $request.Content 'name="revision" value="([a-f0-9]+)"'
    $null = Open-Page $second '/action/issue' @{csrf=$secondCsrf;id=$secondInvoice;revision=$revision}
    $request = Open-Page $second '/action/invite' @{csrf=$secondCsrf;id=$secondInvoice}
    $secondInvite = Get-Match $request.Content '/p/(v1-[a-f0-9]+)'
    $secondCode = Get-Match $request.Content '<strong>([0-9]{8})</strong>'
    $request = Open-Page $secondReader "/p/$secondInvite" @{code=$secondCode}
    Assert-That ($request.Content.Contains('1.230.000') -and !$request.Content.Contains('2.659.000')) 'tenant verification opens only the second organization invoice'
    $request = Open-Page $secondReader "/p/$invite"
    Assert-That (!$request.Content.Contains('2.159.000')) 'tenant session cannot cross between organizations'
}
