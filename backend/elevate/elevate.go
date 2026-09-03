package elevate

// Explain returns a short English reason shown before an elevation prompt.
func Explain() string {
	return "Changing system DNS servers requires administrator or root access so the operating system will accept the new resolver addresses."
}

// IsAdmin reports whether the current process already has elevated rights.
func IsAdmin() bool {
	return isAdmin()
}

// Relaunch requests a new elevated instance of the current executable.
func Relaunch() error {
	return relaunch()
}

// WrapPrefix returns the privilege helper (pkexec) and extra args, if needed.
func WrapPrefix() (helper string, prefix []string) {
	return wrapPrefix()
}
