package libgobuster

import (
	"fmt"
	"runtime/debug"
	"strconv"
)

const (
	// VERSION contains the current gobuster version
	VERSION = "3.7"
)

func GetVersion() string {
	modified := false
	revision := ""
	time := ""
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			if setting.Key == "vcs.revision" {
				revision = setting.Value
			}
			if setting.Key == "vcs.time" {
				time = setting.Value
			}
			if setting.Key == "vcs.modified" {
				if mod, err := strconv.ParseBool(setting.Value); err == nil {
					modified = mod
				}
			}
		}
	}
	version := VERSION
	if revision != "" {
		version = fmt.Sprintf("%s Revision %s", version, revision)
	}

	if modified {
		version = fmt.Sprintf("%s [DIRTY]", version)
	}
	if time != "" {
		version = fmt.Sprintf("%s from %s", version, time)
	}

	return version
}
