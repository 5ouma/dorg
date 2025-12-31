package dock

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/5ouma/dorg/internal/config"
	"github.com/5ouma/dorg/internal/utils"
	"howett.net/plist"
)

const (
	dockPlistPath       = "/Library/Preferences/com.apple.dock.plist"
	dockLaunchAgentID   = "com.apple.Dock.agent"
	dockLaunchAgentPath = "/System/Library/LaunchAgents/com.apple.Dock.agent.plist"
)

type Plist struct {
	PersistentApps        []PAItem `plist:"persistent-apps"`
	PersistentOthers      []POItem `plist:"persistent-others"`
	TileSize              any      `plist:"tilesize,omitempty"`
	LargeSize             any      `plist:"largesize,omitempty"`
	Magnification         bool     `plist:"magnification"`
	MinimizeToApplication bool     `plist:"minimize-to-application"`
	AutoHide              bool     `plist:"autohide"`
	ShowRecents           bool     `plist:"show-recents"`
	SizeImmutable         bool     `plist:"size-immutable"`
}

type FileData struct {
	URLString     string `plist:"_CFURLString"`
	URLStringType int    `plist:"_CFURLStringType"`
}

type TileData struct {
	FileData FileData `plist:"file-data"`
	FileType int      `plist:"file-type"`
}

func (d TileData) GetPath() string {
	out := strings.TrimPrefix(d.FileData.URLString, "file://")
	out = strings.TrimSuffix(out, "/")
	return strings.ReplaceAll(out, "%20", " ")
}

type PAItem struct {
	GUID     int      `plist:"GUID,omitempty"`
	TileType string   `plist:"tile-type"`
	TileData TileData `plist:"tile-data"`
}

type POItem struct {
	GUID     int        `plist:"GUID"`
	TileType string     `plist:"tile-type"`
	TileData POTileData `plist:"tile-data"`
}

type POTileData struct {
	Arrangement int      `plist:"arrangement"`
	DisplayAs   int      `plist:"displayas"`
	ShowAs      int      `plist:"showas"`
	FileData    FileData `plist:"file-data"`
	FileLabel   string   `plist:"file-label"`
	FileType    int      `plist:"file-type"`
	Directory   int      `plist:"directory,omitempty"`
}

func (d POTileData) GetPath() string {
	out := strings.TrimPrefix(d.FileData.URLString, "file://")
	out = strings.TrimSuffix(out, "/")
	return strings.ReplaceAll(out, "%20", " ")
}

func fileNameWithoutExtTrimSuffix(fileName string) string {
	fileName = filepath.Base(fileName)
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func LoadDockPlist() (*Plist, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %v", err)
	}

	dPlist := new(Plist)
	data, err := os.ReadFile(filepath.Join(home, dockPlistPath))
	if err != nil {
		return nil, fmt.Errorf("failed to read dock plist: %v", err)
	}
	if _, err := plist.Unmarshal(data, dPlist); err != nil {
		return nil, fmt.Errorf("failed to unmarshal dock plist: %v", err)
	}

	return dPlist, nil
}

func (p *Plist) AddApp(appPath string) {
	var paItem PAItem
	switch appPath {
	case "":
		paItem = PAItem{TileType: "small-spacer-tile"}
	case " ":
		paItem = PAItem{TileType: "spacer-tile"}
	default:
		paItem = PAItem{
			GUID:     rand.Intn(9999999999),
			TileType: "file-tile",
			TileData: TileData{FileData: FileData{URLString: appPath, URLStringType: 0}, FileType: 41},
		}
	}

	p.PersistentApps = append(p.PersistentApps, paItem)
}

func (p *Plist) AddOther(other config.Folder) error {
	path := other.Path
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	if path == "~" {
		path = home
	} else if after, ok := strings.CutPrefix(path, "~/"); ok {
		path = filepath.Join(home, after)
	} else {
		return fmt.Errorf("invalid path '%s': must be absolute or start with '~/'", other.Path)
	}
	path, err = filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for '%s': %v", path, err)
	}

	poItem := POItem{
		GUID:     rand.Intn(9999999999),
		TileType: "directory-tile",
		TileData: POTileData{
			Directory:   1,
			Arrangement: other.Sort,
			DisplayAs:   other.Display,
			ShowAs:      other.View,
			FileData:    FileData{URLString: path, URLStringType: 0},
			FileLabel:   fileNameWithoutExtTrimSuffix(other.Path),
			FileType:    2,
		},
	}

	p.PersistentOthers = append(p.PersistentOthers, poItem)

	return nil
}

func (p *Plist) ApplySettings(setting config.DockSettings) error {
	p.Magnification = setting.Magnification
	p.MinimizeToApplication = setting.MinimizeToApplication
	p.AutoHide = setting.AutoHide
	p.ShowRecents = setting.ShowRecents
	p.SizeImmutable = setting.SizeImmutable

	switch v := setting.TileSize.(type) {
	case float64:
		if v < 16 && v > 128 {
			return fmt.Errorf("tile size must be between 16 and 128: %d", setting.TileSize)
		}
	case int:
		if v < 16 && v > 128 {
			return fmt.Errorf("tile size must be between 16 and 128: %d", setting.TileSize)
		}
	}

	switch v := setting.LargeSize.(type) {
	case float64:
		if v < 16 && v > 128 {
			return fmt.Errorf("large size must be between 16 and 128: %d", setting.LargeSize)
		}
	case int:
		if v < 16 && v > 128 {
			return fmt.Errorf("large size must be between 16 and 128: %d", setting.LargeSize)
		}
	}

	return nil
}

func (p *Plist) Save() error {
	if err := p.unload(); err != nil {
		return fmt.Errorf("dock save: %w", err)
	}

	file, err := os.CreateTemp("", "dock.plist")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %v", err)
	}
	defer func() {
		if os.Remove(file.Name()) != nil {
			slog.Warn("failed to remove temp dock plist", "plist", file.Name())
		}
	}()

	slog.Debug("writing temp dock plist", "plist", file.Name())
	if err := plist.NewBinaryEncoder(file).Encode(p); err != nil {
		return fmt.Errorf("failed to decode plist: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %v", err)
	}

	if err := p.importPlist(file.Name()); err != nil {
		return fmt.Errorf("failed to import plist: %w", err)
	}
	return p.restart()
}

func (p *Plist) importPlist(path string) error {
	slog.Debug("importing dock plist")
	if _, err := utils.RunCommand(context.Background(), "/usr/bin/defaults", "import", "com.apple.dock", path); err != nil {
		return fmt.Errorf("failed to defaults import dock plist '%s': %v", path, err)
	}
	return nil
}

func (p *Plist) unload() error {
	slog.Debug("unloading Dock launch agent")
	if _, err := utils.RunCommand(context.Background(), "/bin/launchctl", "unload", dockLaunchAgentPath); err != nil {
		return fmt.Errorf("failed to unload Dock launch agent: %v", err)
	}
	return nil
}

func (p *Plist) restart() error {
	slog.Debug("restart Dock launch agent")
	if _, err := utils.RunCommand(context.Background(), "/bin/launchctl", "load", dockLaunchAgentPath); err != nil {
		return fmt.Errorf("failed to load Dock launch agent: %v", err)
	}
	if _, err := utils.RunCommand(context.Background(), "/bin/launchctl", "start", dockLaunchAgentID); err != nil {
		return fmt.Errorf("failed to start Dock launch agent: %v", err)
	}
	return nil
}

func (p *Plist) GenerateConfigFromPlist() (config.Config, error) {
	conf := new(config.Config)

	home, err := os.UserHomeDir()
	if err != nil {
		return *conf, fmt.Errorf("failed to get user home dir: %w", err)
	}

	fmt.Println(utils.H2.Render("Apps"))
	for _, item := range p.PersistentApps {
		fmt.Println(utils.CheckedItem.Render(), item.TileData.GetPath())
		conf.Dock.Apps = append(conf.Dock.Apps, item.TileData.GetPath())
	}

	fmt.Println(utils.H2.Render("Folders"))
	for _, item := range p.PersistentOthers {
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
		TileSize:              p.TileSize,
		LargeSize:             p.LargeSize,
		Magnification:         p.Magnification,
		MinimizeToApplication: p.MinimizeToApplication,
		AutoHide:              p.AutoHide,
		ShowRecents:           p.ShowRecents,
	}

	return *conf, nil
}
