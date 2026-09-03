package dns

import (
	"strings"
	"testing"

	"github.com/saeedshamc/DNSwitch/backend/cmdutil"
)

type fakeEnv struct {
	cmds     map[string]bool
	services map[string]bool
}

func (f fakeEnv) HasCommand(name string) bool    { return f.cmds[name] }
func (f fakeEnv) ServiceActive(name string) bool { return f.services[name] }

func TestSelectLinuxStrategy(t *testing.T) {
	cases := []struct {
		name string
		env  fakeEnv
		want Strategy
	}{
		{
			name: "networkmanager active",
			env: fakeEnv{
				cmds:     map[string]bool{"nmcli": true, "resolvectl": true},
				services: map[string]bool{"NetworkManager": true},
			},
			want: StrategyNetworkManager,
		},
		{
			name: "systemd-resolved active",
			env: fakeEnv{
				cmds:     map[string]bool{"resolvectl": true},
				services: map[string]bool{"systemd-resolved": true},
			},
			want: StrategyResolved,
		},
		{
			name: "resolv.conf fallback",
			env:  fakeEnv{cmds: map[string]bool{}, services: map[string]bool{}},
			want: StrategyResolvConf,
		},
		{
			name: "nmcli present without service",
			env: fakeEnv{
				cmds:     map[string]bool{"nmcli": true},
				services: map[string]bool{},
			},
			want: StrategyNetworkManager,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SelectLinuxStrategy(tc.env)
			if got != tc.want {
				t.Fatalf("got %s want %s", got, tc.want)
			}
		})
	}
}

func TestPatchResolvConf(t *testing.T) {
	original := "search lan\nnameserver 192.168.1.1\noptions edns0\n"
	got := PatchResolvConf(original, []string{"1.1.1.1", "1.0.0.1"})
	if !strings.Contains(got, "search lan") {
		t.Fatalf("lost search directive: %q", got)
	}
	if !strings.Contains(got, "options edns0") {
		t.Fatalf("lost options directive: %q", got)
	}
	if strings.Contains(got, "192.168.1.1") {
		t.Fatalf("old nameserver should be gone: %q", got)
	}
	if !strings.Contains(got, "nameserver 1.1.1.1") || !strings.Contains(got, "nameserver 1.0.0.1") {
		t.Fatalf("new nameservers missing: %q", got)
	}
}

func TestMockRunnerNeverUsesShell(t *testing.T) {
	mock := &cmdutil.MockRunner{
		Handlers: map[string]cmdutil.Result{
			"nmcli con mod Wired ipv4.dns 1.1.1.1 1.0.0.1": {},
		},
	}
	res := mock.Run("nmcli", "con", "mod", "Wired", "ipv4.dns", "1.1.1.1 1.0.0.1")
	if res.Failed() {
		t.Fatal(res.Err)
	}
	if len(mock.Calls) != 1 {
		t.Fatalf("calls: %+v", mock.Calls)
	}
	if strings.Contains(strings.Join(mock.Calls[0].Args, " "), ";") {
		t.Fatal("arguments must not include a shell metacharacter sequence")
	}
}
