package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/5ouma/dorg/internal/command"
	"github.com/5ouma/dorg/internal/dock"
	"github.com/5ouma/dorg/internal/utils"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
)

func newSaveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "save",
		Short: "Save Dock items",
		Long:  "💾 Save the current Dock items and settings to YAML file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return execSaveCmd(cmd, args, dock.LoadDockPlist, (*yaml.Encoder).Close)
		},
	}
	cmd.PersistentFlags().String("file", "dorg.yml", "config file")
	cmd.PersistentFlags().BoolP("verbose", "V", false, "verbose output")
	return cmd
}

func execSaveCmd(cmd *cobra.Command, args []string, loadDock func() (*dock.Plist, error), closeEncoder func(*yaml.Encoder) error) error {
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
	return command.SaveConfig(cfg, loadDock, closeEncoder)
}
