.DEFAULT_GOAL := help
.PHONY: help
help: ## Display defined make tasks
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

gobuster: *.go ## Build the gobuster executable
	go build

.PHONY: bootstrap
bootstrap: ## Get developer tools
	go get -u sourcegraph.com/sqs/goreturns
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install
	go get -u ./...

.PHONY: lint
lint: ## Lint golang source code
	go vet ./...
	gofmt -s -w -l .
	goreturns -b -i -w -l .
	gometalinter --vendored-linters --enable-all --sort=path --aggregate --vendor --disable=lll .

.PHONY: test
test: gobuster ## Verify gobuster base capabilities
	./gobuster -u http://github.com -w fixtures/urls.txt
	./gobuster -m dns -u github.com -w fixtures/subdomains.txt
