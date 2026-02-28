package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

// forbiddenEmailDomains are placeholder domains rejected by Let's Encrypt.
var forbiddenEmailDomains = []string{
	"example.com", "example.org", "example.net",
	"test.com", "localhost", "invalid",
}

// ValidateACMEEmail returns a descriptive error if the email is a placeholder
// or malformed, preventing a confusing ACME API error later.
func ValidateACMEEmail(email string) error {
	if email == "" {
		return fmt.Errorf("acme_email is required — set a real email address in config.json")
	}
	atIdx := strings.LastIndex(email, "@")
	if atIdx < 1 || atIdx == len(email)-1 {
		return fmt.Errorf("acme_email %q is not a valid email address", email)
	}
	domain := strings.ToLower(email[atIdx+1:])
	for _, forbidden := range forbiddenEmailDomains {
		if domain == forbidden {
			return fmt.Errorf(
				"acme_email %q uses placeholder domain %q — "+
					"Let's Encrypt rejects this.\n"+
					"Fix: set a real email in config.json → \"acme_email\": \"you@yourdomain.com\"",
				email, domain,
			)
		}
	}
	return nil
}

type Account struct {
	Username string `json:"username"`
	Password string `json:"password"` // sha256 hashed
}

type Config struct {
	mu          sync.RWMutex
	ListenAddr  string    `json:"listen_addr"`
	DataDir     string    `json:"data_dir"`
	ACMEEmail   string    `json:"acme_email"`
	ACMEServer  string    `json:"acme_server"` // "staging" or "production"
	JWTSecret   string    `json:"jwt_secret"`
	TokenExpiry int       `json:"token_expiry_hours"` // default 24
	Accounts    []Account `json:"accounts"`
	configPath  string
}

var defaultConfig = Config{
	ListenAddr:  ":8080",
	DataDir:     "./data",
	ACMEEmail:   "",
	ACMEServer:  "staging",
	JWTSecret:   "change-this-secret-in-production",
	TokenExpiry: 24,
	Accounts: []Account{
		{
			Username: "admin",
			// default password: "admin123" sha256 hashed
			Password: "240be518fabd2724ddb6f04eeb1da5967448d7e831c08c8fa822809f74c720a9",
		},
	},
}

func Load(path string) (*Config, error) {
	cfg := defaultConfig
	cfg.configPath = path

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := cfg.Save(); err != nil {
				return nil, fmt.Errorf("failed to create default config: %w", err)
			}
			return &cfg, nil
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	cfg.configPath = path

	if cfg.TokenExpiry == 0 {
		cfg.TokenExpiry = 24
	}

	return &cfg, nil
}

func (c *Config) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(c.configPath, data, 0600)
}

func (c *Config) GetAccounts() []Account {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Accounts
}

func (c *Config) IsStaging() bool {
	return c.ACMEServer == "staging" || c.ACMEServer == ""
}
