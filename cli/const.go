//go:build !windows

package cli

const (
	TerminalClearLine = "\r\x1b[2K"
)
