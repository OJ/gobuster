# Gobuster

[![Go Report Card](https://goreportcard.com/badge/github.com/OJ/gobuster/v3)](https://goreportcard.com/report/github.com/OJ/gobuster/v3)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Open Collective](https://opencollective.com/gobuster/backers/badge.svg)](https://opencollective.com/gobuster)

Gobuster is a fast, flexible enumeration tool written in Go. It uses wordlists to
discover directories and files, DNS names, virtual hosts, cloud storage buckets,
and TFTP files. It can also fuzz values in URLs, headers, and request bodies.

> Use Gobuster only against systems you own or have explicit permission to test.

## Modes

| Mode    | Purpose                                             |
| ------- | --------------------------------------------------- |
| `dir`   | Discover directories and files on web servers       |
| `dns`   | Discover DNS subdomains                             |
| `vhost` | Discover virtual hosts on a web server              |
| `fuzz`  | Replace `FUZZ` in URLs, headers, and request bodies |
| `s3`    | Enumerate Amazon S3 buckets                         |
| `gcs`   | Enumerate Google Cloud Storage buckets              |
| `tftp`  | Discover files on TFTP servers                      |

## Installation

### Go

Gobuster requires Go 1.27 or newer.

```console
go install github.com/OJ/gobuster/v3@latest
```

Make sure the Go binary directory is in your `PATH`. You can find it with
`go env GOBIN`; when that value is empty, Go uses `$(go env GOPATH)/bin`.

### Prebuilt binaries

Download an archive for your platform from the
[GitHub releases page](https://github.com/OJ/gobuster/releases).

### Docker

```console
docker pull ghcr.io/oj/gobuster:latest
docker run --rm -it \
  -v "$PWD/wordlists:/wordlists:ro" \
  ghcr.io/oj/gobuster:latest \
  dir -u https://example.com -w /wordlists/common.txt
```

## Quick start

Every mode has its own options. Start with the built-in help when exploring a
new mode:

```console
gobuster --help
gobuster dir --help
```

### Directory and file discovery

```console
gobuster dir -u https://example.com -w wordlist.txt
```

Add extensions, choose accepted status codes, and write results to a file:

```console
gobuster dir \
  -u https://example.com \
  -w wordlist.txt \
  -x php,html,js \
  -s 200,204,301,302,307,401,403 \
  -b "" \
  -o results.txt
```

The status-code blacklist defaults to `404` and overrides the positive list, so
explicitly clear it with `-b ""` when using `-s`.

Useful directory-mode options include:

- `-H 'Name: value'` to add a header; repeat it for multiple headers
- `-c 'name=value'` to send cookies
- `-U user -P password` for HTTP Basic authentication
- `-x php,html` to try file extensions
- `--exclude-length 123,456-500` to ignore response sizes
- `--regex PATTERN` or `--regex-invert PATTERN` to filter response bodies
- `--recursive` to scan discovered directories recursively
- `--recursion-depth N` to limit recursion depth (`0` for unlimited)
- `--recursion-max-targets N` to cap discovered recursive targets (`0` for unlimited)
- `--body-output-dir PATH` to save response bodies
- `-k` to skip TLS certificate verification

### DNS discovery

```console
gobuster dns --domain example.com -w subdomains.txt
```

Use a custom resolver or inspect CNAME records:

```console
gobuster dns --domain example.com -w subdomains.txt --resolver 1.1.1.1
gobuster dns --domain example.com -w subdomains.txt --check-cname
```

### Virtual-host discovery

```console
gobuster vhost -u https://example.com -w hosts.txt --append-domain
```

Point `-u` at the server you want to test. Use `--append-domain` when the
wordlist contains prefixes such as `admin` rather than complete hostnames.

### Fuzzing

Put the literal marker `FUZZ` wherever Gobuster should substitute each wordlist
entry:

```console
gobuster fuzz -u 'https://example.com/?page=FUZZ' -w values.txt
gobuster fuzz -u https://example.com -H 'X-Api-Version: FUZZ' -w versions.txt
gobuster fuzz -u https://example.com/login -m POST \
  -H 'Content-Type: application/x-www-form-urlencoded' \
  -B 'username=admin&password=FUZZ' \
  -w passwords.txt
```

### Cloud storage and TFTP

```console
gobuster s3 -w bucket-names.txt
gobuster gcs -w bucket-names.txt
gobuster tftp -s 192.0.2.10 -w filenames.txt
```

## Controlling a scan

The following options are shared by most modes:

| Option           | Description                                               |
| ---------------- | --------------------------------------------------------- |
| `-w, --wordlist` | Wordlist path; use `-` to read from standard input        |
| `-t, --threads`  | Number of concurrent workers (default: `10`)              |
| `-d, --delay`    | Delay applied by each worker, such as `250ms`             |
| `--timeout`      | Network timeout, such as `15s`                            |
| `-o, --output`   | Write discovered results to a file                        |
| `-q, --quiet`    | Print results without the banner and informational output |
| `--no-progress`  | Disable the progress display                              |
| `--debug`        | Enable diagnostic output                                  |

Start conservatively and increase concurrency only when the target can handle
it. A delay is often more useful than a very high thread count when testing
rate-limited services.

## Patterns

`--pattern` expands every wordlist entry through a pattern file. Each occurrence
of `{GOBUSTER}` is replaced with the current word:

```text
{GOBUSTER}-dev
{GOBUSTER}-staging
api-{GOBUSTER}
```

```console
gobuster dns --domain example.com -w words.txt --pattern patterns.txt
```

Patterns multiply the number of requests, so review the pattern file before a
large scan. `--discover-pattern` applies a pattern file only to successful
guesses.

## Building from source

Clone the repository and use [Task](https://taskfile.dev/) for the standard
development workflow:

```console
git clone https://github.com/OJ/gobuster.git
cd gobuster
task build
```

Common development commands:

```console
task test    # format, vet, and run tests with race detection and coverage
task check   # format, run gofumpt, vet, and apply Go fixes
task lint    # run golangci-lint and verify module files are tidy
task linux   # build a Linux AMD64 binary
task windows # build a Windows AMD64 binary
```

Bug reports should include the Gobuster version, the mode and command used
(with secrets removed), and relevant debug output. Open an issue before making
a large behavioral change.

## Support

Gobuster is maintained by Christian Mehlmauer
([@firefart](https://github.com/firefart)) and OJ Reeves
([@TheColonial](https://github.com/TheColonial)). You can support development
through [Open Collective](https://opencollective.com/gobuster). Project funds
are donated to charity.

## License

Gobuster is available under the [Apache License 2.0](LICENSE).
