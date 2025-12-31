package cmd

import (
	"testing"

	"github.com/5ouma/dorg/internal/utils"
	"github.com/spf13/cobra"
)

func Test_New_has_subcommands_and_version(t *testing.T) {
	cmd := New()
	if cmd == nil {
		t.Fatalf("New() returned nil")
	}
	if cmd.Version != utils.Version() {
		t.Fatalf("version mismatch: got %s want %s", cmd.Version, utils.Version())
	}
	names := map[string]bool{}
	for _, c := range cmd.Commands() {
		names[c.Name()] = true
	}
	if !names["load"] || !names["save"] {
		t.Fatalf("expected load and save subcommands present")
	}
}

func Test_CommandFlags(t *testing.T) {
	tests := map[string]struct {
		makeCmd func() *cobra.Command
	}{
		"load": {makeCmd: newLoadCmd},
		"save": {makeCmd: newSaveCmd},
	}
	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c := tc.makeCmd()
			if c == nil {
				t.Fatalf("%s returned nil", name)
			}
			if f := c.PersistentFlags().Lookup("file"); f == nil || f.DefValue != "dorg.yml" {
				t.Fatalf("file flag missing or default changed: %v", f)
			}
			if f := c.PersistentFlags().Lookup("verbose"); f == nil {
				t.Fatalf("verbose flag missing")
			}
		})
	}
}

func Test_execCommands(t *testing.T) {
	tests := map[string]struct {
		makeCmd func() *cobra.Command
		execFn  func(*cobra.Command, []string) error
	}{
		"load": {makeCmd: newLoadCmd, execFn: execLoadCmd},
		"save": {makeCmd: newSaveCmd, execFn: execSaveCmd},
	}
	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			c := tc.makeCmd()
			if err := c.PersistentFlags().Set("file", ""); err != nil {
				t.Fatalf("failed to set flag: %v", err)
			}
			if err := c.PersistentFlags().Set("verbose", "false"); err != nil {
				t.Fatalf("failed to set verbose flag: %v", err)
			}
			if err := tc.execFn(c, nil); err == nil {
				t.Fatalf("expected error from %s exec with missing resource, got nil", name)
			}
		})
	}
}
