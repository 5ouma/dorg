package cmd

import (
	"github.com/5ouma/dorg/internal/utils"
	"github.com/spf13/cobra"
)

var (
	AppVersion   string
	AppBuildTime string
)

func New() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "dorg",
		Short:        "🚥 Organize macOS Dock Items",
		Long:         "🚥 Organize macOS Dock Items with YAML",
		Version:      utils.Version(),
		SilenceUsage: true,
	}
	cmd.CompletionOptions.HiddenDefaultCmd = true
	cmd.SetVersionTemplate("🚥 {{.Use}} {{.Version}}\n")
	cmd.SetErrPrefix(" 🚨")
	cmd.AddCommand(
		newLoadCmd(),
		newSaveCmd(),
	)

	return cmd
}
