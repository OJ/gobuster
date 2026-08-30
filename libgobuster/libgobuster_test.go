package libgobuster

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

type testPlugin struct{}

func (testPlugin) Name() string                            { return "test" }
func (testPlugin) PreRun(context.Context, *Progress) error { return nil }
func (testPlugin) ProcessWord(context.Context, string, *Progress) (Result, error) {
	return nil, nil //nolint:nilnil // A test plugin intentionally produces no result.
}

type recursiveTestResult string

func (recursiveTestResult) ResultToString() (string, error) { return "", nil }
func (r recursiveTestResult) RecursiveTarget() string       { return string(r) }

type recursiveTestPlugin struct {
	mu              sync.Mutex
	target          string
	started         []string
	distinctTargets bool
}

func (*recursiveTestPlugin) Name() string { return "recursive test" }
func (p *recursiveTestPlugin) PreRun(context.Context, *Progress) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.started = append(p.started, p.target)
	return nil
}
func (p *recursiveTestPlugin) ProcessWord(_ context.Context, word string, _ *Progress) (Result, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	switch p.target {
	case "root":
		if p.distinctTargets {
			return recursiveTestResult(word), nil
		}
		return recursiveTestResult("child"), nil
	case "child":
		return recursiveTestResult("grandchild"), nil
	default:
		return nil, nil //nolint:nilnil
	}
}

func TestRunRecursionEnforcesTargetLimit(t *testing.T) {
	wordlist := t.TempDir() + "/words.txt"
	if err := os.WriteFile(wordlist, []byte("one\ntwo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	plugin := &recursiveTestPlugin{target: "root", distinctTargets: true}
	g, err := NewGobuster(&Options{
		Threads: 1, Wordlist: wordlist, Recursion: true,
		RecursionMaxTargets: 1,
	}, plugin, NewLogger(false))
	if err != nil {
		t.Fatal(err)
	}
	drainProgress(g.Progress)
	if err := g.Run(t.Context()); err == nil || !strings.Contains(err.Error(), "target limit") {
		t.Fatalf("expected recursive target limit error, got %v", err)
	}
}
func (*recursiveTestPlugin) AdditionalWords(string) []string        { return nil }
func (*recursiveTestPlugin) AdditionalWordsLen() int                { return 0 }
func (*recursiveTestPlugin) AdditionalSuccessWords(string) []string { return nil }
func (*recursiveTestPlugin) GetConfigString() (string, error)       { return "", nil }
func (p *recursiveTestPlugin) SetTarget(target string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.target = target
	return nil
}

func TestRunRecursionIsSequentialDeduplicatedAndDepthLimited(t *testing.T) {
	wordlist := t.TempDir() + "/words.txt"
	if err := os.WriteFile(wordlist, []byte("one\ntwo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	plugin := &recursiveTestPlugin{target: "root"}
	g, err := NewGobuster(&Options{
		Threads: 2, Wordlist: wordlist, Recursion: true,
		RecursionDepth: 1, RecursionMaxTargets: 10,
	}, plugin, NewLogger(false))
	if err != nil {
		t.Fatal(err)
	}
	drainProgress(g.Progress)
	if err := g.Run(t.Context()); err != nil {
		t.Fatal(err)
	}
	plugin.mu.Lock()
	defer plugin.mu.Unlock()
	if got, want := strings.Join(plugin.started, ","), "root,child"; got != want {
		t.Fatalf("scanned targets %q, want %q", got, want)
	}
}

func TestRunRejectsRecursionForUnsupportedPlugin(t *testing.T) {
	wordlist := t.TempDir() + "/words.txt"
	if err := os.WriteFile(wordlist, []byte("one\ntwo\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	plugin := &recursiveTestPlugin{target: "root"}
	g, err := NewGobuster(&Options{
		Threads: 1, Wordlist: wordlist, Recursion: true,
		RecursionMaxTargets: 0,
	}, plugin, NewLogger(false))
	if err != nil {
		t.Fatal(err)
	}
	// Use a non-recursive plugin to verify the capability check independently.
	g.plugin = testPlugin{}
	drainProgress(g.Progress)
	if err := g.Run(t.Context()); err == nil || !strings.Contains(err.Error(), "does not support recursion") {
		t.Fatalf("expected unsupported recursion error, got %v", err)
	}
}

func drainProgress(progress *Progress) {
	go func() {
		for range progress.ResultChan {
		}
	}()
	go func() {
		for range progress.ErrorChan {
		}
	}()
	go func() {
		for range progress.MessageChan {
		}
	}()
}
func (testPlugin) AdditionalWords(string) []string        { return nil }
func (testPlugin) AdditionalWordsLen() int                { return 0 }
func (testPlugin) AdditionalSuccessWords(string) []string { return nil }
func (testPlugin) GetConfigString() (string, error)       { return "", nil }

func TestRunRejectsEmptyWordlist(t *testing.T) {
	t.Parallel()
	wordlist := t.TempDir() + "/empty.txt"
	if err := os.WriteFile(wordlist, nil, 0o600); err != nil {
		t.Fatal(err)
	}

	g, err := NewGobuster(&Options{Threads: 1, Wordlist: wordlist}, testPlugin{}, NewLogger(false))
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Run(t.Context()); err == nil {
		t.Fatal("expected an empty wordlist error")
	}
}

func TestRunReturnsOversizedWordlistLineError(t *testing.T) {
	t.Parallel()
	wordlist := t.TempDir() + "/large.txt"
	if err := os.WriteFile(wordlist, []byte(strings.Repeat("x", maxWordlistLineSize+1)), 0o600); err != nil {
		t.Fatal(err)
	}

	g, err := NewGobuster(&Options{Threads: 1, Wordlist: wordlist}, testPlugin{}, NewLogger(false))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	if err := g.Run(ctx); err == nil {
		t.Fatal("expected an oversized wordlist line error")
	}
}
