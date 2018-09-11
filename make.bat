@echo off

SET ARG=%1
SET TARGET=.\build
SET BUILDARGS=-ldflags="-s -w" -gcflags="-trimpath=%GOPATH%\src" -asmflags="-trimpath=%GOPATH%\src"

IF "%ARG%"=="test" (
  go test -v -race ./...
  echo Done.
  GOTO Done
)

IF "%ARG%"=="clean" (
  del /F /Q %TARGET%\*.*
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

IF "%ARG%"=="all" (
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

:Darwin
set GOOS=darwin
set GOARCH=amd64
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
