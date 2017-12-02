.PHONY: test
test: gobuster
	./gobuster -u http://github.com -w urls.txt
	./gobuster -m dns -u github.com -w subdomains.txt

gobuster: *.go
	go build
