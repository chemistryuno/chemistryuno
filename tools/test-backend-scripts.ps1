$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..")
$env:GOCACHE = Join-Path $repoRoot ".gocache"

$scriptDir = Join-Path $repoRoot "tests\_backend\scripts"
$files = Get-ChildItem -Path $scriptDir -Filter *.go | Sort-Object Name

if ($files.Count -eq 0) {
  Write-Host "No Go scripts found in $scriptDir"
  exit 0
}

$failed = $false
$results = @()

foreach ($file in $files) {
  $backendPath = "backend/scripts/$($file.Name)"
  Write-Host "==> node scripts/run-backend-tests.js -tags scripts $backendPath"
  & node scripts/run-backend-tests.js -tags scripts $backendPath
  $exit = $LASTEXITCODE
  $results += [pscustomobject]@{ File = $file.Name; ExitCode = $exit }
  if ($exit -ne 0) {
    $failed = $true
  }
}

Write-Host ""
Write-Host "Results:"
$results | Format-Table -AutoSize

if ($failed) {
  exit 1
}
exit 0
