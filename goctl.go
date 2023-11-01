// Package goctl is a library for CLI Go applications to help interface with the goctl CLI tool,
// and the GitHub API.
//
// Note that the examples in this package assume goctl and git are installed. They do not run in
// the Go Playground used by pkg.go.dev.
package goctl

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/khulnasoft-lab/execsafer"
)

// Exec invokes a goctl command in a subprocess and captures the output and error streams.
func Exec(args ...string) (stdout, stderr bytes.Buffer, err error) {
	goctlExe, err := Path()
	if err != nil {
		return
	}
	err = run(context.Background(), goctlExe, nil, nil, &stdout, &stderr, args)
	return
}

// ExecContext invokes a goctl command in a subprocess and captures the output and error streams.
func ExecContext(ctx context.Context, args ...string) (stdout, stderr bytes.Buffer, err error) {
	goctlExe, err := Path()
	if err != nil {
		return
	}
	err = run(ctx, goctlExe, nil, nil, &stdout, &stderr, args)
	return
}

// Exec invokes a goctl command in a subprocess with its stdin, stdout, and stderr streams connected to
// those of the parent process. This is suitable for running goctl commands with interactive prompts.
func ExecInteractive(ctx context.Context, args ...string) error {
	goctlExe, err := Path()
	if err != nil {
		return err
	}
	return run(ctx, goctlExe, nil, os.Stdin, os.Stdout, os.Stderr, args)
}

// Path searches for an executable named "goctl" in the directories named by the PATH environment variable.
// If the executable is found the result is an absolute path.
func Path() (string, error) {
	if goctlExe := os.Getenv("GOCTL_PATH"); goctlExe != "" {
		return goctlExe, nil
	}
	return safeexec.LookPath("goctl")
}

func run(ctx context.Context, goctlExe string, env []string, stdin io.Reader, stdout, stderr io.Writer, args []string) error {
	cmd := exec.CommandContext(ctx, goctlExe, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if env != nil {
		cmd.Env = env
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("goctl execution failed: %w", err)
	}
	return nil
}
