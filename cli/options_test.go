package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/urfave/cli/v2"
)

func runParseGlobalOptions(t *testing.T, args ...string) (string, error) {
	t.Helper()

	var parseErr error
	var parseResult string
	app := &cli.App{
		Flags: GlobalOptions(),
		Action: func(c *cli.Context) error {
			_, err := ParseGlobalOptions(c)
			if err != nil {
				parseErr = err
				parseResult = err.Error()
			}
			return nil
		},
	}

	if err := app.Run(append([]string{"gobuster"}, args...)); err != nil {
		t.Fatalf("failed to run test app: %v", err)
	}

	return parseResult, parseErr
}

func TestParseGlobalOptionsRejectsNonPositiveThreads(t *testing.T) {
	wordlist := filepath.Join(t.TempDir(), "wordlist.txt")
	if err := os.WriteFile(wordlist, []byte("admin\n"), 0o600); err != nil {
		t.Fatalf("failed to write wordlist: %v", err)
	}

	for _, threads := range []string{"0", "-1"} {
		t.Run("threads="+threads, func(t *testing.T) {
			msg, err := runParseGlobalOptions(t, "--wordlist", wordlist, "--threads", threads)
			if err == nil {
				t.Fatalf("expected an error for --threads %s, got none", threads)
			}
			expected := "threads must be bigger or equal to 1"
			if msg != expected {
				t.Fatalf("expected error %q, got %q", expected, msg)
			}
		})
	}
}

func TestParseGlobalOptionsAcceptsPositiveThreads(t *testing.T) {
	wordlist := filepath.Join(t.TempDir(), "wordlist.txt")
	if err := os.WriteFile(wordlist, []byte("admin\n"), 0o600); err != nil {
		t.Fatalf("failed to write wordlist: %v", err)
	}

	if msg, err := runParseGlobalOptions(t, "--wordlist", wordlist, "--threads", "1"); err != nil {
		t.Fatalf("expected --threads 1 to be accepted, got %q", msg)
	}
}
