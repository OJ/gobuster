package libgobuster

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

type testPlugin struct{}

func (testPlugin) Name() string                            { return "test" }
func (testPlugin) PreRun(context.Context, *Progress) error { return nil }
func (testPlugin) ProcessWord(context.Context, string, *Progress) (Result, error) {
	return nil, nil //nolint:nilnil // A test plugin intentionally produces no result.
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
