package dns

import "testing"

func TestValidateInterfaceName(t *testing.T) {
	valid := []string{"Wi-Fi", "Ethernet", "eth0", "wlan0", "Wired connection 1"}
	for _, name := range valid {
		if err := ValidateInterfaceName(name); err != nil {
			t.Errorf("%q should be valid: %v", name, err)
		}
	}

	invalid := []string{
		"",
		"eth0; rm -rf /",
		`Wi-Fi" && calc`,
		"eth0$(reboot)",
		"a|b",
		"x\nnameserver 1.1.1.1",
		"iface`id`",
	}
	for _, name := range invalid {
		if err := ValidateInterfaceName(name); err == nil {
			t.Errorf("%q should be rejected", name)
		}
	}
}

func TestValidateDNSServers(t *testing.T) {
	if err := ValidateDNSServers([]string{"1.1.1.1", "1.0.0.1"}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateDNSServers([]string{"2606:4700:4700::1111"}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateDNSServers([]string{"127.0.0.1"}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateDNSServers(nil); err != ErrEmptyDNS {
		t.Fatalf("expected empty error, got %v", err)
	}
	if err := ValidateDNSServers([]string{"not-an-ip"}); err != ErrInvalidDNS {
		t.Fatalf("expected invalid error, got %v", err)
	}
	if err := ValidateDNSServers([]string{"0.0.0.0"}); err != ErrInvalidDNS {
		t.Fatalf("expected unspecified rejection, got %v", err)
	}
	if err := ValidateDNSServers([]string{"224.0.0.1"}); err != ErrInvalidDNS {
		t.Fatalf("expected multicast rejection, got %v", err)
	}
}

func TestSplitIPv4v6(t *testing.T) {
	v4, v6 := SplitIPv4v6([]string{"1.1.1.1", "2606:4700:4700::1111", "1.0.0.1"})
	if len(v4) != 2 || v4[0] != "1.1.1.1" || v4[1] != "1.0.0.1" {
		t.Fatalf("unexpected v4: %v", v4)
	}
	if len(v6) != 1 {
		t.Fatalf("unexpected v6: %v", v6)
	}
}

func TestSkipAdapter(t *testing.T) {
	if !skipAdapter("Bluetooth Network Connection") {
		t.Fatal("bluetooth adapters should be skipped")
	}
	if skipAdapter("Wi-Fi") || skipAdapter("Ethernet") || skipAdapter("eth0") {
		t.Fatal("normal adapters should not be skipped")
	}
}

func TestPresets(t *testing.T) {
	list, err := Presets()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"automatic", "google", "cloudflare", "quad9", "opendns", "shecan", "electro"}
	if len(list) != len(want) {
		t.Fatalf("got %d presets, want %d", len(list), len(want))
	}
	seen := map[string]bool{}
	for _, p := range list {
		seen[p.ID] = true
		if !p.IsPreset {
			t.Errorf("%s should be marked as a preset", p.ID)
		}
	}
	for _, id := range want {
		if !seen[id] {
			t.Errorf("missing preset %s", id)
		}
	}
	cf := list[2]
	if cf.ID != "cloudflare" || cf.IPv4[0] != "1.1.1.1" {
		t.Fatalf("cloudflare preset mismatch: %+v", cf)
	}
}
