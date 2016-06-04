Gobuster v1.1 (OJ Reeves @TheColonial)
======================================

Alternative directory and file busting tool written in Go. DNS support recently added after inspiration and effort from [Peleus](https://twitter.com/0x42424242).

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

Yes, you're probably correct. Feel free to :

* Not use it.
* Show me how to do it better.

### Common Command line options

* `-m <mode>` - which mode to use, either `dir` or `dns` (default: `dir`)
* `-t <threads>` - number of threads to run (default: `10`).
* `-u <url/domain>` - full URL (including scheme), or base domain name.
* `-v` - verbose output (show all results).
* `-w <wordlist>` - path to the wordlist used for brute forcing.

### Command line options for `dns` mode

* `-i` - show all IP addresses for the result.

### Command line options for `dir` mode

* `-a <user agent string>` - specify a user agent string to send in the request header
* `-c <http cookies>` - use this to specify any cookies that you might need (simulating auth).
* `-f` - append `/` for directory brute forces.
* `-l` - show the length of the response.
* `-n` - "no status" mode, disables the output of the result's status code.
* `-p <proxy url>` - specify a proxy to use for all requests (scheme much match the URL scheme)
* `-q` - disables banner/underline output.
* `-r` - follow redirects.
* `-s <status codes>` - comma-separated set of the list of status codes to be deemed a "positive" (default: `200,204,301,302,307`).
* `-x <extensions>` - list of extensions to check for, if any.
* `-P <password>` - HTTP Authorization password (Basic Auth only, prompted if missing).
* `-U <username>` - HTTP Authorization username (Basic Auth only).

### Building

Since this tool is written in [Go](https://golang.org/) you need install the Go language/compiler/etc. Full details of installation and set up can be found [on the Go language website](https://golang.org/doc/install). Once installed you have two options.

#### Compiling
```
gobuster$ go build
```
This will create a `gobuster` binary for you.

#### Running as a script
```
gobuster$ go run main.go <parameters>
```

### Examples

#### `dir` mode

Command line might look like this:
```
$ ./gobuster -u https://mysite.com/path/to/folder -c 'session=123456' -t 50 -w common-files.txt -x .php,.html
```
Default options looks like this:
```
$ ./gobuster -u http://buffered.io/ -w words.txt

Gobuster v1.1                OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dir
[+] Url/Domain   : http://buffered.io/
[+] Threads      : 10
[+] Wordlist     : words.txt
[+] Status codes : 200,204,301,302,307
=====================================================
/index (Status: 200)
/posts (Status: 301)
/contact (Status: 301)
=====================================================
```
Default options with status codes disabled looks like this:
```
$ ./gobuster -u http://buffered.io/ -w words.txt -n

Gobuster v1.1                OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dir
[+] Url/Domain   : http://buffered.io/
[+] Threads      : 10
[+] Wordlist     : words.txt
[+] Status codes : 200,204,301,302,307
[+] No status    : true
=====================================================
/index
/posts
/contact
=====================================================
```
Verbose output looks like this:
```
$ ./gobuster -u http://buffered.io/ -w words.txt -v

Gobuster v1.1                OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dir
[+] Url/Domain   : http://buffered.io/
[+] Threads      : 10
[+] Wordlist     : words.txt
[+] Status codes : 200,204,301,302,307
[+] Verbose      : true
=====================================================
Found : /index (Status: 200)
Missed: /derp (Status: 404)
Found : /posts (Status: 301)
Found : /contact (Status: 301)
=====================================================
```
Example showing content length:
```
$ ./gobuster -u http://buffered.io/ -w words.txt -l

Gobuster v1.1                OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dir
[+] Url/Domain   : http://buffered.io/
[+] Threads      : 10
[+] Wordlist     : /tmp/words
[+] Status codes : 301,302,307,200,204
[+] Show length  : true
=====================================================
/contact (Status: 301)
/posts (Status: 301)
/index (Status: 200) [Size: 61481]
=====================================================
```
Quiet output, with status disabled and expanded mode looks like this ("grep mode"):
```
$ ./gobuster -u http://buffered.io/ -w words.txt -q -n -e
http://buffered.io/posts
http://buffered.io/contact
http://buffered.io/index
```

#### `dns` mode

Command line might look like this:
```
$ ./gobuster -m dns -u mysite.com -t 50 -w common-names.txt
```
Normal sample run goes like this:
```
$ ./gobuster -m dns -w subdomains.txt -u google.com

Gobuster v1.1                OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dns
[+] Url/Domain   : google.com
[+] Threads      : 10
[+] Wordlist     : subdomains.txt
=====================================================
Found: m.google.com
Found: admin.google.com
Found: mobile.google.com
Found: www.google.com
Found: search.google.com
Found: chrome.google.com
Found: ns1.google.com
Found: store.google.com
Found: wap.google.com
Found: support.google.com
Found: directory.google.com
Found: translate.google.com
Found: news.google.com
Found: music.google.com
Found: mail.google.com
Found: blog.google.com
Found: cse.google.com
Found: local.google.com
=====================================================
```
Show IP sample run goes like this:
```
$ ./gobuster -m dns -w subdomains.txt -u google.com -i

Gobuster v1.1                OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dns
[+] Url/Domain   : google.com
[+] Threads      : 10
[+] Wordlist     : subdomains.txt
[+] Verbose      : true
=====================================================
Found: chrome.google.com [2404:6800:4006:801::200e, 216.58.220.110]
Found: m.google.com [216.58.220.107, 2404:6800:4006:801::200b]
Found: www.google.com [74.125.237.179, 74.125.237.177, 74.125.237.178, 74.125.237.180, 74.125.237.176, 2404:6800:4006:801::2004]
Found: search.google.com [2404:6800:4006:801::200e, 216.58.220.110]
Found: admin.google.com [216.58.220.110, 2404:6800:4006:801::200e]
Found: store.google.com [216.58.220.110, 2404:6800:4006:801::200e]
Found: mobile.google.com [216.58.220.107, 2404:6800:4006:801::200b]
Found: ns1.google.com [216.239.32.10]
Found: directory.google.com [216.58.220.110, 2404:6800:4006:801::200e]
Found: translate.google.com [216.58.220.110, 2404:6800:4006:801::200e]
Found: cse.google.com [216.58.220.110, 2404:6800:4006:801::200e]
Found: local.google.com [2404:6800:4006:801::200e, 216.58.220.110]
Found: music.google.com [2404:6800:4006:801::200e, 216.58.220.110]
Found: wap.google.com [216.58.220.110, 2404:6800:4006:801::200e]
Found: blog.google.com [216.58.220.105, 2404:6800:4006:801::2009]
Found: support.google.com [216.58.220.110, 2404:6800:4006:801::200e]
Found: news.google.com [216.58.220.110, 2404:6800:4006:801::200e]
Found: mail.google.com [216.58.220.101, 2404:6800:4006:801::2005]
=====================================================
```
Base domain validation warning when the base domain fails to resolve. This is a warning rather than a failure in case the user fat-fingers while typing the domain.
```
$ ./gobuster -m dns -w subdomains.txt -u yp.to -i

Gobuster v1.1                OJ Reeves (@TheColonial)
=====================================================
[+] Mode         : dns
[+] Url/Domain   : yp.to
[+] Threads      : 10
[+] Wordlist     : /tmp/test.txt
=====================================================
[!] Unable to validate base domain: yp.to
Found: cr.yp.to [131.155.70.11, 131.155.70.13]
=====================================================
```

### License

See the LICENSE file.

### Thanks

See the THANKS file for people who helped out.
