package libgobuster

// Result represents a single gobuster result
type Result struct {
	Entity string
	Status int
	Extra  string
	Size   *int64
}

// ToString converts the Result to it's textual representation
func (r *Result) ToString(g *Gobuster) (string, error) {
	s, err := g.plugin.ResultToString(r)
	if err != nil {
		return "", err
	}
	return *s, nil
}
