package utils

import (
	"runtime/debug"
	"testing"
)

func Test_Version(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		"missing build info": {want: "unknown"},
		"development build": {
			info: &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}},
			ok:   true,
			want: "unknown",
		},
		"release version": {
			info: &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}},
			ok:   true,
			want: "v1.2.3",
		},
		"commit version": {
			info: &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3-0-abcdef123456"}},
			ok:   true,
			want: "v1.2.3 (#abcdef1)",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := Version(tc.info, tc.ok); got != tc.want {
				t.Fatalf("Version() = %q, want %q", got, tc.want)
			}
		})
	}

	info, ok := debug.ReadBuildInfo()
	if Version(info, ok) == "" {
		t.Fatal("Version() returned empty string for current build info")
	}
}
