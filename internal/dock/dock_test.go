package dock

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/5ouma/dorg/internal/config"
	"howett.net/plist"
)

func Test_FileNameWithoutExtTrimSuffix(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in   string
		want string
	}{
		"/foo/bar/baz.txt": {in: "/foo/bar/baz.txt", want: "baz"},
		"some.file.tar.gz": {in: "some.file.tar.gz", want: "some.file.tar"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := fileNameWithoutExtTrimSuffix(tc.in)
			if got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func Test_TilePathGetters(t *testing.T) {
	t.Parallel()

	tileData := TileData{FileData: FileData{URLString: "file:///Applications/Calculator.app/"}}
	if got := tileData.GetPath(); got != "/Applications/Calculator.app" {
		t.Fatalf("TileData.GetPath = %s", got)
	}

	poTileData := POTileData{FileData: FileData{URLString: "file:///Users/test%20name/Docs/"}}
	if got := poTileData.GetPath(); got != "/Users/test name/Docs" {
		t.Fatalf("POTileData.GetPath = %s", got)
	}
}

func Test_AddApp(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		in       string
		wantType string
	}{
		"empty spacer": {in: "", wantType: "small-spacer-tile"},
		"spacer":       {in: " ", wantType: "spacer-tile"},
		"normal":       {in: "/Applications/Calculator.app", wantType: "file-tile"},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := &Plist{}
			p.AddApp(tc.in)
			if len(p.PersistentApps) != 1 {
				t.Fatalf("expected 1 app, got %d", len(p.PersistentApps))
			}
			if p.PersistentApps[0].TileType != tc.wantType {
				t.Fatalf("tile type = %s, want %s", p.PersistentApps[0].TileType, tc.wantType)
			}
		})
	}
}

func Test_AddOther(t *testing.T) {
	t.Parallel()

	home, _ := os.UserHomeDir()

	tests := map[string]struct {
		in       config.Folder
		wantPath string
		wantErr  bool
	}{
		"tilde":     {in: config.Folder{Path: "~"}, wantPath: home},
		"tilde sub": {in: config.Folder{Path: "~/Documents"}, wantPath: filepath.Join(home, "Documents")},
		"invalid":   {in: config.Folder{Path: "relative/path"}, wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := &Plist{}
			err := p.AddOther(tc.in, filepath.Abs)
			if (err != nil) != tc.wantErr {
				t.Fatalf("%v err=%v, wantErr=%v", tc.in, err, tc.wantErr)
			}
			if err == nil {
				got := p.PersistentOthers[0].TileData.FileData.URLString
				if got != tc.wantPath {
					t.Fatalf("path = %s, want %s", got, tc.wantPath)
				}
			}
		})
	}
}

func Test_ApplySettings(t *testing.T) {
	t.Parallel()

	tests := map[string]config.DockSettings{
		"disabled": {},
		"enabled": {
			Magnification:         true,
			MinimizeToApplication: true,
			AutoHide:              true,
			ShowRecents:           true,
			SizeImmutable:         true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := &Plist{}
			p.ApplySettings(tc)
			if p.Magnification != tc.Magnification ||
				p.MinimizeToApplication != tc.MinimizeToApplication ||
				p.AutoHide != tc.AutoHide ||
				p.ShowRecents != tc.ShowRecents ||
				p.SizeImmutable != tc.SizeImmutable {
				t.Fatalf("settings were not applied: %+v", p)
			}
		})
	}
}

func Test_GenerateConfigFromPlist(t *testing.T) {
	t.Parallel()

	home, _ := os.UserHomeDir()
	tests := map[string]struct {
		plist Plist
		want  config.Config
	}{
		"all": {
			plist: Plist{
				PersistentApps:        []PAItem{{TileData: TileData{FileData: FileData{URLString: "file:///Applications/Calculator.app/"}}}},
				PersistentOthers:      []POItem{{TileData: POTileData{Arrangement: 1, DisplayAs: 2, ShowAs: 3, FileData: FileData{URLString: filepath.Join(home, "Documents") + "/"}}}},
				TileSize:              32,
				LargeSize:             64,
				Magnification:         true,
				MinimizeToApplication: true,
				AutoHide:              true,
				ShowRecents:           true,
				SizeImmutable:         true,
			},
			want: config.Config{Dock: config.Dock{
				Apps:     []string{"/Applications/Calculator.app"},
				Others:   []config.Folder{{Path: "~/Documents", Sort: 1, Display: 2, View: 3}},
				Settings: &config.DockSettings{TileSize: 32, LargeSize: 64, Magnification: true, MinimizeToApplication: true, AutoHide: true, ShowRecents: true, SizeImmutable: true},
			}},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			conf, err := tc.plist.GenerateConfigFromPlist()
			if err != nil {
				t.Fatalf("GenerateConfigFromPlist error: %v", err)
			}

			if !reflect.DeepEqual(conf, tc.want) {
				t.Fatalf("config mismatch\n got: %#v\nwant: %#v", conf, tc.want)
			}
		})
	}
}

func Test_LoadDockPlist(t *testing.T) {
	var binary bytes.Buffer
	if err := plist.NewBinaryEncoder(&binary).Encode(&Plist{TileSize: 40}); err != nil {
		t.Fatal(err)
	}

	tests := map[string]struct {
		write     bool
		emptyHome bool
		data      []byte
		wantErr   bool
		wantSize  int
	}{
		"valid plist":   {write: true, data: []byte("<plist><dict><key>tilesize</key><integer>32</integer></dict></plist>"), wantSize: 32},
		"binary plist":  {write: true, data: binary.Bytes(), wantSize: 40},
		"invalid plist": {write: true, data: []byte("not a plist"), wantErr: true},
		"missing plist": {wantErr: true},
		"missing home":  {emptyHome: true, wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			home := t.TempDir()
			if tc.emptyHome {
				t.Setenv("HOME", "")
			} else {
				t.Setenv("HOME", home)
			}
			if tc.write {
				path := filepath.Join(home, dockPlistPath)
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, tc.data, 0600); err != nil {
					t.Fatal(err)
				}
			}

			got, err := LoadDockPlist()
			if (err != nil) != tc.wantErr {
				t.Fatalf("LoadDockPlist() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if err == nil && fmt.Sprint(got.TileSize) != fmt.Sprint(tc.wantSize) {
				t.Fatalf("TileSize = %v, want %d", got.TileSize, tc.wantSize)
			}
		})
	}
}

func Test_PlistSave(t *testing.T) {
	tests := map[string]struct {
		failAt       string
		removeTemp   bool
		tempDirError bool
		encodeError  bool
		closeError   bool
		wantErr      bool
		wantCommands []string
	}{
		"success": {
			wantCommands: []string{"unload", "import", "load", "start"},
		},
		"unload failure": {
			failAt:       "unload",
			wantErr:      true,
			wantCommands: []string{"unload"},
		},
		"import failure": {
			failAt:       "import",
			wantErr:      true,
			wantCommands: []string{"unload", "import"},
		},
		"load failure": {
			failAt:       "load",
			wantErr:      true,
			wantCommands: []string{"unload", "import", "load"},
		},
		"start failure": {
			failAt:       "start",
			wantErr:      true,
			wantCommands: []string{"unload", "import", "load", "start"},
		},
		"encode failure": {
			encodeError:  true,
			wantErr:      true,
			wantCommands: []string{"unload"},
		},
		"close failure": {
			closeError:   true,
			wantErr:      true,
			wantCommands: []string{"unload"},
		},
		"temp file already removed": {
			removeTemp:   true,
			wantCommands: []string{"unload", "import", "load", "start"},
		},
		"temp directory unavailable": {
			tempDirError: true,
			wantErr:      true,
			wantCommands: []string{"unload"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var commands []string
			runCommand := func(_ context.Context, _ string, args ...string) (string, error) {
				operation := args[0]
				commands = append(commands, operation)
				if tc.removeTemp && operation == "import" {
					if err := os.Remove(args[2]); err != nil {
						t.Fatalf("remove temp plist: %v", err)
					}
				}
				if operation == tc.failAt {
					return "", errors.New("injected command error")
				}
				return "", nil
			}
			if tc.tempDirError {
				t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "missing"))
			}

			p := &Plist{}
			if tc.encodeError {
				p.TileSize = make(chan int)
			}
			var err error
			if tc.closeError {
				err = p.Save(runCommand, func(*os.File) error {
					return errors.New("injected close error")
				})
			} else {
				err = p.Save(runCommand, (*os.File).Close)
			}
			if (err != nil) != tc.wantErr {
				t.Fatalf("Save() error = %v, wantErr=%v", err, tc.wantErr)
			}
			if !reflect.DeepEqual(commands, tc.wantCommands) {
				t.Fatalf("commands = %v, want %v", commands, tc.wantCommands)
			}
		})
	}
}

func Test_AddOther_homeAndAbsErrors(t *testing.T) {
	t.Run("home missing", func(t *testing.T) {
		t.Setenv("HOME", "")
		if err := (&Plist{}).AddOther(config.Folder{Path: "~/Documents"}, filepath.Abs); err == nil {
			t.Fatal("AddOther() expected an error when HOME is missing")
		}
	})

	t.Run("absolute path resolution fails", func(t *testing.T) {
		absPath := func(string) (string, error) {
			return "", errors.New("injected absolute path error")
		}
		t.Setenv("HOME", t.TempDir())

		if err := (&Plist{}).AddOther(config.Folder{Path: "~/Documents"}, absPath); err == nil {
			t.Fatal("AddOther() expected an absolute path resolution error")
		}
	})
}

func Test_GenerateConfigFromPlist_homeAndRelErrors(t *testing.T) {
	t.Run("home missing", func(t *testing.T) {
		t.Setenv("HOME", "")
		if _, err := (&Plist{}).GenerateConfigFromPlist(); err == nil {
			t.Fatal("GenerateConfigFromPlist() expected an error when HOME is missing")
		}
	})

	t.Run("relative home", func(t *testing.T) {
		t.Setenv("HOME", "relative-home")
		p := &Plist{PersistentOthers: []POItem{{
			TileData: POTileData{FileData: FileData{URLString: "/tmp/Folder"}},
		}}}
		cfg, err := p.GenerateConfigFromPlist()
		if err != nil {
			t.Fatal(err)
		}
		if got := cfg.Dock.Others[0].Path; got != "/tmp/Folder" {
			t.Fatalf("path = %q, want absolute path unchanged", got)
		}
	})
}
