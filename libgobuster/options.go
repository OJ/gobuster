package libgobuster

import "time"

// Options holds all options that can be passed to libgobuster
type Options struct {
	Threads        int
	Wordlist       string
	PatternFile    string
	Patterns       []string
	OutputFilename string
	NoStatus       bool
	NoProgress     bool
	NoError        bool
	Quiet          bool
	Verbose        bool
	Delay          time.Duration
}

// NewOptions returns a new initialized Options object
func NewOptions() *Options {
	return &Options{}
}
