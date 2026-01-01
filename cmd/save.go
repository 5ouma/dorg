package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/5ouma/dorg/internal/command"
	"github.com/5ouma/dorg/internal/utils"
	"github.com/spf13/cobra"
)

func newSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save Dock items",
		Long:  "💾 Save the current Dock items and settings to YAML file",
		Args:  cobra.NoArgs,
		RunE:  execSaveCmd,
	}
	cmd.PersistentFlags().String("file", "dorg.yml", "config file")
	cmd.PersistentFlags().BoolP("verbose", "V", false, "verbose output")
	return cmd
}

func execSaveCmd(cmd *cobra.Command, args []string) error {
	file, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}
	verbose, err := cmd.Flags().GetBool("verbose")
	if err != nil {
		return err
	}

	if verbose {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	}

	cfg := &command.Config{
		Cmd:      cmd.Use,
		File:     file,
		LogLevel: utils.SetLogLevel(verbose),
	}

	if err := cfg.Verify(); err != nil {
		return err
	}

	fmt.Println(utils.H1.Render("💾 Save Dock settings"))
	return command.SaveConfig(cfg)
}
