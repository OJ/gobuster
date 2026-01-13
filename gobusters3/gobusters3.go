package gobusters3

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"syscall"
	"text/tabwriter"

	"github.com/OJ/gobuster/v3/libgobuster"
)

// GobusterS3 is the main type to implement the interface
type GobusterS3 struct {
	options     *OptionsS3
	globalopts  *libgobuster.Options
	http        *libgobuster.HTTPClient
	bucketRegex *regexp.Regexp
}

// New creates a new initialized GobusterS3
func New(globalopts *libgobuster.Options, opts *OptionsS3, logger *libgobuster.Logger) (*GobusterS3, error) {
	if globalopts == nil {
		return nil, errors.New("please provide valid global options")
	}

	if opts == nil {
		return nil, errors.New("please provide valid plugin options")
	}

	g := GobusterS3{
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
	g.bucketRegex = regexp.MustCompile(`^[a-z0-9\-.]{3,63}$`)

	return &g, nil
}

// Name should return the name of the plugin
func (s *GobusterS3) Name() string {
	return "S3 bucket enumeration"
}

// PreRun is the pre run implementation of GobusterS3
func (s *GobusterS3) PreRun(_ context.Context, _ *libgobuster.Progress) error {
	return nil
}

// ProcessWord is the process implementation of GobusterS3
func (s *GobusterS3) ProcessWord(ctx context.Context, word string, progress *libgobuster.Progress) (libgobuster.Result, error) {
	// only check for valid bucket names
	if !s.isValidBucketName(word) {
		return nil, nil // nolint:nilnil
	}

	bucketURL := fmt.Sprintf("https://%s.s3.amazonaws.com/?max-keys=%d", url.PathEscape(word), s.options.MaxFilesToList)
	u, err := url.Parse(bucketURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse bucket URL %s: %w", bucketURL, err)
	}

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
		statusCode, _, _, body, err = s.http.Request(ctx, *u, libgobuster.RequestOptions{ReturnBody: true})
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
			}

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
		break
	}

	if statusCode == 0 || body == nil {
		return nil, nil // nolint:nilnil
	}

	// looks like 404 and 400 are the only negative status codes
	found := false
	switch statusCode {
	case http.StatusBadRequest:
	case http.StatusNotFound:
		found = false
	case http.StatusOK:
		// listing enabled
		found = true
		// parse xml
	default:
		// default to found as we use negative status codes
		found = true
	}

	// nothing found, bail out
	// may add the result later if we want to enable verbose output
	if !found {
		return nil, nil // nolint:nilnil
	}

	var extraStrBuilder strings.Builder

	if s.options.ShowFiles {
		// get status
		if bytes.Contains(body, []byte("<Error>")) {
			awsError := AWSError{}
			err := xml.Unmarshal(body, &awsError)
			if err != nil {
				return nil, fmt.Errorf("could not parse error xml: %w", err)
			}
			// https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html#ErrorCodeList
			_, err = fmt.Fprintf(&extraStrBuilder, "Error: %s (%s)", awsError.Message, awsError.Code)
			if err != nil {
				return nil, fmt.Errorf("fmt.Fprintf to strings.Builder failed: %w", err)
			}
		} else if bytes.Contains(body, []byte("<ListBucketResult ")) {
			// bucket listing enabled
			awsListing := AWSListing{}
			err := xml.Unmarshal(body, &awsListing)
			if err != nil {
				return nil, fmt.Errorf("could not parse result xml: %w", err)
			}
			extraStrBuilder.WriteString("Bucket Listing enabled: ")
			for i, x := range awsListing.Contents {
				if i > 0 {
					extraStrBuilder.WriteString(", ")
				}
				_, err := fmt.Fprintf(&extraStrBuilder, "%s (%db)", x.Key, x.Size)
				if err != nil {
					return nil, fmt.Errorf("fmt.Fprintf to strings.Builder failed: %w", err)
				}
			}
		}
	}

	r := Result{
		Found:      found,
		BucketName: word,
		Status:     extraStrBuilder.String(),
	}

	return r, nil
}

func (s *GobusterS3) AdditionalWordsLen() int {
	return 0
}

func (s *GobusterS3) AdditionalWords(_ string) []string {
	return []string{}
}

func (s *GobusterS3) AdditionalSuccessWords(_ string) []string {
	return []string{}
}

// GetConfigString returns the string representation of the current config
func (s *GobusterS3) GetConfigString() (string, error) {
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

	if o.LocalAddr != nil {
		if _, err := fmt.Fprintf(tw, "[+] Local IP:\t%s\n", o.LocalAddr); err != nil {
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
func (s *GobusterS3) isValidBucketName(bucketName string) bool {
	if !s.bucketRegex.MatchString(bucketName) {
		return false
	}
	if strings.HasSuffix(bucketName, "-") ||
		strings.HasPrefix(bucketName, ".") ||
		strings.HasPrefix(bucketName, "-") ||
		strings.Contains(bucketName, "..") ||
		strings.Contains(bucketName, ".-") ||
		strings.Contains(bucketName, "-.") {
		return false
	}
	return true
}
