package libgobuster

import (
	"bufio"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// wordlistTestPlugin exercises the scheduler without making network requests.
type wordlistTestPlugin struct {
	words []string
}

func (p *wordlistTestPlugin) Name() string { return "wordlist-test" }

func (p *wordlistTestPlugin) PreRun(context.Context, *Progress) error { return nil }

func (p *wordlistTestPlugin) ProcessWord(_ context.Context, word string, _ *Progress) (Result, error) {
	p.words = append(p.words, word)
	return nil, nil
}

func (p *wordlistTestPlugin) AdditionalWords(string) []string { return nil }

func (p *wordlistTestPlugin) AdditionalWordsLen() int { return 0 }

func (p *wordlistTestPlugin) AdditionalSuccessWords(string) []string { return nil }

func (p *wordlistTestPlugin) GetConfigString() (string, error) { return "", nil }

func TestRunWordlistScanner(t *testing.T) {
	longLine := strings.Repeat("x", bufio.MaxScanTokenSize) + "\n"
	for _, source := range []string{"file", "stdin"} {
		for _, tc := range []struct {
			name    string
			input   string
			wantErr error
		}{
			{name: "valid", input: "first\n\nsecond\n"},
			{name: "long-first-line", input: longLine, wantErr: bufio.ErrTooLong},
			{name: "long-later-line", input: "first\n" + longLine + "last\n", wantErr: bufio.ErrTooLong},
		} {
			t.Run(source+"/"+tc.name, func(t *testing.T) {
				path := filepath.Join(t.TempDir(), "wordlist.txt")
				if err := os.WriteFile(path, []byte(tc.input), 0o600); err != nil {
					t.Fatal(err)
				}
				if source == "stdin" {
					input, err := os.Open(path)
					if err != nil {
						t.Fatal(err)
					}
					originalStdin := os.Stdin
					os.Stdin = input
					t.Cleanup(func() {
						os.Stdin = originalStdin
						input.Close()
					})
					path = "-"
				}

				plugin := &wordlistTestPlugin{}
				g, err := NewGobuster(&Options{Wordlist: path, Threads: 1}, plugin, nil)
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
				defer cancel()
				err = g.Run(ctx)
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("Run() error = %v, want %v", err, tc.wantErr)
				}
				if ctx.Err() != nil {
					t.Fatal("Run() waited for context cancellation instead of finishing")
				}
				if tc.wantErr == nil && strings.Join(plugin.words, ",") != "first,second" {
					t.Fatalf("processed words = %v, want [first second]", plugin.words)
				}
			})
		}
	}
}
