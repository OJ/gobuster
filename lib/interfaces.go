package lib

// Result is an interface for the Result object
type Result interface {
	ResultToString() (string, error)
}
