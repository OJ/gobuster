@echo off

SET TARGET=.\build

IF %1=="test" (
  go test -v -race ./...
)

IF %1=="clean" (
  del /F %TARGET%\*.*
)

IF %1=="" (
  echo "Building for windows ..."
  set GOOS=windows
  set GOARCH=amd64
  go build -o %TARGET%\gobuster-%GOARCH%.exe
  set GOARCH=386
  go build -o %TARGET%\gobuster-%GOARCH%.exe

  echo "Building for osx ..."
  set GOOS=osx
  set GOARCH=amd64
  go build -o %TARGET%\gobuster-%GOOS%-%GOARCH%
  set GOARCH=386
  go build -o %TARGET%\gobuster-%GOOS%-%GOARCH%

  echo "Building for linux ..."
  set GOOS=osx
  set GOARCH=amd64
  go build -o %TARGET%\gobuster-%GOOS%-%GOARCH%
  set GOARCH=386
  go build -o %TARGET%\gobuster-%GOOS%-%GOARCH%

  echo "Done."
)

