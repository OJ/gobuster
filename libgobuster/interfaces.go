package libgobuster

import "context"

// GobusterPlugin is an interface which plugins must implement
type GobusterPlugin interface {
	Name() string
	RequestsPerRun() int
	PreRun(context.Context) error
	Run(context.Context, string, chan<- Result) error
	GetConfigString() (string, error)
}

// Result is an interface for the Result object
type Result interface {
	ResultToString() (string, error)
}
