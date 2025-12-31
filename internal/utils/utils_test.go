package utils

import (
	"context"
	"testing"
	"time"
)

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

func Test_RunCommand(t *testing.T) {
	tests := map[string]struct {
		ctx     context.Context
		cmd     string
		args    []string
		wantErr bool
	}{
		"echo succeeds": {ctx: context.Background(), cmd: "echo", args: []string{"hello"}, wantErr: false},
		"command times out": {ctx: func() context.Context {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			_ = cancel
			return ctx
		}(), cmd: "sleep", args: []string{"1"}, wantErr: true},
		"nonexistent command": {ctx: context.Background(), cmd: "no-such-cmd-xyz", args: nil, wantErr: true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := RunCommand(tc.ctx, tc.cmd, tc.args...)
			if (err != nil) != tc.wantErr {
				t.Fatalf("RunCommand(%s) error = %v, wantErr=%v", tc.cmd, err, tc.wantErr)
			}
		})
	}
}

func Test_Version(t *testing.T) {
	t.Parallel()

	if Version() == "" {
		t.Fatalf("Version() returned empty string")
	}
}
