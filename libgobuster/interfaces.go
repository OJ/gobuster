package libgobuster

// GobusterPlugin is an interface which plugins must implement
type GobusterPlugin interface {
	PreRun() error
	Run(string) ([]Result, error)
	RequestsPerRun() int
	ResultToString(*Result) (*string, error)
	GetConfigString() (string, error)
}
