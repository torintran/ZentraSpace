param([string]$Base = 'http://localhost:8088')
$ErrorActionPreference = 'Stop'
$config = ConvertFrom-StringData (Get-Content -Raw -LiteralPath (Join-Path (Split-Path $PSScriptRoot -Parent) '.env'))
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
    Invoke-WebRequest @options
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

$p=Page $owner '/bank/account' @{csrf=$csrf;bin='970407';account='9999000000001';holder='TAI KHOAN MAU';confirm='yes'}
Check ($p.Content.Contains('0001')) 'owner configured a test beneficiary'

$p=Page $owner '/action/room' @{csrf=$csrf;period=$period;name=$room;rent='2800000';electricRate='3000';waterRate='10000';electric='0';water='0'}
$p=Page $owner "/?tab=meters&period=$period"
$form=[regex]::Matches($p.Content,'(?s)<form class="meter-row".*?</form>') | Where-Object {$_.Value.Contains("<strong>$room</strong>")} | Select-Object -First 1
$roomId=Capture $form.Value 'name="id" value="([a-f0-9]+)"'
$p=Page $owner '/action/read' @{csrf=$csrf;period=$period;id=$roomId;electric='10';water='2'}
$p=Page $owner '/action/draft' @{csrf=$csrf;period=$period}
$article=[regex]::Matches($p.Content,'(?s)<article class="list-row">.*?</article>') | Where-Object {$_.Value.Contains("<strong>$room</strong>")} | Select-Object -First 1
$invoice=Capture $article.Value 'href="/invoice/([a-f0-9]+)"'
$p=Page $owner '/action/issue' @{csrf=$csrf;id=$invoice}
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
$p=Page $owner '/bank/import' @{csrf=$csrf;csv=$csv}
$transfer=Capture $p.Content ('(?s)'+[regex]::Escape($ref)+'.*?href="/bank/review/([a-f0-9]+)"')
$p=Page $reader "/p/$invite"
Check ($p.Content.Contains('2.850.000')) 'import does not book a payment'
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
$p=Page $owner '/action/revoke' @{csrf=$csrf;id=$invite}
$p=Page $reader "/p/$invite/qr"
Check ($p.StatusCode -eq 404) 'revoking the link also revokes QR access'
