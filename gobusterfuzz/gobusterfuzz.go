package gobusterfuzz

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/OJ/gobuster/v3/helper"
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
func NewGobusterFuzz(globalopts *libgobuster.Options, opts *OptionsFuzz) (*GobusterFuzz, error) {
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
		Proxy:           opts.Proxy,
		Timeout:         opts.Timeout,
		UserAgent:       opts.UserAgent,
		NoTLSValidation: opts.NoTLSValidation,
	}

	httpOpts := libgobuster.HTTPOptions{
		BasicHTTPOptions: basicOptions,
		FollowRedirect:   opts.FollowRedirect,
		Username:         opts.Username,
		Password:         opts.Password,
		Headers:          opts.Headers,
		Cookies:          opts.Cookies,
		Method:           opts.Method,
	}

	h, err := libgobuster.NewHTTPClient(&httpOpts)
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

// RequestsPerRun returns the number of requests this plugin makes per single wordlist item
func (d *GobusterFuzz) RequestsPerRun() int {
	return 1
}

// PreRun is the pre run implementation of gobusterfuzz
func (d *GobusterFuzz) PreRun(ctx context.Context) error {
	return nil
}

// Run is the process implementation of gobusterfuzz
func (d *GobusterFuzz) Run(ctx context.Context, word string, resChannel chan<- libgobuster.Result) error {
	workingURL := strings.ReplaceAll(d.options.URL, "FUZZ", word)
	statusCode, size, _, _, err := d.http.Request(ctx, workingURL, libgobuster.RequestOptions{})
	if err != nil {
		return err
	}
	if statusCode != nil {
		resultStatus := true

		if helper.SliceContains(d.options.ExcludeLength, int(size)) {
			resultStatus = false
		}

		if d.options.ExcludedStatusCodesParsed.Length() > 0 {
			if d.options.ExcludedStatusCodesParsed.Contains(*statusCode) {
				resultStatus = false
			}
		}

		if resultStatus || d.globalopts.Verbose {
			resChannel <- Result{
				Verbose:    d.globalopts.Verbose,
				Found:      resultStatus,
				Path:       workingURL,
				StatusCode: *statusCode,
				Size:       size,
			}
		}
	}
	return nil
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

	if len(o.ExcludeLength) > 0 {
		if _, err := fmt.Fprintf(tw, "[+] Exclude Length:\t%s\n", helper.JoinIntSlice(d.options.ExcludeLength)); err != nil {
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
		if _, err := fmt.Fprintf(tw, "[+] Follow Redirect:\ttrue\n"); err != nil {
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
		return "", fmt.Errorf("error on tostring: %w", err)
	}

	if err := bw.Flush(); err != nil {
		return "", fmt.Errorf("error on tostring: %w", err)
	}

	return strings.TrimSpace(buffer.String()), nil
}
