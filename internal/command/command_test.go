package command

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_Verify(t *testing.T) {
	file := filepath.Join(t.TempDir(), "subdir", "cfg.yml")

	tests := map[string]struct {
		cfg     *Config
		wantErr bool
	}{
		"create dir": {cfg: &Config{File: file}, wantErr: false},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if err := tc.cfg.Verify(); (err != nil) != tc.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tc.wantErr)
			}
			if _, err := os.Stat(filepath.Dir(tc.cfg.File)); os.IsNotExist(err) {
				t.Fatalf("expected dir created")
			}
		})
	}
}
