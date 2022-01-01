package dir

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"text/tabwriter"

	"github.com/OJ/gobuster/v3/helper"
	"github.com/OJ/gobuster/v3/lib"
	"github.com/google/uuid"
)

// nolint:gochecknoglobals
var (
	backupExtensions    = []string{"~", ".bak", ".bak2", ".old", ".1"}
	backupDotExtensions = []string{".swp"}
)

// Error is the implementation of the error interface
func (e *ErrWildcard) Error() string {
	return fmt.Sprintf(
		"the server returns a status code that matches the provided options for non existing urls. %s => %d (Length: %d)",
		e.url,
		e.statusCode,
		e.length)
}

// NewGobusterDir creates a new initialized GobusterDir
func NewGobusterDir(globalopts *lib.Options, opts *OptionsDir) (*GobusterDir, error) {
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

	basicOptions := lib.BasicHTTPOptions{
		Proxy:           opts.Proxy,
		Timeout:         opts.Timeout,
		UserAgent:       opts.UserAgent,
		NoTLSValidation: opts.NoTLSValidation,
	}

	httpOpts := lib.HTTPOptions{
		BasicHTTPOptions: basicOptions,
		Password:         opts.Password,
		Username:         opts.Username,
		Cookies:          opts.Cookies,
		Headers:          opts.Headers,
		FollowRedirect:   opts.FollowRedirect,
		Method:           opts.Method,
	}

	h, err := lib.NewHTTPClient(&httpOpts)
	if err != nil {
		return nil, err
	}
	g.http = h

	return &g, nil
}

// Name should return the name of the plugin
func (d *GobusterDir) Name() string {
	return "directory enumeration"
}

// PreRun is the pre run implementation of gobusterdir
func (d *GobusterDir) PreRun(ctx context.Context) error {
	// add trailing slash
	if !strings.HasSuffix(d.options.URL, "/") {
		d.options.URL = fmt.Sprintf("%s/", d.options.URL)
	}

	_, _, _, _, err := d.http.Request(ctx, d.options.URL, lib.RequestOptions{})
	if err != nil {
		return fmt.Errorf("unable to connect to %s: %w", d.options.URL, err)
	}

	guid := uuid.New()
	url := fmt.Sprintf("%s%s", d.options.URL, guid)
	if d.options.UseSlash {
		url = fmt.Sprintf("%s/", url)
	}

	wildcardResp, wildcardLength, _, _, err := d.http.Request(ctx, url, lib.RequestOptions{})
	if err != nil {
		return err
	}

	if helper.SliceContains(d.options.ExcludeLength, int(wildcardLength)) {
		// we are done and ignore the request as the length is excluded
		return nil
	}

	if d.options.StatusCodesBlacklistParsed.Length() > 0 {
		if !d.options.StatusCodesBlacklistParsed.Contains(*wildcardResp) {
			return &ErrWildcard{url: url, statusCode: *wildcardResp, length: wildcardLength}
		}
	} else if d.options.StatusCodesParsed.Length() > 0 {
		if d.options.StatusCodesParsed.Contains(*wildcardResp) {
			return &ErrWildcard{url: url, statusCode: *wildcardResp, length: wildcardLength}
		}
	} else {
		return fmt.Errorf("StatusCodes and StatusCodesBlacklist are both not set which should not happen")
	}

	return nil
}

func getBackupFilenames(word string) []string {
	ret := make([]string, len(backupExtensions)+len(backupDotExtensions))
	i := 0
	for _, b := range backupExtensions {
		ret[i] = fmt.Sprintf("%s%s", word, b)
		i++
	}
	for _, b := range backupDotExtensions {
		ret[i] = fmt.Sprintf(".%s%s", word, b)
		i++
	}

	return ret
}

func (d *GobusterDir) AdditionalWords(word string) []string {
	var words []string
	// build list of urls to check
	//   1: No extension
	//   2: With extension
	//   3: backupextension
	if d.options.DiscoverBackup {
		words = append(words, getBackupFilenames(word)...)
	}
	for ext := range d.options.ExtensionsParsed.Set {
		filename := fmt.Sprintf("%s.%s", word, ext)
		words = append(words, filename)
		if d.options.DiscoverBackup {
			words = append(words, getBackupFilenames(filename)...)
		}
	}
	return words
}

// ProcessWord is the process implementation of gobusterdir
func (d *GobusterDir) ProcessWord(ctx context.Context, word string, progress *lib.Progress) error {
	suffix := ""
	if d.options.UseSlash {
		suffix = "/"
	}
	entity := fmt.Sprintf("%s%s", word, suffix)
	url := fmt.Sprintf("%s%s", d.options.URL, entity)

	tries := 1
	if d.options.RetryOnTimeout && d.options.RetryAttempts > 0 {
		// add it so it will be the overall max requests
		tries += d.options.RetryAttempts
	}

	var statusCode *int
	var size int64
	var header http.Header
	for i := 1; i <= tries; i++ {
		var err error
		statusCode, size, header, _, err = d.http.Request(ctx, url, lib.RequestOptions{})
		if err != nil {
			// check if it's a timeout and if we should try again and try again
			// otherwise the timeout error is raised
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() && i != tries {
				continue
			} else if strings.Contains(err.Error(), "invalid control character in URL") {
				// put error in error chan so it's printed out and ignore it
				// so gobuster will not quit
				progress.ErrorChan <- err
				continue
			} else {
				return err
			}
		}
		break
	}

	if statusCode != nil {
		resultStatus := false

		if d.options.StatusCodesBlacklistParsed.Length() > 0 {
			if !d.options.StatusCodesBlacklistParsed.Contains(*statusCode) {
				resultStatus = true
			}
		} else if d.options.StatusCodesParsed.Length() > 0 {
			if d.options.StatusCodesParsed.Contains(*statusCode) {
				resultStatus = true
			}
		} else {
			return fmt.Errorf("StatusCodes and StatusCodesBlacklist are both not set which should not happen")
		}

		if (resultStatus && !helper.SliceContains(d.options.ExcludeLength, int(size))) || d.globalopts.Verbose {
			progress.ResultChan <- Result{
				URL:        d.options.URL,
				Path:       entity,
				Verbose:    d.globalopts.Verbose,
				Expanded:   d.options.Expanded,
				NoStatus:   d.options.NoStatus,
				HideLength: d.options.HideLength,
				Found:      resultStatus,
				Header:     header,
				StatusCode: *statusCode,
				Size:       size,
			}
		}
	}

	return nil
}

// GetConfigString returns the string representation of the current config
func (d *GobusterDir) GetConfigString() (string, error) {
	var buffer bytes.Buffer
	bw := bufio.NewWriter(&buffer)
	tw := tabwriter.NewWriter(bw, 0, 5, 3, ' ', 0)
	o := d.options

	var output string

	output = fmt.Sprintf("[+] Url:\t%s\n[+] Method:\t%s\n[+] Threads:\t%d\n", o.URL, o.Method, d.globalopts.Threads)

	if d.globalopts.Delay > 0 {
		output += fmt.Sprintf("[+] Delay:\t%s\n", d.globalopts.Delay)
	}

	wordlist := "stdin (pipe)"
	if d.globalopts.Wordlist != "-" {
		wordlist = d.globalopts.Wordlist
	}
	output += fmt.Sprintf("[+] Wordlist:\t%s\n", wordlist)

	if d.globalopts.PatternFile != "" {
		output += fmt.Sprintf("[+] Patterns:\t%s (%d entries)\n", d.globalopts.PatternFile, len(d.globalopts.Patterns))
	}

	if o.StatusCodesBlacklistParsed.Length() > 0 {
		output += fmt.Sprintf("[+] Negative Status codes:\t%s\n", o.StatusCodesBlacklistParsed.Stringify())

	} else if o.StatusCodesParsed.Length() > 0 {
		output += fmt.Sprintf("[+] Status codes:\t%s\n", o.StatusCodesParsed.Stringify())
	}

	if len(o.ExcludeLength) > 0 {
		output += fmt.Sprintf("[+] Exclude Length:\t%s\n", helper.JoinIntSlice(d.options.ExcludeLength))
	}

	if o.Proxy != "" {
		output += fmt.Sprintf("[+] Proxy:\t%s\n", o.Proxy)
	}

	if o.Cookies != "" {
		output += fmt.Sprintf("[+] Cookies:\t%s\n", o.Cookies)
	}

	if o.UserAgent != "" {
		output += fmt.Sprintf("[+] User Agent:\t%s\n", o.UserAgent)
	}

	if o.HideLength {
		output += "[+] Show length:\tfalse\n"
	}

	if o.Username != "" {
		output += fmt.Sprintf("[+] Auth User:\t%s\n", o.Username)
	}

	if o.Extensions != "" {
		output += fmt.Sprintf("[+] Extensions:\t%s\n", o.ExtensionsParsed.Stringify())
	}

	if o.UseSlash {
		output += "[+] Add Slash:\ttrue\n"
	}

	if o.FollowRedirect {
		output += "[+] Follow Redirect:\ttrue\n"
	}

	if o.Expanded {
		output += "[+] Expanded:\ttrue\n"
	}

	if o.NoStatus {
		output += "[+] No status:\ttrue\n"
	}

	if d.globalopts.Verbose {
		output += "[+] Verbose:\ttrue\n"
	}

	output += fmt.Sprintf("[+] Timeout:\t%s\n", o.Timeout.String())

	if _, err := fmt.Fprint(tw, output); err != nil {
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
