# Gobuster v3.0.1 (OJ Reeves @TheColonial)

Gobuster is a tool used to brute-force:

* URIs (directories and files) in web sites.
* DNS subdomains (with wildcard support).
* Virtual Host names on target web servers.

## Tags, Statuses, etc

[![Build Status](https://travis-ci.com/OJ/gobuster.svg?branch=master)](https://travis-ci.com/OJ/gobuster) [![Backers on Open Collective](https://opencollective.com/gobuster/backers/badge.svg)](#backers) [![Sponsors on Open Collective](https://opencollective.com/gobuster/sponsors/badge.svg)](#sponsors)

## Oh dear God.. WHY!?

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

## But it's shit! And your implementation sucks!

Yes, you're probably correct. Feel free to:

* Not use it.
* Show me how to do it better.

## Love this tool? Back it!

If you're backing us already, you rock. If you're not, that's cool too! Want to back us? [Become a backer](https://opencollective.com/gobuster#backer)!

[![Backers](https://opencollective.com/gobuster/backers.svg?width=890)](https://opencollective.com/gobuster#backers)

All funds that are donated to this project will be donated to charity. A full log of charity donations will be available in this repository as they are processed.

## Changes in 3.0

* New CLI options so modes are strictly seperated (`-m` is now gone!)
* Performance Optimizations and better connection handling
* Ability to bruteforce vhost names
* Option to supply custom HTTP headers

## Available Modes

* dir - the classic directory brute-forcing mode
* dns - DNS subdomain brute-forcing mode
* vhost - virtual host brute-forcing mode (not the same as DNS!)

## Built-in Help

Help is built-in!

* `gobuster help` - outputs the top-level help.
* `gobuster help <mode>` - outputs the help specific to that mode.

## `dns` Mode Help

```text
Usage:
  gobuster dns [flags]

Flags:
  -d, --domain string      The target domain
  -h, --help               help for dns
  -r, --resolver string    Use custom DNS server (format server.com or server.com:port)
  -c, --showcname          Show CNAME records (cannot be used with '-i' option)
  -i, --showips            Show IP addresses
      --timeout duration   DNS resolver timeout (default 1s)
      --wildcard           Force continued operation when wildcard found

Global Flags:
  -z, --noprogress        Don't display progress
  -o, --output string     Output file to write results to (defaults to stdout)
  -q, --quiet             Don't print the banner and other noise
  -t, --threads int       Number of concurrent threads (default 10)
      --delay duration    Time each thread waits between requests (e.g. 1500ms)
  -v, --verbose           Verbose output (errors)
  -w, --wordlist string   Path to the wordlist
```

## `dir` Mode Options

```text
Usage:
  gobuster dir [flags]

Flags:
  -f, --addslash                      Append / to each request
  -c, --cookies string                Cookies to use for the requests
  -e, --expanded                      Expanded mode, print full URLs
  -x, --extensions string             File extension(s) to search for
  -r, --followredirect                Follow redirects
  -H, --headers stringArray           Specify HTTP headers, -H 'Header1: val1' -H 'Header2: val2'
  -h, --help                          help for dir
  -l, --includelength                 Include the length of the body in the output
  -k, --insecuressl                   Skip SSL certificate verification
  -n, --nostatus                      Don't print status codes
  -P, --password string               Password for Basic Auth
  -p, --proxy string                  Proxy to use for requests [http(s)://host:port]
  -s, --statuscodes string            Positive status codes (will be overwritten with statuscodesblacklist if set) (default "200,204,301,302,307,401,403")
  -b, --statuscodesblacklist string   Negative status codes (will override statuscodes if set)
      --timeout duration              HTTP Timeout (default 10s)
  -u, --url string                    The target URL
  -a, --useragent string              Set the User-Agent string (default "gobuster/3.0.1")
  -U, --username string               Username for Basic Auth
      --wildcard                      Force continued operation when wildcard found

Global Flags:
  -z, --noprogress        Don't display progress
  -o, --output string     Output file to write results to (defaults to stdout)
  -q, --quiet             Don't print the banner and other noise
  -t, --threads int       Number of concurrent threads (default 10)
      --delay duration    Time each thread waits between requests (e.g. 1500ms)
  -v, --verbose           Verbose output (errors)
  -w, --wordlist string   Path to the wordlist
```

## `vhost` Mode Options

```text
Usage:
  gobuster vhost [flags]

Flags:
  -c, --cookies string        Cookies to use for the requests
  -r, --followredirect        Follow redirects
  -H, --headers stringArray   Specify HTTP headers, -H 'Header1: val1' -H 'Header2: val2'
  -h, --help                  help for vhost
  -k, --insecuressl           Skip SSL certificate verification
  -P, --password string       Password for Basic Auth
  -p, --proxy string          Proxy to use for requests [http(s)://host:port]
      --timeout duration      HTTP Timeout (default 10s)
  -u, --url string            The target URL
  -a, --useragent string      Set the User-Agent string (default "gobuster/3.0.1")
  -U, --username string       Username for Basic Auth

Global Flags:
  -z, --noprogress        Don't display progress
  -o, --output string     Output file to write results to (defaults to stdout)
  -q, --quiet             Don't print the banner and other noise
  -t, --threads int       Number of concurrent threads (default 10)
      --delay duration    Time each thread waits between requests (e.g. 1500ms)
  -v, --verbose           Verbose output (errors)
  -w, --wordlist string   Path to the wordlist
```

## Easy Installation

### Binary Releases

We are now shipping binaries for each of the releases so that you don't even have to build them yourself! How wonderful is that!

If you're stupid enough to trust binaries that I've put together, you can download them from the [releases](https://github.com/OJ/gobuster/releases) page.

### Using `go get`

If you have a [Go](https://golang.org/) environment ready to go, it's as easy as:

```bash
go get github.com/OJ/gobuster
```

## Building From Source

Since this tool is written in [Go](https://golang.org/) you need to install the Go language/compiler/etc. Full details of installation and set up can be found [on the Go language website](https://golang.org/doc/install). Once installed you have two options.

### Compiling

`gobuster` now has external dependencies, and so they need to be pulled in first:

```bash
go get && go build
```

This will create a `gobuster` binary for you. If you want to install it in the `$GOPATH/bin` folder you can run:

```bash
go install
```

If you have all the dependencies already, you can make use of the build scripts:

* `make` - builds for the current Go configuration (ie. runs `go build`).
* `make windows` - builds 32 and 64 bit binaries for windows, and writes them to the `build` subfolder.
* `make linux` - builds 32 and 64 bit binaries for linux, and writes them to the `build` subfolder.
* `make darwin` - builds 32 and 64 bit binaries for darwin, and writes them to the `build` subfolder.
* `make all` - builds for all platforms and architectures, and writes the resulting binaries to the `build` subfolder.
* `make clean` - clears out the `build` subfolder.
* `make test` - runs the tests.

## Wordlists via STDIN

Wordlists can be piped into `gobuster` via stdin by providing a `-` to the `-w` option:

```bash
hashcat -a 3 --stdout ?l | gobuster dir -u https://mysite.com -w -
```

Note: If the `-w` option is specified at the same time as piping from STDIN, an error will be shown and the program will terminate.

## Examples

### `dir` Mode

Command line might look like this:

```bash
gobuster dir -u https://mysite.com/path/to/folder -c 'session=123456' -t 50 -w common-files.txt -x .php,.html
```

Default options looks like this:

```bash
gobuster dir -u https://buffered.io -w ~/wordlists/shortlist.txt

===============================================================
Gobuster v3.0.1
by OJ Reeves (@TheColonial) & Christian Mehlmauer (@_FireFart_)
===============================================================
[+] Mode         : dir
[+] Url/Domain   : https://buffered.io/
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/shortlist.txt
[+] Status codes : 200,204,301,302,307,401,403
[+] User Agent   : gobuster/3.0.1
[+] Timeout      : 10s
===============================================================
2019/06/21 11:49:43 Starting gobuster
===============================================================
/categories (Status: 301)
/contact (Status: 301)
/posts (Status: 301)
/index (Status: 200)
===============================================================
2019/06/21 11:49:44 Finished
===============================================================
```

Default options with status codes disabled looks like this:

```bash
gobuster dir -u https://buffered.io -w ~/wordlists/shortlist.txt -n

===============================================================
Gobuster v3.0.1
by OJ Reeves (@TheColonial) & Christian Mehlmauer (@_FireFart_)
===============================================================
[+] Mode         : dir
[+] Url/Domain   : https://buffered.io/
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/shortlist.txt
[+] Status codes : 200,204,301,302,307,401,403
[+] User Agent   : gobuster/3.0.1
[+] No status    : true
[+] Timeout      : 10s
===============================================================
2019/06/21 11:50:18 Starting gobuster
===============================================================
/categories
/contact
/index
/posts
===============================================================
2019/06/21 11:50:18 Finished
===============================================================
```

Verbose output looks like this:

```bash
gobuster dir -u https://buffered.io -w ~/wordlists/shortlist.txt -v

===============================================================
Gobuster v3.0.1
by OJ Reeves (@TheColonial) & Christian Mehlmauer (@_FireFart_)
===============================================================
[+] Mode         : dir
[+] Url/Domain   : https://buffered.io/
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/shortlist.txt
[+] Status codes : 200,204,301,302,307,401,403
[+] User Agent   : gobuster/3.0.1
[+] Verbose      : true
[+] Timeout      : 10s
===============================================================
2019/06/21 11:50:51 Starting gobuster
===============================================================
Missed: /alsodoesnotexist (Status: 404)
Found: /index (Status: 200)
Missed: /doesnotexist (Status: 404)
Found: /categories (Status: 301)
Found: /posts (Status: 301)
Found: /contact (Status: 301)
===============================================================
2019/06/21 11:50:51 Finished
===============================================================
```

Example showing content length:

```bash
gobuster dir -u https://buffered.io -w ~/wordlists/shortlist.txt -l

===============================================================
Gobuster v3.0.1
by OJ Reeves (@TheColonial) & Christian Mehlmauer (@_FireFart_)
===============================================================
[+] Mode         : dir
[+] Url/Domain   : https://buffered.io/
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/shortlist.txt
[+] Status codes : 200,204,301,302,307,401,403
[+] User Agent   : gobuster/3.0.1
[+] Show length  : true
[+] Timeout      : 10s
===============================================================
2019/06/21 11:51:16 Starting gobuster
===============================================================
/categories (Status: 301) [Size: 178]
/posts (Status: 301) [Size: 178]
/contact (Status: 301) [Size: 178]
/index (Status: 200) [Size: 51759]
===============================================================
2019/06/21 11:51:17 Finished
===============================================================
```

Quiet output, with status disabled and expanded mode looks like this ("grep mode"):

```bash
gobuster dir -u https://buffered.io -w ~/wordlists/shortlist.txt -q -n -e
https://buffered.io/index
https://buffered.io/contact
https://buffered.io/posts
https://buffered.io/categories
```

### `dns` Mode

Command line might look like this:

```bash
gobuster dns -d mysite.com -t 50 -w common-names.txt
```

Normal sample run goes like this:

```bash
gobuster dns -d google.com -w ~/wordlists/subdomains.txt

===============================================================
Gobuster v3.0.1
by OJ Reeves (@TheColonial) & Christian Mehlmauer (@_FireFart_)
===============================================================
[+] Mode         : dns
[+] Url/Domain   : google.com
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/subdomains.txt
===============================================================
2019/06/21 11:54:20 Starting gobuster
===============================================================
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
===============================================================
2019/06/21 11:54:20 Finished
===============================================================
```

Show IP sample run goes like this:

```bash
gobuster dns -d google.com -w ~/wordlists/subdomains.txt -i

===============================================================
Gobuster v3.0.1
by OJ Reeves (@TheColonial) & Christian Mehlmauer (@_FireFart_)
===============================================================
[+] Mode         : dns
[+] Url/Domain   : google.com
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/subdomains.txt
===============================================================
2019/06/21 11:54:54 Starting gobuster
===============================================================
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
===============================================================
2019/06/21 11:54:55 Finished
===============================================================
```

Base domain validation warning when the base domain fails to resolve. This is a warning rather than a failure in case the user fat-fingers while typing the domain.

```bash
gobuster dns -d yp.to -w ~/wordlists/subdomains.txt -i

===============================================================
Gobuster v3.0.1
by OJ Reeves (@TheColonial) & Christian Mehlmauer (@_FireFart_)
===============================================================
[+] Mode         : dns
[+] Url/Domain   : yp.to
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/subdomains.txt
===============================================================
2019/06/21 11:56:43 Starting gobuster
===============================================================
2019/06/21 11:56:53 [-] Unable to validate base domain: yp.to
Found: cr.yp.to [131.193.32.108, 131.193.32.109]
===============================================================
2019/06/21 11:56:53 Finished
===============================================================
```

Wildcard DNS is also detected properly:

```bash
gobuster dns -d 0.0.1.xip.io -w ~/wordlists/subdomains.txt

===============================================================
Gobuster v3.0.1
by OJ Reeves (@TheColonial) & Christian Mehlmauer (@_FireFart_)
===============================================================
[+] Mode         : dns
[+] Url/Domain   : 0.0.1.xip.io
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/subdomains.txt
===============================================================
2019/06/21 12:13:48 Starting gobuster
===============================================================
2019/06/21 12:13:48 [-] Wildcard DNS found. IP address(es): 1.0.0.0
2019/06/21 12:13:48 [!] To force processing of Wildcard DNS, specify the '--wildcard' switch.
===============================================================
2019/06/21 12:13:48 Finished
===============================================================
```

If the user wants to force processing of a domain that has wildcard entries, use `--wildcard`:

```bash
gobuster dns -d 0.0.1.xip.io -w ~/wordlists/subdomains.txt --wildcard

===============================================================
Gobuster v3.0.1
by OJ Reeves (@TheColonial) & Christian Mehlmauer (@_FireFart_)
===============================================================
[+] Mode         : dns
[+] Url/Domain   : 0.0.1.xip.io
[+] Threads      : 10
[+] Wordlist     : /home/oj/wordlists/subdomains.txt
===============================================================
2019/06/21 12:13:51 Starting gobuster
===============================================================
2019/06/21 12:13:51 [-] Wildcard DNS found. IP address(es): 1.0.0.0
Found: 127.0.0.1.xip.io
Found: test.127.0.0.1.xip.io
===============================================================
2019/06/21 12:13:53 Finished
===============================================================
```

### `vhost` Mode

Command line might look like this:

```bash
gobuster vhost -u https://mysite.com -w common-vhosts.txt
```

Normal sample run goes like this:

```bash
gobuster vhost -u https://mysite.com -w common-vhosts.txt

===============================================================
Gobuster v3.0.1
by OJ Reeves (@TheColonial) & Christian Mehlmauer (@_FireFart_)
===============================================================
[+] Url:          https://mysite.com
[+] Threads:      10
[+] Wordlist:     common-vhosts.txt
[+] User Agent:   gobuster/3.0.1
[+] Timeout:      10s
===============================================================
2019/06/21 08:36:00 Starting gobuster
===============================================================
Found: www.mysite.com
Found: piwik.mysite.com
Found: mail.mysite.com
===============================================================
2019/06/21 08:36:05 Finished
===============================================================
```

## License

See the LICENSE file.

## Thanks

See the THANKS file for people who helped out.
