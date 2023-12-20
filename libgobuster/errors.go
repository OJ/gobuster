package libgobuster

import "errors"

var (
	ErrorTimeout           = errors.New("timeout occurred during the request")
	ErrorEOF               = errors.New("server closed connection without sending any data back. Maybe you are connecting via https to on http port or vice versa?")
	ErrorConnectionRefused = errors.New("connection refused")
)
