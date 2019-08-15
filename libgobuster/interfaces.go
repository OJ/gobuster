package libgobuster

// GobusterPlugin is an interface which plugins must implement
type GobusterPlugin interface {
	Name() string
	PreRun() error
	Run(string) ([]Result, error)
	ResultToString(*Result) (*string, error)
	GetConfigString() (string, error)
}
