package utils

import "log/slog"

func SetLogLevel(verbose bool) int {
	if verbose {
		return int(slog.LevelDebug)
	}
	return int(slog.LevelWarn)
}
