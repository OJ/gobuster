package gobustervhost

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/OJ/gobuster/libgobuster"
)

// GobusterVhost is the main type to implement the interface
type GobusterVhost struct {
	options      *OptionsVhost
	globalopts   *libgobuster.Options
	http         *libgobuster.HTTPClient
	domain       string
	baseResponse string
}

// NewGobusterVhost creates a new initialized GobusterDir
func NewGobusterVhost(cont context.Context, globalopts *libgobuster.Options, opts *OptionsVhost) (*GobusterVhost, error) {
	if globalopts == nil {
		return nil, fmt.Errorf("please provide valid global options")
	}

	if opts == nil {
		return nil, fmt.Errorf("please provide valid plugin options")
	}

	g := GobusterVhost{
		options:    opts,
		globalopts: globalopts,
	}

	httpOpts := libgobuster.HTTPOptions{
		Proxy:          opts.Proxy,
		FollowRedirect: opts.FollowRedirect,
		InsecureSSL:    opts.InsecureSSL,
		Timeout:        opts.Timeout,
		Username:       opts.Username,
		Password:       opts.Password,
		UserAgent:      opts.UserAgent,
	}

	h, err := libgobuster.NewHTTPClient(cont, &httpOpts)
	if err != nil {
		return nil, err
	}
	g.http = h
	return &g, nil
}

// PreRun is the pre run implementation of gobusterdir
func (v *GobusterVhost) PreRun() error {

	// add trailing slash
	if !strings.HasSuffix(v.options.URL, "/") {
		v.options.URL = fmt.Sprintf("%s/", v.options.URL)
	}

	url, err := url.Parse(v.options.URL)
	if err != nil {
		return fmt.Errorf("invalid url %s: %v", v.options.URL, err)
	}
	v.domain = url.Host

	_, bodyBase, err := v.http.GetBody(v.options.URL, "", v.options.Cookies)
	if err != nil {
		return fmt.Errorf("unable to connect to %s: %v", v.options.URL, err)
	}
	v.baseResponse = *bodyBase
	return nil
}

// Run is the process implementation of gobusterdir
func (v *GobusterVhost) Run(word string) ([]libgobuster.Result, error) {
	subdomain := fmt.Sprintf("%s.%s", word, v.domain)
	_, body, err := v.http.GetBody(v.options.URL, subdomain, v.options.Cookies)
	var ret []libgobuster.Result
	if err != nil {
		return ret, err
	}

	if *body != v.baseResponse {
		result := libgobuster.Result{
			Entity: subdomain,
		}
		ret = append(ret, result)
	}
	return ret, nil
}

// ResultToString is the to string implementation of gobusterdir
func (v *GobusterVhost) ResultToString(r *libgobuster.Result) (*string, error) {
	buf := &bytes.Buffer{}
	if _, err := fmt.Fprintf(buf, "Found: %s\n", r.Entity); err != nil {
		return nil, err
	}

	s := buf.String()
	return &s, nil
}

// GetConfigString returns the string representation of the current config
func (v *GobusterVhost) GetConfigString() (string, error) {
	var buffer bytes.Buffer
	bw := bufio.NewWriter(&buffer)
	tw := tabwriter.NewWriter(bw, 0, 5, 3, ' ', 0)
	o := v.options
	if _, err := fmt.Fprintf(tw, "[+] Url:\t%s\n", o.URL); err != nil {
		return "", err
	}
	if _, err := fmt.Fprintf(tw, "[+] Threads:\t%d\n", v.globalopts.Threads); err != nil {
		return "", err
	}

	wordlist := "stdin (pipe)"
	if v.globalopts.Wordlist != "-" {
		wordlist = v.globalopts.Wordlist
	}
	if _, err := fmt.Fprintf(tw, "[+] Wordlist:\t%s\n", wordlist); err != nil {
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

	if o.Username != "" {
		if _, err := fmt.Fprintf(tw, "[+] Auth User:\t%s\n", o.Username); err != nil {
			return "", err
		}
	}

	if v.globalopts.Verbose {
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
