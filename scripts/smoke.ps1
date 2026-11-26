param([string]$Base = 'http://localhost:8088')
$ErrorActionPreference = 'Stop'
$config = ConvertFrom-StringData (Get-Content -Raw -LiteralPath (Join-Path (Split-Path $PSScriptRoot -Parent) '.env'))
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
$owner = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$meter = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$tenant = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$outsider = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$period = '2026-11'
$room = "Test-" + [Guid]::NewGuid().ToString('N').Substring(0,8)
$request = Open-Page $outsider '/healthz'
Assert-That ($request.StatusCode -eq 200) 'health endpoint'
$request = Open-Page $owner '/login' @{account='owner';password=$config.ZS_OWNER_PASSWORD}
Assert-That ($request.StatusCode -eq 200 -and $request.Content.Contains('Chốt kỳ thu')) 'owner sign-in'
$csrf = Get-Match $request.Content 'name="csrf" value="([a-f0-9]+)"'
$request = Open-Page $owner '/action/room' @{csrf=$csrf;period=$period;name=$room;tenant='Người thuê thử';rent='2500000';electricRate='3000';waterRate='15000';electric='10';water='20'}
Assert-That ($request.Content.Contains($room)) 'room and contract'
$importRoom = "$room-CSV"
$csv = "phong,nguoi_dai_dien,tien_thue,gia_dien,gia_nuoc,dien_dau,nuoc_dau`n$importRoom,,2000000,3000,15000,0,0"
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
$request = Open-Page $meter '/login' @{account='meter';password=$config.ZS_METER_PASSWORD}
$meterCsrf = Get-Match $request.Content 'name="csrf" value="([a-f0-9]+)"'
Assert-That ($request.Content.Contains($room) -and -not $request.Content.Contains('Người thuê thử')) 'meter role only receives meter fields'
$request = Open-Page $meter '/export'
Assert-That ($request.StatusCode -eq 403) 'meter role cannot export finances'
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
$request = Open-Page $meter "/invoice/$invoice"
Assert-That ($request.StatusCode -eq 403) 'meter role cannot read invoice'
$request = Open-Page $owner '/action/issue' @{csrf=$csrf;id=$invoice}
Assert-That ($request.Content.Contains('Đã phát hành')) 'invoice issuance'
$request = Open-Page $meter '/action/read' (@{csrf=$meterCsrf;id=$roomId;period=$period;electric='35';water='26'})
Assert-That ($request.Content.Contains('Kỳ này đã phát hành')) 'issued readings protected'
$payment = @{csrf=$csrf;id=$invoice;kind='Thu';amount='500000';key=[Guid]::NewGuid().ToString('N');note='Tiền mặt'}
$request = Open-Page $owner '/action/post' $payment
$request = Open-Page $owner '/action/post' $payment
Assert-That ($request.Content.Contains('2.159.000') -and -not $request.Content.Contains('1.659.000')) 'cash payment is idempotent'
$request = Open-Page $owner "/export?period=$period"
Assert-That ($request.Content.Contains("$room,$period,2659000,500000,2159000,0")) 'CSV matches outstanding balance'
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
