Gobuster v2.0.1 (OJ Reeves @TheColonial)
========================================

Gobuster is a tool used to brute-force:

* URIs (directories and files) in web sites.
* DNS subdomains (with wildcard support).

### Tags, Statuses, etc

[![Build Status](https://travis-ci.com/OJ/gobuster.svg?branch=master)](https://travis-ci.com/OJ/gobuster) [![Backers on Open Collective](https://opencollective.com/gobuster/backers/badge.svg)](#backers) [![Sponsors on Open Collective](https://opencollective.com/gobuster/sponsors/badge.svg)](#sponsors)

### Oh dear God.. WHY!?

Because I wanted:

1. ... something that didn't have a fat Java GUI (console FTW).
1. ... to build something that just worked on the command line.
1. ... something that did not do recursive brute force.
1. ... something that allowed me to brute force folders and multiple extensions at once.
1. ... something that compiled to native on multiple platforms.
1. ... something that was faster than an interpreted script (such as Python).
1. ... something that didn't require a runtime.
1. ... use something that was good with concurrency (hence Go).
1. ... to build something in Go that wasn't totally useless.

### But it's shit! And your implementation sucks!

Yes, you're probably correct. Feel free to:

* Not use it.
* Show me how to do it better.

### Love this tool? Back it!

If you're backing us already, you rock. If you're not, that's cool too! Want to back us? [[Become a backer](https://opencollective.com/gobuster#backer)]

<a href="https://opencollective.com/gobuster#backers" target="_blank"><img src="https://opencollective.com/gobuster/backers.svg?width=890"></a>

### Common Command line options

* `-fw` - force processing of a domain with wildcard results.
* `-np` - hide the progress output.
* `-m <mode>` - which mode to use, either `dir` or `dns` (default: `dir`).
* `-q` - disables banner/underline output.
* `-t <threads>` - number of threads to run (default: `10`).
* `-u <url/domain>` - full URL (including scheme), or base domain name.
* `-v` - verbose output (show all results).
* `-w <wordlist>` - path to the wordlist used for brute forcing (use `-` for stdin).

### Command line options for `dns` mode

* `-cn` - show CNAME records (cannot be used with '-i' option).
* `-i` - show all IP addresses for the result.

### Command line options for `dir` mode

* `-a <user agent string>` - specify a user agent string to send in the request header.
* `-c <http cookies>` - use this to specify any cookies that you might need (simulating auth).
* `-e` - specify extended mode that renders the full URL.
* `-f` - append `/` for directory brute forces.
* `-k` - Skip verification of SSL certificates.
* `-l` - show the length of the response.
* `-n` - "no status" mode, disables the output of the result's status code.
* `-o <file>` - specify a file name to write the output to.
* `-p <proxy url>` - specify a proxy to use for all requests (scheme much match the URL scheme).
* `-r` - follow redirects.
* `-s <status codes>` - comma-separated set of the list of status codes to be deemed a "positive" (default: `200,204,301,302,307`).
* `-x <extensions>` - list of extensions to check for, if any.
* `-P <password>` - HTTP Authorization password (Basic Auth only, prompted if missing).
* `-U <username>` - HTTP Authorization username (Basic Auth only).
* `-to <timeout>` - HTTP timeout. Examples: 10s, 100ms, 1m (default: 10s).

### Building

Since this tool is written in [Go](https://golang.org/) you need install the Go language/compiler/etc. Full details of installation and set up can be found [on the Go language website](https://golang.org/doc/install). Once installed you have two options.

#### Compiling
`gobuster` now has external dependencies, and so they need to be pulled in first:
```
gobuster $ go get && go build
```
This will create a `gobuster` binary for you. If you want to install it in the `$GOPATH/bin` folder you can run:
```
gobuster $ go install
```
If you have all the dependencies already, you can make use of the build scripts:
* `make` - builds for the current Go configuration (ie. runs `go build`).
* `make windows` - builds 32 and 64 bit binaries for windows, and writes them to the `build` subfolder.
* `make linux` - builds 32 and 64 bit binaries for linux, and writes them to the `build` subfolder.
* `make darwin` - builds 32 and 64 bit binaries for darwin, and writes them to the `build` subfolder.
* `make all` - builds for all platforms and architectures, and writes the resulting binaries to the `build` subfolder.
* `make clean` - clears out the `build` subfolder.
* `make test` - runs the tests.

#### Running as a script
```
gobuster $ go run main.go <parameters>
```

### Wordlists via STDIN
Wordlists can be piped into `gobuster` via stdin by providing a `-` to the `-w` option:
```
hashcat -a 3 --stdout ?l | gobuster -u https://mysite.com -w -
```
Note: If the `-w` option is specified at the same time as piping from STDIN, an error will be shown and the program will terminate.

### Examples

#### `dir` mode

Command line might look like this:
```
$ gobuster -u https://mysite.com/path/to/folder -c 'session=123456' -t 50 -w common-files.txt -x .php,.html
```
Default options looks like this:
```
$ gobuster -u https://buffered.io -w ~/wordlists/shortlist.txt

=====================================================
Gobuster v2.0.1              OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dir
[+] Url/Domain   : https://buffered.io/
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/shortlist.txt
[+] Status codes : 200,204,301,302,307,403
[+] Timeout      : 10s
=====================================================
2018/08/27 11:49:43 Starting gobuster
=====================================================
/categories (Status: 301)
/contact (Status: 301)
/posts (Status: 301)
/index (Status: 200)
=====================================================
2018/08/27 11:49:44 Finished
=====================================================
```
Default options with status codes disabled looks like this:
```
$ gobuster -u https://buffered.io -w ~/wordlists/shortlist.txt -n

=====================================================
Gobuster v2.0.1              OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dir
[+] Url/Domain   : https://buffered.io/
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/shortlist.txt
[+] Status codes : 200,204,301,302,307,403
[+] No status    : true
[+] Timeout      : 10s
=====================================================
2018/08/27 11:50:18 Starting gobuster
=====================================================
/categories
/contact
/index
/posts
=====================================================
2018/08/27 11:50:18 Finished
=====================================================
```
Verbose output looks like this:
```
$ gobuster -u https://buffered.io -w ~/wordlists/shortlist.txt -v

=====================================================
Gobuster v2.0.1              OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dir
[+] Url/Domain   : https://buffered.io/
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/shortlist.txt
[+] Status codes : 200,204,301,302,307,403
[+] Verbose      : true
[+] Timeout      : 10s
=====================================================
2018/08/27 11:50:51 Starting gobuster
=====================================================
Missed: /alsodoesnotexist (Status: 404)
Found: /index (Status: 200)
Missed: /doesnotexist (Status: 404)
Found: /categories (Status: 301)
Found: /posts (Status: 301)
Found: /contact (Status: 301)
=====================================================
2018/08/27 11:50:51 Finished
=====================================================
```
Example showing content length:
```
$ gobuster -u https://buffered.io -w ~/wordlists/shortlist.txt -l

=====================================================
Gobuster v2.0.1              OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dir
[+] Url/Domain   : https://buffered.io/
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/shortlist.txt
[+] Status codes : 200,204,301,302,307,403
[+] Show length  : true
[+] Timeout      : 10s
=====================================================
2018/08/27 11:51:16 Starting gobuster
=====================================================
/categories (Status: 301) [Size: 178]
/posts (Status: 301) [Size: 178]
/contact (Status: 301) [Size: 178]
/index (Status: 200) [Size: 51759]
=====================================================
2018/08/27 11:51:17 Finished
=====================================================
```
Quiet output, with status disabled and expanded mode looks like this ("grep mode"):
```
$ gobuster -u https://buffered.io -w ~/wordlists/shortlist.txt -q -n -e
https://buffered.io/index
https://buffered.io/contact
https://buffered.io/posts
https://buffered.io/categories
```

#### `dns` mode

Command line might look like this:
```
$ gobuster -m dns -u mysite.com -t 50 -w common-names.txt
```
Normal sample run goes like this:
```
$ gobuster -m dns -w ~/wordlists/subdomains.txt -u google.com

=====================================================
Gobuster v2.0.1              OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dns
[+] Url/Domain   : google.com
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/subdomains.txt
=====================================================
2018/08/27 11:54:20 Starting gobuster
=====================================================
Found: chrome.google.com
Found: ns1.google.com
Found: admin.google.com
Found: www.google.com
Found: m.google.com
Found: support.google.com
Found: translate.google.com
Found: cse.google.com
Found: news.google.com
Found: music.google.com
Found: mail.google.com
Found: store.google.com
Found: mobile.google.com
Found: search.google.com
Found: wap.google.com
Found: directory.google.com
Found: local.google.com
Found: blog.google.com
=====================================================
2018/08/27 11:54:20 Finished
=====================================================
```
Show IP sample run goes like this:
```
$ gobuster -m dns -w ~/wordlists/subdomains.txt -u google.com -i

=====================================================
Gobuster v2.0.1              OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dns
[+] Url/Domain   : google.com
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/subdomains.txt
=====================================================
2018/08/27 11:54:54 Starting gobuster
=====================================================
Found: www.google.com [172.217.25.36, 2404:6800:4006:802::2004]
Found: admin.google.com [172.217.25.46, 2404:6800:4006:806::200e]
Found: store.google.com [172.217.167.78, 2404:6800:4006:802::200e]
Found: mobile.google.com [172.217.25.43, 2404:6800:4006:802::200b]
Found: ns1.google.com [216.239.32.10, 2001:4860:4802:32::a]
Found: m.google.com [172.217.25.43, 2404:6800:4006:802::200b]
Found: cse.google.com [172.217.25.46, 2404:6800:4006:80a::200e]
Found: chrome.google.com [172.217.25.46, 2404:6800:4006:802::200e]
Found: search.google.com [172.217.25.46, 2404:6800:4006:802::200e]
Found: local.google.com [172.217.25.46, 2404:6800:4006:80a::200e]
Found: news.google.com [172.217.25.46, 2404:6800:4006:802::200e]
Found: blog.google.com [216.58.199.73, 2404:6800:4006:806::2009]
Found: support.google.com [172.217.25.46, 2404:6800:4006:802::200e]
Found: wap.google.com [172.217.25.46, 2404:6800:4006:802::200e]
Found: directory.google.com [172.217.25.46, 2404:6800:4006:802::200e]
Found: translate.google.com [172.217.25.46, 2404:6800:4006:802::200e]
Found: music.google.com [172.217.25.46, 2404:6800:4006:802::200e]
Found: mail.google.com [172.217.25.37, 2404:6800:4006:802::2005]
=====================================================
2018/08/27 11:54:55 Finished
=====================================================
```
Base domain validation warning when the base domain fails to resolve. This is a warning rather than a failure in case the user fat-fingers while typing the domain.
```
$ gobuster -m dns -w ~/wordlists/subdomains.txt -u yp.to -i

=====================================================
Gobuster v2.0.1              OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dns
[+] Url/Domain   : yp.to
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/subdomains.txt
=====================================================
2018/08/27 11:56:43 Starting gobuster
=====================================================
2018/08/27 11:56:53 [-] Unable to validate base domain: yp.to
Found: cr.yp.to [131.193.32.108, 131.193.32.109]
=====================================================
2018/08/27 11:56:53 Finished
=====================================================
```
Wildcard DNS is also detected properly:
```
$ gobuster -m dns -w ~/wordlists/subdomains.txt -u 0.0.1.xip.io        

=====================================================
Gobuster v2.0.1              OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dns
[+] Url/Domain   : 0.0.1.xip.io
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/subdomains.txt
=====================================================
2018/08/27 12:13:48 Starting gobuster
=====================================================
2018/08/27 12:13:48 [-] Wildcard DNS found. IP address(es): 1.0.0.0
2018/08/27 12:13:48 [!] To force processing of Wildcard DNS, specify the '-fw' switch.
=====================================================
2018/08/27 12:13:48 Finished
=====================================================
```
If the user wants to force processing of a domain that has wildcard entries, use `-fw`:
```
$ gobuster -m dns -w ~/wordlists/subdomains.txt -u 0.0.1.xip.io -fw

=====================================================
Gobuster v2.0.1              OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dns
[+] Url/Domain   : 0.0.1.xip.io
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/subdomains.txt
=====================================================
2018/08/27 12:13:51 Starting gobuster
=====================================================
2018/08/27 12:13:51 [-] Wildcard DNS found. IP address(es): 1.0.0.0
Found: 127.0.0.1.xip.io
Found: test.127.0.0.1.xip.io
=====================================================
2018/08/27 12:13:53 Finished
=====================================================
```

### License

See the LICENSE file.

### Thanks

See the THANKS file for people who helped out.
