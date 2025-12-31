package utils

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"runtime/debug"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/pkg/errors"
)

var (
	heading = lipgloss.NewStyle().
		Foreground(lipgloss.CompleteColor{TrueColor: "#007aff", ANSI256: "27"}).
		Bold(true).
		Padding(1)
	H1 = heading
	H2 = heading.SetString("▌")

	Msg = lipgloss.NewStyle().
		Bold(true).
		Padding(1)

	item = lipgloss.NewStyle().
		PaddingLeft(2)
	CheckedItem = item.
			Foreground(lipgloss.CompleteColor{TrueColor: "#63b946", ANSI256: "41"}).
			SetString("✔︎")
)

func RunCommand(ctx context.Context, cmd string, args ...string) (string, error) {
	var c *exec.Cmd

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

func SetLogLevel(verbose bool) int {
	if verbose {
		return int(slog.LevelDebug)
	}
	return int(slog.LevelWarn)
}

func RestartDock() error {
	slog.Debug("restarting Dock")
	if _, err := RunCommand(context.Background(), "killall", "Dock"); err != nil {
		return errors.Wrap(err, "killing Dock process failed")
	}
	time.Sleep(2 * time.Second)
	return nil
}

func Version() string {
	info, ok := debug.ReadBuildInfo()
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
