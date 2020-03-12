package gobusterfuzz

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/OJ/gobuster/v3/libgobuster"
)

// ErrWildcard is returned if a wildcard response is found
type ErrWildcard struct {
	url        string
	statusCode int
}

// Error is the implementation of the error interface
func (e *ErrWildcard) Error() string {
	return fmt.Sprintf("the server returns a status code that matches the provided options for non existing urls. %s => %d", e.url, e.statusCode)
}

// GobusterFuzz is the main type to implement the interface
type GobusterFuzz struct {
	options    *OptionsFuzz
	globalopts *libgobuster.Options
	http       *libgobuster.HTTPClient
}

// NewGobusterFuzz creates a new initialized GobusterFuzz
func NewGobusterFuzz(cont context.Context, globalopts *libgobuster.Options, opts *OptionsFuzz) (*GobusterFuzz, error) {
	if globalopts == nil {
		return nil, fmt.Errorf("please provide valid global options")
	}

	if opts == nil {
		return nil, fmt.Errorf("please provide valid plugin options")
	}

	g := GobusterFuzz{
		options:    opts,
		globalopts: globalopts,
	}

	basicOptions := libgobuster.BasicHTTPOptions{
		Proxy:     opts.Proxy,
		Timeout:   opts.Timeout,
		UserAgent: opts.UserAgent,
	}

	httpOpts := libgobuster.HTTPOptions{
		BasicHTTPOptions: basicOptions,
		FollowRedirect:   opts.FollowRedirect,
		InsecureSSL:      opts.InsecureSSL,
		Username:         opts.Username,
		Password:         opts.Password,
		Headers:          opts.Headers,
		Cookies:          opts.Cookies,
		Method:           opts.Method,
	}

	h, err := libgobuster.NewHTTPClient(cont, &httpOpts)
	if err != nil {
		return nil, err
	}
	g.http = h
	return &g, nil
}

// Name should return the name of the plugin
func (d *GobusterFuzz) Name() string {
	return "fuzzing"
}

// PreRun is the pre run implementation of gobusterfuzz
func (d *GobusterFuzz) PreRun() error {
	return nil
}

// Run is the process implementation of gobusterfuzz
func (d *GobusterFuzz) Run(word string) ([]libgobuster.Result, error) {
	workingURL := strings.ReplaceAll(d.options.URL, "FUZZ", word)
	resp, size, _, err := d.http.Request(workingURL, libgobuster.RequestOptions{})
	if err != nil {
		return nil, err
	}
	var ret []libgobuster.Result
	if resp != nil {
		resultStatus := libgobuster.StatusFound

		if d.options.ExcludeSize >= 0 && size == d.options.ExcludeSize {
			resultStatus = libgobuster.StatusMissed
		}

		if d.options.ExcludedStatusCodesParsed.Length() > 0 {
			if d.options.ExcludedStatusCodesParsed.Contains(*resp) {
				resultStatus = libgobuster.StatusMissed
			}
		}

		if resultStatus == libgobuster.StatusFound || d.globalopts.Verbose {
			ret = append(ret, libgobuster.Result{
				Entity:     workingURL,
				StatusCode: *resp,
				Size:       &size,
				Status:     resultStatus,
			})
		}
	}
	return ret, nil
}

// ResultToString is the to string implementation of gobusterfuzz
func (d *GobusterFuzz) ResultToString(r *libgobuster.Result) (*string, error) {
	buf := &bytes.Buffer{}

	// Prefix if we're in verbose mode
	if d.globalopts.Verbose {
		switch r.Status {
		case libgobuster.StatusFound:
			if _, err := fmt.Fprintf(buf, "Found: "); err != nil {
				return nil, err
			}
		case libgobuster.StatusMissed:
			if _, err := fmt.Fprintf(buf, "Missed: "); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unknown status %d", r.Status)
		}
	}

	if _, err := fmt.Fprintf(buf, "[Status=%d] [Length=%d] %s", r.StatusCode, *r.Size, r.Entity); err != nil {
		return nil, err
	}

	if _, err := fmt.Fprintf(buf, "\n"); err != nil {
		return nil, err
	}

	s := buf.String()
	return &s, nil
}

// GetConfigString returns the string representation of the current config
func (d *GobusterFuzz) GetConfigString() (string, error) {
	var buffer bytes.Buffer
	bw := bufio.NewWriter(&buffer)
	tw := tabwriter.NewWriter(bw, 0, 5, 3, ' ', 0)
	o := d.options
	if _, err := fmt.Fprintf(tw, "[+] Url:\t%s\n", o.URL); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(tw, "[+] Method:\t%s\n", o.Method); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(tw, "[+] Threads:\t%d\n", d.globalopts.Threads); err != nil {
		return "", err
	}

	if d.globalopts.Delay > 0 {
		if _, err := fmt.Fprintf(tw, "[+] Delay:\t%s\n", d.globalopts.Delay); err != nil {
			return "", err
		}
	}

	wordlist := "stdin (pipe)"
	if d.globalopts.Wordlist != "-" {
		wordlist = d.globalopts.Wordlist
	}
	if _, err := fmt.Fprintf(tw, "[+] Wordlist:\t%s\n", wordlist); err != nil {
		return "", err
	}

	if d.globalopts.PatternFile != "" {
		if _, err := fmt.Fprintf(tw, "[+] Patterns:\t%s (%d entries)\n", d.globalopts.PatternFile, len(d.globalopts.Patterns)); err != nil {
			return "", err
		}
	}

	if o.ExcludedStatusCodesParsed.Length() > 0 {
		if _, err := fmt.Fprintf(tw, "[+] Excluded Status codes:\t%s\n", o.ExcludedStatusCodesParsed.Stringify()); err != nil {
			return "", err
		}
	}

	if o.ExcludeSize >= 0 {
		if _, err := fmt.Fprintf(tw, "[+] Excluded Response Size:\t%d\n", o.ExcludeSize); err != nil {
			return "", err
		}
	}

	if o.Proxy != "" {
		if _, err := fmt.Fprintf(tw, "[+] Proxy:\t%s\n", o.Proxy); err != nil {
			return "", err
		}
	}

	if o.Cookies != "" {
		if _, err := fmt.Fprintf(tw, "[+] Cookies:\t%s\n", o.Cookies); err != nil {
			return "", err
		}
	}

	if o.UserAgent != "" {
		if _, err := fmt.Fprintf(tw, "[+] User Agent:\t%s\n", o.UserAgent); err != nil {
			return "", err
		}
	}

	if o.Username != "" {
		if _, err := fmt.Fprintf(tw, "[+] Auth User:\t%s\n", o.Username); err != nil {
			return "", err
		}
	}

	if o.FollowRedirect {
		if _, err := fmt.Fprintf(tw, "[+] Follow Redir:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if d.globalopts.Verbose {
		if _, err := fmt.Fprintf(tw, "[+] Verbose:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(tw, "[+] Timeout:\t%s\n", o.Timeout.String()); err != nil {
		return "", err
	}

	if err := tw.Flush(); err != nil {
		return "", fmt.Errorf("error on tostring: %v", err)
	}

	if err := bw.Flush(); err != nil {
		return "", fmt.Errorf("error on tostring: %v", err)
	}

	return strings.TrimSpace(buffer.String()), nil
}
