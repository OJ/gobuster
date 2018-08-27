@echo off
set GOOS=windows
set GOARCH=amd64

go test -v -race ./...
go build -o gobuster.exe