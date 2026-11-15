package utils

import (
	"context"
	"errors"
	"testing"
	"time"
)

type deadlineContext struct{}

func (deadlineContext) Deadline() (time.Time, bool) { return time.Time{}, true }
func (deadlineContext) Done() <-chan struct{}       { return nil }
func (deadlineContext) Err() error                  { return context.DeadlineExceeded }
func (deadlineContext) Value(any) any               { return nil }

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
