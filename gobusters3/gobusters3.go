package gobusters3

import (
	"bufio"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
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

// GetRequest issues a GET request to the target and returns
// the status code, length and an error
func (s *GobusterS3) get(url string) (*int, *[]byte, error) {
	return s.http.GetWithBody(url, "", s.options.Cookies)
}

// NewGobusterS3 creates a new initialized GobusterS3
func NewGobusterS3(cont context.Context, globalopts *libgobuster.Options, opts *OptionsS3) (*GobusterS3, error) {
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

	httpOpts := libgobuster.HTTPOptions{
		Proxy: opts.Proxy,
		// needed so we can list bucket contents
		FollowRedirect: true,
		Timeout:        opts.Timeout,
		UserAgent:      opts.UserAgent,
	}

	h, err := libgobuster.NewHTTPClient(cont, &httpOpts)
	if err != nil {
		return nil, err
	}
	g.http = h
	g.bucketRegex = regexp.MustCompile(`^[a-z0-9\-\.]{3,63}$`)

	return &g, nil
}

// Name should return the name of the plugin
func (s *GobusterS3) Name() string {
	return "S3 bucket enumeration"
}

// PreRun is the pre run implementation of GobusterS3
func (s *GobusterS3) PreRun() error {
	return nil
}

// Run is the process implementation of GobusterS3
func (s *GobusterS3) Run(word string) ([]libgobuster.Result, error) {
	var ret []libgobuster.Result

	// only check for valid bucket names
	if !s.isValidBucketName(word) {
		return ret, nil
	}

	// this url will return a 307 with the URL including the region if found
	bucketURL := fmt.Sprintf("http://%s.s3.amazonaws.com/?max-keys=%d", word, s.options.MaxFilesToList)
	status, body, err := s.get(bucketURL)
	if err != nil {
		return nil, err
	}

	if status == nil || body == nil {
		return ret, nil
	}

	// looks like 404 and 400 are the only negative status codes
	found := false
	switch *status {
	case 400:
	case 404:
		found = false
	case 200:
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
		return ret, nil
	}

	extraStr := ""
	if s.globalopts.Verbose {
		// get status
		if bytes.Contains(*body, []byte("<Error>")) {
			awsError := AWSError{}
			err := xml.Unmarshal(*body, &awsError)
			if err != nil {
				return nil, fmt.Errorf("could not parse error xml: %v", err)
			}
			// https://docs.aws.amazon.com/AmazonS3/latest/API/ErrorResponses.html#ErrorCodeList
			extraStr = fmt.Sprintf("Error: %s (%s)", awsError.Message, awsError.Code)
		} else if bytes.Contains(*body, []byte("<ListBucketResult ")) {
			// bucket listing enabled
			awsListing := AWSListing{}
			err := xml.Unmarshal(*body, &awsListing)
			if err != nil {
				return nil, fmt.Errorf("could not parse result xml: %v", err)
			}
			extraStr = "Bucket Listing enabled: "
			for _, x := range awsListing.Contents {
				extraStr += fmt.Sprintf("%s (%db), ", x.Key, x.Size)
			}
			extraStr = strings.TrimRight(extraStr, ", ")
		}
	}

	ret = append(ret, libgobuster.Result{
		Entity: word,
		Status: libgobuster.StatusFound,
		Extra:  extraStr,
	})

	return ret, nil
}

// ResultToString is the to string implementation of GobusterS3
func (s *GobusterS3) ResultToString(r *libgobuster.Result) (*string, error) {
	buf := &bytes.Buffer{}

	if s.options.Expanded {
		if _, err := fmt.Fprintf(buf, "http://%s.s3.amazonaws.com/", r.Entity); err != nil {
			return nil, err
		}
	} else {
		if _, err := fmt.Fprintf(buf, "%s", r.Entity); err != nil {
			return nil, err
		}
	}

	if r.Extra != "" {
		if _, err := fmt.Fprintf(buf, " [%s]", r.Extra); err != nil {
			return nil, err
		}
	}

	if _, err := fmt.Fprintf(buf, "\n"); err != nil {
		return nil, err
	}

	str := buf.String()
	return &str, nil
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

	if o.FollowRedirect {
		if _, err := fmt.Fprintf(tw, "[+] Follow Redir:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if s.globalopts.Verbose {
		if _, err := fmt.Fprintf(tw, "[+] Verbose:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(tw, "[+] Timeout:\t%s\n", o.Timeout.String()); err != nil {
		return "", err
	}

	if o.Expanded {
		if _, err := fmt.Fprintf(tw, "[+] Expanded:\ttrue\n"); err != nil {
			return "", err
		}
	}

	if _, err := fmt.Fprintf(tw, "[+] Maximum files to list:\t%d\n", o.MaxFilesToList); err != nil {
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
