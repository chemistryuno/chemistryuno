@echo off
echo Starting backend server...
cd /d %~dp0backend
set CGO_LDFLAGS=-g -O2 -Wl,--no-high-entropy-va
go run main.go
pause
