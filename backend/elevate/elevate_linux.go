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

	var helper string
	var args []string
	if p, err := exec.LookPath("pkexec"); err == nil {
		helper = p
		// Preserve the graphical session so the elevated Wails window can open.
		args = []string{"env"}
		for _, key := range []string{
			"DISPLAY",
			"WAYLAND_DISPLAY",
			"XAUTHORITY",
			"XDG_RUNTIME_DIR",
			"DBUS_SESSION_BUS_ADDRESS",
			"XDG_CURRENT_DESKTOP",
			"XDG_SESSION_TYPE",
			"LANG",
			"LC_ALL",
		} {
			if val := os.Getenv(key); val != "" {
				args = append(args, key+"="+val)
			}
		}
		args = append(args, exe)
	} else if p, err := exec.LookPath("sudo"); err == nil {
		helper = p
		args = []string{exe}
	} else {
		return fmt.Errorf("neither pkexec nor sudo is available")
	}

	cmd := exec.Command(helper, args...)
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
		// Prefer an interactive prompt when a TTY is available; -n would
		// silently fail when credentials are not cached.
		return p, nil
	}
	return "", nil
}
