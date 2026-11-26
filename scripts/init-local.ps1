$ErrorActionPreference = 'Stop'
$path = Join-Path (Split-Path $PSScriptRoot -Parent) '.env'
if (Test-Path -LiteralPath $path) { Write-Output 'Local configuration already exists.'; exit 0 }
function New-Secret {
  $bytes = New-Object byte[] 32
  $random = [System.Security.Cryptography.RandomNumberGenerator]::Create()
  try { $random.GetBytes($bytes) } finally { $random.Dispose() }
  return ($bytes | ForEach-Object { $_.ToString('x2') }) -join ''
}
$keys = 'DB_ADMIN_PASSWORD','DB_RUNTIME_PASSWORD','ZS_MASTER_KEY','ZS_OWNER_PASSWORD','ZS_METER_PASSWORD'
$lines = foreach ($key in $keys) { "$key=$(New-Secret)" }
[IO.File]::WriteAllLines($path, [string[]]$lines, [Text.UTF8Encoding]::new($false))
Write-Output 'Created local credentials in .env.'
