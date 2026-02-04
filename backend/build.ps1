# Chemistry UNO 后端编译脚本 (PowerShell)
# 禁用CGO以避免MinGW链接器版本问题

Write-Host "🔨 正在编译后端 (禁用CGO)..." -ForegroundColor Cyan

$env:CGO_ENABLED = "0"
go build -o chemistryuno.exe

if ($LASTEXITCODE -eq 0) {
    $fileSize = [math]::Round((Get-Item chemistryuno.exe).Length / 1MB, 2)
    Write-Host "✅ 编译成功！" -ForegroundColor Green
    Write-Host "📦 可执行文件: chemistryuno.exe ($fileSize MB)" -ForegroundColor Green
} else {
    Write-Host "❌ 编译失败" -ForegroundColor Red
    exit 1
}

$env:CGO_ENABLED = "1"
