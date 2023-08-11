//go:build !windows

package cli

const (
	TERMINAL_CLEAR_LINE = "\r\x1b[2K"
)
