package cmd

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/5ouma/dorg/internal/config"
	"github.com/5ouma/dorg/internal/dock"
	"howett.net/plist"
)

func Test_loadFileConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		content   string
		path      string
		wantError bool
	}{
		"valid file":        {content: `dock_items: {apps: ["/Applications/Calculator.app"]}`, wantError: false},
		"missing file":      {path: "/non/existing/path.yml", wantError: true},
		"invalid yaml":      {content: "dock_items: [unclosed", wantError: true},
		"non-finite number": {content: "dock_items:\n  settings:\n    tilesize: .nan", wantError: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var path string
			if tc.content != "" {
				path = filepath.Join(t.TempDir(), "dorg.yml")
				if err := os.WriteFile(path, []byte(tc.content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
			} else {
				path = tc.path
			}

			data, err := loadFileConfig(path)
			if (err != nil) != tc.wantError {
				t.Fatalf("%s error=%v, wantErr=%v", path, err, tc.wantError)
			}
			if err == nil {
				var got config.Config
				if err := json.Unmarshal(data, &got); err != nil {
					t.Fatalf("failed to unmarshal json: %v", err)
				}
				if len(got.Dock.Apps) != 1 || got.Dock.Apps[0] != "/Applications/Calculator.app" {
					t.Fatalf("unexpected apps: %#v", got.Dock.Apps)
				}
			}
		})
	}
}

func Test_loadPlistConfig(t *testing.T) {
	tests := map[string]struct {
		plist     *dock.Plist
		loadErr   error
		emptyHome bool
		wantErr   bool
	}{
		"valid plist": {
			plist: &dock.Plist{
				PersistentApps: []dock.PAItem{{
					TileData: dock.TileData{FileData: dock.FileData{URLString: "file:///Applications/Calculator.app/"}},
				}},
				PersistentOthers: []dock.POItem{{
					TileData: dock.POTileData{
						Arrangement: 1,
						DisplayAs:   2,
						ShowAs:      3,
						FileData:    dock.FileData{URLString: "file:///Users/test/Documents/"},
					},
				}},
				TileSize: 32, LargeSize: 64, Magnification: true,
			},
		},
		"load fails": {
			loadErr: errors.New("injected load error"),
			wantErr: true,
		},
		"generation fails": {
			plist:     &dock.Plist{},
			emptyHome: true,
			wantErr:   true,
		},
		"json encoding fails": {
			plist:   &dock.Plist{TileSize: make(chan int)},
			wantErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.emptyHome {
				t.Setenv("HOME", "")
			} else {
				t.Setenv("HOME", t.TempDir())
			}
			loadDock := func() (*dock.Plist, error) {
				return tc.plist, tc.loadErr
			}

			data, err := loadPlistConfig(loadDock)
			if (err != nil) != tc.wantErr {
				t.Fatalf("loadPlistConfig() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err == nil {
				var got config.Config
				if err := json.Unmarshal(data, &got); err != nil {
					t.Fatalf("failed to unmarshal json: %v", err)
				}
				if len(got.Dock.Apps) != 1 || got.Dock.Apps[0] != "/Applications/Calculator.app" {
					t.Fatalf("unexpected apps: %#v", got.Dock.Apps)
				}
				if len(got.Dock.Others) != 1 {
					t.Fatalf("unexpected others length: %d", len(got.Dock.Others))
				}
				o := got.Dock.Others[0]
				if filepath.Base(o.Path) != "Documents" || o.Sort != 1 || o.Display != 2 || o.View != 3 {
					t.Fatalf("unexpected other: %#v", o)
				}
			}
		})
	}
}

func Test_execCheckCmd(t *testing.T) {
	tests := map[string]struct {
		config  string
		plist   bool
		wantErr bool
	}{
		"matching config": {
			config: checkConfig("/Applications/Calculator.app"),
			plist:  true,
		},
		"out of date config": {
			config:  checkConfig("/Applications/Notes.app"),
			plist:   true,
			wantErr: true,
		},
		"config missing": {
			wantErr: true,
		},
		"plist missing": {
			config:  checkConfig("/Applications/Calculator.app"),
			wantErr: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			t.Setenv("HOME", home)
			if tc.plist {
				writeCheckPlist(t, home)
			}
			file := filepath.Join(t.TempDir(), "dorg.yml")
			if tc.config != "" {
				if err := os.WriteFile(file, []byte(tc.config), 0600); err != nil {
					t.Fatal(err)
				}
			}

			cmd := newCheckCmd()
			cmd.Flags().AddFlagSet(cmd.PersistentFlags())
			if err := cmd.Flags().Set("file", file); err != nil {
				t.Fatal(err)
			}
			if err := cmd.Flags().Set("verbose", "true"); err != nil {
				t.Fatal(err)
			}
			originalLogger := slog.Default()
			defer slog.SetDefault(originalLogger)

			err := execCheckCmd(cmd, nil, dock.LoadDockPlist)
			if (err != nil) != tc.wantErr {
				t.Fatalf("execCheckCmd() error = %v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

func checkConfig(app string) string {
	return `dock_items:
  apps:
    - ` + app + `
  settings:
    tilesize: 32
    largesize: 64
    magnification: true
    minimize-to-application: true
    autohide: true
    show-recents: true
    size-immutable: true
`
}

func writeCheckPlist(t *testing.T, home string) {
	t.Helper()
	path := filepath.Join(home, "Library", "Preferences", "com.apple.dock.plist")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	p := &dock.Plist{
		PersistentApps: []dock.PAItem{{
			TileType: "file-tile",
			TileData: dock.TileData{FileData: dock.FileData{URLString: "/Applications/Calculator.app"}},
		}},
		TileSize:              32,
		LargeSize:             64,
		Magnification:         true,
		MinimizeToApplication: true,
		AutoHide:              true,
		ShowRecents:           true,
		SizeImmutable:         true,
	}
	if err := plist.NewBinaryEncoder(file).Encode(p); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}
