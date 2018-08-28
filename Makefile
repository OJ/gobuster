TARGET=./build
OSES=darwin linux windows
ARCHS=amd64 386

current: outputdir
	@go build -o ./gobuster; \
	echo "Done."

outputdir:
	@mkdir -p ${TARGET}

windows: outputdir
	@for GOARCH in ${ARCHS}; do \
		echo "Building for windows $${GOARCH} ..." ; \
		GOOS=windows GARCH=$${GOARCH} go build -o ${TARGET}/gobuster-$${GOARCH}.exe ; \
	done; \
	echo "Done."

linux: outputdir
	@for GOARCH in ${ARCHS}; do \
		echo "Building for linux $${GOARCH} ..." ; \
		GOOS=linux GARCH=$${GOARCH} go build -o ${TARGET}/gobuster-linux-$${GOARCH} ; \
	done; \
	echo "Done."

darwin: outputdir
	@for GOARCH in ${ARCHS}; do \
		echo "Building for darwin $${GOARCH} ..." ; \
		GOOS=darwin GARCH=$${GOARCH} go build -o ${TARGET}/gobuster-darwin-$${GOARCH} ; \
	done; \
	echo "Done."


all: darwin linux windows

test:
	@go test -v -race ./... ; \
	echo "Done."

clean:
	@rm -rf ${TARGET}/* ; \
	echo "Done."
