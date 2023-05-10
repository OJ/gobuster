package libgobuster

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Set is a set of Ts
type Set[T comparable] struct {
	Set map[T]bool
}

// NewSSet creates a new initialized Set
func NewSet[T comparable]() Set[T] {
	return Set[T]{Set: map[T]bool{}}
}

// Add an element to a set
func (set *Set[T]) Add(s T) bool {
	_, found := set.Set[s]
	set.Set[s] = true
	return !found
}

// AddRange adds a list of elements to a set
func (set *Set[T]) AddRange(ss []T) {
	for _, s := range ss {
		set.Set[s] = true
	}
}

// Contains tests if an element is in a set
func (set *Set[T]) Contains(s T) bool {
	_, found := set.Set[s]
	return found
}

// ContainsAny checks if any of the elements exist
func (set *Set[T]) ContainsAny(ss []T) bool {
	for _, s := range ss {
		if set.Set[s] {
			return true
		}
	}
	return false
}

// Length returns the length of the Set
func (set *Set[T]) Length() int {
	return len(set.Set)
}

// Stringify the set
func (set *Set[T]) Stringify() string {
	values := make([]string, len(set.Set))
	i := 0
	for s := range set.Set {
		values[i] = fmt.Sprint(s)
		i++
	}
	return strings.Join(values, ",")
}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 1
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case errors.Is(err, io.EOF):
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

// DefaultUserAgent returns the default user agent to use in HTTP requests
func DefaultUserAgent() string {
	return fmt.Sprintf("gobuster/%s", VERSION)
}

// ParseExtensions parses the extensions provided as a comma separated list
func ParseExtensions(extensions string) (Set[string], error) {
	ret := NewSet[string]()

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
func ParseCommaSeparatedInt(inputString string) (Set[int], error) {
	ret := NewSet[int]()

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
				return NewSet[int](), fmt.Errorf("invalid range given: %s", part)
			}
			from := strings.TrimSpace(match[1])
			to := strings.TrimSpace(match[2])
			fromI, err := strconv.Atoi(from)
			if err != nil {
				return NewSet[int](), fmt.Errorf("invalid string in range %s: %s", part, from)
			}
			toI, err := strconv.Atoi(to)
			if err != nil {
				return NewSet[int](), fmt.Errorf("invalid string in range %s: %s", part, to)
			}
			if toI < fromI {
				return NewSet[int](), fmt.Errorf("invalid range given: %s", part)
			}
			for i := fromI; i <= toI; i++ {
				ret.Add(i)
			}
		} else {
			i, err := strconv.Atoi(part)
			if err != nil {
				return NewSet[int](), fmt.Errorf("invalid string given: %s", part)
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
