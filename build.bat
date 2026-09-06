@echo off
go build -ldflags="-s -w" -trimpath -o ./output/xd.exe
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -trimpath -o ./output/xd-linux

