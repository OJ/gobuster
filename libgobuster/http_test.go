package libgobuster

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func httpServer(t *testing.T, content string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, content)
	}))
	return ts
}

func TestMakeRequest(t *testing.T) {
	h := httpServer(t, "test")
	defer h.Close()
	o := NewOptions()
	c, err := newHTTPClient(context.Background(), o)
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	a, b, err := c.makeRequest(h.URL, "")
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
