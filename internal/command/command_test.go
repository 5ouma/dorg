package command

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/5ouma/dorg/internal/dock"
	"go.yaml.in/yaml/v4"
)

func Test_Verify(t *testing.T) {
	t.Parallel()

	cfg := &Config{File: filepath.Join(t.TempDir(), "subdir", "cfg.yml")}
	if err := cfg.Verify(); err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if _, err := os.Stat(filepath.Dir(cfg.File)); os.IsNotExist(err) {
		t.Fatal("expected directory to be created")
	}
}

func Test_Verify_createDirectoryError(t *testing.T) {
	dir := t.TempDir()
	blockingFile := filepath.Join(dir, "file")
	if err := os.WriteFile(blockingFile, nil, 0600); err != nil {
		t.Fatal(err)
	}
	cfg := &Config{File: filepath.Join(blockingFile, "child", "config.yml")}
	if err := cfg.Verify(); err == nil {
		t.Fatal("Verify() expected a directory creation error")
	}
}

func Test_SaveConfig(t *testing.T) {
	tests := map[string]struct {
		loadErr     error
		encodeError bool
		closeErr    error
		blockDir    bool
		emptyHome   bool
		wantErr     bool
		wantWritten bool
	}{
		"writes yaml":         {wantWritten: true},
		"load fails":          {loadErr: errors.New("injected load error"), wantErr: true},
		"generation fails":    {emptyHome: true, wantErr: true},
		"yaml encoding fails": {encodeError: true, wantErr: true},
		"yaml close fails":    {closeErr: errors.New("injected close error"), wantErr: true},
		"create dir fails":    {blockDir: true, wantErr: true},
		"write fails":         {wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.emptyHome {
				t.Setenv("HOME", "")
			} else {
				t.Setenv("HOME", t.TempDir())
			}
			loadDock := func() (*dock.Plist, error) {
				if tc.loadErr != nil {
					return nil, tc.loadErr
				}
				p := &dock.Plist{
					PersistentApps: []dock.PAItem{{
						TileType: "file-tile",
						TileData: dock.TileData{FileData: dock.FileData{URLString: "/Applications/Calculator.app"}},
					}},
				}
				if tc.encodeError {
					p.TileSize = yamlEncodeError{}
				}
				return p, nil
			}
			closeEncoder := (*yaml.Encoder).Close
			if tc.closeErr != nil {
				closeEncoder = func(*yaml.Encoder) error { return tc.closeErr }
			}

			dir := t.TempDir()
			file := filepath.Join(dir, "config.yml")
			switch {
			case tc.blockDir:
				file = filepath.Join(dir, "existing", "nested", "config.yml")
				if err := os.WriteFile(filepath.Join(dir, "existing"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			case name == "write fails":
				file = dir
			}

			err := SaveConfig(&Config{File: file}, loadDock, closeEncoder)
			if (err != nil) != tc.wantErr {
				t.Fatalf("SaveConfig() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if tc.wantWritten {
				data, err := os.ReadFile(file)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(data), "/Applications/Calculator.app") {
					t.Fatalf("saved YAML missing app path: %s", data)
				}
			}
		})
	}
}

func Test_LoadConfig(t *testing.T) {
	tests := map[string]struct {
		content      string
		loadErr      error
		failCommand  string
		wantErr      bool
		wantApps     int
		wantOthers   int
		wantCommands []string
	}{
		"missing config": {wantErr: true},
		"empty config":   {content: "dock_items: {}\n", wantErr: true},
		"dock load fails": {
			content: "dock_items:\n  apps: [/Applications/Calculator.app]\n",
			loadErr: errors.New("injected load error"),
			wantErr: true,
		},
		"invalid folder": {
			content: "dock_items:\n  others:\n    - path: relative\n",
			wantErr: true,
		},
		"save fails": {
			content:      "dock_items:\n  apps: [/Applications/Calculator.app]\n",
			failCommand:  "unload",
			wantErr:      true,
			wantCommands: []string{"unload"},
		},
		"restart fails": {
			content:      "dock_items:\n  apps: [/Applications/Calculator.app]\n",
			failCommand:  "restart",
			wantErr:      true,
			wantCommands: []string{"unload", "import", "load", "start", "restart"},
		},
		"loads apps folders and settings": {
			content:      "dock_items:\n  apps: [/Applications/Calculator.app]\n  others:\n    - path: ~/Documents\n      sort: 1\n      display: 2\n      view: 3\n  settings:\n    tilesize: 32\n    largesize: 64\n    magnification: true\n",
			wantApps:     1,
			wantOthers:   1,
			wantCommands: []string{"unload", "import", "load", "start", "restart"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			plist := &dock.Plist{
				PersistentApps:   []dock.PAItem{{GUID: 1}},
				PersistentOthers: []dock.POItem{{GUID: 2}},
			}
			loadDock := func() (*dock.Plist, error) {
				if tc.loadErr != nil {
					return nil, tc.loadErr
				}
				return plist, nil
			}
			var commands []string
			runCommand := func(_ context.Context, cmd string, args ...string) (string, error) {
				op := cmd
				switch cmd {
				case "/bin/launchctl", "/usr/bin/defaults":
					op = args[0]
				case "killall":
					op = "restart"
				}
				commands = append(commands, op)
				if op == tc.failCommand {
					return "", errors.New("injected command error")
				}
				return "", nil
			}

			file := filepath.Join(t.TempDir(), "config.yml")
			if tc.content != "" {
				if err := os.WriteFile(file, []byte(tc.content), 0600); err != nil {
					t.Fatal(err)
				}
			}

			err := LoadConfig(&Config{File: file}, loadDock, runCommand)
			if (err != nil) != tc.wantErr {
				t.Fatalf("LoadConfig() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if !reflect.DeepEqual(commands, tc.wantCommands) {
				t.Fatalf("commands = %v, want %v", commands, tc.wantCommands)
			}
			if !tc.wantErr {
				if len(plist.PersistentApps) != tc.wantApps || len(plist.PersistentOthers) != tc.wantOthers {
					t.Fatalf("apps=%d, others=%d; want apps=%d, others=%d", len(plist.PersistentApps), len(plist.PersistentOthers), tc.wantApps, tc.wantOthers)
				}
				if tc.wantApps > 0 && !plist.Magnification {
					t.Fatalf("settings were not applied: %+v", plist)
				}
			}
		})
	}
}

type yamlEncodeError struct{}

func (yamlEncodeError) MarshalYAML() (any, error) {
	return nil, errors.New("injected YAML encoding error")
}
