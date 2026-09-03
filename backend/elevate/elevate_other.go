//go:build !windows && !linux

package elevate

import "fmt"

func isAdmin() bool { return false }

func relaunch() error {
	return fmt.Errorf("elevation is not supported on this operating system")
}

func wrapPrefix() (string, []string) { return "", nil }
