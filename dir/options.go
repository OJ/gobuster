package dir

import "github.com/OJ/gobuster/v3/lib"

// NewOptionsDir returns a new initialized OptionsDir
func NewOptionsDir() *OptionsDir {
	return &OptionsDir{
		StatusCodesParsed:          lib.NewIntSet(),
		StatusCodesBlacklistParsed: lib.NewIntSet(),
		ExtensionsParsed:           lib.NewStringSet(),
	}
}
