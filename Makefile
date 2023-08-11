.DEFAULT_GOAL := linux

.PHONY: linux
linux:
	go build -o ./gobuster

.PHONY: windows
windows:
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o ./gobuster.exe

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: update
update:
	go get -u
	go mod tidy -v

.PHONY: all
all: fmt update linux windows test lint

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

.PHONY: tag
tag:
	@[ "${TAG}" ] && echo "Tagging a new version ${TAG}" || ( echo "TAG is not set"; exit 1 )
	git tag -a "${TAG}" -m "${TAG}"
	git push origin "${TAG}"
