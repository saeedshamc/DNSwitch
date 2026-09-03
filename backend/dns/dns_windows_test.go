//go:build windows

package dns

import (
	"strings"
	"testing"

	"github.com/saeedshamc/DNSwitch/backend/cmdutil"
)

func TestWindowsSetDNSUsesArgumentList(t *testing.T) {
	mock := &cmdutil.MockRunner{Default: cmdutil.Result{}}
	m := newPlatformManager(mock).(*windowsManager)
	if err := m.SetDNS("Ethernet", []string{"1.1.1.1", "1.0.0.1"}); err != nil {
		t.Fatal(err)
	}
	if len(mock.Calls) < 2 {
		t.Fatalf("expected netsh set + add, got %+v", mock.Calls)
	}
	for _, call := range mock.Calls {
		joined := strings.Join(call.Args, " ")
		if strings.ContainsAny(joined, ";&|") && !strings.Contains(joined, "name=Ethernet") {
			t.Fatalf("unexpected shell metacharacters: %q", joined)
		}
		if call.Name == "cmd" || call.Name == "powershell" {
			t.Fatalf("must not invoke a shell: %s", call.Name)
		}
	}
}

func TestWindowsRejectsInjectedInterface(t *testing.T) {
	mock := &cmdutil.MockRunner{Default: cmdutil.Result{}}
	m := newPlatformManager(mock)
	if err := m.SetDNS(`Ethernet" && calc`, []string{"1.1.1.1"}); err == nil {
		t.Fatal("expected invalid interface error")
	}
	if len(mock.Calls) != 0 {
		t.Fatalf("command should not run: %+v", mock.Calls)
	}
}
