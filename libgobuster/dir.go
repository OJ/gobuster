package libgobuster

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	uuid "github.com/satori/go.uuid"
)

type RedirectHandler struct {
	Transport http.RoundTripper
	State     *State
}

type RedirectError struct {
	StatusCode int
}

func (e *RedirectError) Error() string {
	return fmt.Sprintf("Redirect code: %d", e.StatusCode)
}

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

// Make a request to the given URL.
func MakeRequest(s *State, fullUrl, cookie string) (*int, *int64, *[]string) {
	req, err := http.NewRequest("GET", fullUrl, nil)

	if err != nil {
		return nil, nil, nil
	}

	if cookie != "" {
		req.Header.Set("Cookie", cookie)
	}

	if s.UserAgent != "" {
		req.Header.Set("User-Agent", s.UserAgent)
	}

	if s.Username != "" {
		req.SetBasicAuth(s.Username, s.Password)
	}

	resp, err := s.Client.Do(req)

	if err != nil {
		if ue, ok := err.(*url.Error); ok {

			if strings.HasPrefix(ue.Err.Error(), "x509") {
				fmt.Println("[-] Invalid certificate")
			}

			if re, ok := ue.Err.(*RedirectError); ok {
				return &re.StatusCode, nil, nil
			}
		}
		return nil, nil, nil
	}

	defer resp.Body.Close()

	var statusCode = &resp.StatusCode
	var length *int64 = nil

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
	} else {
		// DO NOT REMOVE!
		// absolutely needed so golang will reuse connections!
		_, err = io.Copy(ioutil.Discard, resp.Body)
		if err != nil {
			return nil, nil
		}
	}

	options := new([]string)

	if s.IncludeOptions {
		// Make an OPTIONS request to the given URL.
		req.Method = "OPTIONS"
		resp, err = s.Client.Do(req)
		if err == nil {
			*options = resp.Header["Allow"]
		}
	}

	return statusCode, length, options
}

// Small helper to combine URL with URI then make a
// request to the generated location.
func GoGet(s *State, url, uri, cookie string) (*int, *int64, *[]string) {
	return MakeRequest(s, url+uri, cookie)
}

func SetupDir(s *State) bool {
	guid := uuid.Must(uuid.NewV4())
	wildcardResp, _, _ := GoGet(s, s.Url, fmt.Sprintf("%s", guid), s.Cookies)

	if s.StatusCodes.Contains(*wildcardResp) {
		s.IsWildcard = true
		fmt.Println("[-] Wildcard response found:", fmt.Sprintf("%s%s", s.Url, guid), "=>", *wildcardResp)
		if !s.WildcardForced {
			fmt.Println("[-] To force processing of Wildcard responses, specify the '-fw' switch.")
		}
		return s.WildcardForced
	}

	return true
}

func ProcessDirEntry(s *State, word string, resultChan chan<- Result) {
	suffix := ""
	if s.UseSlash {
		suffix = "/"
	}

	// Try the DIR first
	dirResp, dirSize, dirOptions := GoGet(s, s.Url, word+suffix, s.Cookies)
	if dirResp != nil {
		resultChan <- Result{
			Entity:  word + suffix,
			Status:  *dirResp,
			Size:    dirSize,
			Options: dirOptions,
		}
	}

	// Follow up with files using each ext.
	for ext := range s.Extensions {
		file := word + s.Extensions[ext]
		fileResp, fileSize, fileOptions := GoGet(s, s.Url, file, s.Cookies)

		if fileResp != nil {
			resultChan <- Result{
				Entity:  file,
				Status:  *fileResp,
				Size:    fileSize,
				Options: fileOptions,
			}
		}
	}
}

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
			output += s.Url
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

		if r.Options != nil && len(*r.Options) > 0 {
			output += fmt.Sprintf(" [Methods: %s]", strings.Join(*r.Options, ","))
		}
		output += "\n"

		fmt.Printf(output)

		if s.OutputFile != nil {
			WriteToFile(output, s)
		}
	}
}
