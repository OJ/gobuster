package helper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/OJ/gobuster/v3/libgobuster"
)

// ParseExtensions parses the extensions provided as a comma separated list
func ParseExtensions(extensions string) (libgobuster.StringSet, error) {
	if extensions == "" {
		return libgobuster.StringSet{}, fmt.Errorf("invalid extension string provided")
	}

	ret := libgobuster.NewStringSet()
	exts := strings.Split(extensions, ",")
	for _, e := range exts {
		e = strings.TrimSpace(e)
		// remove leading . from extensions
		ret.Add(strings.TrimPrefix(e, "."))
	}
	return ret, nil
}

// ParseStatusCodes parses the status codes provided as a comma separated list
func ParseStatusCodes(statuscodes string) (libgobuster.IntSet, error) {
	if statuscodes == "" {
		return libgobuster.IntSet{}, fmt.Errorf("invalid status code string provided")
	}

	ret := libgobuster.NewIntSet()
	for _, c := range strings.Split(statuscodes, ",") {
		c = strings.TrimSpace(c)
		i, err := strconv.Atoi(c)
		if err != nil {
			return libgobuster.IntSet{}, fmt.Errorf("invalid status code given: %s", c)
		}
		ret.Add(i)
	}
	return ret, nil
}
