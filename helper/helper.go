package helper

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
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

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return ret, nil
}

// ParseCommaSeparatedInt parses the status codes provided as a comma separated list
func ParseCommaSeparatedInt(inputString string) (libgobuster.Set[int], error) {
	ret := libgobuster.NewSet[int]()

	if inputString == "" {
		return ret, nil
	}

	for _, part := range strings.Split(inputString, ",") {
		part = strings.TrimSpace(part)
		// check for range
		if strings.Contains(part, "-") {
			re := regexp.MustCompile(`^\s*(\d+)\s*-\s*(\d+)\s*$`)
			match := re.FindStringSubmatch(part)
			if match == nil || len(match) != 3 {
				return libgobuster.NewSet[int](), fmt.Errorf("invalid range given: %s", part)
			}
			from := strings.TrimSpace(match[1])
			to := strings.TrimSpace(match[2])
			fromI, err := strconv.Atoi(from)
			if err != nil {
				return libgobuster.NewSet[int](), fmt.Errorf("invalid string in range %s: %s", part, from)
			}
			toI, err := strconv.Atoi(to)
			if err != nil {
				return libgobuster.NewSet[int](), fmt.Errorf("invalid string in range %s: %s", part, to)
			}
			if toI < fromI {
				return libgobuster.NewSet[int](), fmt.Errorf("invalid range given: %s", part)
			}
			for i := fromI; i <= toI; i++ {
				ret.Add(i)
			}
		} else {
			i, err := strconv.Atoi(part)
			if err != nil {
				return libgobuster.NewSet[int](), fmt.Errorf("invalid string given: %s", part)
			}
			ret.Add(i)
		}
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
