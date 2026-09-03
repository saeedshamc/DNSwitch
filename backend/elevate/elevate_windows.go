//go:build windows

package elevate

import (
	"os"
	"syscall"

	"golang.org/x/sys/windows"
)

func isAdmin() bool {
	var sid *windows.SID
	err := windows.AllocateAndInitializeSid(
		&windows.SECURITY_NT_AUTHORITY,
		2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0,
		&sid,
	)
	if err != nil {
		return false
	}
	defer windows.FreeSid(sid)

	token := windows.Token(0)
	member, err := token.IsMember(sid)
	return err == nil && member
}

func relaunch() error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	verb, err := syscall.UTF16PtrFromString("runas")
	if err != nil {
		return err
	}
	file, err := syscall.UTF16PtrFromString(exe)
	if err != nil {
		return err
	}

	var show int32 = 1 // SW_SHOWNORMAL
	return windows.ShellExecute(0, verb, file, nil, nil, show)
}

func wrapPrefix() (string, []string) {
	return "", nil
}
