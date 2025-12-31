// Package command provides the command line interface functionality for lporg.
package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/apex/log"
	"github.com/blacktop/lporg/internal/database"
	"github.com/blacktop/lporg/internal/dock"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

const bold = "\033[1m%s\033[0m"

// Config is the command config
type Config struct {
	Cmd      string
	File     string
	Cloud    bool
	Backup   bool
	LogLevel int
}

// Verify will verify the command config
func (c *Config) Verify() error {
	if c.Cloud && len(c.File) > 0 {
		return fmt.Errorf("cannot use --config with --icloud")
	}

	switch c.Cmd {
	case "revert":
		if c.Cloud {
			iCloudPath, err := getiCloudDrivePath()
			if err != nil {
				return fmt.Errorf("get iCloud drive path failed")
			}
			host, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("failed to get hostname")
			}
			c.File = filepath.Join(iCloudPath, ".config", "lporg", strings.TrimRight(host, ".local")+".yml.bak")
		} else {
			if len(c.File) == 0 { // set DEFAULT config file
				confDir, err := os.UserConfigDir()
				if err != nil {
					return fmt.Errorf("failed to get user config dir")
				}
				c.File = filepath.Join(confDir, "lporg", "config.yml.bak")
			}
		}
	case "load":
		if len(c.File) == 0 && !c.Cloud {
			return fmt.Errorf("must supply --config file OR use --icloud")
		}
		fallthrough
	default:
		if c.Cloud { // use iCloud to store config
			iCloudPath, err := getiCloudDrivePath()
			if err != nil {
				return fmt.Errorf("get iCloud drive path failed")
			}
			host, err := os.Hostname()
			if err != nil {
				return fmt.Errorf("failed to get hostname")
			}
			c.File = filepath.Join(iCloudPath, ".config", "lporg", strings.TrimRight(host, ".local")+".yml")
		} else {
			if len(c.File) == 0 { // set DEFAULT config file
				confDir, err := os.UserConfigDir()
				if err != nil {
					return fmt.Errorf("failed to get user config dir")
				}
				c.File = filepath.Join(confDir, "lporg", "config.yml")
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(c.File), 0750); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	log.Info("using config file: " + c.File)

	return nil
}

func SaveConfig(c *Config) (err error) {
	var conf database.Config

	dPlist, err := dock.LoadDockPlist()
	if err != nil {
		return errors.Wrap(err, "unable to load dock plist")
	}

	home, _ := os.UserHomeDir()

	for _, item := range dPlist.PersistentApps {
		conf.Dock.Apps = append(conf.Dock.Apps, item.TileData.GetPath())
	}
	for _, item := range dPlist.PersistentOthers {
		abspath := item.TileData.GetPath()
		if relPath, err := filepath.Rel(home, abspath); err == nil {
			abspath = filepath.Join("~", relPath)
		}
		conf.Dock.Others = append(conf.Dock.Others, database.Folder{
			Path:    abspath,
			Display: int(item.TileData.DisplayAs),
			View:    int(item.TileData.ShowAs),
			Sort:    int(item.TileData.Arrangement),
		})
	}
	conf.Dock.Settings = &database.DockSettings{
		AutoHide:              dPlist.AutoHide,
		LargeSize:             dPlist.LargeSize,
		Magnification:         dPlist.Magnification,
		MinimizeToApplication: dPlist.MinimizeToApplication,
		MruSpaces:             dPlist.MruSpaces,
		ShowRecents:           dPlist.ShowRecents,
		TileSize:              dPlist.TileSize,
	}

	if err := os.MkdirAll(filepath.Dir(c.File), 0750); err != nil {
		return fmt.Errorf("failed to create config dir: %w", err)
	}

	f, err := os.Create(c.File)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	// write out config YAML file
	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(&conf); err != nil {
		return errors.Wrap(err, "unable to marshal YAML")
	}
	if err := enc.Close(); err != nil {
		return errors.Wrap(err, "unable to close YAML encoder")
	}

	log.Infof(bold, "successfully wrote settings to: "+c.File)
	return nil
}

// LoadConfig will apply dock settings from a config file
func LoadConfig(c *Config) (err error) {
	var conf database.Config

	conf, err = database.LoadConfig(c.File)
	if err != nil {
		return fmt.Errorf("failed to load config file: %v", err)
	}

	if len(conf.Dock.Apps) == 0 && len(conf.Dock.Others) == 0 && conf.Dock.Settings == nil {
		log.Info("no dock configuration found in config file")
		return nil
	}

	dPlist, err := dock.LoadDockPlist()
	if err != nil {
		return errors.Wrap(err, "unable to load dock plist")
	}

	if len(dPlist.PersistentApps) > 0 {
		dPlist.PersistentApps = nil
	}
	for _, app := range conf.Dock.Apps {
		log.WithField("app", app).Info("adding to dock")
		dPlist.AddApp(app)
	}

	if len(dPlist.PersistentOthers) > 0 {
		dPlist.PersistentOthers = nil
	}
	for _, other := range conf.Dock.Others {
		log.WithField("other", other.Path).Info("adding to dock")
		dPlist.AddOther(other)
	}

	if conf.Dock.Settings != nil {
		if err := dPlist.ApplySettings(*conf.Dock.Settings); err != nil {
			return fmt.Errorf("failed to apply dock settings: %w", err)
		}
	}

	if err := dPlist.Save(); err != nil {
		return fmt.Errorf("failed to save dock plist: %w", err)
	}

	return restartDock()
}

// DefaultOrg is removed; keep stub for compatibility
func DefaultOrg(c *Config) (err error) {
	return fmt.Errorf("Default organization (apps/widgets/desktop) removed; dock-only mode")
}
