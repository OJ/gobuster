package libgobuster

import (
	"bytes"
	"errors"
	"fmt"
	"io"
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
