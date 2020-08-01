package libgobuster

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
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

func randomString(length int) string {
	var letter = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	letterLen := len(letter)

	b := make([]byte, length)
	for i := range b {
		b[i] = letter[rand.Intn(letterLen)]
	}
	return string(b)
}

func TestRequest(t *testing.T) {
	ret := randomString(100)
	h := httpServerT(t, ret)
	defer h.Close()
	var o HTTPOptions
	c, err := NewHTTPClient(context.Background(), &o)
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	status, length, _, body, err := c.Request(h.URL, RequestOptions{ReturnBody: true})
	if err != nil {
		t.Fatalf("Got Error: %v", err)
	}
	if *status != 200 {
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
	h := httpServerB(b, randomString(10000))
	defer h.Close()
	var o HTTPOptions
	c, err := NewHTTPClient(context.Background(), &o)
	if err != nil {
		b.Fatalf("Got Error: %v", err)
	}
	for x := 0; x < b.N; x++ {
		_, _, _, _, err := c.Request(h.URL, RequestOptions{ReturnBody: false})
		if err != nil {
			b.Fatalf("Got Error: %v", err)
		}
	}
}

func BenchmarkRequestWitBody(b *testing.B) {
	h := httpServerB(b, randomString(10000))
	defer h.Close()
	var o HTTPOptions
	c, err := NewHTTPClient(context.Background(), &o)
	if err != nil {
		b.Fatalf("Got Error: %v", err)
	}
	for x := 0; x < b.N; x++ {
		_, _, _, _, err := c.Request(h.URL, RequestOptions{ReturnBody: true})
		if err != nil {
			b.Fatalf("Got Error: %v", err)
		}
	}
}

func BenchmarkNewHTTPClient(b *testing.B) {
	h := httpServerB(b, randomString(500))
	defer h.Close()
	var o HTTPOptions
	for x := 0; x < b.N; x++ {
		_, err := NewHTTPClient(context.Background(), &o)
		if err != nil {
			b.Fatalf("Got Error: %v", err)
		}
	}
}
