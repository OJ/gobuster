package gobusterdir

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/tabwriter"

	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/google/uuid"
)

// GobusterDir is the main type to implement the interface
type GobusterDir struct {
	options    *OptionsDir
	globalopts *libgobuster.Options
	http       *libgobuster.HTTPClient
	WildCard   *libgobuster.WildCard
}

// GetRequest issues a GET request to the target and returns
// the status code, length and an error
func (d *GobusterDir) get(url string) (*int, *int64, error) {
	return d.http.Get(url, "", d.options.Cookies)
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
		WildCard:   libgobuster.WildCardInit(),
	}

	httpOpts := libgobuster.HTTPOptions{
		Proxy:          opts.Proxy,
		FollowRedirect: opts.FollowRedirect,
		InsecureSSL:    opts.InsecureSSL,
		Timeout:        opts.Timeout,
		Username:       opts.Username,
		Password:       opts.Password,
		UserAgent:      opts.UserAgent,
		Headers:        opts.Headers,
	}

	h, err := libgobuster.NewHTTPClient(cont, &httpOpts)
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

	if d.options.StatusCodesParsed.Contains(*wildcardResp) && !d.options.WildcardForced {
		d.WildCard.Set(d.options.URL, *wildcardResp, "")
		return fmt.Errorf("wildcard found")
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
		resultStatus := libgobuster.StatusMissed
		if d.options.StatusCodesParsed.Contains(*dirResp) {
			resultStatus = libgobuster.StatusFound
		}

		if resultStatus == libgobuster.StatusFound || d.globalopts.Verbose {
			ret = append(ret, libgobuster.Result{
				Entity:     fmt.Sprintf("%s%s", word, suffix),
				StatusCode: *dirResp,
				Size:       dirSize,
				Status:     resultStatus,
			})
		}
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
			resultStatus := libgobuster.StatusMissed
			if d.options.StatusCodesParsed.Contains(*fileResp) {
				resultStatus = libgobuster.StatusFound
			}

			if resultStatus == libgobuster.StatusFound || d.globalopts.Verbose {
				ret = append(ret, libgobuster.Result{
					Entity:     file,
					StatusCode: *fileResp,
					Size:       fileSize,
					Status:     resultStatus,
				})
			}

		}
	}
	return ret, nil
}

// ResultToString is the to string implementation of gobusterdir
func (d *GobusterDir) ResultToString(r *libgobuster.Result) (*string, error) {
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
		if _, err := fmt.Fprintf(buf, " (Status: %d)", r.StatusCode); err != nil {
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

	s := buf.String()
	return &s, nil
}

// GetConfigString returns the string representation of the current config
func (d *GobusterDir) GetConfigString() (string, error) {
	var buffer bytes.Buffer
	bw := bufio.NewWriter(&buffer)
	tw := tabwriter.NewWriter(bw, 0, 5, 3, ' ', 0)
	o := d.options
	if _, err := fmt.Fprintf(tw, "[+] Url:\t%s\n", o.URL); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(tw, "[+] Threads:\t%d\n", d.globalopts.Threads); err != nil {
		return "", err
	}

	wordlist := "stdin (pipe)"
	if d.globalopts.Wordlist != "-" {
		wordlist = d.globalopts.Wordlist
	}
	if _, err := fmt.Fprintf(tw, "[+] Wordlist:\t%s\n", wordlist); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(tw, "[+] Status codes:\t%s\n", o.StatusCodesParsed.Stringify()); err != nil {
		return "", err
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

	if o.IncludeLength {
		if _, err := fmt.Fprintf(tw, "[+] Show length:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if o.Username != "" {
		if _, err := fmt.Fprintf(tw, "[+] Auth User:\t%s\n", o.Username); err != nil {
			return "", err
		}
	}

	if o.Extensions != "" {
		if _, err := fmt.Fprintf(tw, "[+] Extensions:\t%s\n", o.ExtensionsParsed.Stringify()); err != nil {
			return "", err
		}
	}

	if o.UseSlash {
		if _, err := fmt.Fprintf(tw, "[+] Add Slash:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if o.FollowRedirect {
		if _, err := fmt.Fprintf(tw, "[+] Follow Redir:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if o.Expanded {
		if _, err := fmt.Fprintf(tw, "[+] Expanded:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if o.NoStatus {
		if _, err := fmt.Fprintf(tw, "[+] No status:\ttrue\n"); err != nil {
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

// GetRequestString returns the string representation of the current request
func (d *GobusterDir) GetRequestString() string {
	o := d.options
	return o.URL
}
