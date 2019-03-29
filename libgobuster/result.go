package libgobuster

// ResultStatus is a status enum
type ResultStatus int

const (
	// StatusFound represents a found item
	StatusFound ResultStatus = iota
	// StatusMissed represents a missed item
	StatusMissed ResultStatus = iota
)

// Result represents a single gobuster result
type Result struct {
	Entity     string
	StatusCode int
	Status     ResultStatus
	Extra      string
	Size       *int64
}

// ToString converts the Result to it's textual representation
func (r *Result) ToString(g *Gobuster) (string, error) {
	s, err := g.plugin.ResultToString(r)
	if err != nil {
		return "", err
	}
	return *s, nil
}
