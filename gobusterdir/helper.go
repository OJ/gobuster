package gobusterdir

import (
	"fmt"
	"strconv"
	"strings"
)

// ParseExtensions parses the extensions provided as a comma seperated list
func (opt *OptionsDir) ParseExtensions() error {
	if opt.Extensions == "" {
		return fmt.Errorf("invalid extension string provided")
	}

	exts := strings.Split(opt.Extensions, ",")
	for _, e := range exts {
		e = strings.TrimSpace(e)
		// remove leading . from extensions
		opt.ExtensionsParsed.Add(strings.TrimPrefix(e, "."))
	}
	return nil
}

// ParseStatusCodes parses the status codes provided as a comma seperated list
func (opt *OptionsDir) ParseStatusCodes() error {
	if opt.StatusCodes == "" {
		return fmt.Errorf("invalid status code string provided")
	}

	for _, c := range strings.Split(opt.StatusCodes, ",") {
		c = strings.TrimSpace(c)
		i, err := strconv.Atoi(c)
		if err != nil {
			return fmt.Errorf("invalid status code given: %s", c)
		}
		opt.StatusCodesParsed.Add(i)
	}
	return nil
}
