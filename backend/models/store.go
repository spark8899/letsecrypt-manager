package models

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type DomainStatus string

const (
	StatusPending   DomainStatus = "pending"
	StatusVerifying DomainStatus = "verifying"
	StatusVerified  DomainStatus = "verified"
	StatusIssued    DomainStatus = "issued"
	StatusFailed    DomainStatus = "failed"
)

type DNSChallenge struct {
	Type       string `json:"type"`
	Domain     string `json:"domain"`       // _acme-challenge.domain
	CNAMETarget string `json:"cname_target"` // target for CNAME
	Token      string `json:"token"`
	KeyAuth    string `json:"key_auth"`
}

type DomainInfo struct {
	Domain      string       `json:"domain"`
	Status      DomainStatus `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Challenge   *DNSChallenge `json:"challenge,omitempty"`
	CertExpiry  *time.Time   `json:"cert_expiry,omitempty"`
	Error       string       `json:"error,omitempty"`
}

type Store struct {
	mu      sync.RWMutex
	dataDir string
	domains map[string]*DomainInfo
}

func NewStore(dataDir string) (*Store, error) {
	dirs := []string{
		dataDir,
		filepath.Join(dataDir, "domains"),
		filepath.Join(dataDir, "certs"),
		filepath.Join(dataDir, "accounts"),
		filepath.Join(dataDir, "logs"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return nil, fmt.Errorf("failed to create dir %s: %w", d, err)
		}
	}

	s := &Store{
		dataDir: dataDir,
		domains: make(map[string]*DomainInfo),
	}

	if err := s.loadDomains(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) loadDomains() error {
	pattern := filepath.Join(s.dataDir, "domains", "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var info DomainInfo
		if err := json.Unmarshal(data, &info); err != nil {
			continue
		}
		s.domains[info.Domain] = &info
	}
	return nil
}

func (s *Store) saveDomain(info *DomainInfo) error {
	path := filepath.Join(s.dataDir, "domains", info.Domain+".json")
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func (s *Store) AddDomain(domain string) (*DomainInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.domains[domain]; exists {
		return nil, fmt.Errorf("domain %s already exists", domain)
	}

	info := &DomainInfo{
		Domain:    domain,
		Status:    StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.domains[domain] = info
	return info, s.saveDomain(info)
}

func (s *Store) GetDomain(domain string) (*DomainInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, ok := s.domains[domain]
	if !ok {
		return nil, fmt.Errorf("domain %s not found", domain)
	}
	// return a copy
	copy := *info
	return &copy, nil
}

func (s *Store) UpdateDomain(info *DomainInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info.UpdatedAt = time.Now()
	s.domains[info.Domain] = info
	return s.saveDomain(info)
}

func (s *Store) ListDomains() []*DomainInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]*DomainInfo, 0, len(s.domains))
	for _, v := range s.domains {
		copy := *v
		list = append(list, &copy)
	}
	return list
}

// Certificate file helpers

func (s *Store) CertDir(domain string) string {
	return filepath.Join(s.dataDir, "certs", domain)
}

func (s *Store) SaveCert(domain string, fullchain, privkey []byte) error {
	dir := s.CertDir(domain)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, "fullchain.cer"), fullchain, 0600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "private.key"), privkey, 0600)
}

func (s *Store) ReadCert(domain string) (fullchain, privkey []byte, err error) {
	dir := s.CertDir(domain)
	fullchain, err = os.ReadFile(filepath.Join(dir, "fullchain.cer"))
	if err != nil {
		return nil, nil, fmt.Errorf("fullchain not found: %w", err)
	}
	privkey, err = os.ReadFile(filepath.Join(dir, "private.key"))
	if err != nil {
		return nil, nil, fmt.Errorf("private key not found: %w", err)
	}
	return fullchain, privkey, nil
}

// ACME account key storage

func (s *Store) AccountKeyPath() string {
	return filepath.Join(s.dataDir, "accounts", "account.key")
}

func (s *Store) AccountInfoPath() string {
	return filepath.Join(s.dataDir, "accounts", "account.json")
}
