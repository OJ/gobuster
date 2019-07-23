TARGET=./build
ARCHS=amd64 386
LDFLAGS="-s -w"
GCFLAGS="all=-trimpath=$(shell pwd)"
ASMFLAGS="all=-trimpath=$(shell pwd)"

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
		GOOS=windows GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -gcflags=${GCFLAGS} -asmflags=${ASMFLAGS} -o ${TARGET}/gobuster-windows-$${GOARCH}/gobuster.exe ; \
	done; \
	echo "Done."

linux:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for linux $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gobuster-linux-$${GOARCH} ; \
		GOOS=linux GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -gcflags=${GCFLAGS} -asmflags=${ASMFLAGS} -o ${TARGET}/gobuster-linux-$${GOARCH}/gobuster ; \
	done; \
	echo "Done."

darwin:
	@for GOARCH in ${ARCHS}; do \
		echo "Building for darwin $${GOARCH} ..." ; \
		mkdir -p ${TARGET}/gobuster-darwin-$${GOARCH} ; \
		GOOS=darwin GOARCH=$${GOARCH} GO111MODULE=on CGO_ENABLED=0 go build -ldflags=${LDFLAGS} -gcflags=${GCFLAGS} -asmflags=${ASMFLAGS} -o ${TARGET}/gobuster-darwin-$${GOARCH}/gobuster ; \
	done; \
	echo "Done."

all: clean fmt update lint test darwin linux windows

test:
	@go test -v -race ./... ; \
	echo "Done."

lint:
	@go get -u github.com/golangci/golangci-lint@master ; \
	golangci-lint run ./... ; \
	go mod tidy ; \
	echo Done

clean:
	@rm -rf ${TARGET}/* ; \
	go clean ./... ; \
	echo "Done."
