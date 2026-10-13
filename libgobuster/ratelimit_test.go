package libgobuster

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
)

// ratePlugin is a minimal GobusterPlugin used to exercise the stop on rate
// limit behaviour without doing any real network work. It returns
// ErrRateLimited once it has processed failAfter words.
type ratePlugin struct {
	processed atomic.Int64
	failAfter int64
}

func (p *ratePlugin) Name() string { return "ratemock" }

func (p *ratePlugin) PreRun(_ context.Context, _ *Progress) error { return nil }

func (p *ratePlugin) ProcessWord(_ context.Context, _ string, _ *Progress) (Result, error) {
	n := p.processed.Add(1)
	if p.failAfter > 0 && n >= p.failAfter {
		return nil, ErrRateLimited
	}
	return nil, nil // nolint:nilnil
}

func (p *ratePlugin) AdditionalWords(_ string) []string { return nil }

func (p *ratePlugin) AdditionalWordsLen() int { return 0 }

func (p *ratePlugin) AdditionalSuccessWords(_ string) []string { return nil }

func (p *ratePlugin) GetConfigString() (string, error) { return "", nil }

func writeWordlist(t *testing.T, words int) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "wordlist.txt")
	var b []byte
	for i := range words {
		b = append(b, []byte(fmt.Sprintf("word%d\n", i))...)
	}
	if err := os.WriteFile(f, b, 0o600); err != nil {
		t.Fatalf("could not write wordlist: %v", err)
	}
	return f
}

// drain reads from the progress channels until they are closed by Run so the
// worker never blocks on a send.
func drain(g *Gobuster) *sync.WaitGroup {
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		for r := range g.Progress.ResultChan {
			_ = r
		}
	}()
	go func() {
		defer wg.Done()
		for e := range g.Progress.ErrorChan {
			_ = e
		}
	}()
	go func() {
		defer wg.Done()
		for m := range g.Progress.MessageChan {
			_ = m
		}
	}()
	return &wg
}

func TestStopOnRateLimit(t *testing.T) {
	t.Parallel()

	const words = 100
	wordlist := writeWordlist(t, words)

	opts := &Options{
		Threads:         1,
		Wordlist:        wordlist,
		StopOnRateLimit: true,
	}
	plugin := &ratePlugin{failAfter: 2}
	g, err := NewGobuster(opts, plugin, NewLogger(false))
	if err != nil {
		t.Fatalf("could not create gobuster: %v", err)
	}

	wg := drain(g)
	if err := g.Run(context.Background()); err != nil {
		t.Fatalf("run returned an error: %v", err)
	}
	wg.Wait()

	if !g.rateLimited.Load() {
		t.Fatal("expected the run to be flagged as rate limited")
	}
	if got := plugin.processed.Load(); got >= words {
		t.Fatalf("expected the run to stop early, but processed %d of %d words", got, words)
	}
}

func TestStopOnRateLimitDisabled(t *testing.T) {
	t.Parallel()

	const words = 100
	wordlist := writeWordlist(t, words)

	opts := &Options{
		Threads:         1,
		Wordlist:        wordlist,
		StopOnRateLimit: false,
	}
	// the plugin still returns ErrRateLimited, but with the flag disabled it
	// is treated as a normal error and the run keeps going
	plugin := &ratePlugin{failAfter: 2}
	g, err := NewGobuster(opts, plugin, NewLogger(false))
	if err != nil {
		t.Fatalf("could not create gobuster: %v", err)
	}

	wg := drain(g)
	if err := g.Run(context.Background()); err != nil {
		t.Fatalf("run returned an error: %v", err)
	}
	wg.Wait()

	if g.rateLimited.Load() {
		t.Fatal("run should not be flagged as rate limited when the flag is disabled")
	}
	if got := plugin.processed.Load(); got != words {
		t.Fatalf("expected all %d words to be processed, got %d", words, got)
	}
}
