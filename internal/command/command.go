package command

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/5ouma/dorg/internal/config"
	"github.com/5ouma/dorg/internal/dock"
	"github.com/5ouma/dorg/internal/utils"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Cmd      string
	File     string
	LogLevel int
}

func (c *Config) Verify() error {
	if err := os.MkdirAll(filepath.Dir(c.File), 0750); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	slog.Debug(fmt.Sprintf("📄 Using config file: %s", c.File))

	return nil
}

func SaveConfig(c *Config) (err error) {
	var conf config.Config

	dPlist, err := dock.LoadDockPlist()
	if err != nil {
		return errors.Wrap(err, "unable to load dock plist")
	}

	fmt.Println(utils.H2.Render("Apps"))
	for _, item := range dPlist.PersistentApps {
		fmt.Println(utils.CheckedItem.Render(), item.TileData.GetPath())
		conf.Dock.Apps = append(conf.Dock.Apps, item.TileData.GetPath())
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home dir: %w", err)
	}

	fmt.Println(utils.H2.Render("Folders"))
	for _, item := range dPlist.PersistentOthers {
		path := item.TileData.GetPath()
		if relPath, err := filepath.Rel(home, path); err == nil {
			path = filepath.Join("~", relPath)
		}
		fmt.Println(utils.CheckedItem.Render(), path)
		conf.Dock.Others = append(conf.Dock.Others, config.Folder{
			Path:    path,
			Sort:    item.TileData.Arrangement,
			Display: item.TileData.DisplayAs,
			View:    item.TileData.ShowAs,
		})
	}

	conf.Dock.Settings = &config.DockSettings{
		TileSize:              dPlist.TileSize,
		LargeSize:             dPlist.LargeSize,
		Magnification:         dPlist.Magnification,
		MinimizeToApplication: dPlist.MinimizeToApplication,
		AutoHide:              dPlist.AutoHide,
		ShowRecents:           dPlist.ShowRecents,
		SizeImmutable:         dPlist.SizeImmutable,
	}

	if err := os.MkdirAll(filepath.Dir(c.File), 0750); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&conf); err != nil {
		if err := enc.Close(); err != nil {
			slog.Warn("failed to close encoder after encode error", "error", err)
		}
		return errors.Wrap(err, "unable to encode YAML for logging")
	}
	if err := enc.Close(); err != nil {
		return errors.Wrap(err, "unable to close encoder")
	}
	data := buf.Bytes()
	if err := os.WriteFile(c.File, data, 0644); err != nil {
		return err
	}

	fmt.Println(utils.Msg.Render("✅", c.File))
	return nil
}

func LoadConfig(c *Config) (err error) {
	var conf config.Config

	conf, err = config.Load(c.File)
	if err != nil {
		return fmt.Errorf("failed to load config file: %v", err)
	}

	if len(conf.Dock.Apps) == 0 && len(conf.Dock.Others) == 0 && conf.Dock.Settings == nil {
		return errors.Errorf("no dock configuration found in config file")
	}

	dPlist, err := dock.LoadDockPlist()
	if err != nil {
		return errors.Wrap(err, "unable to load dock plist")
	}

	if len(dPlist.PersistentApps) > 0 {
		dPlist.PersistentApps = nil
	}
	fmt.Println(utils.H2.Render("Apps"))
	for _, app := range conf.Dock.Apps {
		fmt.Println(utils.CheckedItem.Render(), app)
		dPlist.AddApp(app)
	}

	if len(dPlist.PersistentOthers) > 0 {
		dPlist.PersistentOthers = nil
	}
	fmt.Println(utils.H2.Render("Folders"))
	for _, other := range conf.Dock.Others {
		fmt.Println(utils.CheckedItem.Render(), other.Path)
		if err := dPlist.AddOther(other); err != nil {
			return errors.Wrapf(err, "unable to add other %s", other.Path)
		}
	}

	if conf.Dock.Settings != nil {
		if err := dPlist.ApplySettings(*conf.Dock.Settings); err != nil {
			return fmt.Errorf("failed to apply dock settings: %w", err)
		}
	}

	if err := dPlist.Save(); err != nil {
		return fmt.Errorf("failed to save dock plist: %w", err)
	}

	return utils.RestartDock()
}
