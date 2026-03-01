package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_DefaultsCreatedWhenMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ListenAddr != ":8080" {
		t.Errorf("expected default ListenAddr :8080, got %s", cfg.ListenAddr)
	}
	if cfg.TokenExpiry != 24 {
		t.Errorf("expected default TokenExpiry 24, got %d", cfg.TokenExpiry)
	}
	if len(cfg.Accounts) == 0 {
		t.Error("expected at least one default account")
	}

	// Config file should have been written to disk
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected config file to be created on disk")
	}
}

func TestLoad_ReadsExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	raw := Config{
		ListenAddr:  ":9090",
		DataDir:     "/tmp/mydata",
		ACMEEmail:   "test@example.com",
		ACMEServer:  "production",
		JWTSecret:   "supersecret",
		TokenExpiry: 48,
		Accounts: []Account{
			{Username: "alice", Password: "abc123"},
			{Username: "bob", Password: "def456"},
		},
	}
	data, _ := json.MarshalIndent(raw, "", "  ")
	os.WriteFile(path, data, 0600)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.ListenAddr != ":9090" {
		t.Errorf("ListenAddr = %s, want :9090", cfg.ListenAddr)
	}
	if cfg.ACMEEmail != "test@example.com" {
		t.Errorf("ACMEEmail = %s, want test@example.com", cfg.ACMEEmail)
	}
	if cfg.TokenExpiry != 48 {
		t.Errorf("TokenExpiry = %d, want 48", cfg.TokenExpiry)
	}
	if len(cfg.Accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(cfg.Accounts))
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	os.WriteFile(path, []byte("{invalid json"), 0600)

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

func TestLoad_TokenExpiryDefaultsTo24WhenZero(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	raw := map[string]interface{}{
		"listen_addr": ":8080",
		"jwt_secret":  "s",
		// token_expiry_hours omitted (zero value)
	}
	data, _ := json.Marshal(raw)
	os.WriteFile(path, data, 0600)

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.TokenExpiry != 24 {
		t.Errorf("expected TokenExpiry=24 when zero, got %d", cfg.TokenExpiry)
	}
}

func TestIsStaging(t *testing.T) {
	cases := []struct {
		server   string
		expected bool
	}{
		{"staging", true},
		{"", true},
		{"production", false},
		{"PRODUCTION", false},
	}
	for _, c := range cases {
		cfg := &Config{ACMEServer: c.server}
		if got := cfg.IsStaging(); got != c.expected {
			t.Errorf("IsStaging(%q) = %v, want %v", c.server, got, c.expected)
		}
	}
}

func TestGetAccounts_Concurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	cfg, _ := Load(path)

	// Concurrent reads should not race
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			_ = cfg.GetAccounts()
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestSave_PersistsToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	cfg.ListenAddr = ":7777"
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Reload and verify
	cfg2, err := Load(path)
	if err != nil {
		t.Fatalf("Load after Save() error = %v", err)
	}
	if cfg2.ListenAddr != ":7777" {
		t.Errorf("ListenAddr after save = %s, want :7777", cfg2.ListenAddr)
	}
}

// ---- ValidateACMEEmail ----

func TestValidateACMEEmail_Valid(t *testing.T) {
	cases := []string{
		"user@gmail.com",
		"admin@mycompany.io",
		"cert-bot@sub.domain.org",
	}
	for _, email := range cases {
		if err := ValidateACMEEmail(email); err != nil {
			t.Errorf("ValidateACMEEmail(%q) unexpected error: %v", email, err)
		}
	}
}

func TestValidateACMEEmail_Empty(t *testing.T) {
	if err := ValidateACMEEmail(""); err == nil {
		t.Error("expected error for empty email, got nil")
	}
}

func TestValidateACMEEmail_ForbiddenDomains(t *testing.T) {
	cases := []string{
		"admin@example.com",
		"user@example.org",
		"test@test.com",
		"foo@localhost",
		"a@invalid",
	}
	for _, email := range cases {
		if err := ValidateACMEEmail(email); err == nil {
			t.Errorf("ValidateACMEEmail(%q): expected error for forbidden domain, got nil", email)
		}
	}
}

func TestValidateACMEEmail_MalformedEmail(t *testing.T) {
	cases := []string{
		"notanemail",
		"@nodomain",
		"noat.",
	}
	for _, email := range cases {
		if err := ValidateACMEEmail(email); err == nil {
			t.Errorf("ValidateACMEEmail(%q): expected error for malformed email, got nil", email)
		}
	}
}
