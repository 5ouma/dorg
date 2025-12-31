package config

import (
	"os"

	yaml "gopkg.in/yaml.v3"
)

type Config struct {
	Dock Dock `yaml:"dock_items"`
}

type Dock struct {
	Apps     []string      `yaml:"apps,omitempty"`
	Others   []Folder      `yaml:"others,omitempty"`
	Settings *DockSettings `yaml:"settings,omitempty"`
}

type Folder struct {
	Path    string `yaml:"path,omitempty"`
	Sort    int    `yaml:"sort,omitempty"`
	Display int    `yaml:"display,omitempty"`
	View    int    `yaml:"view,omitempty"`
}

type DockSettings struct {
	TileSize              any  `yaml:"tilesize"`
	LargeSize             any  `yaml:"largesize"`
	Magnification         bool `yaml:"magnification"`
	MinimizeToApplication bool `yaml:"minimize-to-application"`
	AutoHide              bool `yaml:"autohide"`
	ShowRecents           bool `yaml:"show-recents"`
	SizeImmutable         bool `yaml:"size-immutable"`
}

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
