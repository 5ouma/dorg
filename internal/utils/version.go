package utils

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// Returns the build version for the current dorg binary.
func Version(info *debug.BuildInfo, ok bool) string {
	if !ok {
		return "unknown"
	}
	versions := strings.Split(info.Main.Version, "-")
	if len(versions) < 3 {
		if versions[0] == "(devel)" {
			return "unknown"
		}
		return versions[0]
	}
	return fmt.Sprintf("%s (#%s)", versions[0], versions[2][:7])
}
