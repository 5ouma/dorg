package cmd

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/5ouma/dorg/internal/dock"
)

func Test_execLoadCmd(t *testing.T) {
	tests := map[string]struct {
		content     string
		loadErr     error
		failCommand string
		verbose     bool
		wantErr     bool
	}{
		"success": {
			content: "dock_items:\n  apps: [/Applications/Calculator.app]\n",
			verbose: true,
		},
		"load fails": {
			content: "dock_items:\n  apps: [/Applications/Calculator.app]\n",
			loadErr: errors.New("injected load error"),
			wantErr: true,
		},
		"save fails": {
			content:     "dock_items:\n  apps: [/Applications/Calculator.app]\n",
			failCommand: "unload",
			wantErr:     true,
		},
		"restart fails": {
			content:     "dock_items:\n  apps: [/Applications/Calculator.app]\n",
			failCommand: "restart",
			wantErr:     true,
		},
		"verify fails": {wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			file := filepath.Join(t.TempDir(), "config.yml")
			if name == "verify fails" {
				parent := filepath.Join(t.TempDir(), "file")
				if err := os.WriteFile(parent, nil, 0600); err != nil {
					t.Fatal(err)
				}
				file = filepath.Join(parent, "child", "config.yml")
			} else if err := os.WriteFile(file, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			loadDock := func() (*dock.Plist, error) {
				if tc.loadErr != nil {
					return nil, tc.loadErr
				}
				return &dock.Plist{}, nil
			}
			runCommand := func(_ context.Context, cmd string, args ...string) (string, error) {
				op := cmd
				switch cmd {
				case "/bin/launchctl", "/usr/bin/defaults":
					op = args[0]
				case "killall":
					op = "restart"
				}
				if op == tc.failCommand {
					return "", errors.New("injected command error")
				}
				return "", nil
			}

			cmd := newLoadCmd()
			cmd.Flags().AddFlagSet(cmd.PersistentFlags())
			if err := cmd.Flags().Set("file", file); err != nil {
				t.Fatal(err)
			}
			if err := cmd.Flags().Set("verbose", strconv.FormatBool(tc.verbose)); err != nil {
				t.Fatal(err)
			}
			originalLogger := slog.Default()
			defer slog.SetDefault(originalLogger)

			err := execLoadCmd(cmd, nil, loadDock, runCommand)
			if (err != nil) != tc.wantErr {
				t.Fatalf("execLoadCmd() error = %v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
