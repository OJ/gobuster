package helper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/OJ/gobuster/v3/lib"
)

// ParseExtensions parses the extensions provided as a comma separated list
func ParseExtensions(extensions string) (lib.StringSet, error) {
	if extensions == "" {
		return lib.StringSet{}, nil
	}

	ret := lib.NewStringSet()
	for _, e := range strings.Split(extensions, ",") {
		e = strings.TrimSpace(e)
		// remove leading . from extensions
		ret.Add(strings.TrimPrefix(e, "."))
	}
	return ret, nil
}

// ParseCommaSeparatedInt parses the status codes provided as a comma separated list
func ParseCommaSeparatedInt(inputString string) (lib.IntSet, error) {
	if inputString == "" {
		return lib.IntSet{}, nil
	}

	ret := lib.NewIntSet()
	for _, c := range strings.Split(inputString, ",") {
		c = strings.TrimSpace(c)
		i, err := strconv.Atoi(c)
		if err != nil {
			return lib.IntSet{}, fmt.Errorf("invalid string given: %s", c)
		}
		ret.Add(i)
	}
	return ret, nil
}

// SliceContains checks if an integer slice contains a specific value
func SliceContains(s []int, e int) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// JoinIntSlice joins an int slice by ,
func JoinIntSlice(s []int) string {
	valuesText := make([]string, len(s))
	for i, number := range s {
		text := strconv.Itoa(number)
		valuesText[i] = text
	}
	result := strings.Join(valuesText, ",")
	return result
}
