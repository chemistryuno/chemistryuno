$ErrorActionPreference = "Stop"

Write-Host "Starting CI test script..."

function Invoke-Step {
  param(
    [string]$Name,
    [scriptblock]$Action
  )

  Write-Host ""
  Write-Host "==> $Name"
  & $Action

  if ($LASTEXITCODE -ne 0) {
    throw "$Name failed with exit code $LASTEXITCODE"
  }
}

Invoke-Step "Run backend tests" { node scripts/run-backend-tests.js ./backend/... }
Invoke-Step "Run frontend type-check" { pnpm -C frontend type-check }
Invoke-Step "Run frontend build" { pnpm -C frontend build }

Write-Host ""
Write-Host "CI test script completed successfully."
