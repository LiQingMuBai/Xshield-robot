package tron

import "testing"

func TestGetTronAddressRequiresConfiguredMnemonic(t *testing.T) {
	t.Setenv("TRON_MNEMONIC", "")

	_, _, err := GetTronAddress(0)
	if err == nil {
		t.Fatalf("expected GetTronAddress to fail when mnemonic is missing")
	}
}
