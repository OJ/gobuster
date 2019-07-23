package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobusterdir"
	"github.com/OJ/gobuster/v3/helper"
	"github.com/OJ/gobuster/v3/libgobuster"
)

func httpServer(b *testing.B, content string) *httptest.Server {
	b.Helper()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, content)
	}))
	return ts
}
func BenchmarkDirMode(b *testing.B) {
	h := httpServer(b, "test")
	defer h.Close()

	pluginopts := gobusterdir.NewOptionsDir()
	pluginopts.URL = h.URL
	pluginopts.Timeout = 10 * time.Second
	pluginopts.WildcardForced = true

	pluginopts.Extensions = ".php,.csv"
	tmpExt, err := helper.ParseExtensions(pluginopts.Extensions)
	if err != nil {
		b.Fatalf("could not parse extensions: %v", err)
	}
	pluginopts.ExtensionsParsed = tmpExt

	pluginopts.StatusCodes = "200,204,301,302,307,401,403"
	tmpStat, err := helper.ParseStatusCodes(pluginopts.StatusCodes)
	if err != nil {
		b.Fatalf("could not parse status codes: %v", err)
	}
	pluginopts.StatusCodesParsed = tmpStat

	wordlist, err := ioutil.TempFile("", "")
	if err != nil {
		b.Fatalf("could not create tempfile: %v", err)
	}
	defer os.Remove(wordlist.Name())
	for w := 0; w < 1000; w++ {
		_, _ = wordlist.WriteString(fmt.Sprintf("%d\n", w))
	}
	wordlist.Close()

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
	log.SetFlags(0)
	log.SetOutput(ioutil.Discard)

	// Run the real benchmark
	for x := 0; x < b.N; x++ {
		os.Stdout = devnull
		os.Stderr = devnull
		plugin, err := gobusterdir.NewGobusterDir(ctx, &globalopts, pluginopts)
		if err != nil {
			b.Fatalf("error on creating gobusterdir: %v", err)
		}

		if err := cli.Gobuster(ctx, &globalopts, plugin); err != nil {
			b.Fatalf("error on running gobuster: %v", err)
		}
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}
}
