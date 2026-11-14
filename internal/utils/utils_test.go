package utils

import (
	"context"
	"errors"
	"runtime/debug"
	"testing"
	"time"
)

type deadlineContext struct{}

func (deadlineContext) Deadline() (time.Time, bool) { return time.Time{}, true }
func (deadlineContext) Done() <-chan struct{}       { return nil }
func (deadlineContext) Err() error                  { return context.DeadlineExceeded }
func (deadlineContext) Value(any) any               { return nil }

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
		"echo succeeds":          {ctx: context.Background(), cmd: "echo", args: []string{"hello"}, wantErr: false},
		"nil context":            {cmd: "echo", args: []string{"hello"}},
		"deadline after command": {ctx: deadlineContext{}, cmd: "echo", args: []string{"hello"}, wantErr: true},
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

func Test_RestartDock(t *testing.T) {
	tests := map[string]struct {
		commandErr error
		wantErr    bool
	}{
		"restart succeeds": {},
		"kill command fails": {
			commandErr: errors.New("injected command error"),
			wantErr:    true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			runner := func(context.Context, string, ...string) (string, error) {
				return "", tc.commandErr
			}

			err := RestartDock(runner)
			if (err != nil) != tc.wantErr {
				t.Fatalf("RestartDock() error = %v, wantErr=%v", err, tc.wantErr)
			}
		})
	}
}

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
