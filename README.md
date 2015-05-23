Gobuster v0.6 (OJ Reeves @TheColonial)
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

* `-m <mode>`       - which mode to use, either `dir` or `dns` (default: `dir`)
* `-u <url/domain>` - full URL (including scheme), or base domain name.
* `-t <threads>`    - number of threads to run (default: `10`).
* `-w <wordlist>`   - path to the wordlist used for brute forcing.
* `-v`              - verbose output (show error codes, and IP addresses).

### Command line options for `dir` mode

* `-c <http cookies>` - use this to specify any cookies that you might need (simulating auth).
* `-f <true|false>`   - set to `true` if you want to append `/` for directory brute forces.
* `-s <status codes>` - comma-separated set of the list of status codes to be deemed a "positive" (default: `200,204,301,302,307`).
* `-x <extensions>`   - list of extensions to check for, if any.

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
Sample run goes like this:
```
$ ./gobuster -w words.txt -u http://buffered.io/ -x .html -v

=====================================================
Gobuster v0.6 (DIR support by OJ Reeves @TheColonial)
              (DNS support by Peleus     @0x42424242)
=====================================================
[+] Mode         : dir
[+] Url/Domain   : http://buffered.io/
[+] Threads      : 10
[+] Wordlist     : words.txt
[+] Status codes : 200,204,301,302,307
[+] Extensions   : .html
[+] Verbose      : true
=====================================================
Result: /download (404)
Result: /2006 (404)
Result: /news (404)
Found: /index (200)
Result: /crack (404)
Result: /warez (404)
Result: /serial (404)
Result: /full (404)
Result: /download.html (404)
Result: /images (404)
Result: /news.html (404)
Result: /2006.html (404)
Result: /crack.html (404)
Result: /warez.html (404)
Found: /index.html (200)
```

#### `dns` mode

Command line might look like this:
```
$ ./gobuster -m dns -u mysite.com -t 50 -w common-names.txt
```
Normal sample run goes like this:
```
$ ./gobuster -m dns -w subdomains.txt -u google.com

=====================================================
Gobuster v0.6 (DIR support by OJ Reeves @TheColonial)
              (DNS support by Peleus     @0x42424242)
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
Verbose sample run goes like this:
```
$ ./gobuster -m dns -w subdomains.txt -u google.com -v

=====================================================
Gobuster v0.6 (DIR support by OJ Reeves @TheColonial)
              (DNS support by Peleus     @0x42424242)
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

### License

See the LICENSE file.
