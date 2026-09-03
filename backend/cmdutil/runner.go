package cmdutil

import (
	"bytes"
	"context"
	"os/exec"
	"time"
)

const defaultTimeout = 20 * time.Second

// Result is the outcome of an executed command.
type Result struct {
	Stdout string
	Stderr string
	Err    error
}

// Runner executes programs with an explicit argument list (never a shell).
type Runner interface {
	Run(name string, args ...string) Result
	RunContext(ctx context.Context, name string, args ...string) Result
	LookPath(name string) (string, error)
}

// ExecRunner is the production Runner backed by os/exec.
type ExecRunner struct{}

// Run executes name with args using a default timeout.
func (r ExecRunner) Run(name string, args ...string) Result {
	return r.RunContext(context.Background(), name, args...)
}

// RunContext executes name with args. Arguments are passed directly to the
// process; they are never concatenated into a shell command string.
func (r ExecRunner) RunContext(ctx context.Context, name string, args ...string) Result {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, defaultTimeout)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, args...)
	applyPlatformAttr(cmd)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	return Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
		Err:    err,
	}
}

// LookPath resolves an executable on PATH.
func (r ExecRunner) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// Combined returns stdout and stderr joined, useful for error messages.
func (res Result) Combined() string {
	out := res.Stdout
	if res.Stderr != "" {
		if out != "" {
			out += "\n"
		}
		out += res.Stderr
	}
	return out
}

// Failed reports whether the command did not succeed.
func (res Result) Failed() bool {
	return res.Err != nil
}
