package vhost

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobustervhost"
	"github.com/OJ/gobuster/v3/libgobuster"
)

func httpServer(b *testing.B, content string) *httptest.Server {
	b.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, content); err != nil {
			b.Fatalf("%v", err)
		}
	}))
	return ts
}

func BenchmarkVhostMode(b *testing.B) {
	h := httpServer(b, "test")
	defer h.Close()

	pluginopts := gobustervhost.NewOptions()
	pluginopts.URL = h.URL
	pluginopts.Timeout = 10 * time.Second

	wordlist, err := os.CreateTemp("", "")
	if err != nil {
		b.Fatalf("could not create tempfile: %v", err)
	}
	defer os.Remove(wordlist.Name())
	for w := 0; w < 1000; w++ {
		_, _ = wordlist.WriteString(fmt.Sprintf("%d\n", w))
	}
	if err := wordlist.Close(); err != nil {
		b.Fatalf("%v", err)
	}

	globalopts := libgobuster.Options{
		Threads:    10,
		Wordlist:   wordlist.Name(),
		NoProgress: true,
	}

	ctx := context.Background()
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func(out, err *os.File) { os.Stdout = out; os.Stderr = err }(oldStdout, oldStderr)
	devnull, err := os.Open(os.DevNull)
	if err != nil {
		b.Fatalf("could not get devnull %v", err)
	}
	defer devnull.Close()
	log := libgobuster.NewLogger(false)

	// Run the real benchmark
	for x := 0; x < b.N; x++ {
		os.Stdout = devnull
		os.Stderr = devnull
		plugin, err := gobustervhost.New(&globalopts, pluginopts, log)
		if err != nil {
			b.Fatalf("error on creating gobusterdir: %v", err)
		}

		if err := cli.Gobuster(ctx, &globalopts, plugin, log); err != nil {
			b.Fatalf("error on running gobuster: %v", err)
		}
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}
}
