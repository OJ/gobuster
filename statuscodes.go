// status codes

package main

import (
	"strconv"
	"strings"
)

type statuscodes struct {
	sc map[int]bool
}

func (set *statuscodes) add(i int) {
	set.sc[i] = true
}

func (set *statuscodes) contains(i int) bool {
	_, found := set.sc[i]
	return found
}

func (set *statuscodes) string() string {
	values := []string{}
	for s := range set.sc {
		values = append(values, strconv.Itoa(s))
	}
	return strings.Join(values, ",")
}
