//go:build !windows

package cmdutil

import "os/exec"

func applyPlatformAttr(_ *exec.Cmd) {}
