package helper

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/OJ/gobuster/v3/libgobuster"
)

// ParseExtensions parses the extensions provided as a comma separated list
func ParseExtensions(extensions string) (libgobuster.Set[string], error) {
	ret := libgobuster.NewSet[string]()

	if extensions == "" {
		return ret, nil
	}

	for _, e := range strings.Split(extensions, ",") {
		e = strings.TrimSpace(e)
		// remove leading . from extensions
		ret.Add(strings.TrimPrefix(e, "."))
	}
	return ret, nil
}

func ParseExtensionsFile(file string) ([]string, error) {
	var ret []string

	stream, err := os.Open(file)
    if err != nil {
        return ret, err
    }
    defer stream.Close()

    scanner := bufio.NewScanner(stream)
    for scanner.Scan() {
		e := scanner.Text()
		e = strings.TrimSpace(e)
		// remove leading . from extensions
		ret = append(ret, (strings.TrimPrefix(e, ".")))
    }
    return ret, nil
}

// ParseCommaSeparatedInt parses the status codes provided as a comma separated list
func ParseCommaSeparatedInt(inputString string) (libgobuster.Set[int], error) {
	ret := libgobuster.NewSet[int]()

	if inputString == "" {
		return ret, nil
	}

	for _, c := range strings.Split(inputString, ",") {
		c = strings.TrimSpace(c)
		i, err := strconv.Atoi(c)
		if err != nil {
			return libgobuster.NewSet[int](), fmt.Errorf("invalid string given: %s", c)
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
