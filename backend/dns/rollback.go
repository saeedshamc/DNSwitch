package dns

import "fmt"

// ApplyWithRollback captures the current DNS, applies servers, and restores
// the previous state if the change fails.
func ApplyWithRollback(m DNSManager, interfaceName string, servers []string) error {
	snap, err := m.Snapshot(interfaceName)
	if err != nil {
		return err
	}
	if err := m.SetDNS(interfaceName, servers); err != nil {
		_ = restore(m, interfaceName, snap)
		return err
	}
	return nil
}

// ResetWithRollback restores DHCP and rolls back if the reset fails.
func ResetWithRollback(m DNSManager, interfaceName string) error {
	snap, err := m.Snapshot(interfaceName)
	if err != nil {
		return err
	}
	if err := m.ResetToDHCP(interfaceName); err != nil {
		_ = restore(m, interfaceName, snap)
		return err
	}
	return nil
}

func restore(m DNSManager, interfaceName string, snap DNSState) error {
	if snap.DHCP || len(snap.Servers) == 0 {
		if err := m.ResetToDHCP(interfaceName); err != nil {
			return fmt.Errorf("rollback failed: %w", err)
		}
		return nil
	}
	if err := m.SetDNS(interfaceName, snap.Servers); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}
	return nil
}
