package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit_CreatesLogDir(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "logs")

	if err := Init(dir); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("expected log directory to be created")
	}
}

func TestInit_CreatesLogFile(t *testing.T) {
	dir := t.TempDir()

	if err := Init(dir); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("expected at least one log file in log dir")
	}

	// File should be named app-YYYY-MM-DD.log
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "app-") && strings.HasSuffix(e.Name(), ".log") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected app-YYYY-MM-DD.log file, got: %v", entries)
	}
}

func TestInit_LoggersNotNil(t *testing.T) {
	dir := t.TempDir()
	Init(dir)

	if Info == nil {
		t.Error("Info logger is nil after Init()")
	}
	if Warn == nil {
		t.Error("Warn logger is nil after Init()")
	}
	if Error == nil {
		t.Error("Error logger is nil after Init()")
	}
}

func TestInit_LoggersWriteToFile(t *testing.T) {
	dir := t.TempDir()
	Init(dir)

	Info.Println("test info message")
	Warn.Println("test warn message")
	Error.Println("test error message")

	entries, _ := os.ReadDir(dir)
	var logFile string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "app-") {
			logFile = filepath.Join(dir, e.Name())
		}
	}
	if logFile == "" {
		t.Fatal("no log file found")
	}

	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	for _, expected := range []string{"test info message", "test warn message", "test error message"} {
		if !strings.Contains(string(content), expected) {
			t.Errorf("log file should contain %q", expected)
		}
	}
}

func TestInit_LogPrefixes(t *testing.T) {
	dir := t.TempDir()
	Init(dir)

	if !strings.Contains(Info.Prefix(), "INFO") {
		t.Errorf("Info prefix = %q, expected to contain INFO", Info.Prefix())
	}
	if !strings.Contains(Warn.Prefix(), "WARN") {
		t.Errorf("Warn prefix = %q, expected to contain WARN", Warn.Prefix())
	}
	if !strings.Contains(Error.Prefix(), "ERROR") {
		t.Errorf("Error prefix = %q, expected to contain ERROR", Error.Prefix())
	}
}
