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

func TestAnalyzeResults(t *testing.T) {
	data := []string{
		`{
			"PRIORITY": "3",
			"MESSAGE": "nginx: worker process exited unexpectedly",
			"_SYSTEMD_UNIT": "nginx.service",
			"SYSLOG_IDENTIFIER": "nginx"
		}`,
		`{
			"PRIORITY": "2",
			"MESSAGE": "nginx: critical error occurred",
			"_SYSTEMD_UNIT": "nginx.service",
			"SYSLOG_IDENTIFIER": "nginx"
		}`,
		`{
			"PRIORITY": "4",
			"MESSAGE": "postgresql: warning about database connection",
			"_SYSTEMD_UNIT": "postgresql.service",
			"SYSLOG_IDENTIFIER": "postgresql"
		}`,
		`{
			"PRIORITY": "6",
			"MESSAGE": "postgresql: database connection established",
			"_SYSTEMD_UNIT": "postgresql.service",
			"SYSLOG_IDENTIFIER": "postgresql"
		}`,
	}

	var a Analyzer
	ar, err := a.Analyze(data)

	if err != nil {
		t.Fatal("expected no error, got %w", err)
	}

	if ar.TotalEntries != 4 {
		t.Errorf("expected 4 total entries, got %d", ar.TotalEntries)
	}
	if ar.ErrorCount != 2 {
		t.Errorf("expected 2 errors, got %d", ar.ErrorCount)
	}
	if ar.WarningCount != 1 {
		t.Errorf("expected 1 warning, got %d", ar.WarningCount)
	}
	if len(ar.ErrorDetails) != 2 {
		t.Errorf("expected 2 error details, got %d", len(ar.ErrorDetails))
	}
	if ar.ErrorsByUnit["nginx.service"] != 2 {
		t.Errorf("expected 2 nginx.service errors, got %d", ar.ErrorsByUnit["nginx.service"])
	}
	if ar.ErrorsByUnit["postgresql.service"] != 0 {
		t.Errorf("expected 0 postgresql.service errors, got %d", ar.ErrorsByUnit["postgresql.service"])
	}
}

func TestAnalyzeEmptyInput(t *testing.T) {
	data := []string{}
	var a Analyzer
	_, err := a.Analyze(data)
	if err == nil {
		t.Fatal("expected error for empty input")
	}
}

func TestAnalyzeInvalidJSON(t *testing.T) {
	data := []string{`{PRIORITY: }`}
	var a Analyzer
	_, err := a.Analyze(data)
	if err == nil {
		t.Fatal("expected error for invalid json")
	}
}
