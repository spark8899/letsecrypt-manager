package models

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	store, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	return store
}

// ---- NewStore ----

func TestNewStore_CreatesDirectories(t *testing.T) {
	dir := t.TempDir()
	_, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	for _, sub := range []string{"domains", "certs", "accounts", "logs"} {
		path := filepath.Join(dir, sub)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected directory %s to exist", path)
		}
	}
}

func TestNewStore_LoadsExistingDomains(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(dir)
	store.AddDomain("example.com")
	store.AddDomain("another.com")

	// Re-open same dir
	store2, err := NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore() reload error = %v", err)
	}

	list := store2.ListDomains()
	if len(list) != 2 {
		t.Errorf("expected 2 domains after reload, got %d", len(list))
	}
}

// ---- AddDomain ----

func TestAddDomain_Success(t *testing.T) {
	store := newTestStore(t)

	info, err := store.AddDomain("example.com")
	if err != nil {
		t.Fatalf("AddDomain() error = %v", err)
	}
	if info.Domain != "example.com" {
		t.Errorf("Domain = %s, want example.com", info.Domain)
	}
	if info.Status != StatusPending {
		t.Errorf("Status = %s, want %s", info.Status, StatusPending)
	}
	if info.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
}

func TestAddDomain_DuplicateReturnsError(t *testing.T) {
	store := newTestStore(t)
	store.AddDomain("example.com")

	_, err := store.AddDomain("example.com")
	if err == nil {
		t.Error("expected error for duplicate domain, got nil")
	}
}

func TestAddDomain_PersistsToDisk(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(dir)
	store.AddDomain("persist.com")

	path := filepath.Join(dir, "domains", "persist.com.json")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected domain JSON file to be created on disk")
	}
}

// ---- GetDomain ----

func TestGetDomain_ExistingDomain(t *testing.T) {
	store := newTestStore(t)
	store.AddDomain("get.com")

	info, err := store.GetDomain("get.com")
	if err != nil {
		t.Fatalf("GetDomain() error = %v", err)
	}
	if info.Domain != "get.com" {
		t.Errorf("Domain = %s, want get.com", info.Domain)
	}
}

func TestGetDomain_NotFound(t *testing.T) {
	store := newTestStore(t)

	_, err := store.GetDomain("notexist.com")
	if err == nil {
		t.Error("expected error for non-existent domain, got nil")
	}
}

func TestGetDomain_ReturnsCopy(t *testing.T) {
	store := newTestStore(t)
	store.AddDomain("copy.com")

	info, _ := store.GetDomain("copy.com")
	info.Status = StatusIssued // mutate the copy

	// Original should be unchanged
	orig, _ := store.GetDomain("copy.com")
	if orig.Status != StatusPending {
		t.Errorf("original domain status was mutated: got %s, want %s", orig.Status, StatusPending)
	}
}

// ---- UpdateDomain ----

func TestUpdateDomain_ChangesStatus(t *testing.T) {
	store := newTestStore(t)
	store.AddDomain("update.com")

	info, _ := store.GetDomain("update.com")
	info.Status = StatusVerified
	if err := store.UpdateDomain(info); err != nil {
		t.Fatalf("UpdateDomain() error = %v", err)
	}

	updated, _ := store.GetDomain("update.com")
	if updated.Status != StatusVerified {
		t.Errorf("Status = %s, want %s", updated.Status, StatusVerified)
	}
}

func TestUpdateDomain_SetsUpdatedAt(t *testing.T) {
	store := newTestStore(t)
	store.AddDomain("timestamp.com")

	info, _ := store.GetDomain("timestamp.com")
	before := info.UpdatedAt
	time.Sleep(2 * time.Millisecond)

	info.Status = StatusVerified
	store.UpdateDomain(info)

	after, _ := store.GetDomain("timestamp.com")
	if !after.UpdatedAt.After(before) {
		t.Error("UpdatedAt should be updated after UpdateDomain()")
	}
}

// ---- ListDomains ----

func TestListDomains_Empty(t *testing.T) {
	store := newTestStore(t)
	list := store.ListDomains()
	if len(list) != 0 {
		t.Errorf("expected empty list, got %d", len(list))
	}
}

func TestListDomains_Multiple(t *testing.T) {
	store := newTestStore(t)
	store.AddDomain("a.com")
	store.AddDomain("b.com")
	store.AddDomain("c.com")

	list := store.ListDomains()
	if len(list) != 3 {
		t.Errorf("expected 3 domains, got %d", len(list))
	}
}

// ---- SaveCert / ReadCert ----

func TestSaveCert_AndReadBack(t *testing.T) {
	store := newTestStore(t)

	fullchain := []byte("-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----\n")
	privkey := []byte("-----BEGIN EC PRIVATE KEY-----\ntest\n-----END EC PRIVATE KEY-----\n")

	if err := store.SaveCert("certtest.com", fullchain, privkey); err != nil {
		t.Fatalf("SaveCert() error = %v", err)
	}

	gotChain, gotKey, err := store.ReadCert("certtest.com")
	if err != nil {
		t.Fatalf("ReadCert() error = %v", err)
	}
	if string(gotChain) != string(fullchain) {
		t.Error("fullchain mismatch")
	}
	if string(gotKey) != string(privkey) {
		t.Error("private key mismatch")
	}
}

func TestReadCert_NotFound(t *testing.T) {
	store := newTestStore(t)

	_, _, err := store.ReadCert("notexist.com")
	if err == nil {
		t.Error("expected error reading non-existent cert, got nil")
	}
}

func TestSaveCert_FilePermissions(t *testing.T) {
	store := newTestStore(t)
	store.SaveCert("perm.com", []byte("chain"), []byte("key"))

	dir := store.CertDir("perm.com")
	for _, name := range []string{"fullchain.cer", "private.key"} {
		info, err := os.Stat(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if perm := info.Mode().Perm(); perm != 0600 {
			t.Errorf("%s permissions = %o, want 0600", name, perm)
		}
	}
}

// ---- Concurrent access ----

func TestStore_ConcurrentAddAndList(t *testing.T) {
	store := newTestStore(t)
	var wg sync.WaitGroup

	// Concurrent adds with unique domains
	for i := 0; i < 20; i++ {
		wg.Add(1)
		i := i
		go func() {
			defer wg.Done()
			domain := "concurrent" + string(rune('a'+i)) + ".com"
			store.AddDomain(domain)
		}()
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.ListDomains()
		}()
	}
	wg.Wait()
}

// ---- AccountKeyPath / AccountInfoPath ----

func TestAccountPaths(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(dir)

	keyPath := store.AccountKeyPath()
	if keyPath == "" {
		t.Error("AccountKeyPath() returned empty string")
	}

	infoPath := store.AccountInfoPath()
	if infoPath == "" {
		t.Error("AccountInfoPath() returned empty string")
	}

	// Paths should be inside the data dir
	rel, err := filepath.Rel(dir, keyPath)
	if err != nil || rel == "" {
		t.Error("AccountKeyPath should be inside data dir")
	}
}
