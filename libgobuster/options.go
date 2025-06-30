package libgobuster

import (
	"time"
)

// Options holds all options that can be passed to libgobuster
type Options struct {
	Threads             int
	Debug               bool
	Wordlist            string
	WordlistOffset      int
	PatternFile         string
	DiscoverPatternFile string
	Patterns            []string
	DiscoverPatterns    []string
	OutputFilename      string
	NoProgress          bool
	NoError             bool
	Quiet               bool
	Delay               time.Duration
}
