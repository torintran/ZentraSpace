#Requires -Version 7.3
param([ValidateRange(1, 65535)][int]$Port = 8088, [ValidateRange(1, 65535)][int]$MockPort = 8094, [switch]$Browser)

$ErrorActionPreference = 'Stop'
$PSNativeCommandUseErrorActionPreference = $true
$root = Split-Path $PSScriptRoot -Parent
$project = 'zentraspace-check-' + [Guid]::NewGuid().ToString('N').Substring(0, 12)
$envFile = Join-Path ([IO.Path]::GetTempPath()) "$project.env"
$compose = @('compose', '-f', (Join-Path $root 'compose.yaml'), '-f', (Join-Path $root 'compose.mock-bank.yaml'), '--env-file', $envFile, '-p', $project)
$base = "http://localhost:$Port"
$previous = @{}

try {
    & "$PSScriptRoot/init-local.ps1" -Path $envFile
    $config = ConvertFrom-StringData (Get-Content -Raw -LiteralPath $envFile)
    $config['ZS_PORT'] = "$Port"
    $config['ZS_ORIGIN'] = $base
    $config['ZS_LOCAL_HTTP'] = '1'
    $config['ZS_MOCK_BANK_PORT'] = "$MockPort"
    $config['ZS_MOCK_BANK_SECRET'] = [Convert]::ToHexString([Security.Cryptography.RandomNumberGenerator]::GetBytes(32))
    if ($Browser) {
        $config['ZS_BROWSER_BASE'] = $base
        $config['ZS_BROWSER_MOCK'] = "http://localhost:$MockPort"
        $config['ZS_BROWSER_ENV_FILE'] = $envFile
    }
    $secondOwnerPassword = [Guid]::NewGuid().ToString('N')
    $config['ZS_STAFF_ACCOUNTS'] = ConvertTo-Json -Compress -InputObject @(@{
        Username='owner-secondary'; Organization='secondary'; Role='owner'; Password=$secondOwnerPassword
    })
    foreach ($name in $config.Keys) {
        $previous[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
        [Environment]::SetEnvironmentVariable($name, $config[$name], 'Process')
    }

    docker @compose up --build -d --wait
    docker @compose exec -T web zentraspace seed-demo 2026-11
    docker @compose exec -T web zentraspace seed-demo 2026-11
    & "$PSScriptRoot/smoke.ps1" -Base $base -EnvFile $envFile -CheckDemo -SecondOwnerPassword $secondOwnerPassword
    & "$PSScriptRoot/bank-smoke.ps1" -Base $base -EnvFile $envFile -MockBase "http://localhost:$MockPort" -MockSecret $config.ZS_MOCK_BANK_SECRET
    if ($Browser) {
        npm --prefix (Join-Path $root 'tests/browser') test
    }

    @'
BEGIN;
SELECT set_config('app.organization', 'restore-check-org', true);
INSERT INTO books (organization, sealed) VALUES ('restore-check-org', decode('00', 'hex'));
SET LOCAL ROLE zentra_runtime;
DO $$
BEGIN
    IF (SELECT array_agg(organization) FROM books) IS DISTINCT FROM ARRAY['restore-check-org'] THEN
        RAISE EXCEPTION 'Runtime role can see rows outside its organization or cannot see its own row';
    END IF;
    BEGIN
        INSERT INTO books (organization, sealed) VALUES ('other-org', decode('00', 'hex'));
        RAISE EXCEPTION 'Runtime role can insert under another organization';
    EXCEPTION WHEN insufficient_privilege THEN
        NULL;
    END;
END
$$;
SELECT set_config('app.organization', '', true);
DO $$
BEGIN
    IF EXISTS (SELECT FROM books) THEN
        RAISE EXCEPTION 'Runtime role without an organization can read books';
    END IF;
END
$$;
ROLLBACK;
'@ | docker @compose exec -T db psql -X -U postgres -d zentraspace -v ON_ERROR_STOP=1 -q
    'PASS PostgreSQL isolates organization reads and writes'

    docker @compose exec -T db pg_dump -U postgres -Fc -f /tmp/zentraspace-check.dump zentraspace
    docker @compose exec -T db createdb -U postgres zentraspace_restore
    docker @compose exec -T db pg_restore -U postgres --exit-on-error -d zentraspace_restore /tmp/zentraspace-check.dump
    $snapshot = "SELECT organization, '' AS receipt, sealed FROM books UNION ALL SELECT organization, entry_key, sealed FROM cash_receipts ORDER BY 1, 2"
    $original = docker @compose exec -T db psql -X -U postgres -d zentraspace -At -v ON_ERROR_STOP=1 -c $snapshot
    $restored = docker @compose exec -T db psql -X -U postgres -d zentraspace_restore -At -v ON_ERROR_STOP=1 -c $snapshot
    if (!$original -or ($original -join "`n") -cne ($restored -join "`n")) {
        throw 'Restored books differ from the source database.'
    }
    $visible = docker @compose exec -T db psql -X -U postgres -d zentraspace_restore -qAt -v ON_ERROR_STOP=1 -c 'SET ROLE zentra_runtime; SELECT (SELECT count(*) FROM books) + (SELECT count(*) FROM cash_receipts)'
    if ($visible -ne '0') {
        throw 'Restored database exposes books to the runtime role without an organization.'
    }
    'PASS isolated restore preserves encrypted books and runtime read restrictions'
}
finally {
    try {
        if (Test-Path -LiteralPath $envFile) {
            docker @compose down --volumes
        }
    }
    finally {
        foreach ($name in $previous.Keys) {
            [Environment]::SetEnvironmentVariable($name, $previous[$name], 'Process')
        }
        if (Test-Path -LiteralPath $envFile) {
            Remove-Item -LiteralPath $envFile
        }
    }
}
