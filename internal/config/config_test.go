package config

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_Load(t *testing.T) {
	tmp := t.TempDir()

	tests := map[string]struct {
		file    string
		content string
		wantErr bool
	}{
		"valid file": {
			file: "valid.yml",
			content: `dock_items:
  apps:
    - /Applications/Calculator.app
  others:
    - path: ~/Documents
      sort: 1
      display: 2
      view: 3`,
			wantErr: false,
		},
		"invalid yaml": {file: "invalid.yml", content: "dock_items: [unclosed", wantErr: true},
		"missing file": {file: filepath.Join(tmp, "missing.yml"), wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if tc.content != "" {
				path := filepath.Join(tmp, tc.file)
				if err := os.WriteFile(path, []byte(tc.content), 0644); err != nil {
					t.Fatalf("failed to write test file: %v", err)
				}
				tc.file = path
			}

			_, err := Load(tc.file)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Load(%s) error = %v, wantErr=%v", tc.file, err, tc.wantErr)
			}
		})
	}
}
