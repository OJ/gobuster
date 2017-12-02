// IP wildcards

package main

import "strings"

type ipwildcards struct {
	ipw map[string]bool
}

func (ipw *ipwildcards) add(s string) {
	ipw.ipw[s] = true
}

func (ipw *ipwildcards) addRange(ss []string) {
	for _, s := range ss {
		ipw.ipw[s] = true
	}
}

func (ipw *ipwildcards) contains(s string) bool {
	_, found := ipw.ipw[s]
	return found
}

func (ipw *ipwildcards) containsAny(ss []string) bool {
	for _, s := range ss {
		if ipw.ipw[s] {
			return true
		}
	}
	return false
}

func (ipw *ipwildcards) string() string {
	values := []string{}
	for s := range ipw.ipw {
		values = append(values, s)
	}
	return strings.Join(values, ",")
}
