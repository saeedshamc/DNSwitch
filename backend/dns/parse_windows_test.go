package dns

import "testing"

func TestParseNetshDNS(t *testing.T) {
	dhcpOut := `
Configuration for interface "Wi-Fi"
    DNS servers configured through DHCP:  192.168.1.1
`
	servers, dhcp, err := parseNetshDNS(dhcpOut)
	if err != nil {
		t.Fatal(err)
	}
	if !dhcp {
		t.Fatal("expected DHCP")
	}
	if len(servers) != 1 || servers[0] != "192.168.1.1" {
		t.Fatalf("servers: %v", servers)
	}

	staticOut := `
Configuration for interface "Ethernet"
    Statically Configured DNS Servers:    8.8.8.8
                                          8.8.4.4
`
	servers, dhcp, err = parseNetshDNS(staticOut)
	if err != nil {
		t.Fatal(err)
	}
	if dhcp {
		t.Fatal("expected static DNS")
	}
	if len(servers) != 2 {
		t.Fatalf("servers: %v", servers)
	}
}
