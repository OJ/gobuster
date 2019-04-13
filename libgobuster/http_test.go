package libgobuster

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func httpServerB(b *testing.B, content string) *httptest.Server {
	b.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, content)
	}))
	return ts
}

func httpServerT(t *testing.T, content string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, content)
	}))
	return ts
}

func TestGet(t *testing.T) {
	h := httpServerT(t, "test")
	defer h.Close()
	var o HTTPOptions
	c, err := NewHTTPClient(context.Background(), &o)
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	a, b, err := c.Get(h.URL, "", "")
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	if *a != 200 {
		t.Fatalf("Invalid status returned: %d", a)
	}
	if b != nil && *b != int64(len("test")) {
		t.Fatalf("Invalid length returned: %d", b)
	}
}

func BenchmarkGet(b *testing.B) {
	h := httpServerB(b, "test")
	defer h.Close()
	var o HTTPOptions
	c, err := NewHTTPClient(context.Background(), &o)
	if err != nil {
		b.Fatalf("Got Error: %v", err)
	}
	for x := 0; x < b.N; x++ {
		_, _, err := c.Get(h.URL, "", "")
		if err != nil {
			b.Fatalf("Got Error: %v", err)
		}
	}
}

func BenchmarkNewHTTPClient(b *testing.B) {
	h := httpServerB(b, "test")
	defer h.Close()
	var o HTTPOptions
	for x := 0; x < b.N; x++ {
		_, err := NewHTTPClient(context.Background(), &o)
		if err != nil {
			b.Fatalf("Got Error: %v", err)
		}
	}
}
