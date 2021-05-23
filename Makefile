TARGET=./build
ARCHS=amd64 386
LDFLAGS="-s -w"

.PHONY: current
current:
	@go build -o ./gobuster; \
	echo "Done."

.PHONY: fmt
fmt:
	@go fmt ./...; \
	echo "Done."

.PHONY: update
update:
	@go get -u; \
	go mod tidy -v; \
	echo "Done."

.PHONY: windows
windows:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for windows $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gobuster-windows-$${GOARCH} ; \
		GOOS=windows GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-windows-$${GOARCH}/gobuster.exe ; \
	done; \
	echo "Done."

.PHONY: linux
linux:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for linux $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gobuster-linux-$${GOARCH} ; \
		GOOS=linux GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-linux-$${GOARCH}/gobuster ; \
	done; \
	echo "Done."

.PHONY: darwin
darwin:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for darwin $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gobuster-darwin-$${GOARCH} ; \
		GOOS=darwin GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-darwin-$${GOARCH}/gobuster ; \
	done; \
	echo "Done."

.PHONY: all
all: clean fmt update test lint darwin linux windows

.PHONY: test
test:
	@go test -v -race ./... ; \
	echo "Done."

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
	@rm -rf ${TARGET}/* ; \
	go clean ./... ; \
	echo "Done."
