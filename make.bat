@echo off

SET ARG=%1
SET TARGET=.\build
SET BUILDARGS=-ldflags="-s -w" -gcflags="all=-trimpath=%GOPATH%\src" -asmflags="all=-trimpath=%GOPATH%\src"

IF "%ARG%"=="test" (
  CALL :Test
  GOTO Done
)

IF "%ARG%"=="clean" (
  del /F /Q %TARGET%\*.*
  go clean ./...
  echo Done.
  GOTO Done
)

IF "%ARG%"=="windows" (
  CALL :Windows
  GOTO Done
)

IF "%ARG%"=="darwin" (
  CALL :Darwin
  GOTO Done
)

IF "%ARG%"=="linux" (
  CALL :Linux
  GOTO Done
)

IF "%ARG%"=="update" (
  CALL :Update
  GOTO Done
)

IF "%ARG%"=="fmt" (
  CALL :Fmt
  GOTO Done
)

IF "%ARG%"=="all" (
  CALL :Fmt
  CALL :Update
  CALL :Lint
  CALL :Test
  CALL :Darwin
  CALL :Linux
  CALL :Windows
  GOTO Done
)

IF "%ARG%"=="" (
  go build -o .\gobuster.exe
  GOTO Done
)

GOTO Done

:Test
set GO111MODULE=on
set CGO_ENABLED=0
echo Testing ...
go test -v ./...
echo Done
EXIT /B 0

:Lint
set GO111MODULE=on
echo Linting ...
go get -u github.com/golangci/golangci-lint@master
golangci-lint run ./...
rem remove test deps
go mod tidy
echo Done

:Fmt
set GO111MODULE=on
echo Formatting ...
go fmt ./...
echo Done.
EXIT /B 0

:Update
set GO111MODULE=on
echo Updating ...
go get -u
go mod tidy -v
echo Done.
EXIT /B 0

:Darwin
set GOOS=darwin
set GOARCH=amd64
set GO111MODULE=on
set CGO_ENABLED=0
echo Building for %GOOS% %GOARCH% ...
set DIR=%TARGET%\gobuster-%GOOS%-%GOARCH%
mkdir %DIR% 2> NUL
go build %BUILDARGS% -o %DIR%\gobuster
set GOARCH=386
echo Building for %GOOS% %GOARCH% ...
set DIR=%TARGET%\gobuster-%GOOS%-%GOARCH%
mkdir %DIR% 2> NUL
go build %BUILDARGS% -o %DIR%\gobuster
echo Done.
EXIT /B 0

:Linux
set GOOS=linux
set GOARCH=amd64
set GO111MODULE=on
set CGO_ENABLED=0
echo Building for %GOOS% %GOARCH% ...
set DIR=%TARGET%\gobuster-%GOOS%-%GOARCH%
mkdir %DIR% 2> NUL
go build %BUILDARGS% -o %DIR%\gobuster
set GOARCH=386
echo Building for %GOOS% %GOARCH% ...
set DIR=%TARGET%\gobuster-%GOOS%-%GOARCH%
mkdir %DIR% 2> NUL
go build %BUILDARGS% -o %DIR%\gobuster
echo Done.
EXIT /B 0

:Windows
set GOOS=windows
set GOARCH=amd64
set GO111MODULE=on
set CGO_ENABLED=0
echo Building for %GOOS% %GOARCH% ...
set DIR=%TARGET%\gobuster-%GOOS%-%GOARCH%
mkdir %DIR% 2> NUL
go build %BUILDARGS% -o %DIR%\gobuster.exe
set GOARCH=386
echo Building for %GOOS% %GOARCH% ...
set DIR=%TARGET%\gobuster-%GOOS%-%GOARCH%
mkdir %DIR% 2> NUL
go build %BUILDARGS% -o %DIR%\gobuster.exe
echo Done.
EXIT /B 0

:Done
