package gobusterdir

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/OJ/gobuster/libgobuster"
	"github.com/google/uuid"
)

// GobusterDir is the main type to implement the interface
type GobusterDir struct {
	options    *OptionsDir
	globalopts *libgobuster.Options
	http       *httpClient
}

// GetRequest issues a GET request to the target and returns
// the status code, length and an error
func (d *GobusterDir) get(url string) (*int, *int64, error) {
	return d.http.makeRequest(url, d.options.Cookies)
}

// NewGobusterDir creates a new initialized GobusterDir
func NewGobusterDir(cont context.Context, globalopts *libgobuster.Options, opts *OptionsDir) (*GobusterDir, error) {
	if globalopts == nil {
		return nil, fmt.Errorf("please provide valid global options")
	}

	if opts == nil {
		return nil, fmt.Errorf("please provide valid plugin options")
	}

	g := GobusterDir{
		options:    opts,
		globalopts: globalopts,
	}
	h, err := newHTTPClient(cont, opts)
	if err != nil {
		return nil, err
	}
	g.http = h
	return &g, nil
}

// PreRun is the pre run implementation of gobusterdir
func (d *GobusterDir) PreRun() error {
	// add trailing slash
	if !strings.HasSuffix(d.options.URL, "/") {
		d.options.URL = fmt.Sprintf("%s/", d.options.URL)
	}

	_, _, err := d.get(d.options.URL)
	if err != nil {
		return fmt.Errorf("unable to connect to %s: %v", d.options.URL, err)
	}

	guid := uuid.New()
	url := fmt.Sprintf("%s%s", d.options.URL, guid)
	wildcardResp, _, err := d.get(url)
	if err != nil {
		return err
	}

	if d.options.StatusCodesParsed.Contains(*wildcardResp) {
		d.options.IsWildcard = true
		log.Printf("[-] Wildcard response found: %s => %d", url, *wildcardResp)
		if !d.options.WildcardForced {
			return fmt.Errorf("To force processing of Wildcard responses, specify the '--wildcard' switch.")
		}
	}

	return nil
}

// Run is the process implementation of gobusterdir
func (d *GobusterDir) Run(word string) ([]libgobuster.Result, error) {
	suffix := ""
	if d.options.UseSlash {
		suffix = "/"
	}

	// Try the DIR first
	url := fmt.Sprintf("%s%s%s", d.options.URL, word, suffix)
	dirResp, dirSize, err := d.get(url)
	if err != nil {
		return nil, err
	}
	var ret []libgobuster.Result
	if dirResp != nil {
		ret = append(ret, libgobuster.Result{
			Entity: fmt.Sprintf("%s%s", word, suffix),
			Status: *dirResp,
			Size:   dirSize,
		})
	}

	// Follow up with files using each ext.
	for ext := range d.options.ExtensionsParsed.Set {
		file := fmt.Sprintf("%s.%s", word, ext)
		url = fmt.Sprintf("%s%s", d.options.URL, file)
		fileResp, fileSize, err := d.get(url)
		if err != nil {
			return nil, err
		}

		if fileResp != nil {
			ret = append(ret, libgobuster.Result{
				Entity: file,
				Status: *fileResp,
				Size:   fileSize,
			})
		}
	}
	return ret, nil
}

// ResultToString is the to string implementation of gobusterdir
func (d *GobusterDir) ResultToString(r *libgobuster.Result) (*string, error) {
	buf := &bytes.Buffer{}

	// Prefix if we're in verbose mode
	if d.globalopts.Verbose {
		if d.options.StatusCodesParsed.Contains(r.Status) {
			if _, err := fmt.Fprintf(buf, "Found: "); err != nil {
				return nil, err
			}
		} else {
			if _, err := fmt.Fprintf(buf, "Missed: "); err != nil {
				return nil, err
			}
		}
	}

	if d.options.StatusCodesParsed.Contains(r.Status) || d.globalopts.Verbose {
		if d.options.Expanded {
			if _, err := fmt.Fprintf(buf, "%s", d.options.URL); err != nil {
				return nil, err
			}
		} else {
			if _, err := fmt.Fprintf(buf, "/"); err != nil {
				return nil, err
			}
		}
		if _, err := fmt.Fprintf(buf, "%s", r.Entity); err != nil {
			return nil, err
		}

		if !d.options.NoStatus {
			if _, err := fmt.Fprintf(buf, " (Status: %d)", r.Status); err != nil {
				return nil, err
			}
		}

		if r.Size != nil {
			if _, err := fmt.Fprintf(buf, " [Size: %d]", *r.Size); err != nil {
				return nil, err
			}
		}
		if _, err := fmt.Fprintf(buf, "\n"); err != nil {
			return nil, err
		}
	}
	s := buf.String()
	return &s, nil
}

// GetConfigString returns the string representation of the current config
func (d *GobusterDir) GetConfigString() (string, error) {
	buf := &bytes.Buffer{}
	o := d.options
	if _, err := fmt.Fprintf(buf, "[+] Url          : %s\n", o.URL); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(buf, "[+] Threads      : %d\n", d.globalopts.Threads); err != nil {
		return "", err
	}

	wordlist := "stdin (pipe)"
	if d.globalopts.Wordlist != "-" {
		wordlist = d.globalopts.Wordlist
	}
	if _, err := fmt.Fprintf(buf, "[+] Wordlist     : %s\n", wordlist); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(buf, "[+] Status codes : %s\n", o.StatusCodesParsed.Stringify()); err != nil {
		return "", err
	}

	if o.Proxy != "" {
		if _, err := fmt.Fprintf(buf, "[+] Proxy        : %s\n", o.Proxy); err != nil {
			return "", err
		}
	}

	if o.Cookies != "" {
		if _, err := fmt.Fprintf(buf, "[+] Cookies      : %s\n", o.Cookies); err != nil {
			return "", err
		}
	}

	if o.UserAgent != "" {
		if _, err := fmt.Fprintf(buf, "[+] User Agent   : %s\n", o.UserAgent); err != nil {
			return "", err
		}
	}

	if o.IncludeLength {
		if _, err := fmt.Fprintf(buf, "[+] Show length  : true\n"); err != nil {
			return "", err
		}
	}

	if o.Username != "" {
		if _, err := fmt.Fprintf(buf, "[+] Auth User    : %s\n", o.Username); err != nil {
			return "", err
		}
	}

	if o.Extensions != "" {
		if _, err := fmt.Fprintf(buf, "[+] Extensions   : %s\n", o.ExtensionsParsed.Stringify()); err != nil {
			return "", err
		}
	}

	if o.UseSlash {
		if _, err := fmt.Fprintf(buf, "[+] Add Slash    : true\n"); err != nil {
			return "", err
		}
	}

	if o.FollowRedirect {
		if _, err := fmt.Fprintf(buf, "[+] Follow Redir : true\n"); err != nil {
			return "", err
		}
	}

	if o.Expanded {
		if _, err := fmt.Fprintf(buf, "[+] Expanded     : true\n"); err != nil {
			return "", err
		}
	}

	if o.NoStatus {
		if _, err := fmt.Fprintf(buf, "[+] No status    : true\n"); err != nil {
			return "", err
		}
	}

	if d.globalopts.Verbose {
		if _, err := fmt.Fprintf(buf, "[+] Verbose      : true\n"); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(buf, "[+] Timeout      : %s\n", o.Timeout.String()); err != nil {
		return "", err
	}

	return strings.TrimSpace(buf.String()), nil
}
