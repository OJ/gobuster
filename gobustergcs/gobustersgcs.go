package gobustergcs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/OJ/gobuster/v3/libgobuster"
)

// GobusterGCS is the main type to implement the interface
type GobusterGCS struct {
	options     *OptionsGCS
	globalopts  *libgobuster.Options
	http        *libgobuster.HTTPClient
	bucketRegex *regexp.Regexp
}

// New creates a new initialized GobusterGCS
func New(globalopts *libgobuster.Options, opts *OptionsGCS, logger *libgobuster.Logger) (*GobusterGCS, error) {
	if globalopts == nil {
		return nil, fmt.Errorf("please provide valid global options")
	}

	if opts == nil {
		return nil, fmt.Errorf("please provide valid plugin options")
	}

	g := GobusterGCS{
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
		BasicHTTPOptions: basicOptions,
		// needed so we can list bucket contents
		FollowRedirect: true,
	}

	h, err := libgobuster.NewHTTPClient(&httpOpts, logger)
	if err != nil {
		return nil, err
	}
	g.http = h
	// https://cloud.google.com/storage/docs/naming-buckets
	g.bucketRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9\-_.]{1,61}[a-z0-9](\.[a-z0-9][a-z0-9\-_.]{1,61}[a-z0-9])*$`)

	return &g, nil
}

// Name should return the name of the plugin
func (s *GobusterGCS) Name() string {
	return "GCS bucket enumeration"
}

// PreRun is the pre run implementation of GobusterS3
func (s *GobusterGCS) PreRun(_ context.Context, _ *libgobuster.Progress) error {
	return nil
}

// ProcessWord is the process implementation of GobusterS3
func (s *GobusterGCS) ProcessWord(ctx context.Context, word string, progress *libgobuster.Progress) error {
	// only check for valid bucket names
	if !s.isValidBucketName(word) {
		return nil
	}

	bucketURL := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o?maxResults=%d", word, s.options.MaxFilesToList)

	// add some debug output
	if s.globalopts.Debug {
		progress.MessageChan <- libgobuster.Message{
			Level:   libgobuster.LevelDebug,
			Message: fmt.Sprintf("trying word %s", word),
		}
	}

	tries := 1
	if s.options.RetryOnTimeout && s.options.RetryAttempts > 0 {
		// add it so it will be the overall max requests
		tries += s.options.RetryAttempts
	}

	var statusCode int
	var body []byte
	for i := 1; i <= tries; i++ {
		var err error
		statusCode, _, _, body, err = s.http.Request(ctx, bucketURL, libgobuster.RequestOptions{ReturnBody: true})
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

	if statusCode == 0 || body == nil {
		return nil
	}

	// looks like 401, 403, and 404 are the only negative status codes
	found := false
	switch statusCode {
	case http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound:
		found = false
	case http.StatusOK:
		// listing enabled
		found = true
	default:
		// default to found as we use negative status codes
		found = true
	}

	// nothing found, bail out
	// may add the result later if we want to enable verbose output
	if !found {
		return nil
	}

	extraStr := ""
	if s.options.ShowFiles {
		// get status
		var result map[string]interface{}
		err := json.Unmarshal(body, &result)

		if err != nil {
			return fmt.Errorf("could not parse response json: %w", err)
		}

		if _, exist := result["error"]; exist {
			// https://cloud.google.com/storage/docs/json_api/v1/status-codes
			gcsError := GCSError{}
			err := json.Unmarshal(body, &gcsError)
			if err != nil {
				return fmt.Errorf("could not parse error json: %w", err)
			}
			extraStr = fmt.Sprintf("Error: %s (%d)", gcsError.Error.Message, gcsError.Error.Code)
		} else if v, exist := result["kind"]; exist && v == "storage#objects" {
			// https://cloud.google.com/storage/docs/json_api/v1/status-codes
			// bucket listing enabled
			gcsListing := GCSListing{}
			err := json.Unmarshal(body, &gcsListing)
			if err != nil {
				return fmt.Errorf("could not parse result json: %w", err)
			}
			extraStr = "Bucket Listing enabled: "
			for _, x := range gcsListing.Items {
				extraStr += fmt.Sprintf("%s (%sb), ", x.Name, x.Size)
			}
			extraStr = strings.TrimRight(extraStr, ", ")
		}
	}

	progress.ResultChan <- Result{
		Found:      found,
		BucketName: word,
		Status:     extraStr,
	}

	return nil
}

func (s *GobusterGCS) AdditionalWords(_ string) []string {
	return []string{}
}

// GetConfigString returns the string representation of the current config
func (s *GobusterGCS) GetConfigString() (string, error) {
	var buffer bytes.Buffer
	bw := bufio.NewWriter(&buffer)
	tw := tabwriter.NewWriter(bw, 0, 5, 3, ' ', 0)
	o := s.options

	if _, err := fmt.Fprintf(tw, "[+] Threads:\t%d\n", s.globalopts.Threads); err != nil {
		return "", err
	}

	if s.globalopts.Delay > 0 {
		if _, err := fmt.Fprintf(tw, "[+] Delay:\t%s\n", s.globalopts.Delay); err != nil {
			return "", err
		}
	}

	wordlist := "stdin (pipe)"
	if s.globalopts.Wordlist != "-" {
		wordlist = s.globalopts.Wordlist
	}
	if _, err := fmt.Fprintf(tw, "[+] Wordlist:\t%s\n", wordlist); err != nil {
		return "", err
	}

	if s.globalopts.PatternFile != "" {
		if _, err := fmt.Fprintf(tw, "[+] Patterns:\t%s (%d entries)\n", s.globalopts.PatternFile, len(s.globalopts.Patterns)); err != nil {
			return "", err
		}
	}

	if o.Proxy != "" {
		if _, err := fmt.Fprintf(tw, "[+] Proxy:\t%s\n", o.Proxy); err != nil {
			return "", err
		}
	}

	if o.UserAgent != "" {
		if _, err := fmt.Fprintf(tw, "[+] User Agent:\t%s\n", o.UserAgent); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(tw, "[+] Timeout:\t%s\n", o.Timeout.String()); err != nil {
		return "", err
	}

	if s.options.ShowFiles {
		if _, err := fmt.Fprintf(tw, "[+] Show Files:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(tw, "[+] Maximum files to list:\t%d\n", o.MaxFilesToList); err != nil {
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

// https://docs.aws.amazon.com/AmazonS3/latest/dev/BucketRestrictions.html
func (s *GobusterGCS) isValidBucketName(bucketName string) bool {
	if len(bucketName) > 222 || !s.bucketRegex.MatchString(bucketName) {
		return false
	}
	if strings.HasPrefix(bucketName, "-") || strings.HasSuffix(bucketName, "-") ||
		strings.HasPrefix(bucketName, "_") || strings.HasSuffix(bucketName, "_") ||
		strings.HasPrefix(bucketName, ".") || strings.HasSuffix(bucketName, ".") {
		return false
	}
	return true
}
