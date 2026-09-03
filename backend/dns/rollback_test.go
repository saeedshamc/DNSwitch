package dns

import "testing"

type stubManager struct {
	snap     DNSState
	failSet  bool
	sets     int
	resets   int
	restored bool
}

func (s *stubManager) GetCurrentDNS(string) ([]string, error) { return s.snap.Servers, nil }
func (s *stubManager) Snapshot(string) (DNSState, error)      { return s.snap, nil }
func (s *stubManager) ListInterfaces() ([]NetworkInterface, error) {
	return []NetworkInterface{{Name: "eth0", IsUp: true}}, nil
}
func (s *stubManager) FlushCache() error { return nil }
func (s *stubManager) SetDNS(_ string, _ []string) error {
	s.sets++
	if s.failSet {
		return ErrApplyFailed
	}
	return nil
}
func (s *stubManager) ResetToDHCP(string) error {
	s.resets++
	s.restored = true
	return nil
}

func TestApplyWithRollback(t *testing.T) {
	m := &stubManager{snap: DNSState{DHCP: true}, failSet: true}
	err := ApplyWithRollback(m, "eth0", []string{"1.1.1.1"})
	if err == nil {
		t.Fatal("expected failure")
	}
	if m.sets != 1 || m.resets != 1 {
		t.Fatalf("set=%d reset=%d", m.sets, m.resets)
	}
}
