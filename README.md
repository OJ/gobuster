# Gobuster

[![Go Report Card](https://goreportcard.com/badge/github.com/OJ/gobuster/v3)](https://goreportcard.com/report/github.com/OJ/gobuster/v3) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/OJ/gobuster/blob/master/LICENSE) [![Backers on Open Collective](https://opencollective.com/gobuster/backers/badge.svg)](https://opencollective.com/gobuster) [![Sponsors on Open Collective](https://opencollective.com/gobuster/sponsors/badge.svg)](https://opencollective.com/gobuster)

## 💻 Introduction

> A fast and flexible brute-forcing tool written in Go

**Gobuster** is a high-performance directory/file, DNS and virtual host brute-forcing tool written in Go. It's designed to be fast, reliable, and easy to use for security professionals and penetration testers.

## ✨ Features

- 🚀 **High Performance**: Multi-threaded scanning with configurable concurrency
- 🔍 **Multiple Modes**: Directory, DNS, virtual host, S3, GCS, TFTP, and fuzzing modes
- 🛡️ **Security Focused**: Built for penetration testing and security assessments
- 🐳 **Docker Support**: Available as a Docker container
- 🔧 **Extensible**: Pattern-based scanning and custom wordlists

## 🎯 What Can Gobuster Do?

- **Web Directory/File Enumeration**: Discover hidden directories and files on web servers
- **DNS Subdomain Discovery**: Find subdomains with wildcard support
- **Virtual Host Detection**: Identify virtual hosts on target web servers
- **Cloud Storage Enumeration**: Discover open Amazon S3 and Google Cloud Storage buckets
- **TFTP File Discovery**: Find files on TFTP servers
- **Custom Fuzzing**: Flexible fuzzing with customizable parameters

## 🚀 Quick Start

```bash
# Install gobuster
go install github.com/OJ/gobuster/v3@latest

# Basic directory enumeration
gobuster dir -u https://example.com -w /path/to/wordlist.txt

# DNS subdomain enumeration
gobuster dns -do example.com -w /path/to/wordlist.txt

# Virtual host discovery
gobuster vhost -u https://example.com -w /path/to/wordlist.txt

# S3 bucket enumeration
gobuster s3 -w /path/to/bucket-names.txt
```

## 📦 Installation

### Quick Install (Recommended)

```bash
go install github.com/OJ/gobuster/v3@latest
```

**Requirements**: Go 1.24 or higher

### Alternative Installation Methods

#### Using Binary Releases

Download pre-compiled binaries from the [releases page](https://github.com/OJ/gobuster/releases).

#### Using Docker

```bash
# Pull the latest image
docker pull ghcr.io/oj/gobuster:latest

# Run gobuster in Docker
docker run --rm -it ghcr.io/oj/gobuster:latest dir -u https://example.com -w /usr/share/wordlists/dirb/common.txt
```

#### Building from Source

```bash
git clone https://github.com/OJ/gobuster.git
cd gobuster
go mod tidy
go build
```

### Troubleshooting Installation

If you encounter issues:

- Ensure Go version 1.24+ is installed: `go version`
- Check your `$GOPATH` and `$GOBIN` environment variables
- Verify `$GOPATH/bin` is in your `$PATH`

## 🎯 Usage

Gobuster uses a mode-based approach. Each mode is designed for specific enumeration tasks:

```bash
gobuster [mode] [options]
```

### Getting Help

```bash
gobuster help                   # Show general help
gobuster help [mode]            # Show help for specific mode
gobuster [mode] --help          # Alternative help syntax
```

### 📊 Available Modes

#### 🌐 Directory Mode (`dir`)

Enumerate directories and files on web servers.

**Basic Usage:**

```bash
gobuster dir -u https://example.com -w wordlist.txt
```

**Advanced Options:**

```bash
# With file extensions
gobuster dir -u https://example.com -w wordlist.txt -x php,html,js,txt

# With custom headers and cookies
gobuster dir -u https://example.com -w wordlist.txt -H "Authorization: Bearer token" -c "session=value"

# Show response length
gobuster dir -u https://example.com -w wordlist.txt -l

# Filter by status codes
gobuster dir -u https://example.com -w wordlist.txt -s 200,301,302
```

#### 🔍 DNS Mode (`dns`)

Discover subdomains through DNS resolution.

**Basic Usage:**

```bash
gobuster dns -do example.com -w wordlist.txt
```

**Advanced Options:**

```bash
# Use custom DNS server
gobuster dns -do example.com -w wordlist.txt -r 8.8.8.8:53

# Increase threads for faster scanning
gobuster dns -do example.com -w wordlist.txt -t 50
```

#### 🏠 Virtual Host Mode (`vhost`)

Discover virtual hosts on web servers.

**Basic Usage:**

```bash
gobuster vhost -u https://example.com --append-domain -w wordlist.txt
```

#### ☁️ S3 Mode (`s3`)

Enumerate Amazon S3 buckets.

**Basic Usage:**

```bash
gobuster s3 -w bucket-names.txt
```

**With Debug Output:**

```bash
gobuster s3 -w bucket-names.txt --debug
```

#### 🖥️ TFTP Mode (`tftp`)

Enumerate files on tftp servers.

**Basic Usage:**

```bash
gobuster tftp -s 10.0.0.1 -w wordlist.txt
```

#### ☁️ GCS Mode (`gcs`)

Enumerate Google Cloud Storage Buckets.

**Basic Usage:**

```bash
gobuster gcs -w bucket-names.txt
```

**With Debug Output:**

```bash
gobuster gcs -w bucket-names.txt --debug
```

#### 🔧 Fuzz Mode (`fuzz`)

Custom fuzzing with the `FUZZ` keyword.

**Basic Usage:**

```bash
gobuster fuzz -u https://example.com?FUZZ=test -w wordlist.txt
```

**Advanced Examples:**

```bash
# Fuzz URL parameters
gobuster fuzz -u https://example.com?param=FUZZ -w wordlist.txt

# Fuzz headers
gobuster fuzz -u https://example.com -H "X-Custom-Header: FUZZ" -w wordlist.txt

# Fuzz POST data
gobuster fuzz -u https://example.com -d "username=admin&password=FUZZ" -w passwords.txt
```

## 💰 Support

[![Backers on Open Collective](https://opencollective.com/gobuster/backers/badge.svg)](https://opencollective.com/gobuster) [![Sponsors on Open Collective](https://opencollective.com/gobuster/sponsors/badge.svg)](https://opencollective.com/gobuster)

### Love this tool? Back it!

If you're backing us already, you rock. If you're not, that's cool too! Want to back us? [Become a backer](https://opencollective.com/gobuster#backer)!

[![Backers](https://opencollective.com/gobuster/backers.svg?width=890)](https://opencollective.com/gobuster#backers)

All funds that are donated to this project will be donated to charity. A full log of charity donations will be available in this repository as they are processed.

## 💡 Common Use Cases

### Web Application Security Testing

```bash
# Comprehensive directory enumeration
gobuster dir -u https://target.com -w /usr/share/wordlists/dirbuster/directory-list-2.3-medium.txt -x php,html,js,txt,asp,aspx,jsp

# API endpoint discovery
gobuster dir -u https://api.target.com -w /usr/share/wordlists/dirb/common.txt -x json

# Admin panel discovery
gobuster dir -u https://target.com -w admin-panels.txt -s 200,301,302,403
```

### DNS Reconnaissance

```bash
# Comprehensive subdomain enumeration
gobuster dns -do target.com -w /usr/share/wordlists/dnsrecon/subdomains-top1mil-5000.txt -t 50
```

### Cloud Storage Assessment

```bash
# S3 bucket enumeration with patterns
gobuster s3 -w company-names.txt -v

# GCS bucket enumeration
gobuster gcs -w company-names.txt -v
```

## 🔧 Troubleshooting

### Common Issues

#### "Permission Denied" or "Access Denied"

- Try reducing thread count with `-t` flag
- Add delays between requests with `--delay`
- Use different user agent with `-a` flag

#### "Connection Timeout"

- Increase timeout with `--timeout` flag
- Reduce thread count for slower targets
- Check your internet connection

#### "No Results Found"

- Verify the target URL is accessible
- Try different wordlists
- Check status code filtering with `-s` flag

### Performance Issues

#### Slow Scanning

- Increase thread count with `-t` flag (but be careful not to overwhelm the target)
- Use smaller, more targeted wordlists

## 🎯 Best Practices

### Security Testing Guidelines

1. **Always get proper authorization** before testing any target
2. **Start with low thread counts** to avoid overwhelming servers
3. **Use appropriate wordlists** for the target technology
4. **Respect rate limits** and implement delays if needed
5. **Monitor your network traffic** to avoid detection

### Wordlist Selection

- **For web applications**: Use technology-specific wordlists (PHP, ASP.NET, etc.)
- **For APIs**: Focus on common API endpoints and versioning patterns
- **For DNS**: Use subdomain-specific wordlists with common patterns
- **For cloud storage**: Use company/brand-specific patterns

### Output Management

```bash
# Save results to file
gobuster dir -u https://example.com -w wordlist.txt -o results.txt

# Use quiet mode for clean output
gobuster dir -u https://example.com -w wordlist.txt -q
```

## 📚 Additional Resources

### Recommended Wordlists

- **SecLists**: [https://github.com/danielmiessler/SecLists](https://github.com/danielmiessler/SecLists)
- **FuzzDB**: [https://github.com/fuzzdb-project/fuzzdb](https://github.com/fuzzdb-project/fuzzdb)
- **Seclists DNS**: [https://github.com/danielmiessler/SecLists/tree/master/Discovery/DNS](https://github.com/danielmiessler/SecLists/tree/master/Discovery/DNS)

---

**Happy hacking! 🚀**

_Remember: Always test responsibly and with proper authorization._

# Changes

<details>

<summary>3.8.2</summary>

## 3.8.2

- Fix expanded mode to show the full url again

</details>

<details>

<summary>3.8.1</summary>

## 3.8.1

- Fix expanded mode showing the entries twice

</details>

<details>

<summary>3.8</summary>

## 3.8

- Add exclude-hostname-length flag to dynamically adjust exclude-length by @0xyy66
- Fix Fuzzing query parameters
- Add `--force` flag in `dir` mode to continue execution if precheck errors occur

</details>

<details>

<summary>3.7</summary>

## 3.7

- use new cli library
- a lot more short options due to the new cli library
- more user friendly error messages
- clean up DNS mode
- renamed `show-cname` to `check-cname` in dns mode
- got rid of `verbose` flag and introduced `debug` instead
- the version command now also shows some build variables for more info
- switched to another pkcs12 library to support p12s generated with openssl3 that use SHA256 HMAC
- comments in wordlists (strings starting with #) are no longer ignored
- warn in vhost mode if the --append-domain switch might have been forgotten
- allow to exclude status code and length in vhost mode
- added automaxprocs for use in docker with cpu limits
- log http requests with debug enabled
- allow fuzzing of Host header in fuzz mode
- automatically disable progress output when output is redirected
- fix extra special characters when run with `--no-progress`
- warn when using vhost mode with a proxy and http based urls as this might not work as expected
- add `interface` and `local-ip` parameters to specify the outgoing interface for http requests
- add support for tls renegotiation
- fix progress with patterns by @acammack
- fix backup discovery by @acammack
- support tcp protocol on dns servers
- add support for URL query parameters

</details>

<details>
<summary>3.6</summary>

## 3.6

- Wordlist offset parameter to skip x lines from the wordlist
- prevent double slashes when building up an url in dir mode
- allow for multiple values and ranges on `--exclude-length`
- `no-fqdn` parameter on dns bruteforce to disable the use of the systems search domains. This should speed up the run if you have configured some search domains. [https://github.com/OJ/gobuster/pull/418](https://github.com/OJ/gobuster/pull/418)

</details>

<details>
<summary>3.5</summary>

## 3.5

- Allow Ranges in status code and status code blacklist. Example: 200,300-305,404

</details>

<details>
<summary>3.4</summary>

## 3.4

- Enable TLS1.0 and TLS1.1 support
- Add TFTP mode to search for files on tftp servers

</details>

<details>
<summary>3.3</summary>

## 3.3

- Support TLS client certificates / mtls
- support loading extensions from file
- support fuzzing POST body, HTTP headers and basic auth
- new option to not canonicalize header names

</details>

<details>
<summary>3.2</summary>

## 3.2

- Use go 1.19
- use contexts in the correct way
- get rid of the wildcard flag (except in DNS mode)
- color output
- retry on timeout
- google cloud bucket enumeration
- fix nil reference errors

</details>

<details>
<summary>3.1</summary>

## 3.1

- enumerate public AWS S3 buckets
- fuzzing mode
- specify HTTP method
- added support for patterns. You can now specify a file containing patterns that are applied to every word, one by line. Every occurrence of the term `{GOBUSTER}` in it will be replaced with the current wordlist item. Please use with caution as this can cause increase the number of requests issued a lot.
- The shorthand `p` flag which was assigned to proxy is now used by the pattern flag

</details>

<details>
<summary>3.0</summary>

## 3.0

- New CLI options so modes are strictly separated (`-m` is now gone!)
- Performance Optimizations and better connection handling
- Ability to enumerate vhost names
- Option to supply custom HTTP headers

</details>
