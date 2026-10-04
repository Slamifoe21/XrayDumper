@echo off
setlocal

for /f "delims=" %%i in ('git rev-parse --short HEAD 2^>nul') do set COMMIT=%%i
if "%COMMIT%"=="" set COMMIT=none

go build -ldflags="-s -w -X main.version=1.0.0 -X main.commit=%COMMIT%" -trimpath -o ./output/xd.exe

set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w -X main.version=1.0.0 -X main.commit=%COMMIT%" -trimpath -o ./output/xd-linux

endlocal
