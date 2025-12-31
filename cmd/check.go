package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/5ouma/dorg/internal/config"
	"github.com/5ouma/dorg/internal/dock"
	"github.com/5ouma/dorg/internal/utils"
	"github.com/spf13/cobra"
)

func newCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check Login Items",
		Long:  "🔍 Check the Login Items are up-to-date",
		Args:  cobra.NoArgs,
		RunE:  execCheckCmd,
	}
	cmd.PersistentFlags().String("file", "dorg.yml", "config file")
	cmd.PersistentFlags().BoolP("verbose", "V", false, "verbose output")
	return cmd
}

func execCheckCmd(cmd *cobra.Command, args []string) error {
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

	fmt.Println(utils.H1.Render("🔍 Check Login Items"))

	cfg, err := loadFileConfig(file)
	if err != nil {
		return err
	}

	plistCfg, err := loadPlistConfig()
	if err != nil {
		return err
	}

	if !bytes.Equal(cfg, plistCfg) {
		return fmt.Errorf("dock items are out-of-date")
	}

	fmt.Println(utils.Msg.Render("✅ Dock Items are up-to-date!"))
	return nil
}

func loadFileConfig(path string) ([]byte, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func loadPlistConfig() ([]byte, error) {
	plist, err := dock.LoadDockPlist()
	if err != nil {
		return nil, err
	}

	cfg, err := plist.GenerateConfigFromPlist()
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return nil, err
	}

	return data, nil
}
