package cmd

import (
	"runtime/debug"
	"testing"

	"github.com/5ouma/dorg/internal/dock"
	"github.com/5ouma/dorg/internal/utils"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

func Test_New_has_subcommands_and_version(t *testing.T) {
	t.Parallel()

	cmd := New()
	if cmd == nil {
		t.Fatal("New() returned nil")
	}
	info, ok := debug.ReadBuildInfo()
	if cmd.Version != utils.Version(info, ok) {
		t.Fatalf("version mismatch: got %s want %s", cmd.Version, utils.Version(info, ok))
	}
	names := map[string]bool{}
	for _, c := range cmd.Commands() {
		names[c.Name()] = true
	}
	if !names["check"] || !names["load"] || !names["save"] {
		t.Fatalf("expected check, load, and save subcommands present")
	}
}

func Test_CommandFlags(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		makeCmd func() *cobra.Command
	}{
		"check": {makeCmd: newCheckCmd},
		"load":  {makeCmd: newLoadCmd},
		"save":  {makeCmd: newSaveCmd},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c := tc.makeCmd()
			if f := c.PersistentFlags().Lookup("file"); f == nil || f.DefValue != "dorg.yml" {
				t.Fatalf("file flag missing or default changed: %v", f)
			}
			if f := c.PersistentFlags().Lookup("verbose"); f == nil {
				t.Fatal("verbose flag missing")
			}
		})
	}
}

func Test_RunEHandlers(t *testing.T) {
	tests := map[string]func() *cobra.Command{
		"check": newCheckCmd,
		"load":  newLoadCmd,
		"save":  newSaveCmd,
	}
	for name, makeCmd := range tests {
		t.Run(name, func(t *testing.T) {
			cmd := makeCmd()
			if err := cmd.RunE(cmd, nil); err == nil {
				t.Fatal("expected missing local flag error")
			}
		})
	}
}

func Test_execCommandFlagErrors(t *testing.T) {
	tests := map[string]struct {
		exec func(*cobra.Command, []string) error
	}{
		"check": {
			exec: func(cmd *cobra.Command, args []string) error {
				return execCheckCmd(cmd, args, dock.LoadDockPlist)
			},
		},
		"load": {
			exec: func(cmd *cobra.Command, args []string) error {
				return execLoadCmd(cmd, args, dock.LoadDockPlist, utils.RunCommand)
			},
		},
		"save": {
			exec: func(cmd *cobra.Command, args []string) error {
				return execSaveCmd(cmd, args, dock.LoadDockPlist, (*yaml.Encoder).Close)
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if err := tc.exec(&cobra.Command{}, nil); err == nil {
				t.Fatal("expected missing file flag error")
			}

			cmd := &cobra.Command{}
			cmd.Flags().String("file", "dorg.yml", "config file")
			if err := tc.exec(cmd, nil); err == nil {
				t.Fatal("expected missing verbose flag error")
			}
		})
	}
}
