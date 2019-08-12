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

// ParseCommaSeperated parses the status codes provided as a comma separated list
func ParseCommaSeperatedInt(inputString string) (libgobuster.IntSet, error) {
	if inputString == "" {
		return libgobuster.IntSet{}, fmt.Errorf("invalid string provided")
	}

	ret := libgobuster.NewIntSet()
	for _, c := range strings.Split(inputString, ",") {
		c = strings.TrimSpace(c)
		i, err := strconv.Atoi(c)
		if err != nil {
			return libgobuster.IntSet{}, fmt.Errorf("invalid string given: %s", c)
		}
		ret.Add(i)
	}
	return ret, nil
}

// ParseCommaSeperatedString parses the status codes provided as a comma separated list
func ParseCommaSeperatedString(inputString string) (libgobuster.StringSet, error) {
	if inputString == "" {
		return libgobuster.StringSet{}, fmt.Errorf("invalid string provided")
	}

	ret := libgobuster.NewStringSet()
	for _, c := range strings.Split(inputString, ",") {
		c = strings.TrimSpace(c)
		ret.Add(c)
	}
	return ret, nil
}
