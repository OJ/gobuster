// Redirect error

package main

import "fmt"

type redirectError struct {
	StatusCode int
}

func (e *redirectError) Error() string {
	return fmt.Sprintf("Redirect code: %d", e.StatusCode)
}
