package libgobuster

import "context"

// GobusterPlugin is an interface which plugins must implement
type GobusterPlugin interface {
	Name() string
	PreRun(context.Context, *Progress) error
	ProcessWord(context.Context, string, *Progress) (Result, error)
	AdditionalWords(string) []string
	AdditionalWordsLen() int
	AdditionalSuccessWords(string) []string
	GetConfigString() (string, error)
}

// Result is an interface for the Result object
type Result interface {
	ResultToString() (string, error)
}
