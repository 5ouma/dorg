package cmd

import (
	"encoding/json"
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
		"valid file":   {content: `dock_items: {apps: ["/Applications/Calculator.app"]}`, wantError: false},
		"missing file": {path: "/non/existing/path.yml", wantError: true},
		"invalid yaml": {content: "dock_items: [unclosed", wantError: true},
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
	t.Parallel()

	tests := map[string]struct {
		createPlist bool
		wantError   bool
	}{
		"valid plist":   {createPlist: true, wantError: false},
		"missing plist": {createPlist: false, wantError: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			tmp := t.TempDir()
			oldHome := os.Getenv("HOME")
			if err := os.Setenv("HOME", tmp); err != nil {
				t.Fatalf("failed to set HOME: %v", err)
			}
			defer func() {
				if err := os.Setenv("HOME", oldHome); err != nil {
					t.Fatalf("failed to restore HOME: %v", err)
				}
			}()

			if tc.createPlist {
				prefsDir := filepath.Join(tmp, "Library", "Preferences")
				if err := os.MkdirAll(prefsDir, 0755); err != nil {
					t.Fatalf("failed to create prefs dir: %v", err)
				}
				p := &dock.Plist{
					PersistentApps:   []dock.PAItem{{TileType: "file-tile", TileData: dock.TileData{FileData: dock.FileData{URLString: "file:///Applications/Calculator.app/", URLStringType: 0}}}},
					PersistentOthers: []dock.POItem{{TileType: "directory-tile", TileData: dock.POTileData{Arrangement: 1, DisplayAs: 2, ShowAs: 3, FileData: dock.FileData{URLString: "file://" + filepath.Join(tmp, "Documents") + "/", URLStringType: 0}, FileLabel: "Documents", FileType: 2}}},
					TileSize:         32,
					LargeSize:        64,
					Magnification:    true,
					AutoHide:         false,
					ShowRecents:      true,
				}
				plistPath := filepath.Join(prefsDir, "com.apple.dock.plist")
				f, err := os.Create(plistPath)
				if err != nil {
					t.Fatalf("failed to create plist file: %v", err)
				}
				if err := plist.NewBinaryEncoder(f).Encode(p); err != nil {
					_ = f.Close()
					t.Fatalf("failed to encode plist: %v", err)
				}
				if err := f.Close(); err != nil {
					t.Fatalf("failed to close plist file: %v", err)
				}
			}

			data, err := loadPlistConfig()
			if (err != nil) != tc.wantError {
				t.Fatalf("loadPlistConfig() error = %v, wantErr=%v", err, tc.wantError)
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
