package network

import "testing"

func TestMeasureResolverRejectsInvalidAddress(t *testing.T) {
	res := MeasureResolver("not-a-dns", 0)
	if res.Success {
		t.Fatal("invalid address must not succeed")
	}
	if res.Error == "" {
		t.Fatal("expected an error message")
	}
}
