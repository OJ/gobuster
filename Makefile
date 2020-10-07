TARGET=./build
ARCHS=amd64 386
LDFLAGS="-s -w"

current:
	@go build -o ./gobuster; \
	echo "Done."

fmt:
	@go fmt ./...; \
	echo "Done."

update:
	@go get -u; \
	go mod tidy -v; \
	echo "Done."

windows:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for windows $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gobuster-windows-$${GOARCH} ; \
		GOOS=windows GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-windows-$${GOARCH}/gobuster.exe ; \
	done; \
	echo "Done."

linux:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for linux $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gobuster-linux-$${GOARCH} ; \
		GOOS=linux GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-linux-$${GOARCH}/gobuster ; \
	done; \
	echo "Done."

darwin:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for darwin $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gobuster-darwin-$${GOARCH} ; \
		GOOS=darwin GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -trimpath -o ${TARGET}/gobuster-darwin-$${GOARCH}/gobuster ; \
	done; \
	echo "Done."

all: clean fmt update test lint darwin linux windows

test:
	@go test -v -race ./... ; \
	echo "Done."

lint:
	@if [ ! -f "$$(go env GOPATH)/bin/golangci-lint" ]; then \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.29.0; \
	fi
	"$$(go env GOPATH)/bin/golangci-lint" run ./...
	go mod tidy

clean:
	@rm -rf ${TARGET}/* ; \
	go clean ./... ; \
	echo "Done."
