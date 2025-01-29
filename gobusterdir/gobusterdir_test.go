package gobusterdir

import (
	"testing"

	"github.com/OJ/gobuster/v3/libgobuster"
)

func TestAdditionalWordsLen(t *testing.T) {
	t.Parallel()

	tt := []struct {
		testName   string
		backups    bool
		extensions map[string]bool
	}{
		{"Backups no extensions", true, map[string]bool{}},
		{"No backups no extensions", false, map[string]bool{}},
		{"Backups and extensions", true, map[string]bool{"htm": true, "html": true, "php": true}},
		{"No Backups and some extensions", false, map[string]bool{"htm": true, "html": true, "php": true}},
	}

	globalOpts := libgobuster.Options{}
	for _, x := range tt {
		opts := OptionsDir{}
		opts.DiscoverBackup = x.backups
		opts.ExtensionsParsed.Set = x.extensions

		d, _ := New(&globalOpts, &opts, nil)

		calculatedLen := d.AdditionalWordsLen()
		wordsLen := len(d.AdditionalWords("dummy"))

		if calculatedLen != wordsLen {
			t.Fatalf("Mismatched additional words length: %d got %d generated words %v", calculatedLen, wordsLen, d.AdditionalWords("dummy"))
		}
	}
}
