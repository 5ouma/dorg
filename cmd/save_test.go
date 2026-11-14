package cmd

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/5ouma/dorg/internal/dock"
	"go.yaml.in/yaml/v4"
)

func Test_execSaveCmd(t *testing.T) {
	tests := map[string]struct {
		loadErr   error
		verbose   bool
		verifyErr bool
		wantErr   bool
	}{
		"success":      {verbose: true},
		"save fails":   {loadErr: errors.New("injected load error"), wantErr: true},
		"verify fails": {verifyErr: true, wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("HOME", t.TempDir())
			file := filepath.Join(t.TempDir(), "dorg.yml")
			if tc.verifyErr {
				parent := filepath.Join(t.TempDir(), "file")
				if err := os.WriteFile(parent, nil, 0600); err != nil {
					t.Fatal(err)
				}
				file = filepath.Join(parent, "child", "config.yml")
			}
			loadDock := func() (*dock.Plist, error) {
				if tc.loadErr != nil {
					return nil, tc.loadErr
				}
				return &dock.Plist{
					PersistentApps: []dock.PAItem{{
						TileType: "file-tile",
						TileData: dock.TileData{FileData: dock.FileData{URLString: "/Applications/Calculator.app"}},
					}},
				}, nil
			}

			cmd := newSaveCmd()
			cmd.Flags().AddFlagSet(cmd.PersistentFlags())
			if err := cmd.Flags().Set("file", file); err != nil {
				t.Fatal(err)
			}
			if err := cmd.Flags().Set("verbose", strconv.FormatBool(tc.verbose)); err != nil {
				t.Fatal(err)
			}
			originalLogger := slog.Default()
			defer slog.SetDefault(originalLogger)

			err := execSaveCmd(cmd, nil, loadDock, (*yaml.Encoder).Close)
			if (err != nil) != tc.wantErr {
				t.Fatalf("execSaveCmd() error = %v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
