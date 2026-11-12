package gobusterdir

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/OJ/gobuster/v3/libgobuster"
)

// This test makes exactly four loopback requests per case: two from PreRun,
// then two direct ProcessWord calls, even when PreRun rejects the scan.
// Calling ProcessWord independently verifies that both methods agree about
// whether the same response qualifies as a result.
func TestPreRunWildcardFilters(t *testing.T) {
	const body = "fallback"
	regexCases := []struct {
		name    string
		pattern string
		invert  bool
		accept  bool
	}{
		{name: "no_regex", accept: true},
		{name: "matching_regex", pattern: "^fallback$", accept: true},
		{name: "nonmatching_regex", pattern: "^real$"},
		{name: "inverted_matching_regex", pattern: "^fallback$", invert: true},
		{name: "inverted_nonmatching_regex", pattern: "^real$", invert: true, accept: true},
	}
	scenarios := []struct {
		name          string
		status        int
		excludeLength bool
	}{
		{name: "catch_all_200", status: http.StatusOK},
		{name: "excluded_length", status: http.StatusOK, excludeLength: true},
		{name: "rejected_status", status: http.StatusNotFound},
	}
	for _, mode := range []string{"whitelist_200", "blacklist_404"} {
		for _, regexCase := range regexCases {
			for _, scenario := range scenarios {
				t.Run(mode+"/"+regexCase.name+"/"+scenario.name, func(t *testing.T) {
					var requests atomic.Int32
					server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
						requests.Add(1)
						w.WriteHeader(scenario.status)
						_, _ = fmt.Fprint(w, body)
					}))
					defer server.Close()
					opts := NewOptions()
					var err error
					opts.URL, err = url.Parse(server.URL)
					if err != nil || opts.URL.Hostname() != "127.0.0.1" {
						t.Fatalf("expected loopback test server: %s, %v", server.URL, err)
					}
					opts.Timeout = time.Second
					if mode == "blacklist_404" {
						opts.StatusCodesBlacklistParsed.Add(http.StatusNotFound)
					} else {
						opts.StatusCodesParsed.Add(http.StatusOK)
					}
					if regexCase.pattern != "" {
						opts.Regex = regexp.MustCompile(regexCase.pattern)
					}
					opts.RegexInvert = regexCase.invert
					if scenario.excludeLength {
						opts.ExcludeLengthParsed.Add(len(body))
					}
					d, err := New(&libgobuster.Options{Threads: 1}, opts, libgobuster.NewLogger(false))
					if err != nil {
						t.Fatal(err)
					}
					ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
					defer cancel()
					progress := libgobuster.NewProgress()
					preErr := d.PreRun(ctx, progress)
					var wildcard *WildcardError
					gotWildcard := errors.As(preErr, &wildcard)
					wantAccept := scenario.status == http.StatusOK && !scenario.excludeLength && regexCase.accept
					if preErr != nil && !gotWildcard {
						t.Fatalf("unexpected precheck error: %v", preErr)
					}
					if gotWildcard != wantAccept {
						t.Errorf("PreRun wildcard=%v, want %v (error=%v)", gotWildcard, wantAccept, preErr)
					}
					hits := 0
					for _, word := range []string{"alpha", "bravo"} {
						result, err := d.ProcessWord(ctx, word, progress)
						if err != nil {
							t.Fatal(err)
						}
						if (result != nil) != wantAccept {
							t.Errorf("ProcessWord(%q) result=%v, want accepted=%v", word, result, wantAccept)
						}
						if result != nil {
							hits++
							r, ok := result.(Result)
							if !ok || strings.TrimSpace(r.Path) != word || r.StatusCode != scenario.status || r.Size != int64(len(body)) {
								t.Errorf("unexpected result: %#v", result)
							}
						}
					}
					if requests.Load() != 4 {
						t.Errorf("requests=%d, want exactly 4", requests.Load())
					}
					t.Logf("PreRun wildcard=%v; scan eligible=%v; direct ProcessWord hits=%d/2; requests=%d", gotWildcard, preErr == nil, hits, requests.Load())
				})
			}
		}
	}
}
