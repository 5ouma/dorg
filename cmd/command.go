package cmd

import (
	"runtime/debug"

	"github.com/5ouma/dorg/internal/utils"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	info, ok := debug.ReadBuildInfo()
	cmd := &cobra.Command{
		Use:          "dorg",
		Short:        "🚥 Organize macOS Dock Items",
		Long:         "🚥 Organize macOS Dock Items with YAML",
		Version:      utils.Version(info, ok),
		SilenceUsage: true,
	}
	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.SetVersionTemplate("🚥 {{.Use}} {{.Version}}\n")
	cmd.SetErrPrefix(" 🚨")
	cmd.AddCommand(
		newCheckCmd(),
		newLoadCmd(),
		newSaveCmd(),
	)

	return cmd
}
