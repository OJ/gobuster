package gobusterdir

import (
	"testing"

	"github.com/OJ/gobuster/v3/libgobuster"
)

func TestAdditionalWordsLen(t *testing.T) {
	t.Parallel()

	tt := []struct {
		testName   string
		extensions map[string]bool
	}{
		{"No extensions", map[string]bool{}},
		{"Some extensions", map[string]bool{"htm": true, "html": true, "php": true}},
	}

	globalOpts := libgobuster.Options{}
	for _, x := range tt {
		opts := OptionsDir{}
		opts.ExtensionsParsed.Set = x.extensions

		d, _ := New(&globalOpts, &opts, nil)

		calculatedLen := d.AdditionalWordsLen()
		wordsLen := len(d.AdditionalWords("dummy"))

		if calculatedLen != wordsLen {
			t.Fatalf("Mismatched additional words length: %d got %d generated words %v", calculatedLen, wordsLen, d.AdditionalWords("dummy"))
		}
	}
}
