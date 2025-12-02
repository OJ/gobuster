package gobustervhost

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"

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
	once         sync.Once
}

// New creates a new initialized GobusterDir
func New(globalopts *libgobuster.Options, opts *OptionsVhost, logger *libgobuster.Logger) (*GobusterVhost, error) {
	if globalopts == nil {
		return nil, errors.New("please provide valid global options")
	}

	if opts == nil {
		return nil, errors.New("please provide valid plugin options")
	}

	g := GobusterVhost{
		options:    opts,
		globalopts: globalopts,
	}

	basicOptions := libgobuster.BasicHTTPOptions{
		Proxy:            opts.Proxy,
		Timeout:          opts.Timeout,
		UserAgent:        opts.UserAgent,
		NoTLSValidation:  opts.NoTLSValidation,
		RetryOnTimeout:   opts.RetryOnTimeout,
		RetryAttempts:    opts.RetryAttempts,
		TLSCertificate:   opts.TLSCertificate,
		LocalAddr:        opts.LocalAddr,
		TLSRenegotiation: opts.TLSRenegotiation,
	}

	httpOpts := libgobuster.HTTPOptions{
		BasicHTTPOptions:      basicOptions,
		FollowRedirect:        opts.FollowRedirect,
		Username:              opts.Username,
		Password:              opts.Password,
		Headers:               opts.Headers,
		NoCanonicalizeHeaders: opts.NoCanonicalizeHeaders,
		Cookies:               opts.Cookies,
		Method:                opts.Method,
		BodyOutputDir:         opts.BodyOutputDir,
	}

	h, err := libgobuster.NewHTTPClient(&httpOpts, logger)
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
func (v *GobusterVhost) PreRun(ctx context.Context, _ *libgobuster.Progress) error {
	// add trailing slash
	if !strings.HasSuffix(v.options.URL.Path, "/") {
		v.options.URL.Path = fmt.Sprintf("%s/", v.options.URL.Path)
	}

	if v.options.Domain != "" {
		v.domain = v.options.Domain
	} else {
		v.domain = v.options.URL.Host
	}

	// request default vhost for normalBody
	_, _, _, body, err := v.http.Request(ctx, *v.options.URL, libgobuster.RequestOptions{ReturnBody: true})
	if err != nil {
		switch {
		case errors.Is(err, io.EOF):
			return libgobuster.ErrEOF
		case os.IsTimeout(err):
			return libgobuster.ErrTimeout
		case errors.Is(err, syscall.ECONNREFUSED):
			return libgobuster.ErrConnectionRefused
		}
		return fmt.Errorf("unable to connect to %s: %w", v.options.URL, err)
	}
	v.normalBody = body

	// request non existent vhost for abnormalBody
	subdomain := fmt.Sprintf("%s.%s", uuid.New(), v.domain)
	_, _, _, body, err = v.http.Request(ctx, *v.options.URL, libgobuster.RequestOptions{Host: subdomain, ReturnBody: true})
	if err != nil {
		switch {
		case errors.Is(err, io.EOF):
			return libgobuster.ErrEOF
		case os.IsTimeout(err):
			return libgobuster.ErrTimeout
		case errors.Is(err, syscall.ECONNREFUSED):
			return libgobuster.ErrConnectionRefused
		}
		return fmt.Errorf("unable to connect to %s: %w", v.options.URL, err)
	}
	v.abnormalBody = body
	return nil
}

// ProcessWord is the process implementation of gobusterdir
func (v *GobusterVhost) ProcessWord(ctx context.Context, word string, progress *libgobuster.Progress) (libgobuster.Result, error) {
	var subdomain string
	var hostnameLength int
	if v.options.AppendDomain {
		subdomain = fmt.Sprintf("%s.%s", word, v.domain)
	} else {
		// wordlist needs to include full domains
		subdomain = word
	}
	if v.options.ExcludeHostnameLength {
		hostnameLength = len(subdomain)
	} else {
		hostnameLength = 0
	}

	// warn people when there is no . detected so they might want to use the other options
	v.once.Do(func() {
		if !strings.Contains(subdomain, ".") {
			progress.MessageChan <- libgobuster.Message{
				Level:   libgobuster.LevelWarn,
				Message: fmt.Sprintf("the first subdomain to try does not contain a dot (%s). You might want to use the option to append the base domain otherwise the vhost will be tried as is", subdomain),
			}
		}
	})

	// add some debug output
	if v.globalopts.Debug {
		progress.MessageChan <- libgobuster.Message{
			Level:   libgobuster.LevelDebug,
			Message: fmt.Sprintf("trying vhost %s", subdomain),
		}
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
		statusCode, size, header, body, err = v.http.Request(ctx, *v.options.URL, libgobuster.RequestOptions{Host: subdomain, ReturnBody: true})
		if err != nil {
			// check if it's a timeout and if we should try again and try again
			// otherwise the timeout error is raised
			switch {
			case os.IsTimeout(err) && i != tries:
				continue
			case strings.Contains(err.Error(), "invalid control character in URL"):
				// put error in error chan, so it's printed out and ignore it
				// so gobuster will not quit
				progress.ErrorChan <- err
				continue
			default:
				switch {
				case errors.Is(err, io.EOF):
					return nil, libgobuster.ErrEOF
				case os.IsTimeout(err):
					return nil, libgobuster.ErrTimeout
				case errors.Is(err, syscall.ECONNREFUSED):
					return nil, libgobuster.ErrConnectionRefused
				}
				return nil, err
			}
		}
		break
	}

	if v.options.BodyOutputDir != "" && body != nil {
		fname := libgobuster.SanitizeFilename(fmt.Sprintf("%s_%d.html", strings.Trim(word, "/"), statusCode))
		fpath := filepath.Join(v.options.BodyOutputDir, fname)
		err := os.WriteFile(fpath, body, 0o600)
		if err != nil {
			progress.MessageChan <- libgobuster.Message{
				Level:   libgobuster.LevelError,
				Message: fmt.Sprintf("Could not write body to file %s: %v", fpath, err),
			}
		}
	}

	// subdomain must not match default vhost and non existent vhost
	// or verbose mode is enabled
	found := body != nil && !bytes.Equal(body, v.normalBody) && !bytes.Equal(body, v.abnormalBody)
	if found && !v.options.ExcludeLengthParsed.Contains(int(size)-hostnameLength) && !v.options.ExcludeStatusParsed.Contains(statusCode) {
		r := Result{
			Vhost:      subdomain,
			StatusCode: statusCode,
			Size:       size,
			Header:     header,
		}
		return r, nil
	}
	return nil, nil // nolint:nilnil
}

func (v *GobusterVhost) AdditionalWordsLen() int {
	return 0
}

func (v *GobusterVhost) AdditionalWords(_ string) []string {
	return []string{}
}

func (v *GobusterVhost) AdditionalSuccessWords(_ string) []string {
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

	if o.LocalAddr != nil {
		if _, err := fmt.Fprintf(tw, "[+] Local IP:\t%s\n", o.LocalAddr); err != nil {
			return "", err
		}
	}

	if o.Username != "" {
		if _, err := fmt.Fprintf(tw, "[+] Auth User:\t%s\n", o.Username); err != nil {
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
		if _, err := fmt.Fprintf(tw, "[+] Exclude Length:\t%s\n", v.options.ExcludeLengthParsed.Stringify()); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(tw, "[+] Exclude Hostname Length:\t%t\n", v.options.ExcludeHostnameLength); err != nil {
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
