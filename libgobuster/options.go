package libgobuster

// Options holds all options that can be passed to libgobuster
type Options struct {
	Threads        int
	Wordlist       string
	OutputFilename string
	HeadRequests   bool
	NoStatus       bool
	NoProgress     bool
	Quiet          bool
	WildcardForced bool
	Verbose        bool
}

// NewOptions returns a new initialized Options object
func NewOptions() *Options {
	return &Options{}
}
