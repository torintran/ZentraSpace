#Requires -Version 7.3
param(
    [ValidateRange(1024, 65535)][int]$AppPort = 8092,
    [ValidateRange(1024, 65535)][int]$MockPort = 8093,
    [string]$Period = '2026-11'
)

$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true
$root = Split-Path $PSScriptRoot -Parent
$compose = @('compose', '-f', (Join-Path $root 'compose.yaml'), '-f', (Join-Path $root 'compose.mock-bank.yaml'),
    '--env-file', (Join-Path $root '.env'), '-p', 'zentraspace-bank-demo')

if ($AppPort -eq $MockPort) { throw 'Choose different app and mock bank ports.' }
$settings = @{
    ZS_PORT = "$AppPort"
    ZS_ORIGIN = "http://localhost:$AppPort"
    ZS_LOCAL_HTTP = '1'
    ZS_MOCK_BANK_PORT = "$MockPort"
    ZS_MOCK_BANK_SECRET = [Convert]::ToHexString([Security.Cryptography.RandomNumberGenerator]::GetBytes(32))
    ZS_STAFF_ACCOUNTS = '[]'
}
$previous = @{}
try {
    & "$PSScriptRoot/init-local.ps1"
    foreach ($name in $settings.Keys) {
        $previous[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
        [Environment]::SetEnvironmentVariable($name, $settings[$name], 'Process')
    }
    docker @compose up --build -d --wait
    docker @compose exec -T web zentraspace seed-bank-demo $Period
    "ZentraSpace: http://localhost:$AppPort/?tab=billing&period=$Period"
    "Mock bank: http://localhost:$MockPort/"
}
finally {
    foreach ($name in $previous.Keys) {
        [Environment]::SetEnvironmentVariable($name, $previous[$name], 'Process')
    }
}
