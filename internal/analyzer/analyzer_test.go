package analyzer

import (
	"testing"
)

func TestParseLogEntry(t *testing.T) {
	data := []byte(`{
	"PRIORITY": "6",
	"MESSAGE": "Starting demo.service - Demo service.",
	"_SYSTEMD_UNIT": "user-demo.service",
	"SYSLOG_IDENTIFIER": "demo"
	}`)

	entry, err := ParseLogEntry(data)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if entry.Priority != 6 {
		t.Errorf("expected priority 6, got %d", entry.Priority)
	}

	if entry.Message != "Starting demo.service - Demo service." {
		t.Errorf("expected message 'Starting demo.service - Demo service.', got '%s'", entry.Message)
	}

	if entry.Unit != "user-demo.service" {
		t.Errorf("expected unit 'user-demo.service', got '%s'", entry.Unit)
	}

	if entry.Source != "demo" {
		t.Errorf("expected source 'demo', got '%s'", entry.Source)
	}
}

func TestValidateLogEntry(t *testing.T) {
	data := []byte(`{
	"PRIORITY": "9",
	"MESSAGE": "Demo service failed",
	"_SYSTEMD_UNIT": "user-demo.service",
	"SYSLOG_IDENTIFIER": "demo"
	}`)

	_, err := ParseLogEntry(data)

	if err == nil {
		t.Fatal("expected error for invalid priority")
	}
}

func TestMessageValidation(t *testing.T) {
	data := []byte(`{
	"PRIORITY": "3",
	"MESSAGE": "   ",
	"_SYSTEMD_UNIT": "user-demo.service",
	"SYSLOG_IDENTIFIER": "demo"
	}`)

	_, err := ParseLogEntry(data)

	if err == nil {
		t.Fatal("expected error for empty message")
	}
}

func TestJsonValidation(t *testing.T) {
	data := []byte(`{"PRIORITY":`)

	_, err := ParseLogEntry(data)

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
