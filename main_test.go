package main

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

func Test_main(t *testing.T) {
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()
	os.Args = []string{"dorg"}

	main()
}

func Test_mainExitsOnCommandError(t *testing.T) {
	if os.Getenv("DORG_TEST_MAIN") == "1" {
		os.Args = []string{"dorg", "--invalid"}
		main()
		return
	}

	child := exec.Command(os.Args[0], "-test.run=^Test_mainExitsOnCommandError$")
	child.Env = append(os.Environ(), "DORG_TEST_MAIN=1")
	err := child.Run()
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("child error = %v, want exit error", err)
	}
	if exitErr.ExitCode() != 1 {
		t.Fatalf("child exit code = %d, want 1", exitErr.ExitCode())
	}
}
