package config

import "testing"

func TestFirstEnvReturnsPrimaryKeyValue(t *testing.T) {
	t.Setenv("TRXFEE_API_KEY", "primary-key")
	t.Setenv("TRXFEE_APIKEY", "legacy-key")

	value := firstEnv("TRXFEE_API_KEY", "TRXFEE_APIKEY")
	if value != "primary-key" {
		t.Fatalf("expected primary key value, got %q", value)
	}
}

func TestFirstEnvFallsBackToLegacyKey(t *testing.T) {
	t.Setenv("TRXFEE_API_KEY", "")
	t.Setenv("TRXFEE_APIKEY", "legacy-key")

	value := firstEnv("TRXFEE_API_KEY", "TRXFEE_APIKEY")
	if value != "legacy-key" {
		t.Fatalf("expected legacy key value, got %q", value)
	}
}

func TestFirstEnvReturnsEmptyWhenNoKeyIsSet(t *testing.T) {
	t.Setenv("TRXFEE_API_KEY", "")
	t.Setenv("TRXFEE_APIKEY", "")

	value := firstEnv("TRXFEE_API_KEY", "TRXFEE_APIKEY")
	if value != "" {
		t.Fatalf("expected empty value, got %q", value)
	}
}
