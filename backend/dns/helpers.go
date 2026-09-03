package dns

import "strings"

func looksLikeLoopback(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	return n == "lo" || n == "loopback" || strings.Contains(n, "loopback")
}

func skipAdapter(name string) bool {
	n := strings.ToLower(strings.TrimSpace(name))
	for _, part := range []string{
		"bluetooth",
		"isatap",
		"teredo",
		"pseudo-interface",
		"loopback",
		"vmware",
		"virtualbox",
	} {
		if strings.Contains(n, part) {
			return true
		}
	}
	return false
}
