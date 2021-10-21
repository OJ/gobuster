package gobustergcs

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
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

// NewGobusterGCS creates a new initialized GobusterGCS
func NewGobusterGCS(globalopts *libgobuster.Options, opts *OptionsGCS) (*GobusterGCS, error) {
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
	// https://cloud.google.com/storage/docs/naming-buckets
	g.bucketRegex = regexp.MustCompile(`^[a-z0-9][a-z0-9\-_.]{1,61}[a-z0-9](\.[a-z0-9][a-z0-9\-_.]{1,61}[a-z0-9])*$`)

	return &g, nil
}

// Name should return the name of the plugin
func (s *GobusterGCS) Name() string {
	return "GCS bucket enumeration"
}

// RequestsPerRun returns the number of requests this plugin makes per single wordlist item
func (s *GobusterGCS) RequestsPerRun() int {
	return 1
}

// PreRun is the pre run implementation of GobusterS3
func (s *GobusterGCS) PreRun(ctx context.Context) error {
	return nil
}

// Run is the process implementation of GobusterS3
func (s *GobusterGCS) Run(ctx context.Context, word string, resChannel chan<- libgobuster.Result) error {
	// only check for valid bucket names
	if !s.isValidBucketName(word) {
		return nil
	}

	bucketURL := fmt.Sprintf("https://storage.googleapis.com/storage/v1/b/%s/o?maxResults=%d", word, s.options.MaxFilesToList)
	status, _, _, body, err := s.http.Request(ctx, bucketURL, libgobuster.RequestOptions{ReturnBody: true})
	if err != nil {
		return err
	}

	if status == nil || body == nil {
		return nil
	}

	// looks like 401, 403, and 404 are the only negative status codes
	found := false
	switch *status {
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
	if s.globalopts.Verbose {
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

	resChannel <- Result{
		Found:      found,
		BucketName: word,
		Status:     extraStr,
	}

	return nil
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
