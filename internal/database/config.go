package database

import (
	"os"

	"github.com/apex/log"
	"github.com/blacktop/lporg/internal/utils"
	yaml "gopkg.in/yaml.v3"
)

// Config is the Dock config
type Config struct {
	Dock Dock `yaml:"dock_items" json:"dock_items,omitempty" mapstructure:"dock_items"`
}

// Folder is a launchpad folder object used for Dock 'others'
type Folder struct {
	Path    string `yaml:"path,omitempty" json:"path,omitempty"`
	Display int    `yaml:"display,omitempty" json:"display,omitempty"`
	View    int    `yaml:"view,omitempty" json:"view,omitempty"`
	Sort    int    `yaml:"sort,omitempty" json:"sort,omitempty"`
}

// DockSettings is the launchpad dock settings object
type DockSettings struct {
	AutoHide              bool `yaml:"autohide" json:"autohide,omitempty"`
	LargeSize             any  `yaml:"largesize" json:"largesize,omitempty"`
	Magnification         bool `yaml:"magnification" json:"magnification,omitempty"`
	MinimizeToApplication bool `yaml:"minimize-to-application" json:"minimize-to-application,omitempty"`
	MruSpaces             bool `yaml:"mru-spaces" json:"mru-spaces,omitempty"`
	ShowRecents           bool `yaml:"show-recents" json:"show-recents,omitempty"`
	TileSize              any  `yaml:"tilesize" json:"tilesize,omitempty"`
}

// Dock is the launchpad dock config object
type Dock struct {
	Apps     []string      `yaml:"apps,omitempty" json:"apps,omitempty"`
	Others   []Folder      `yaml:"others,omitempty" json:"others,omitempty"`
	Settings *DockSettings `yaml:"settings,omitempty" json:"settings,omitempty"`
}

// LoadConfig loads the Dock config from the config file
func LoadConfig(filename string) (Config, error) {
	var conf Config

	utils.Indent(log.WithField("path", filename).Info, 2)("parsing launchpad config YAML")
	data, err := os.ReadFile(filename)
	if err != nil {
		utils.Indent(log.WithError(err).WithField("path", filename).Fatal, 3)("config file not found")
		return conf, err
	}

	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		utils.Indent(log.WithError(err).WithField("path", filename).Fatal, 3)("unmarshalling yaml failed")
		return conf, err
	}

	return conf, nil
}
