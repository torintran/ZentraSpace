param(
    [string]$Base = 'http://localhost:8088',
    [string]$EnvFile = (Join-Path (Split-Path $PSScriptRoot -Parent) '.env'),
    [string]$MockBase,
    [string]$MockSecret
)
$ErrorActionPreference = 'Stop'
$config = ConvertFrom-StringData (Get-Content -Raw -LiteralPath $EnvFile)
function Check([bool]$Condition, [string]$Label) {
    if (-not $Condition) { throw "Failed: $Label" }
    "PASS $Label"
}
function Page($Session, [string]$Path, [hashtable]$Form) {
    $options = @{Uri="$Base$Path";WebSession=$Session;SkipHttpErrorCheck=$true;TimeoutSec=10}
    if ($null -ne $Form) {
        $options.Method='Post'
        $options.Headers=@{Origin=$Base}
        $options.Body=$Form
    }
    $response = Invoke-WebRequest @options
    $content = $response.Content
    if ($response.Headers['Content-Type'] -like 'text/html*') {
        $content = $content.Replace('<!-- -->', '')
    }
    [pscustomobject]@{
        StatusCode = $response.StatusCode
        Headers = $response.Headers
        Content = $content
        RawContentStream = $response.RawContentStream
    }
}
function Capture([string]$Text,[string]$Pattern) {
    $m=[regex]::Match($Text,$Pattern)
    if (-not $m.Success) { throw "Expected markup missing: $Pattern" }
    $m.Groups[1].Value
}
$owner = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$meter = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$guest = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$reader = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$room='B-'+[Guid]::NewGuid().ToString('N').Substring(0,8)
$period='2026-11'
$p=Page $owner '/login' @{account='owner';password=$config.ZS_OWNER_PASSWORD}
$csrf=Capture $p.Content 'name="csrf" value="([a-f0-9]+)"'
$p=Page $meter '/login' @{account='meter';password=$config.ZS_METER_PASSWORD}
$p=Page $meter '/bank'
Check ($p.StatusCode -eq 403) 'meter role cannot access bank details or transfers'

$account = if ($MockBase) { '0000000000000' } else { '9999000000001' }
$p=Page $owner '/bank/account' @{csrf=$csrf;bin='970407';account=$account;holder='TAI KHOAN MAU';confirm='yes'}
Check ($p.StatusCode -eq 200 -and $p.Content.Contains('Đã lưu tài khoản nhận tiền')) 'owner configured a test beneficiary'

$p=Page $owner '/action/room' @{csrf=$csrf;period=$period;name=$room;tenant='Khách thuê mô phỏng';rent='2800000';electricRate='3000';waterRate='10000';electric='0';water='0'}
$p=Page $owner "/?tab=meters&period=$period"
$form=[regex]::Matches($p.Content,'(?s)<form class="meter-row".*?</form>') | Where-Object {$_.Value.Contains("<strong>$room</strong>")} | Select-Object -First 1
$roomId=Capture $form.Value 'name="id" value="([a-f0-9]+)"'
$p=Page $owner '/action/read' @{csrf=$csrf;period=$period;id=$roomId;electric='10';water='2'}
$p=Page $owner '/action/draft' @{csrf=$csrf;period=$period}
$article=[regex]::Matches($p.Content,'(?s)<article class="list-row">.*?</article>') | Where-Object {$_.Value.Contains("<strong>$room</strong>")} | Select-Object -First 1
$invoice=Capture $article.Value 'href="/invoice/([a-f0-9]+)"'
$p=Page $owner "/invoice/$invoice"
$revision=Capture $p.Content 'name="revision" value="([a-f0-9]+)"'
$p=Page $owner '/action/issue' @{csrf=$csrf;id=$invoice;revision=$revision}
$p=Page $owner '/action/invite' @{csrf=$csrf;id=$invoice}
$invite=Capture $p.Content '/p/([a-f0-9]{48})'
$code=Capture $p.Content '<strong>([0-9]{8})</strong>'
$p=Page $guest "/p/$invite/qr"
Check ($p.StatusCode -eq 404) 'unverified link cannot fetch QR'
$p=Page $reader "/p/$invite" @{code=$code}
$memo=Capture $p.Content '(ZS[0-9A-HJKMNP-TV-Z]{8})'
Check ($p.Content.Contains('2.850.000')) 'verified reader sees amount and memo'
$qr=Page $reader "/p/$invite/qr"
Check ($qr.StatusCode -eq 200 -and [string]$qr.Headers['Content-Type'] -eq 'image/png') 'verified reader receives a private PNG'
$qrBeforeHash=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($qr.RawContentStream.ToArray()))
$ref='BANK-'+[Guid]::NewGuid().ToString('N').Substring(0,8)
$csv="ma_giao_dich,ngay,so_tien,noi_dung`n$ref,2026-11-20,350000,CK $memo"
if ($MockBase) {
    $event = @{ transaction_id=$ref; date='2026-11-20'; amount_vnd=350000; account='9999000000001'; bin='970407'; direction='credit'; memo="CK $memo" } | ConvertTo-Json -Compress
    $endpoint = "$Base/webhooks/mock-bank"
    $stamp = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
    $headers = @{'X-ZS-Mock-Timestamp'="$stamp"; 'X-ZS-Mock-Signature'=('0' * 64)}
    $result = Invoke-WebRequest $endpoint -Method Post -ContentType 'application/json' -Body $event -Headers $headers -SkipHttpErrorCheck
    Check ($result.StatusCode -eq 401) 'mock receiver rejects a bad signature'
    $signature = [Security.Cryptography.HMACSHA256]::HashData([Text.Encoding]::UTF8.GetBytes($MockSecret), [Text.Encoding]::UTF8.GetBytes("$stamp.$event"))
    $headers['X-ZS-Mock-Signature'] = [Convert]::ToHexString($signature)
    $result = Invoke-WebRequest $endpoint -Method Post -ContentType 'application/json' -Body $event -Headers $headers -SkipHttpErrorCheck
    Check ($result.StatusCode -eq 422) 'mock receiver rejects a signed transfer into another account'
    $sample = @{reference=$ref;date='2026-11-20';amount='350000';memo="CK $memo"}
    $result = Invoke-WebRequest "$MockBase/transfer" -Method Post -Headers @{Origin=$MockBase} -Body $sample
    Check ($result.Content.Contains('Đã nhận giao dịch')) 'mock service delivers a signed payment to the review queue'
    $result = Invoke-WebRequest "$MockBase/transfer" -Method Post -Headers @{Origin=$MockBase} -Body $sample
    Check ($result.Content.Contains('Không ghi trùng')) 'mock webhook retry preserves the transaction'
    $sample.amount = '350001'
    $result = Invoke-WebRequest "$MockBase/transfer" -Method Post -Headers @{Origin=$MockBase} -Body $sample -SkipHttpErrorCheck
    Check ($result.StatusCode -eq 502 -and $result.Content.Contains('HTTP 409')) 'a repeated transaction id with a changed amount is rejected'
}
$p=Page $owner '/bank/import' @{csrf=$csrf;csv=$csv}
if ($MockBase) {
    Check ($p.Content.Contains('Đã thêm 0 giao dịch, bỏ qua 1 dòng')) 'statement backfill deduplicates the webhook transaction'
}
$transfer=Capture $p.Content ('(?s)'+[regex]::Escape($ref)+'.*?href="/bank/review/([a-f0-9]+)"')
$p=Page $reader "/p/$invite"
Check ($p.Content.Contains('2.850.000')) 'ingestion does not book a payment before approval'
$p=Page $owner "/bank/review/$transfer"
Check ($p.Content.Contains($room)) 'exact reference suggests the intended room'
$form=@{csrf=$csrf;invoice=$invoice;confirm='yes';existing='new'}
$p=Page $owner "/bank/approve/$transfer" $form
$p=Page $owner "/bank/approve/$transfer" $form
$p=Page $reader "/p/$invite"
Check ($p.Content.Contains('2.500.000') -and -not $p.Content.Contains('2.150.000') -and $p.Content.Contains($memo)) 'approved partial payment updates balance once and keeps memo'
$qrAfter=Page $reader "/p/$invite/qr"
$qrAfterHash=[Convert]::ToHexString([Security.Cryptography.SHA256]::HashData($qrAfter.RawContentStream.ToArray()))
Check ($qrAfter.StatusCode -eq 200 -and $qrBeforeHash -ne $qrAfterHash) 'QR changes when the balance changes'
if ($MockBase) {
    $p = Page $owner '/bank/mock-settings' @{csrf=$csrf;mode='auto'}
    Check ($p.Content.Contains('Tự đối soát đang bật')) 'owner can enable automatic matching in the mock sandbox'
    $sample = @{reference="$ref-AUTO";date='2026-11-20';amount='2500000';memo="CK $memo"}
    $result = Invoke-WebRequest "$MockBase/transfer" -Method Post -Headers @{Origin=$MockBase} -Body $sample
    Check ($result.Content.Contains('Đã tự đối soát')) 'exact mock amount and memo are matched automatically'
    $result = Invoke-WebRequest "$MockBase/transfer" -Method Post -Headers @{Origin=$MockBase} -Body $sample
    Check ($result.Content.Contains('Không ghi trùng')) 'automatic matching retry does not double-book'
    $p = Page $reader "/p/$invite"
    Check ($p.Content.Contains('Còn phải trả</dt><dd>0 ₫')) 'automatic match clears the verified invoice balance'
    $sample.reference = "$ref-NOMATCH"
    $sample.memo = 'NO MATCH'
    $result = Invoke-WebRequest "$MockBase/transfer" -Method Post -Headers @{Origin=$MockBase} -Body $sample
    $p = Page $owner '/bank'
    Check ($result.Content.Contains('Đã nhận giao dịch') -and $p.Content.Contains($sample.reference)) 'unmatched transaction remains in manual review'
    $p = Page $owner '/bank/mock-settings' @{csrf=$csrf;mode='manual'}
    Check ($p.Content.Contains('Giao dịch mô phỏng đang chờ chủ trọ duyệt')) 'automatic matching can be switched back to manual'
}
$p=Page $owner '/action/revoke' @{csrf=$csrf;id=$invite}
$p=Page $reader "/p/$invite/qr"
Check ($p.StatusCode -eq 404) 'revoking the link also revokes QR access'
