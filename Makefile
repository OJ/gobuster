TARGET=./build
ARCHS=amd64 386
LDFLAGS="-s -w"

.PHONY: current
current:
	go build -o ./gobuster

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: update
update:
	go get -u
	go mod tidy -v

.PHONY: windows
windows:
	mkdir -p ${TARGET}
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-windows-386.exe
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-windows-amd64.exe

.PHONY: linux
linux:
	mkdir -p ${TARGET}
	GOOS=linux GOARCH=386 CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-linux-386
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-linux-amd64
	GOOS=linux GOARCH=arm CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-linux-arm
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-linux-arm64

.PHONY: darwin
darwin:
	mkdir -p ${TARGET}
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-darwin-arm64
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-darwin-amd64

.PHONY: all
all: clean fmt update test lint darwin linux windows

.PHONY: test
test:
	go test -v -race ./...

.PHONY: lint
lint:
	"$$(go env GOPATH)/bin/golangci-lint" run ./...
	go mod tidy

.PHONY: lint-update
lint-update:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin
	$$(go env GOPATH)/bin/golangci-lint --version

.PHONY: lint-docker
lint-docker:
	docker pull golangci/golangci-lint:latest
	docker run --rm -v $$(pwd):/app -w /app golangci/golangci-lint:latest golangci-lint run

.PHONY: clean
clean:
	rm -rf ${TARGET}/*
	go clean ./...
