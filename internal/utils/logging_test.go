package utils

import "testing"

func Test_SetLogLevel(t *testing.T) {
	tests := map[string]struct {
		verbose bool
		want    int
	}{
		"verbose": {verbose: true, want: SetLogLevel(true)},
		"minimal": {verbose: false, want: SetLogLevel(false)},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := SetLogLevel(tc.verbose)
			if got != tc.want {
				t.Fatalf("SetLogLevel(%v) = %d, want %d", tc.verbose, got, tc.want)
			}
		})
	}
}
