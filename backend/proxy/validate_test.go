package proxy

import "testing"

func TestNormalizeAddress(t *testing.T) {
	got, err := NormalizeAddress("http://127.0.0.1:8080")
	if err != nil || got != "127.0.0.1:8080" {
		t.Fatalf("got %q err %v", got, err)
	}
	if _, err := NormalizeAddress("bad"); err == nil {
		t.Fatal("expected error")
	}
	if _, err := NormalizeAddress("host;rm:80"); err == nil {
		t.Fatal("expected rejection")
	}
}

func TestValidateConfig(t *testing.T) {
	if err := ValidateConfig(Config{Enabled: true}); err == nil {
		t.Fatal("expected empty proxy error")
	}
	if err := ValidateConfig(Config{Enabled: true, HTTP: "1.1.1.1:8080"}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateConfig(Config{Enabled: false}); err != nil {
		t.Fatal(err)
	}
}
