package gobustervhost

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/OJ/gobuster/v3/helper"
	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/google/uuid"
)

// GobusterVhost is the main type to implement the interface
type GobusterVhost struct {
	options      *OptionsVhost
	globalopts   *libgobuster.Options
	http         *libgobuster.HTTPClient
	domain       string
	normalBody   []byte
	abnormalBody []byte
}

// NewGobusterVhost creates a new initialized GobusterDir
func NewGobusterVhost(globalopts *libgobuster.Options, opts *OptionsVhost) (*GobusterVhost, error) {
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
func (v *GobusterVhost) Name() string {
	return "VHOST enumeration"
}

// PreRun is the pre run implementation of gobusterdir
func (v *GobusterVhost) PreRun(ctx context.Context) error {
	// add trailing slash
	if !strings.HasSuffix(v.options.URL, "/") {
		v.options.URL = fmt.Sprintf("%s/", v.options.URL)
	}

	urlParsed, err := url.Parse(v.options.URL)
	if err != nil {
		return fmt.Errorf("invalid url %s: %w", v.options.URL, err)
	}
	if v.options.Domain != "" {
		v.domain = v.options.Domain
	} else {
		v.domain = urlParsed.Host
	}

	// request default vhost for normalBody
	_, _, _, body, err := v.http.Request(ctx, v.options.URL, libgobuster.RequestOptions{ReturnBody: true})
	if err != nil {
		return fmt.Errorf("unable to connect to %s: %w", v.options.URL, err)
	}
	v.normalBody = body

	// request non existent vhost for abnormalBody
	subdomain := fmt.Sprintf("%s.%s", uuid.New(), v.domain)
	_, _, _, body, err = v.http.Request(ctx, v.options.URL, libgobuster.RequestOptions{Host: subdomain, ReturnBody: true})
	if err != nil {
		return fmt.Errorf("unable to connect to %s: %w", v.options.URL, err)
	}
	v.abnormalBody = body
	return nil
}

// ProcessWord is the process implementation of gobusterdir
func (v *GobusterVhost) ProcessWord(ctx context.Context, word string, progress *libgobuster.Progress) error {
	var subdomain string
	if v.options.AppendDomain {
		subdomain = fmt.Sprintf("%s.%s", word, v.domain)
	} else {
		// wordlist needs to include full domains
		subdomain = word
	}

	tries := 1
	if v.options.RetryOnTimeout && v.options.RetryAttempts > 0 {
		// add it so it will be the overall max requests
		tries += v.options.RetryAttempts
	}

	var statusCode int
	var size int64
	var header http.Header
	var body []byte
	for i := 1; i <= tries; i++ {
		var err error
		statusCode, size, header, body, err = v.http.Request(ctx, v.options.URL, libgobuster.RequestOptions{Host: subdomain, ReturnBody: true})
		if err != nil {
			// check if it's a timeout and if we should try again and try again
			// otherwise the timeout error is raised
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() && i != tries {
				continue
			} else {
				return err
			}
		}
		break
	}

	// subdomain must not match default vhost and non existent vhost
	// or verbose mode is enabled
	found := body != nil && !bytes.Equal(body, v.normalBody) && !bytes.Equal(body, v.abnormalBody)
	if (found && !helper.SliceContains(v.options.ExcludeLength, int(size))) || v.globalopts.Verbose {
		resultStatus := false
		if found {
			resultStatus = true
		}
		progress.ResultChan <- Result{
			Found:      resultStatus,
			Vhost:      subdomain,
			StatusCode: statusCode,
			Size:       size,
			Header:     header,
		}
	}
	return nil
}

func (v *GobusterVhost) AdditionalWords(word string) []string {
	return []string{}
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

	if _, err := fmt.Fprintf(tw, "[+] Method:\t%s\n", o.Method); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(tw, "[+] Threads:\t%d\n", v.globalopts.Threads); err != nil {
		return "", err
	}

	if v.globalopts.Delay > 0 {
		if _, err := fmt.Fprintf(tw, "[+] Delay:\t%s\n", v.globalopts.Delay); err != nil {
			return "", err
		}
	}

	wordlist := "stdin (pipe)"
	if v.globalopts.Wordlist != "-" {
		wordlist = v.globalopts.Wordlist
	}
	if _, err := fmt.Fprintf(tw, "[+] Wordlist:\t%s\n", wordlist); err != nil {
		return "", err
	}

	if v.globalopts.PatternFile != "" {
		if _, err := fmt.Fprintf(tw, "[+] Patterns:\t%s (%d entries)\n", v.globalopts.PatternFile, len(v.globalopts.Patterns)); err != nil {
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

	if v.globalopts.Verbose {
		if _, err := fmt.Fprintf(tw, "[+] Verbose:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(tw, "[+] Timeout:\t%s\n", o.Timeout.String()); err != nil {
		return "", err
	}

	if _, err := fmt.Fprintf(tw, "[+] Append Domain:\t%t\n", v.options.AppendDomain); err != nil {
		return "", err
	}

	if len(o.ExcludeLength) > 0 {
		if _, err := fmt.Fprintf(tw, "[+] Exclude Length:\t%s\n", helper.JoinIntSlice(v.options.ExcludeLength)); err != nil {
			return "", err
		}
	}

	if err := tw.Flush(); err != nil {
		return "", fmt.Errorf("error on tostring: %w", err)
	}

	if err := bw.Flush(); err != nil {
		return "", fmt.Errorf("error on tostring: %w", err)
	}

	return strings.TrimSpace(buffer.String()), nil
}
