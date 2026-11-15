package config

import (
	"os"

	"go.yaml.in/yaml/v4"
)

// Represents the YAML structure persisted for a Dock layout.
type Config struct {
	Dock Dock `yaml:"dock_items"`
}

// Groups the app tiles, folder tiles, and global Dock settings.
type Dock struct {
	Apps     []string      `yaml:"apps,omitempty"`
	Others   []Folder      `yaml:"others,omitempty"`
	Settings *DockSettings `yaml:"settings,omitempty"`
}

// Describes a Dock folder tile and its placement metadata.
type Folder struct {
	Path    string `yaml:"path,omitempty"`
	Sort    int    `yaml:"sort,omitempty"`
	Display int    `yaml:"display,omitempty"`
	View    int    `yaml:"view,omitempty"`
}

// Stores the visual and behavior settings for the Dock.
type DockSettings struct {
	TileSize              any  `yaml:"tilesize"`
	LargeSize             any  `yaml:"largesize"`
	Magnification         bool `yaml:"magnification"`
	MinimizeToApplication bool `yaml:"minimize-to-application"`
	AutoHide              bool `yaml:"autohide"`
	ShowRecents           bool `yaml:"show-recents"`
	SizeImmutable         bool `yaml:"size-immutable"`
}

// Reads a Dock configuration from a YAML file.
func Load(file string) (Config, error) {
	conf := new(Config)

	data, err := os.ReadFile(file)
	if err != nil {
		return *conf, err
	}

	if err := yaml.Unmarshal(data, conf); err != nil {
		return *conf, err
	}

	return *conf, nil
}
