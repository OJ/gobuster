package gobusters3

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"
	"strings"
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

// NewGobusterS3 creates a new initialized GobusterS3
func NewGobusterS3(globalopts *libgobuster.Options, opts *OptionsS3) (*GobusterS3, error) {
	if globalopts == nil {
		return nil, fmt.Errorf("please provide valid global options")
	}

	if opts == nil {
		return nil, fmt.Errorf("please provide valid plugin options")
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
	}

	httpOpts := libgobuster.HTTPOptions{
		BasicHTTPOptions: basicOptions,
		// needed so we can list bucket contents
		FollowRedirect: true,
	}

	h, err := libgobuster.NewHTTPClient(&httpOpts)
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

// RequestsPerRun returns the number of requests this plugin makes per single wordlist item
func (s *GobusterS3) RequestsPerRun() int {
	return 1
}

// PreRun is the pre run implementation of GobusterS3
func (s *GobusterS3) PreRun(ctx context.Context) error {
	return nil
}

// Run is the process implementation of GobusterS3
func (s *GobusterS3) Run(ctx context.Context, word string, resChannel chan<- libgobuster.Result) error {
	// only check for valid bucket names
	if !s.isValidBucketName(word) {
		return nil
	}

	bucketURL := fmt.Sprintf("https://%s.s3.amazonaws.com/?max-keys=%d", word, s.options.MaxFilesToList)
	status, _, _, body, err := s.http.Request(ctx, bucketURL, libgobuster.RequestOptions{ReturnBody: true})
	if err != nil {
		return err
	}

	if status == nil || body == nil {
		return nil
	}

	// looks like 404 and 400 are the only negative status codes
	found := false
	switch *status {
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
		return nil
	}

	extraStr := ""
	if s.globalopts.Verbose {
		// get status
		if bytes.Contains(body, []byte("<Error>")) {
			awsError := AWSError{}
			err := xml.Unmarshal(body, &awsError)
			if err != nil {
				return fmt.Errorf("could not parse error xml: %w", err)
			}
			// https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html#ErrorCodeList
			extraStr = fmt.Sprintf("Error: %s (%s)", awsError.Message, awsError.Code)
		} else if bytes.Contains(body, []byte("<ListBucketResult ")) {
			// bucket listing enabled
			awsListing := AWSListing{}
			err := xml.Unmarshal(body, &awsListing)
			if err != nil {
				return fmt.Errorf("could not parse result xml: %w", err)
			}
			extraStr = "Bucket Listing enabled: "
			for _, x := range awsListing.Contents {
				extraStr += fmt.Sprintf("%s (%db), ", x.Key, x.Size)
			}
			extraStr = strings.TrimRight(extraStr, ", ")
		}
	}

	resChannel <- Result{
		Found:      found,
		BucketName: word,
		Status:     extraStr,
	}

	return nil
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

	if _, err := fmt.Fprintf(tw, "[+] Timeout:\t%s\n", o.Timeout.String()); err != nil {
		return "", err
	}

	if s.globalopts.Verbose {
		if _, err := fmt.Fprintf(tw, "[+] Verbose:\ttrue\n"); err != nil {
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
