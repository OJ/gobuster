package libgobuster

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	multierror "github.com/hashicorp/go-multierror"
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
	ExtensionsParsed  []string
	Mode              string
	Password          string
	StatusCodes       string
	StatusCodesParsed intSet
	Threads           int
	URL               string
	UserAgent         string
	Username          string
	Wordlist          string
	Proxy             string
	Cookies           string
	Timeout           time.Duration
	FollowRedirect    bool
	IncludeLength     bool
	NoStatus          bool
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
func NewOptions() *Options {
	return &Options{
		StatusCodesParsed: intSet{Set: map[int]bool{}},
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
		errorList = multierror.Append(errorList, fmt.Errorf("WordList (-w): Must be specified"))
	} else if opt.Wordlist == "-" {
		// STDIN
	} else if _, err := os.Stat(opt.Wordlist); os.IsNotExist(err) {
		errorList = multierror.Append(errorList, fmt.Errorf("Wordlist (-w): File does not exist: %s", opt.Wordlist))
	}

	if opt.URL == "" {
		errorList = multierror.Append(errorList, fmt.Errorf("Url/Domain (-u): Must be specified"))
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

	if !strings.HasSuffix(opt.URL, "/") {
		opt.URL = fmt.Sprintf("%s/", opt.URL)
	}

	if opt.Mode == ModeDir {
		if err := opt.validateDirMode(); err != nil {
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
		// remove leading . from extensions
		opt.ExtensionsParsed = append(opt.ExtensionsParsed, strings.TrimPrefix(e, "."))
	}
	return nil
}

// ParseStatusCodes parses the status codes provided as a comma seperated list
func (opt *Options) parseStatusCodes() error {
	if opt.StatusCodes == "" {
		return fmt.Errorf("invalid status code string provided")
	}

	for _, c := range strings.Split(opt.StatusCodes, ",") {
		i, err := strconv.Atoi(c)
		if err != nil {
			return fmt.Errorf("invalid status code given: %s", c)
		}
		opt.StatusCodesParsed.Add(i)
	}
	return nil
}

func (opt *Options) validateDirMode() error {
	// bail out if we are not in dir mode
	if opt.Mode != ModeDir {
		return nil
	}
	if !strings.HasPrefix(opt.URL, "http") {
		// check to see if a port was specified
		re := regexp.MustCompile(`^[^/]+:(\d+)`)
		match := re.FindStringSubmatch(opt.URL)

		if len(match) < 2 {
			// no port, default to http on 80
			opt.URL = fmt.Sprintf("http://%s", opt.URL)
		} else {
			port, err := strconv.Atoi(match[1])
			if err != nil || (port != 80 && port != 443) {
				return fmt.Errorf("url scheme not specified")
			} else if port == 80 {
				opt.URL = fmt.Sprintf("http://%s", opt.URL)
			} else {
				opt.URL = fmt.Sprintf("https://%s", opt.URL)
			}
		}
	}

	if opt.Username != "" && opt.Password == "" {
		return fmt.Errorf("username was provided but password is missing")
	}

	return nil
}
