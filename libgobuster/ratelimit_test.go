package libgobuster

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// rateLimitServer returns 429 for the first `limited` requests and 200 afterwards
func rateLimitServer(t *testing.T, limited int64, wantBody string) (*httptest.Server, *atomic.Int64) {
	t.Helper()
	var count atomic.Int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := count.Add(1)
		if wantBody != "" {
			b, err := io.ReadAll(r.Body)
			if err != nil || string(b) != wantBody {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
		if n <= limited {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)
	return ts, &count
}

func newRateLimitClient(t *testing.T, sleep time.Duration) *HTTPClient {
	t.Helper()
	o := HTTPOptions{RateLimitSleep: sleep}
	c, err := NewHTTPClient(&o, NewLogger(false))
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	return c
}

func mustParseURL(t *testing.T, s string) url.URL {
	t.Helper()
	u, err := url.Parse(s)
	if err != nil {
		t.Fatalf("could not parse URL %v: %v", s, err)
	}
	return *u
}

func TestRateLimitDisabled(t *testing.T) {
	t.Parallel()
	ts, count := rateLimitServer(t, 1, "")
	c := newRateLimitClient(t, 0)

	status, _, _, _, err := c.Request(t.Context(), mustParseURL(t, ts.URL), RequestOptions{})
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	if status != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", status)
	}
	if count.Load() != 1 {
		t.Fatalf("expected 1 request, got %d", count.Load())
	}
}

func TestRateLimitPersistent429(t *testing.T) {
	t.Parallel()
	body := "some=body"
	ts, count := rateLimitServer(t, 4, body)
	c := newRateLimitClient(t, 10*time.Millisecond)

	start := time.Now()
	status, _, _, _, err := c.Request(t.Context(), mustParseURL(t, ts.URL), RequestOptions{Body: strings.NewReader(body)})
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if count.Load() != 5 {
		t.Fatalf("expected 5 requests, got %d", count.Load())
	}
	if elapsed := time.Since(start); elapsed < 40*time.Millisecond {
		t.Fatalf("expected to sleep at least 40ms, slept %s", elapsed)
	}
}

func TestRateLimitPausesAllWorkers(t *testing.T) {
	t.Parallel()
	const workers = 10
	const limited = 3
	sleep := 100 * time.Millisecond
	ts, count := rateLimitServer(t, limited, "")
	c := newRateLimitClient(t, sleep)
	u := mustParseURL(t, ts.URL)

	var wg sync.WaitGroup
	statuses := make([]int, workers)
	errs := make([]error, workers)
	start := make(chan struct{})
	for i := range workers {
		wg.Go(func() {
			<-start
			statuses[i], _, _, _, errs[i] = c.Request(t.Context(), u, RequestOptions{})
		})
	}
	close(start)
	wg.Wait()

	for i := range workers {
		if errs[i] != nil {
			t.Fatalf("worker %d got error: %v", i, errs[i])
		}
		if statuses[i] != http.StatusOK {
			t.Fatalf("worker %d expected status 200, got %d", i, statuses[i])
		}
	}

	// every worker sends one request, the limited ones are retried once each
	// and only the leader probes during the pause
	if got := count.Load(); got != workers+limited {
		t.Fatalf("expected %d requests, got %d", workers+limited, got)
	}
}

func TestRateLimitContextCanceled(t *testing.T) {
	t.Parallel()
	ts, _ := rateLimitServer(t, 1000, "")
	c := newRateLimitClient(t, time.Hour)

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		status, _, _, _, err := c.Request(ctx, mustParseURL(t, ts.URL), RequestOptions{})
		if err != nil {
			t.Errorf("Got Error: %v", err)
		}
		if status != 0 {
			t.Errorf("expected status 0 on canceled context, got %d", status)
		}
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("request did not return after the context was canceled")
	}

	// the gate must be open again after the leader gave up
	if c.rateLimit.paused != nil {
		t.Fatal("rate limit gate is still paused")
	}
}
