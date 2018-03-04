package libgobuster

import (
	"fmt"
	"os"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"
	uuid "github.com/satori/go.uuid"
)

// RedirectHandler ... A handler structure for HTTP 3xx responses
type RedirectHandler struct {
	Transport http.RoundTripper
	State     *State
}

// RedirectError ... A simple structure for an HTTP response status code
type RedirectError struct {
	StatusCode int
}

func (e *RedirectError) Error() string {
	return fmt.Sprintf("Redirect code: %d", e.StatusCode)
}


// RoundTrip ... handle an HTTP 3xx (Redirect)
func (rh *RedirectHandler) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	if rh.State.FollowRedirect {
		return rh.Transport.RoundTrip(req)
	}

	resp, err = rh.Transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	switch resp.StatusCode {
	case http.StatusMovedPermanently, http.StatusFound, http.StatusSeeOther,
		http.StatusNotModified, http.StatusUseProxy, http.StatusTemporaryRedirect:
		return nil, &RedirectError{StatusCode: resp.StatusCode}
	}

	return resp, err
}

// MakeRequest ... Make a request to the given URL
func MakeRequest(s *State, fullURL, cookie string) (*int, *int64) {

	request, err := http.NewRequest(s.Verb, fullURL, strings.NewReader(s.Body))

	if err != nil {
		os.Exit(1)
	}

	if cookie != "" {
		request.Header.Set("Cookie", cookie)
	}

	if s.ContentType != "" {
		request.Header.Set("Content-Type", s.ContentType)
	}

	if s.UserAgent != "" {
		request.Header.Set("User-Agent", s.UserAgent)
	}

	if s.Username != "" {
		request.SetBasicAuth(s.Username, s.Password)
	}

	if s.Headers != "" {
	    headers := strings.Split(s.Headers, "|")
	    for i := range headers {
			headerPair := strings.Split(headers[i], ": ")
			request.Header.Set(headerPair[0], headerPair[1])
	    }
	}

	resp, err := s.Client.Do(request)

	if err != nil {
		if ue, ok := err.(*url.Error); ok {

			if strings.HasPrefix(ue.Err.Error(), "x509") {
				fmt.Println("[-] Invalid certificate")
			}

			if re, ok := ue.Err.(*RedirectError); ok {
				return &re.StatusCode, nil
			}
		}
		return nil, nil
	}

	defer resp.Body.Close()

	var length *int64

	if s.IncludeLength {
		length = new(int64)
		if resp.ContentLength <= 0 {
			body, err := ioutil.ReadAll(resp.Body)
			if err == nil {
				*length = int64(utf8.RuneCountInString(string(body)))
			}
		} else {
			*length = resp.ContentLength
		}
	}

	return &resp.StatusCode, length
}

// GoGet ... Small helper to combine URL with URI then make a
// request to the generated location.
func GoGet(s *State, url, uri, cookie string) (*int, *int64) {
	return MakeRequest(s, url+uri, cookie)
}

// SetupDir ... Make an initial request with a random GUID to identify wildcard
// responses
func SetupDir(s *State) bool {
	guid := uuid.Must(uuid.NewV4())
	wildcardResp, _ := GoGet(s, s.URL, guid.String(), s.Cookies)

	if s.StatusCodes.Contains(*wildcardResp) {
		s.IsWildcard = true
		fmt.Println("[-] Wildcard response found:", fmt.Sprintf("%s%s", s.URL, guid), "=>", *wildcardResp)
		if !s.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard responses, specify the '-fw' switch.")
		}
		return s.WildcardForced
	}

	return true
}

// ProcessDirEntry ... Make a request to see if a URL is present
func ProcessDirEntry(s *State, word string, resultChan chan<- Result) {
	suffix := ""
	if s.UseSlash {
		suffix = "/"
	}

	// Try the DIR first
	dirResp, dirSize := GoGet(s, s.URL, word+suffix, s.Cookies)
	if dirResp != nil {
		resultChan <- Result{
			Entity: word + suffix,
			Status: *dirResp,
			Size:   dirSize,
		}
	}

	// Follow up with files using each ext.
	for ext := range s.Extensions {
		file := word + s.Extensions[ext]
		fileResp, fileSize := GoGet(s, s.URL, file, s.Cookies)

		if fileResp != nil {
			resultChan <- Result{
				Entity: file,
				Status: *fileResp,
				Size:   fileSize,
			}
		}
	}
}

// PrintDirResult ... Print various metadata about an HTTP response
func PrintDirResult(s *State, r *Result) {
	output := ""

	// Prefix if we're in verbose mode
	if s.Verbose {
		if s.StatusCodes.Contains(r.Status) {
			output = "Found : "
		} else {
			output = "Missed: "
		}
	}

	if s.StatusCodes.Contains(r.Status) || s.Verbose {
		if s.Expanded {
			output += s.URL
		} else {
			output += "/"
		}
		output += r.Entity

		if !s.NoStatus {
			output += fmt.Sprintf(" (Status: %d)", r.Status)
		}

		if r.Size != nil {
			output += fmt.Sprintf(" [Size: %d]", *r.Size)
		}
		output += "\n"

		fmt.Printf("%s", output)

		if s.OutputFile != nil {
			WriteToFile(output, s)
		}
	}
}
