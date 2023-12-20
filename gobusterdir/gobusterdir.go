package gobusterdir

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"
	"text/tabwriter"
	"unicode/utf8"

	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/google/uuid"
)

// nolint:gochecknoglobals
var (
	backupExtensions    = []string{"~", ".bak", ".bak2", ".old", ".1"}
	backupDotExtensions = []string{".swp"}
)

// ErrWildcard is returned if a wildcard response is found
type ErrWildcard struct {
	url        string
	location   string
	statusCode int
	length     int64
}

// Error is the implementation of the error interface
func (e *ErrWildcard) Error() string {
	addInfo := ""
	if e.location != "" {
		addInfo = fmt.Sprintf("%s => %d (redirect to %s) (Length: %d)", e.url, e.statusCode, e.location, e.length)
	} else {
		addInfo = fmt.Sprintf("%s => %d (Length: %d)", e.url, e.statusCode, e.length)
	}
	return fmt.Sprintf("the server returns a status code that matches the provided options for non existing urls. %s. Please exclude the response length or the status code or set the wildcard option.", addInfo)
}

// GobusterDir is the main type to implement the interface
type GobusterDir struct {
	options    *OptionsDir
	globalopts *libgobuster.Options
	http       *libgobuster.HTTPClient
}

// New creates a new initialized GobusterDir
func New(globalopts *libgobuster.Options, opts *OptionsDir, logger *libgobuster.Logger) (*GobusterDir, error) {
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

	basicOptions := libgobuster.BasicHTTPOptions{
		Proxy:           opts.Proxy,
		Timeout:         opts.Timeout,
		UserAgent:       opts.UserAgent,
		NoTLSValidation: opts.NoTLSValidation,
		RetryOnTimeout:  opts.RetryOnTimeout,
		RetryAttempts:   opts.RetryAttempts,
		TLSCertificate:  opts.TLSCertificate,
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
	}

	h, err := libgobuster.NewHTTPClient(&httpOpts, logger)
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
func (d *GobusterDir) PreRun(ctx context.Context, _ *libgobuster.Progress) error {
	// add trailing slash
	if !strings.HasSuffix(d.options.URL, "/") {
		d.options.URL = fmt.Sprintf("%s/", d.options.URL)
	}

	_, _, _, _, err := d.http.Request(ctx, d.options.URL, libgobuster.RequestOptions{})
	if err != nil {
		if errors.Is(err, io.EOF) {
			return libgobuster.ErrorEOF
		} else if os.IsTimeout(err) {
			return libgobuster.ErrorTimeout
		} else if errors.Is(err, syscall.ECONNREFUSED) {
			return libgobuster.ErrorConnectionRefused
		}
		return fmt.Errorf("unable to connect to %s: %w", d.options.URL, err)
	}

	guid := uuid.New()
	url := fmt.Sprintf("%s%s", d.options.URL, guid)
	if d.options.UseSlash {
		url = fmt.Sprintf("%s/", url)
	}

	wildcardResp, wildcardLength, wildcardHeader, _, err := d.http.Request(ctx, url, libgobuster.RequestOptions{})
	if err != nil {
		if errors.Is(err, io.EOF) {
			return libgobuster.ErrorEOF
		} else if os.IsTimeout(err) {
			return libgobuster.ErrorTimeout
		} else if errors.Is(err, syscall.ECONNREFUSED) {
			return libgobuster.ErrorConnectionRefused
		}
		return err
	}

	if d.options.ExcludeLengthParsed.Contains(int(wildcardLength)) {
		// we are done and ignore the request as the length is excluded
		return nil
	}

	if d.options.StatusCodesBlacklistParsed.Length() > 0 {
		if !d.options.StatusCodesBlacklistParsed.Contains(wildcardResp) {
			return &ErrWildcard{url: url, statusCode: wildcardResp, length: wildcardLength, location: wildcardHeader.Get("Location")}
		}
	} else if d.options.StatusCodesParsed.Length() > 0 {
		if d.options.StatusCodesParsed.Contains(wildcardResp) {
			return &ErrWildcard{url: url, statusCode: wildcardResp, length: wildcardLength, location: wildcardHeader.Get("Location")}
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
func (d *GobusterDir) ProcessWord(ctx context.Context, word string, progress *libgobuster.Progress) error {
	suffix := ""
	if d.options.UseSlash {
		suffix = "/"
	}
	entity := fmt.Sprintf("%s%s", word, suffix)

	// make sure the url ends with a slash
	if !strings.HasSuffix(d.options.URL, "/") {
		d.options.URL = fmt.Sprintf("%s/", d.options.URL)
	}
	// prevent double slashes by removing leading /
	if strings.HasPrefix(entity, "/") {
		// get size of first rune and trim it
		_, i := utf8.DecodeRuneInString(entity)
		entity = entity[i:]
	}
	url := fmt.Sprintf("%s%s", d.options.URL, entity)

	// add some debug output
	if d.globalopts.Debug {
		progress.MessageChan <- libgobuster.Message{
			Level:   libgobuster.LevelDebug,
			Message: fmt.Sprintf("trying %s", entity),
		}
	}

	tries := 1
	if d.options.RetryOnTimeout && d.options.RetryAttempts > 0 {
		// add it so it will be the overall max requests
		tries += d.options.RetryAttempts
	}

	var statusCode int
	var size int64
	var header http.Header
	for i := 1; i <= tries; i++ {
		var err error
		statusCode, size, header, _, err = d.http.Request(ctx, url, libgobuster.RequestOptions{})
		if err != nil {
			// check if it's a timeout and if we should try again and try again
			// otherwise the timeout error is raised
			if os.IsTimeout(err) && i != tries {
				continue
			} else if strings.Contains(err.Error(), "invalid control character in URL") {
				// put error in error chan, so it's printed out and ignore it
				// so gobuster will not quit
				progress.ErrorChan <- err
				continue
			} else {
				if errors.Is(err, io.EOF) {
					return libgobuster.ErrorEOF
				} else if os.IsTimeout(err) {
					return libgobuster.ErrorTimeout
				} else if errors.Is(err, syscall.ECONNREFUSED) {
					return libgobuster.ErrorConnectionRefused
				}
				return err
			}
		}
		break
	}

	if statusCode != 0 {
		resultStatus := false

		if d.options.StatusCodesBlacklistParsed.Length() > 0 {
			if !d.options.StatusCodesBlacklistParsed.Contains(statusCode) {
				resultStatus = true
			}
		} else if d.options.StatusCodesParsed.Length() > 0 {
			if d.options.StatusCodesParsed.Contains(statusCode) {
				resultStatus = true
			}
		} else {
			return fmt.Errorf("StatusCodes and StatusCodesBlacklist are both not set which should not happen")
		}

		if resultStatus && !d.options.ExcludeLengthParsed.Contains(int(size)) {
			path := "/"
			if d.options.Expanded {
				path = d.options.URL
			}
			path = fmt.Sprintf("%s%-20s", path, entity)

			r := Result{
				Path:       path,
				Header:     header,
				StatusCode: -1,
				Size:       -1,
			}
			if !d.options.NoStatus {
				r.StatusCode = statusCode
			}
			if !d.options.HideLength {
				r.Size = size
			}
			progress.ResultChan <- r
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

	if o.StatusCodesBlacklistParsed.Length() > 0 {
		if _, err := fmt.Fprintf(tw, "[+] Negative Status codes:\t%s\n", o.StatusCodesBlacklistParsed.Stringify()); err != nil {
			return "", err
		}
	} else if o.StatusCodesParsed.Length() > 0 {
		if _, err := fmt.Fprintf(tw, "[+] Status codes:\t%s\n", o.StatusCodesParsed.Stringify()); err != nil {
			return "", err
		}
	}

	if len(o.ExcludeLength) > 0 {
		if _, err := fmt.Fprintf(tw, "[+] Exclude Length:\t%s\n", d.options.ExcludeLengthParsed.Stringify()); err != nil {
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

	if o.HideLength {
		if _, err := fmt.Fprintf(tw, "[+] Show length:\tfalse\n"); err != nil {
			return "", err
		}
	}

	if o.Username != "" {
		if _, err := fmt.Fprintf(tw, "[+] Auth User:\t%s\n", o.Username); err != nil {
			return "", err
		}
	}

	if o.Extensions != "" || o.ExtensionsFile != "" {
		if _, err := fmt.Fprintf(tw, "[+] Extensions:\t%s\n", o.ExtensionsParsed.Stringify()); err != nil {
			return "", err
		}
	}

	if o.ExtensionsFile != "" {
		if _, err := fmt.Fprintf(tw, "[+] Extensions file:\t%s\n", o.ExtensionsFile); err != nil {
			return "", err
		}
	}

	if o.UseSlash {
		if _, err := fmt.Fprintf(tw, "[+] Add Slash:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if o.FollowRedirect {
		if _, err := fmt.Fprintf(tw, "[+] Follow Redirect:\ttrue\n"); err != nil {
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
