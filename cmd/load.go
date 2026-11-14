package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/5ouma/dorg/internal/command"
	"github.com/5ouma/dorg/internal/dock"
	"github.com/5ouma/dorg/internal/utils"
	"github.com/spf13/cobra"
)

func newLoadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "Load Dock items",
		Long:  "📂 Load Dock items and settings from YAML file",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return execLoadCmd(cmd, args, dock.LoadDockPlist, utils.RunCommand)
		},
	}
	cmd.PersistentFlags().String("file", "dorg.yml", "config file")
	cmd.PersistentFlags().BoolP("verbose", "V", false, "verbose output")
	return cmd
}

func execLoadCmd(cmd *cobra.Command, args []string, loadDock func() (*dock.Plist, error), runCommand func(context.Context, string, ...string) (string, error)) error {
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

	fmt.Println(utils.H1.Render("📁 Load Dock settings"))
	if err := command.LoadConfig(cfg, loadDock, runCommand); err != nil {
		return err
	}
	fmt.Println(utils.Msg.Render("✅ Dock settings loaded successfully"))
	return nil
}
