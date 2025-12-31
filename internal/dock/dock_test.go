package dock

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/5ouma/dorg/internal/config"
)

func Test_FileNameWithoutExtTrimSuffix(t *testing.T) {
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
	home, _ := os.UserHomeDir()

	tests := map[string]struct {
		in      config.Folder
		wantErr bool
	}{
		"tilde":     {in: config.Folder{Path: "~"}, wantErr: false},
		"tilde sub": {in: config.Folder{Path: "~/Documents"}, wantErr: false},
		"invalid":   {in: config.Folder{Path: "relative/path"}, wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := &Plist{}
			err := p.AddOther(tc.in)
			if (err != nil) != tc.wantErr {
				t.Fatalf("%v err=%v, wantErr=%v", tc.in, err, tc.wantErr)
			}
			if err == nil {
				got := p.PersistentOthers[0].TileData.FileData.URLString
				if tc.in.Path == "~" && got != home {
					t.Fatalf("expected %s got %s", home, got)
				}
				if tc.in.Path == "~/Documents" {
					if got != filepath.Join(home, "Documents") {
						t.Fatalf("expected %s got %s", filepath.Join(home, "Documents"), got)
					}
				}
			}
		})
	}
}

func Test_ApplySettings(t *testing.T) {

	tests := map[string]struct {
		in      config.DockSettings
		wantErr bool
	}{
		"valid sizes int":   {in: config.DockSettings{TileSize: 32, LargeSize: 64, Magnification: true}, wantErr: false},
		"valid sizes float": {in: config.DockSettings{TileSize: 32.0, LargeSize: 64.0}, wantErr: false},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			p := &Plist{}
			if err := p.ApplySettings(tc.in); (err != nil) != tc.wantErr {
				t.Fatalf("ApplySettings error = %v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
