@echo off
cd backend
echo 正在测试Session ID生成...
go test -v ./utils -run TestGenerateSessionID 2>&1
echo.
echo 测试完成！
pause
