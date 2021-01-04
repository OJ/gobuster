package libgobuster

// GobusterPlugin is an interface which plugins must implement
type GobusterPlugin interface {
	Name() string
	RequestsPerRun() int
	PreRun() error
	Run(string, chan<- Result) error
	GetConfigString() (string, error)
}

// Result is an interface for the Result object
type Result interface {
	ResultToString() (string, error)
}
