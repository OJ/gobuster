package gobusterfuzz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/OJ/gobuster/v3/libgobuster"
)

// TestProcessWordPreservesPercentEncoding is a regression test for a bug where
// wordlist entries containing literal percent-encoding (e.g. "%2e%2e", used to
// bypass path-traversal filters such as CVE-2021-41773) were re-escaped to
// "%25..." before being sent, because the FUZZ keyword was substituted into
// url.URL.Path (the decoded path) instead of the escaped wire path.
func TestProcessWordPreservesPercentEncoding(t *testing.T) {
	tt := []struct {
		testName     string
		word         string
		expectedPath string
	}{
		{"single dot-dot traversal", "%2e%2e/opt/passwords", "/cgi-bin/%2e%2e/opt/passwords"},
		{"nested dot-dot traversal", "%2e%2e/%2e%2e/opt/passwords", "/cgi-bin/%2e%2e/%2e%2e/opt/passwords"},
		{"plain word", "plainword", "/cgi-bin/plainword"},
	}

	for _, x := range tt {
		t.Run(x.testName, func(t *testing.T) {
			var gotRequestURI string
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotRequestURI = r.URL.EscapedPath()
				w.WriteHeader(http.StatusOK)
			}))
			defer ts.Close()

			baseURL, err := url.Parse(ts.URL + "/cgi-bin/FUZZ")
			if err != nil {
				t.Fatalf("could not parse test server url: %v", err)
			}

			globalOpts := &libgobuster.Options{}
			opts := NewOptions()
			opts.URL = baseURL
			opts.Method = http.MethodGet
			opts.Timeout = ts.Client().Timeout

			logger := libgobuster.NewLogger(false)
			d, err := New(globalOpts, opts, logger)
			if err != nil {
				t.Fatalf("could not create GobusterFuzz: %v", err)
			}

			progress := libgobuster.NewProgress()
			if _, err := d.ProcessWord(context.Background(), x.word, progress); err != nil {
				t.Fatalf("ProcessWord returned an error: %v", err)
			}

			if gotRequestURI != x.expectedPath {
				t.Errorf("word %q: expected request path %q, got %q", x.word, x.expectedPath, gotRequestURI)
			}
		})
	}
}
