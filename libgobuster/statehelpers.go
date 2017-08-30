package libgobuster

//----------------------------------------------------
// Gobuster -- by OJ Reeves
//
// A crap attempt at building something that resembles
// dirbuster or dirb using Go. The goal was to build
// a tool that would help learn Go and to actually do
// something useful. The idea of having this compile
// to native code is also appealing.
//
// Run: gobuster -h
//
// Please see THANKS file for contributors.
// Please see LICENSE file for license details.
//
//----------------------------------------------------

import (
	"crypto/tls"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"golang.org/x/crypto/ssh/terminal"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

func InitState() State {
	return State{
		StatusCodes: IntSet{Set: map[int]bool{}},
		WildcardIps: StringSet{Set: map[string]bool{}},
		IsWildcard:  false,
		StdIn:       false,
	}
}

func ValidateState(
	s *State,
	extensions string,
	codes string,
	proxy string) *multierror.Error {

	var errorList *multierror.Error

	switch strings.ToLower(s.Mode) {
	case "dir":
		s.Printer = PrintDirResult
		s.Processor = ProcessDirEntry
		s.Setup = SetupDir
	case "dns":
		s.Printer = PrintDnsResult
		s.Processor = ProcessDnsEntry
		s.Setup = SetupDns
	default:
		errorList = multierror.Append(errorList, fmt.Errorf("[!] Mode (-m): Invalid value: %s", s.Mode))
	}

	if s.Threads < 0 {
		errorList = multierror.Append(errorList, fmt.Errorf("[!] Threads (-t): Invalid value: %s", s.Threads))
	}

	stdin, err := os.Stdin.Stat()
	if err != nil {
		fmt.Println("[!] Unable to stat stdin, falling back to wordlist file.")
	} else if (stdin.Mode()&os.ModeCharDevice) == 0 && stdin.Size() > 0 {
		s.StdIn = true
	}

	if !s.StdIn {
		if s.Wordlist == "" {
			errorList = multierror.Append(errorList, fmt.Errorf("[!] WordList (-w): Must be specified"))
		} else if _, err := os.Stat(s.Wordlist); os.IsNotExist(err) {
			errorList = multierror.Append(errorList, fmt.Errorf("[!] Wordlist (-w): File does not exist: %s", s.Wordlist))
		}
	} else if s.Wordlist != "" {
		errorList = multierror.Append(errorList, fmt.Errorf("[!] Wordlist (-w) specified with pipe from stdin. Can't have both!"))
	}

	if s.Url == "" {
		errorList = multierror.Append(errorList, fmt.Errorf("[!] Url/Domain (-u): Must be specified"))
	}

	if s.Mode == "dir" {
		if err := ValidateDirModeState(s, extensions, codes, proxy, errorList); err.ErrorOrNil() != nil {
			errorList = multierror.Append(errorList, err)
		}
	}

	return errorList
}

func ValidateDirModeState(
	s *State,
	extensions string,
	codes string,
	proxy string,
	previousErrors *multierror.Error) *multierror.Error {

	// If we had previous errors, copy them into the current errorList.
	// This is an easier to understand solution compared to double pointer black magick
	var errorList *multierror.Error
	if previousErrors != nil {
		errorList = multierror.Append(errorList, previousErrors)
	}

	if strings.HasSuffix(s.Url, "/") == false {
		s.Url = s.Url + "/"
	}

	if strings.HasPrefix(s.Url, "http") == false {
		// check to see if a port was specified
		re := regexp.MustCompile(`^[^/]+:(\d+)`)
		match := re.FindStringSubmatch(s.Url)

		if len(match) < 2 {
			// no port, default to http on 80
			s.Url = "http://" + s.Url
		} else {
			port, err := strconv.Atoi(match[1])
			if err != nil || (port != 80 && port != 443) {
				errorList = multierror.Append(errorList, fmt.Errorf("[!] Url/Domain (-u): Scheme not specified."))
			} else if port == 80 {
				s.Url = "http://" + s.Url
			} else {
				s.Url = "https://" + s.Url
			}
		}
	}

	// extensions are comma separated
	if extensions != "" {
		s.Extensions = strings.Split(extensions, ",")
		for i := range s.Extensions {
			if s.Extensions[i][0] != '.' {
				s.Extensions[i] = "." + s.Extensions[i]
			}
		}
	}

	// status codes are comma separated
	if codes != "" {
		for _, c := range strings.Split(codes, ",") {
			i, err := strconv.Atoi(c)
			if err != nil {
				errorList = multierror.Append(errorList, fmt.Errorf("[!] Invalid status code given: %s", c))
			} else {
				s.StatusCodes.Add(i)
			}
		}
	}

	// prompt for password if needed
	if errorList.ErrorOrNil() == nil && s.Username != "" && s.Password == "" {
		fmt.Printf("[?] Auth Password: ")
		passBytes, err := terminal.ReadPassword(int(syscall.Stdin))

		// print a newline to simulate the newline that was entered
		// this means that formatting/printing after doesn't look bad.
		fmt.Println("")

		if err == nil {
			s.Password = string(passBytes)
		} else {
			errorList = multierror.Append(errorList, fmt.Errorf("[!] Auth username given but reading of password failed"))
		}
	}

	if errorList.ErrorOrNil() == nil {
		var proxyUrlFunc func(*http.Request) (*url.URL, error)
		proxyUrlFunc = http.ProxyFromEnvironment

		if proxy != "" {
			proxyUrl, err := url.Parse(proxy)
			if err != nil {
				errorList = multierror.Append(errorList, fmt.Errorf("[!] Proxy URL is invalid"))
				panic("[!] Proxy URL is invalid") // TODO: Does this need to be a panic? Could be a standard error?
			}
			s.ProxyUrl = proxyUrl
			proxyUrlFunc = http.ProxyURL(s.ProxyUrl)
		}

		s.Client = &http.Client{
			Transport: &RedirectHandler{
				State: s,
				Transport: &http.Transport{
					Proxy: proxyUrlFunc,
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: s.InsecureSSL,
					},
				},
			}}

		code, _ := GoGet(s, s.Url, "", s.Cookies)
		if code == nil {
			errorList = multierror.Append(errorList, fmt.Errorf("[-] Unable to connect: %s", s.Url))
		}
	}

	return errorList
}
