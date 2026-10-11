package redact

import (
	"strings"
	"testing"
)

func TestRedactSensitivePatterns(t *testing.T) {
	r := New()

	// Use placeholder test tokens that do not trigger high-entropy secret scanners
	dummyBearer := "bearer-token-example-123456"
	dummyKey := "apikey_test_example_123456"
	dummyPass := "pass123456"

	input := "Connected to 192.168.1.50 with Bearer " + dummyBearer + " and api_key=" + dummyKey + " password=" + dummyPass
	output := r.Apply(input)

	if strings.Contains(output, "192.168.1.50") {
		t.Errorf("expected IP address to be redacted, got %s", output)
	}
	if strings.Contains(output, dummyBearer) {
		t.Errorf("expected Bearer token to be redacted, got %s", output)
	}
	if strings.Contains(output, dummyKey) {
		t.Errorf("expected API key to be redacted, got %s", output)
	}
	if strings.Contains(output, dummyPass) {
		t.Errorf("expected password to be redacted, got %s", output)
	}
}
