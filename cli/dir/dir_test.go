package dir

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterdir"
	"github.com/OJ/gobuster/v3/libgobuster"
)

func httpServer(b *testing.B, content string) *httptest.Server {
	b.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if _, err := fmt.Fprint(w, content); err != nil {
			b.Fatalf("%v", err)
		}
	}))
	return ts
}

func BenchmarkDirMode(b *testing.B) {
	h := httpServer(b, "test")
	defer h.Close()

	u, err := url.Parse(h.URL)
	if err != nil {
		b.Fatalf("could not parse URL: %v", err)
	}
	pluginopts := gobusterdir.NewOptions()
	pluginopts.URL = u
	pluginopts.Timeout = 10 * time.Second

	pluginopts.Extensions = ".php,.csv"
	tmpExt, err := libgobuster.ParseExtensions(pluginopts.Extensions)
	if err != nil {
		b.Fatalf("could not parse extensions: %v", err)
	}
	pluginopts.ExtensionsParsed = tmpExt

	pluginopts.StatusCodes = "200,204,301,302,307,401,403"
	tmpStat, err := libgobuster.ParseCommaSeparatedInt(pluginopts.StatusCodes)
	if err != nil {
		b.Fatalf("could not parse status codes: %v", err)
	}
	pluginopts.StatusCodesParsed = tmpStat

	wordlist, err := os.CreateTemp(b.TempDir(), "")
	if err != nil {
		b.Fatalf("could not create tempfile: %v", err)
	}
	defer os.Remove(wordlist.Name())
	for w := range 1000 {
		_, _ = fmt.Fprintf(wordlist, "%d\n", w)
	}
	if err := wordlist.Close(); err != nil {
		b.Fatalf("%v", err)
	}

	globalopts := libgobuster.Options{
		Threads:    10,
		Wordlist:   wordlist.Name(),
		NoProgress: true,
	}

	ctx := b.Context()
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
	for b.Loop() {
		os.Stdout = devnull
		os.Stderr = devnull
		plugin, err := gobusterdir.New(&globalopts, pluginopts, log)
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
