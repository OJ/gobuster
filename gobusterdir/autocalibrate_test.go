package gobusterdir

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/OJ/gobuster/v3/libgobuster"
)

func TestAutoCalibrate(t *testing.T) {
	// Create a test server that returns 301 for a specific path (the wildcard)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 1 {
			// This is our "wildcard" or any path longer than /
			w.Header().Set("Location", "http://localhost:3000"+r.URL.Path)
			w.WriteHeader(http.StatusMovedPermanently)
			fmt.Fprint(w, "redirecting...") // Length will be 14
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer ts.Close()

	parsedURL, _ := url.Parse(ts.URL)
	globalOpts := libgobuster.Options{Threads: 1}
	opts := NewOptions()
	opts.URL = parsedURL
	opts.AutoCalibrate = true
	opts.StatusCodesBlacklistParsed.Add(404) // Default

	log := libgobuster.NewLogger(false)
	d, err := New(&globalOpts, opts, log)
	if err != nil {
		t.Fatalf("failed to create gobusterdir: %v", err)
	}

	ctx := context.Background()
	pr := libgobuster.NewProgress()
	go func() {
		for range pr.MessageChan {
		}
	}()

	err = d.PreRun(ctx, pr)
	if err != nil {
		t.Fatalf("PreRun failed: %v", err)
	}

	if !opts.ExcludeLengthParsed.Contains(14) {
		t.Errorf("expected length 14 to be excluded, but it wasn't. Excluded lengths: %v", opts.ExcludeLengthParsed.Stringify())
	}

	// Now check if ProcessWord correctly ignores a word that results in the same length
	res, err := d.ProcessWord(ctx, "anyword", pr)
	if err != nil {
		t.Fatalf("ProcessWord failed: %v", err)
	}

	if res != nil {
		t.Errorf("expected result to be nil (excluded by length), but got %v", res)
	}
}
