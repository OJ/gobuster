package gobusterdir

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/OJ/gobuster/v3/libgobuster"
)

func newDirForRateLimit(t *testing.T, target string, stopOn429 bool) *GobusterDir {
	t.Helper()

	u, err := url.Parse(target)
	if err != nil {
		t.Fatalf("could not parse url: %v", err)
	}

	globalOpts := &libgobuster.Options{StopOnRateLimit: stopOn429}
	opts := NewOptions()
	opts.URL = u
	opts.Timeout = 10 * time.Second
	// treat 429 as a matching status so, without the flag, it is reported normally
	opts.StatusCodesParsed.Add(http.StatusTooManyRequests)

	d, err := New(globalOpts, opts, libgobuster.NewLogger(false))
	if err != nil {
		t.Fatalf("could not create gobusterdir: %v", err)
	}
	return d
}

func TestProcessWordStopsOn429(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	d := newDirForRateLimit(t, ts.URL, true)

	_, err := d.ProcessWord(context.Background(), "admin", libgobuster.NewProgress())
	if !errors.Is(err, libgobuster.ErrRateLimited) {
		t.Fatalf("expected ErrRateLimited, got %v", err)
	}
}

func TestProcessWord429WithoutFlag(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer ts.Close()

	d := newDirForRateLimit(t, ts.URL, false)

	res, err := d.ProcessWord(context.Background(), "admin", libgobuster.NewProgress())
	if err != nil {
		t.Fatalf("did not expect an error, got %v", err)
	}
	if res == nil {
		t.Fatal("expected a result for the 429 response when the flag is disabled")
	}
}
