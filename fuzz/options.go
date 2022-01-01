package fuzz

import "github.com/OJ/gobuster/v3/lib"

// NewOptionsFuzz returns a new initialized OptionsFuzz
func NewOptionsFuzz() *OptionsFuzz {
	return &OptionsFuzz{
		ExcludedStatusCodesParsed: lib.NewIntSet(),
	}
}
