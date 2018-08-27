// +build !windows

package libgobuster

func resetTerminal() string {
	return "\r\x1b[2K"
}
