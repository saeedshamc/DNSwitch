//go:build linux

package elevate

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func isAdmin() bool {
	return os.Geteuid() == 0
}

func relaunch() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	helper, err := exec.LookPath("pkexec")
	if err != nil {
		helper, err = exec.LookPath("sudo")
		if err != nil {
			return fmt.Errorf("neither pkexec nor sudo is available")
		}
	}
	cmd := exec.Command(helper, exe)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	return cmd.Start()
}

func wrapPrefix() (string, []string) {
	if isAdmin() {
		return "", nil
	}
	if p, err := exec.LookPath("pkexec"); err == nil {
		return p, nil
	}
	if p, err := exec.LookPath("sudo"); err == nil {
		return p, []string{"-n"}
	}
	return "", nil
}
