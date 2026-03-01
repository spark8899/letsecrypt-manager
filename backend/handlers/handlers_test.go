package handlers

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"letsencrypt-manager/config"
	"letsencrypt-manager/logger"
	"letsencrypt-manager/middleware"
	"letsencrypt-manager/models"
)

func init() {
	gin.SetMode(gin.TestMode)
	// handlers.go calls logger.Info/Warn/Error — must init before any handler runs
	logDir := os.TempDir() + "/letsencrypt-test-logs"
	os.MkdirAll(logDir, 0755)
	if err := logger.Init(logDir); err != nil {
		panic("failed to init logger for tests: " + err.Error())
	}
}

// testApp builds a complete router identical to main.go
func testApp(t *testing.T) (*gin.Engine, *Handler) {
	t.Helper()

	dir := t.TempDir()
	store, err := models.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	adminHash := fmt.Sprintf("%x", sha256.Sum256([]byte("admin123")))
	cfg := &config.Config{
		ListenAddr:  ":8080",
		DataDir:     dir,
		ACMEEmail:   "test@example.com",
		ACMEServer:  "staging",
		JWTSecret:   "test-jwt-secret",
		TokenExpiry: 24,
		Accounts: []config.Account{
			{Username: "admin", Password: adminHash},
		},
	}

	h := New(cfg, store)

	r := gin.New()
	r.POST("/api/auth/login", h.Login)
	r.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := r.Group("/api", middleware.AuthMiddleware(cfg.JWTSecret))
	api.POST("/domains", h.AddDomain)
	api.GET("/domains", h.ListDomains)
	api.POST("/domains/:domain/dns-challenge", h.GetDNSChallenge)
	api.GET("/domains/:domain/dns-verify", h.VerifyDNS)
	api.POST("/domains/:domain/issue", h.IssueCertificate)
	api.GET("/domains/:domain/cert", h.GetCertificate)

	return r, h
}

// getToken logs in and returns a JWT token
func getToken(t *testing.T, r *gin.Engine, username, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": username, "password": password})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp["token"].(string)
}

func doRequest(r *gin.Engine, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(data)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, bodyReader)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func parseJSON(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("failed to parse JSON response: %v\nbody: %s", err, w.Body.String())
	}
	return m
}

// =========================================================
// Login tests
// =========================================================

func TestLogin_Success(t *testing.T) {
	r, _ := testApp(t)

	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin123"})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}

	resp := parseJSON(t, w)
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected non-empty token in response")
	}
	if resp["token_type"] != "Bearer" {
		t.Errorf("token_type = %v, want Bearer", resp["token_type"])
	}
	if resp["expires_in"] == nil {
		t.Error("expected expires_in in response")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	r, _ := testApp(t)
	w := doRequest(r, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "admin",
		"password": "wrongpassword",
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestLogin_UnknownUser(t *testing.T) {
	r, _ := testApp(t)
	w := doRequest(r, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "nobody",
		"password": "admin123",
	})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestLogin_MissingFields(t *testing.T) {
	r, _ := testApp(t)

	// Missing password
	w := doRequest(r, http.MethodPost, "/api/auth/login", "", map[string]string{
		"username": "admin",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestLogin_EmptyBody(t *testing.T) {
	r, _ := testApp(t)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// =========================================================
// AddDomain tests
// =========================================================

func TestAddDomain_Success(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	w := doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{
		"domain": "example.com",
	})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(t, w)
	if resp["domain"] != "example.com" {
		t.Errorf("domain = %v, want example.com", resp["domain"])
	}
	if resp["status"] != string(models.StatusPending) {
		t.Errorf("status = %v, want pending", resp["status"])
	}
}

func TestAddDomain_WithWildcardPrefix(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	// User inputs "*.example.com" — should normalize to "example.com"
	w := doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{
		"domain": "*.example.com",
	})
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200: %s", w.Code, w.Body.String())
	}
	resp := parseJSON(t, w)
	if resp["domain"] != "example.com" {
		t.Errorf("domain = %v, want example.com after wildcard strip", resp["domain"])
	}
}

func TestAddDomain_Duplicate(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{"domain": "dup.com"})
	w := doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{"domain": "dup.com"})

	if w.Code != http.StatusConflict {
		t.Errorf("status = %d, want 409", w.Code)
	}
}

func TestAddDomain_InvalidDomain(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	cases := []string{"nodot", "", "   "}
	for _, d := range cases {
		w := doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{"domain": d})
		if w.Code != http.StatusBadRequest {
			t.Errorf("domain=%q: status = %d, want 400", d, w.Code)
		}
	}
}

func TestAddDomain_Unauthorized(t *testing.T) {
	r, _ := testApp(t)
	w := doRequest(r, http.MethodPost, "/api/domains", "", map[string]string{"domain": "example.com"})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

func TestAddDomain_UppercaseNormalized(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	w := doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{
		"domain": "UPPER.COM",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", w.Code, w.Body.String())
	}
	resp := parseJSON(t, w)
	if resp["domain"] != "upper.com" {
		t.Errorf("domain = %v, want upper.com", resp["domain"])
	}
}

// =========================================================
// ListDomains tests
// =========================================================

func TestListDomains_Empty(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	w := doRequest(r, http.MethodGet, "/api/domains", token, nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	resp := parseJSON(t, w)
	domains := resp["domains"].([]interface{})
	if len(domains) != 0 {
		t.Errorf("expected 0 domains, got %d", len(domains))
	}
}

func TestListDomains_AfterAdding(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{"domain": "a.com"})
	doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{"domain": "b.com"})
	doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{"domain": "c.com"})

	w := doRequest(r, http.MethodGet, "/api/domains", token, nil)
	resp := parseJSON(t, w)

	total := int(resp["total"].(float64))
	if total != 3 {
		t.Errorf("total = %d, want 3", total)
	}
}

// =========================================================
// VerifyDNS tests
// =========================================================

func TestVerifyDNS_NoChallengeSet(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{"domain": "nodns.com"})

	w := doRequest(r, http.MethodGet, "/api/domains/nodns.com/dns-verify", token, nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestVerifyDNS_DomainNotFound(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	w := doRequest(r, http.MethodGet, "/api/domains/notexist.com/dns-verify", token, nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestVerifyDNS_WithChallengeSetsVerifying(t *testing.T) {
	r, h := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	// Add domain and manually inject challenge
	h.store.AddDomain("manual.com")
	info, _ := h.store.GetDomain("manual.com")
	info.Challenge = &models.DNSChallenge{
		Type:   "dns-01",
		Domain: "_acme-challenge.manual.com",
	}
	h.store.UpdateDomain(info)

	w := doRequest(r, http.MethodGet, "/api/domains/manual.com/dns-verify", token, nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(t, w)
	// verified field must exist
	if _, ok := resp["verified"]; !ok {
		t.Error("response should contain 'verified' field")
	}
	// challenge_record must match
	if resp["challenge_record"] != "_acme-challenge.manual.com" {
		t.Errorf("challenge_record = %v", resp["challenge_record"])
	}
}

// =========================================================
// GetCertificate tests
// =========================================================

func TestGetCertificate_NotIssued(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{"domain": "nocert.com"})

	w := doRequest(r, http.MethodGet, "/api/domains/nocert.com/cert", token, nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestGetCertificate_DomainNotFound(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	w := doRequest(r, http.MethodGet, "/api/domains/ghost.com/cert", token, nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestGetCertificate_IssuedReturnsContent(t *testing.T) {
	r, h := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	// Inject a domain with issued status + cert files directly via store
	h.store.AddDomain("issued.com")

	fullchain := []byte("-----BEGIN CERTIFICATE-----\nMIItest\n-----END CERTIFICATE-----\n")
	privkey := []byte("-----BEGIN EC PRIVATE KEY-----\nMIItest\n-----END EC PRIVATE KEY-----\n")
	h.store.SaveCert("issued.com", fullchain, privkey)

	expiry := time.Now().Add(90 * 24 * time.Hour)
	info, _ := h.store.GetDomain("issued.com")
	info.Status = models.StatusIssued
	info.CertExpiry = &expiry
	h.store.UpdateDomain(info)

	w := doRequest(r, http.MethodGet, "/api/domains/issued.com/cert", token, nil)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200: %s", w.Code, w.Body.String())
	}

	resp := parseJSON(t, w)
	if resp["fullchain_cer"] == nil || resp["fullchain_cer"] == "" {
		t.Error("expected fullchain_cer in response")
	}
	if resp["private_key"] == nil || resp["private_key"] == "" {
		t.Error("expected private_key in response")
	}
	if resp["cert_expiry"] == nil {
		t.Error("expected cert_expiry in response")
	}

	covered := resp["domains_covered"].([]interface{})
	if len(covered) != 2 {
		t.Errorf("expected 2 domains_covered, got %d", len(covered))
	}
}

// =========================================================
// IssueCertificate tests (error paths only — no real ACME)
// =========================================================

func TestIssueCertificate_DomainNotFound(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	w := doRequest(r, http.MethodPost, "/api/domains/ghost.com/issue", token, nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestIssueCertificate_Unauthorized(t *testing.T) {
	r, _ := testApp(t)
	w := doRequest(r, http.MethodPost, "/api/domains/example.com/issue", "", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", w.Code)
	}
}

// =========================================================
// normalizeDomain unit tests
// =========================================================

func TestNormalizeDomain(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"example.com", "example.com"},
		{"*.example.com", "example.com"},
		{"EXAMPLE.COM", "example.com"},
		{"  example.com  ", "example.com"},
		{"*.UPPER.COM", "upper.com"},
		{"/example.com/", "example.com"},
	}
	for _, c := range cases {
		got := normalizeDomain(c.input)
		if got != c.expected {
			t.Errorf("normalizeDomain(%q) = %q, want %q", c.input, got, c.expected)
		}
	}
}

// =========================================================
// Health check
// =========================================================

func TestHealthCheck(t *testing.T) {
	r, _ := testApp(t)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

// =========================================================
// End-to-end domain workflow (without real ACME)
// =========================================================

func TestWorkflow_AddAndList(t *testing.T) {
	r, _ := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	// Add several domains
	domains := []string{"foo.com", "bar.com", "baz.io"}
	for _, d := range domains {
		w := doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{"domain": d})
		if w.Code != http.StatusOK {
			t.Errorf("add %s: status = %d", d, w.Code)
		}
	}

	// List and verify count
	w := doRequest(r, http.MethodGet, "/api/domains", token, nil)
	resp := parseJSON(t, w)
	if int(resp["total"].(float64)) != len(domains) {
		t.Errorf("total = %v, want %d", resp["total"], len(domains))
	}
}

func TestWorkflow_ManualCertInjection(t *testing.T) {
	r, h := testApp(t)
	token := getToken(t, r, "admin", "admin123")

	// 1. Add domain
	doRequest(r, http.MethodPost, "/api/domains", token, map[string]string{"domain": "workflow.com"})

	// 2. Simulate challenge set
	info, _ := h.store.GetDomain("workflow.com")
	info.Challenge = &models.DNSChallenge{
		Type:        "dns-01",
		Domain:      "_acme-challenge.workflow.com",
		CNAMETarget: "_acme-challenge.workflow.com.acme-dns.io.",
		KeyAuth:     "test-key-auth-value",
	}
	info.Status = models.StatusVerifying
	h.store.UpdateDomain(info)

	// 3. Check DNS verify returns challenge info
	w := doRequest(r, http.MethodGet, "/api/domains/workflow.com/dns-verify", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("dns-verify status = %d", w.Code)
	}
	resp := parseJSON(t, w)
	if resp["challenge_record"] != "_acme-challenge.workflow.com" {
		t.Errorf("challenge_record = %v", resp["challenge_record"])
	}

	// 4. Simulate certificate issuance
	fullchain := []byte("-----BEGIN CERTIFICATE-----\ncert-data\n-----END CERTIFICATE-----\n")
	privkey := []byte("-----BEGIN EC PRIVATE KEY-----\nkey-data\n-----END EC PRIVATE KEY-----\n")
	h.store.SaveCert("workflow.com", fullchain, privkey)

	expiry := time.Now().Add(90 * 24 * time.Hour)
	info, _ = h.store.GetDomain("workflow.com")
	info.Status = models.StatusIssued
	info.CertExpiry = &expiry
	h.store.UpdateDomain(info)

	// 5. Retrieve cert
	w = doRequest(r, http.MethodGet, "/api/domains/workflow.com/cert", token, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("cert status = %d: %s", w.Code, w.Body.String())
	}
	resp = parseJSON(t, w)
	if resp["status"] != string(models.StatusIssued) {
		t.Errorf("status = %v, want issued", resp["status"])
	}
	if resp["fullchain_cer"] == "" {
		t.Error("fullchain_cer should not be empty")
	}
}
