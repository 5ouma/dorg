package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/pkg/errors"
)

// Executes a command and returns its stdout output and error.
func RunCommand(ctx context.Context, cmd string, args ...string) (string, error) {
	c := new(exec.Cmd)

	if ctx != nil {
		c = exec.CommandContext(ctx, cmd, args...)
	} else {
		c = exec.Command(cmd, args...)
	}

	output, err := c.Output()
	if err != nil {
		return string(output), err
	}

	if ctx != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("command %s timed out", cmd)
		}
	}

	return string(output), nil
}

// Restarts the Dock process so Dock changes are applied.
func RestartDock(runCommand func(context.Context, string, ...string) (string, error)) error {
	slog.Debug("restarting Dock")
	if _, err := runCommand(context.Background(), "killall", "Dock"); err != nil {
		return errors.Wrap(err, "killing Dock process failed")
	}
	time.Sleep(2 * time.Second)
	return nil
}
