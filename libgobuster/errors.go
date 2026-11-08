package libgobuster

import "errors"

var (
	ErrTimeout           = errors.New("timeout occurred during the request")
	ErrEOF               = errors.New("server closed connection without sending any data back. Maybe you are connecting via https to on http port or vice versa?")
	ErrConnectionRefused = errors.New("connection refused")
	ErrRateLimited       = errors.New("server responded with HTTP 429 Too Many Requests")
)
