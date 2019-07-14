package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/OJ/gobuster/v3/cli"
	"github.com/OJ/gobuster/v3/gobustervhost"
	"github.com/OJ/gobuster/v3/libgobuster"
)

func BenchmarkVhostMode(b *testing.B) {
	h := httpServer(b, "test")
	defer h.Close()

	pluginopts := gobustervhost.OptionsVhost{}
	pluginopts.URL = h.URL
	pluginopts.Timeout = 10 * time.Second

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
		plugin, err := gobustervhost.NewGobusterVhost(ctx, &globalopts, &pluginopts)
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
