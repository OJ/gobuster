package libgobuster

import (
	"fmt"
	"runtime/debug"
)

const (
	// VERSION contains the current gobuster version
	VERSION = "3.7"
)

func GetVersion() string {
	version := VERSION
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				version = fmt.Sprintf("%s Revision %s", version, setting.Value)
			}
			if setting.Key == "vcs.time" {
				version = fmt.Sprintf("%s from %s", version, setting.Value)
			}
		}
	}
	return version
}
