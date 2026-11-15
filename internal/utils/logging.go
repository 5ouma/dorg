package utils

import "log/slog"

// Configures the logging level for the current CLI session.
func SetLogLevel(verbose bool) int {
	if verbose {
		return int(slog.LevelDebug)
	}
	return int(slog.LevelWarn)
}
