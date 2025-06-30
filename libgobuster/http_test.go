package libgobuster

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func httpServerB(b *testing.B, content string) *httptest.Server {
	b.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprint(w, content); err != nil {
			b.Fatalf("%v", err)
		}
	}))
	return ts
}

func httpServerT(t *testing.T, content string) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprint(w, content); err != nil {
			t.Fatalf("%v", err)
		}
	}))
	return ts
}

func randomString(length int) (string, error) {
	letter := []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	letterLen := len(letter)

	b := make([]byte, length)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(letterLen)))
		if err != nil {
			return "", err
		}
		b[i] = letter[n.Int64()]
	}
	return string(b), nil
}

func TestRequest(t *testing.T) {
	t.Parallel()
	ret, err := randomString(100)
	if err != nil {
		t.Fatal(err)
	}
	h := httpServerT(t, ret)
	defer h.Close()
	var o HTTPOptions
	log := NewLogger(false)
	c, err := NewHTTPClient(&o, log)
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	u, err := url.Parse(h.URL)
	if err != nil {
		t.Fatalf("could not parse URL %v: %v", h.URL, err)
	}
	status, length, _, body, err := c.Request(t.Context(), *u, RequestOptions{ReturnBody: true})
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	if status != 200 {
		t.Fatalf("Invalid status returned: %d", status)
	}
	if length != int64(len(ret)) {
		t.Fatalf("Invalid length returned: %d", length)
	}
	if body == nil || !bytes.Equal(body, []byte(ret)) {
		t.Fatalf("Invalid body returned: %d", body)
	}
}

func BenchmarkRequestWithoutBody(b *testing.B) {
	r, err := randomString(10000)
	if err != nil {
		b.Fatal(err)
	}
	h := httpServerB(b, r)
	defer h.Close()
	var o HTTPOptions
	log := NewLogger(false)
	c, err := NewHTTPClient(&o, log)
	if err != nil {
		b.Fatalf("Got Error: %v", err)
	}
	u, err := url.Parse(h.URL)
	if err != nil {
		b.Fatalf("could not parse URL %v: %v", h.URL, err)
	}
	for b.Loop() {
		_, _, _, _, err := c.Request(b.Context(), *u, RequestOptions{ReturnBody: false})
		if err != nil {
			b.Fatalf("Got Error: %v", err)
		}
	}
}

func BenchmarkRequestWitBody(b *testing.B) {
	r, err := randomString(10000)
	if err != nil {
		b.Fatal(err)
	}
	h := httpServerB(b, r)
	defer h.Close()
	var o HTTPOptions
	log := NewLogger(false)
	c, err := NewHTTPClient(&o, log)
	if err != nil {
		b.Fatalf("Got Error: %v", err)
	}
	u, err := url.Parse(h.URL)
	if err != nil {
		b.Fatalf("could not parse URL %v: %v", h.URL, err)
	}
	for b.Loop() {
		_, _, _, _, err := c.Request(b.Context(), *u, RequestOptions{ReturnBody: true})
		if err != nil {
			b.Fatalf("Got Error: %v", err)
		}
	}
}

func BenchmarkNewHTTPClient(b *testing.B) {
	r, err := randomString(500)
	if err != nil {
		b.Fatal(err)
	}
	h := httpServerB(b, r)
	defer h.Close()
	var o HTTPOptions
	log := NewLogger(false)
	for b.Loop() {
		_, err := NewHTTPClient(&o, log)
		if err != nil {
			b.Fatalf("Got Error: %v", err)
		}
	}
}
