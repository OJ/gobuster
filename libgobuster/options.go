package libgobuster

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
)

const (
	// ModeDir represents -m dir
	ModeDir = "dir"
	// ModeDNS represents -m dns
	ModeDNS = "dns"
)

// Options helds all options that can be passed to libgobuster
type Options struct {
	Extensions        string
	ExtensionsParsed  stringSet
	Mode              string
	Password          string
	StatusCodes       string
	StatusCodesParsed intSet
	Threads           int
	URLGroup 		  uint
	URL               string
	URLFile			  string
	UserAgent         string
	Username          string
	Wordlist          string
	Proxy             string
	Cookies           string
	Timeout           time.Duration
	Recursive		  bool
	FollowRedirect    bool
	IncludeLength     bool
	NoStatus          bool
	NoProgress        bool
	Expanded          bool
	Quiet             bool
	ShowIPs           bool
	ShowCNAME         bool
	InsecureSSL       bool
	WildcardForced    bool
	Verbose           bool
	UseSlash          bool
}

// NewOptions returns a new initialized Options object
func NewOptions() Options {
	return Options{
		StatusCodesParsed: newIntSet(),
		ExtensionsParsed:  newStringSet(),
	}
}

// Validate validates the given options
func (opt *Options) validate() *multierror.Error {
	var errorList *multierror.Error

	if strings.ToLower(opt.Mode) != ModeDir && strings.ToLower(opt.Mode) != ModeDNS {
		errorList = multierror.Append(errorList, fmt.Errorf("Mode (-m): Invalid value: %s", opt.Mode))
	}

	if opt.Threads < 0 {
		errorList = multierror.Append(errorList, fmt.Errorf("Threads (-t): Invalid value: %d", opt.Threads))
	}

	if opt.Wordlist == "" {
		errorList = multierror.Append(errorList, fmt.Errorf("WordList (-w): Must be specified (use `-w -` for stdin)"))
	} else if opt.Wordlist == "-" {
		// STDIN
	} else if _, err := os.Stat(opt.Wordlist); os.IsNotExist(err) {
		errorList = multierror.Append(errorList, fmt.Errorf("Wordlist (-w): File does not exist: %s", opt.Wordlist))
	}

	if opt.URLFile != "" {
		if _, err := os.Stat(opt.URLFile); os.IsNotExist(err) {
			errorList = multierror.Append(errorList, fmt.Errorf("URLFile (-file): File does not exist: %s", opt.URLFile))
		}
	} else if opt.URL == "" {
		errorList = multierror.Append(errorList, fmt.Errorf("Url/Domain (-u) or URLFile (-file): Must be specified"))
	}

	if opt.StatusCodes != "" {
		if err := opt.parseStatusCodes(); err != nil {
			errorList = multierror.Append(errorList, err)
		}
	}

	if opt.Extensions != "" {
		if err := opt.parseExtensions(); err != nil {
			errorList = multierror.Append(errorList, err)
		}
	}

	if opt.Mode == ModeDir && opt.URL != "" {
		if err := FixUrl(&opt.URL); err != nil {
			errorList = multierror.Append(errorList, err)
		}
	}

	return errorList
}

// ParseExtensions parses the extensions provided as a comma seperated list
func (opt *Options) parseExtensions() error {
	if opt.Extensions == "" {
		return fmt.Errorf("invalid extension string provided")
	}

	exts := strings.Split(opt.Extensions, ",")
	for _, e := range exts {
		e = strings.TrimSpace(e)
		// remove leading . from extensions
		opt.ExtensionsParsed.Add(strings.TrimPrefix(e, "."))
	}
	return nil
}

// ParseStatusCodes parses the status codes provided as a comma seperated list
func (opt *Options) parseStatusCodes() error {
	if opt.StatusCodes == "" {
		return fmt.Errorf("invalid status code string provided")
	}

	for _, c := range strings.Split(opt.StatusCodes, ",") {
		c = strings.TrimSpace(c)
		i, err := strconv.Atoi(c)
		if err != nil {
			return fmt.Errorf("invalid status code given: %s", c)
		}
		opt.StatusCodesParsed.Add(i)
	}
	return nil
}

func (opt *Options) ReadUrls() []string {
	urls := []string{}

	data, err := ioutil.ReadFile(opt.URLFile)

	if err != nil {
		return urls
	}

	text := string(data)

	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if err = FixUrl(&line); err != nil {
			continue
		}

		urls = append(urls, line)
	}

	return urls
}