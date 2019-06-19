package libgobuster

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"
)

// IntSet is a set of Ints
type IntSet struct {
	Set map[int]bool
}

// StringSet is a set of Strings
type StringSet struct {
	Set map[string]bool
}

// NewStringSet creates a new initialized StringSet
func NewStringSet() StringSet {
	return StringSet{Set: map[string]bool{}}
}

// Add an element to a set
func (set *StringSet) Add(s string) bool {
	_, found := set.Set[s]
	set.Set[s] = true
	return !found
}

// AddRange adds a list of elements to a set
func (set *StringSet) AddRange(ss []string) {
	for _, s := range ss {
		set.Set[s] = true
	}
}

// Contains tests if an element is in a set
func (set *StringSet) Contains(s string) bool {
	_, found := set.Set[s]
	return found
}

// ContainsAny checks if any of the elements exist
func (set *StringSet) ContainsAny(ss []string) bool {
	for _, s := range ss {
		if set.Set[s] {
			return true
		}
	}
	return false
}

// Length returns the length of the Set
func (set *StringSet) Length() int {
	return len(set.Set)
}

// Stringify the set
func (set *StringSet) Stringify() string {
	values := []string{}
	for s := range set.Set {
		values = append(values, s)
	}
	return strings.Join(values, ",")
}

// NewIntSet creates a new initialized IntSet
func NewIntSet() IntSet {
	return IntSet{Set: map[int]bool{}}
}

// Add adds an element to a set
func (set *IntSet) Add(i int) bool {
	_, found := set.Set[i]
	set.Set[i] = true
	return !found
}

// Contains tests if an element is in a set
func (set *IntSet) Contains(i int) bool {
	_, found := set.Set[i]
	return found
}

// Stringify the set
func (set *IntSet) Stringify() string {
	values := []int{}
	for s := range set.Set {
		values = append(values, s)
	}
	sort.Ints(values)

	delim := ","
	return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(values)), delim), "[]")
}

// Length returns the length of the Set
func (set *IntSet) Length() int {
	return len(set.Set)
}

func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 1
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
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
