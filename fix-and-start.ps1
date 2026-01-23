# Chemistry UNO Alpha - 自动安装MinGW修复脚本
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Chemistry UNO Alpha - GCC修复工具" -ForegroundColor Cyan  
Write-Host "============================================`n" -ForegroundColor Cyan

$toolsDir = Join-Path $PSScriptRoot "tools"
$mingwDir = Join-Path $toolsDir "mingw64"
$mingwBinDir = Join-Path $mingwDir "bin"

# 检查是否已有新版GCC
if (Test-Path (Join-Path $mingwBinDir "gcc.exe")) {
    Write-Host "✓ 检测到已安装的MinGW" -ForegroundColor Green
    $gccVersion = & "$mingwBinDir\gcc.exe" --version | Select-Object -First 1
    Write-Host "  版本: $gccVersion`n" -ForegroundColor Gray
} else {
    Write-Host "需要下载并安装MinGW-w64..." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "请选择安装方式:" -ForegroundColor Cyan
    Write-Host "1. 自动下载安装 (推荐, 约120MB)" -ForegroundColor White
    Write-Host "2. 手动下载安装" -ForegroundColor White
    Write-Host "3. 退出" -ForegroundColor White
    Write-Host ""
    
    $choice = Read-Host "请输入选择 (1-3)"
    
    switch ($choice) {
        "1" {
            Write-Host "`n开始自动下载..." -ForegroundColor Green
            
            # WinLibs GCC 下载链接 (GCC 13.2.0)
            $downloadUrl = "https://github.com/brechtsanders/winlibs_mingw/releases/download/13.2.0posix-17.0.6-11.0.1-ucrt-r5/winlibs-x86_64-posix-seh-gcc-13.2.0-mingw-w64ucrt-11.0.1-r5.7z"
            $zipFile = Join-Path $toolsDir "mingw.7z"
            
            if (-not (Test-Path $toolsDir)) {
                New-Item -ItemType Directory -Path $toolsDir | Out-Null
            }
            
            try {
                Write-Host "正在下载 MinGW-w64 (这可能需要几分钟)..." -ForegroundColor Yellow
                Invoke-WebRequest -Uri $downloadUrl -OutFile $zipFile -UseBasicParsing
                
                Write-Host "下载完成! 正在解压..." -ForegroundColor Green
                
                # 检查是否有7z命令
                if (Get-Command 7z -ErrorAction SilentlyContinue) {
                    & 7z x $zipFile -o"$toolsDir" -y
                } else {
                    Write-Host "未找到7-Zip，尝试使用PowerShell解压..." -ForegroundColor Yellow
                    Write-Host "注意: 7z格式需要7-Zip，如果失败请手动安装" -ForegroundColor Red
                    Write-Host "7-Zip下载: https://www.7-zip.org/download.html" -ForegroundColor Yellow
                    pause
                    exit 1
                }
                
                Remove-Item $zipFile -Force
                Write-Host "安装完成!" -ForegroundColor Green
                
            } catch {
                Write-Host "下载失败: $_" -ForegroundColor Red
                Write-Host "请选择手动安装方式" -ForegroundColor Yellow
                pause
                exit 1
            }
        }
        
        "2" {
            Write-Host "`n手动安装步骤:" -ForegroundColor Yellow
            Write-Host "1. 访问: https://winlibs.com/" -ForegroundColor White
            Write-Host "2. 下载: UCRT runtime 版本 (推荐 GCC 13.x)" -ForegroundColor White
            Write-Host "3. 解压到: $mingwDir" -ForegroundColor White
            Write-Host "4. 确保存在: $mingwBinDir\gcc.exe`n" -ForegroundColor White
            
            Write-Host "完成后按任意键继续..." -ForegroundColor Cyan
            pause
            
            if (-not (Test-Path (Join-Path $mingwBinDir "gcc.exe"))) {
                Write-Host "错误: 未找到GCC，请检查安装路径" -ForegroundColor Red
                pause
                exit 1
            }
        }
        
        default {
            Write-Host "已取消" -ForegroundColor Gray
            exit 0
        }
    }
}

# 设置环境变量并启动
Write-Host "`n正在配置环境..." -ForegroundColor Cyan
$env:PATH = "$mingwBinDir;$env:PATH"

Write-Host "验证GCC版本..." -ForegroundColor Cyan
gcc --version

Write-Host "`n正在启动后端服务器..." -ForegroundColor Green
Set-Location (Join-Path $PSScriptRoot "backend")
go run main.go
