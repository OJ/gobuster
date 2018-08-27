TARGET=./build
OSES=darwin linux windows
ARCHS=amd64 386

all:
	@mkdir -p ${TARGET}
	@for GOOS in ${OSES}; do \
		for GOARCH in ${ARCHS}; do \
			echo "Building for $${GOOS} $${GOARCH} ..." ; \
			if [ "$${GOOS}" == "windows" ]; then \
				GOOS=$${GOOS} GARCH=$${GOARCH} go build -o ${TARGET}/gobuster-$${GOARCH}.exe ; \
			else \
				GOOS=$${GOOS} GARCH=$${GOARCH} go build -o ${TARGET}/gobuster-$${GOOS}-$${GOARCH} ; \
			fi; \
		done; \
	done; \
	echo "Done."

test:
	@go test -v -race ./...

clean:
	@rm -rf ${TARGET}/* ; \
	echo "Done."
