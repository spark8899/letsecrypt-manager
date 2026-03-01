package handlers

import (
	"crypto/sha256"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"letsencrypt-manager/acme"
	"letsencrypt-manager/config"
	"letsencrypt-manager/logger"
	"letsencrypt-manager/middleware"
	"letsencrypt-manager/models"
)

type Handler struct {
	cfg   *config.Config
	store *models.Store

	// Track active ACME clients per domain (one order at a time)
	mu      sync.Mutex
	clients map[string]*acme.Client
}

func New(cfg *config.Config, store *models.Store) *Handler {
	return &Handler{
		cfg:     cfg,
		store:   store,
		clients: make(map[string]*acme.Client),
	}
}

func (h *Handler) acmeClient() (*acme.Client, error) {
	if err := config.ValidateACMEEmail(h.cfg.ACMEEmail); err != nil {
		return nil, err
	}
	return acme.NewClient(h.store, h.cfg.ACMEEmail, h.cfg.IsStaging())
}

// POST /api/auth/login
func (h *Handler) Login(ctx *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPw := fmt.Sprintf("%x", sha256.Sum256([]byte(req.Password)))

	for _, acc := range h.cfg.GetAccounts() {
		if acc.Username == req.Username && acc.Password == hashedPw {
			token, err := middleware.GenerateToken(req.Username, h.cfg.JWTSecret, h.cfg.TokenExpiry)
			if err != nil {
				logger.Error.Printf("Token generation error: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "token generation failed"})
				return
			}
			logger.Info.Printf("User logged in: %s", req.Username)
			ctx.JSON(http.StatusOK, gin.H{
				"token":      token,
				"expires_in": h.cfg.TokenExpiry * 3600,
				"token_type": "Bearer",
			})
			return
		}
	}

	logger.Warn.Printf("Failed login attempt: %s", req.Username)
	ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
}

// POST /api/domains — Add a wildcard domain
func (h *Handler) AddDomain(ctx *gin.Context) {
	var req struct {
		Domain string `json:"domain" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain := normalizeDomain(req.Domain)
	if domain == "" || !strings.Contains(domain, ".") {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain format"})
		return
	}

	info, err := h.store.AddDomain(domain)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Printf("Domain added: %s", domain)
	ctx.JSON(http.StatusOK, gin.H{
		"domain":  info.Domain,
		"status":  info.Status,
		"message": "domain added successfully, call /dns-challenge to start certificate issuance",
	})
}

// GET /api/domains — List all domains
func (h *Handler) ListDomains(ctx *gin.Context) {
	domains := h.store.ListDomains()
	ctx.JSON(http.StatusOK, gin.H{
		"total":   len(domains),
		"domains": domains,
	})
}

// POST /api/domains/:domain/dns-challenge
// Start ACME order and return DNS records that must be set
func (h *Handler) GetDNSChallenge(ctx *gin.Context) {
	domain := normalizeDomain(ctx.Param("domain"))

	info, err := h.store.GetDomain(domain)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "domain not found, add it first"})
		return
	}

	client, err := h.acmeClient()
	if err != nil {
		logger.Error.Printf("ACME client init error for %s: %v", domain, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "ACME client error: " + err.Error()})
		return
	}

	logger.Info.Printf("Starting ACME order for: %s", domain)
	ch, err := client.StartOrder(domain)
	if err != nil {
		logger.Error.Printf("ACME order error for %s: %v", domain, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start ACME order: " + err.Error()})
		return
	}

	// Save challenge info and store client
	h.mu.Lock()
	h.clients[domain] = client
	h.mu.Unlock()

	info.Challenge = ch
	info.Status = models.StatusVerifying
	_ = h.store.UpdateDomain(info)

	logger.Info.Printf("DNS challenge obtained for: %s -> %s", domain, ch.Domain)
	ctx.JSON(http.StatusOK, gin.H{
		"domain": domain,
		"status": info.Status,
		"challenge": gin.H{
			"type":          ch.Type,
			"record_name":   ch.Domain,
			"txt_value":     ch.KeyAuth,
			"cname_target":  ch.CNAMETarget,
			"instructions": fmt.Sprintf(
				"Option 1 — Add a TXT record:\n  Name:  %s\n  Value: %s\n\nOption 2 — Add a CNAME record (acme-dns delegation):\n  Name:  %s\n  Target: %s\n\nAfter setting DNS, call GET /api/domains/%s/dns-verify",
				ch.Domain, ch.KeyAuth, ch.Domain, ch.CNAMETarget, domain,
			),
		},
	})
}

// GET /api/domains/:domain/dns-verify
// Check if DNS TXT/CNAME record is set and signal ACME to proceed
func (h *Handler) VerifyDNS(ctx *gin.Context) {
	domain := normalizeDomain(ctx.Param("domain"))

	info, err := h.store.GetDomain(domain)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		return
	}

	if info.Challenge == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "no challenge found, call POST /dns-challenge first",
		})
		return
	}

	verified, detail := checkDNS(info.Challenge.Domain)
	logger.Info.Printf("DNS verify for %s: verified=%v detail=%s", domain, verified, detail)

	if verified {
		info.Status = models.StatusVerified
		_ = h.store.UpdateDomain(info)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"domain":              domain,
		"verified":            verified,
		"detail":              detail,
		"status":              info.Status,
		"challenge_record":    info.Challenge.Domain,
		"next_step":           "If verified=true, call POST /api/domains/" + domain + "/issue",
	})
}

// POST /api/domains/:domain/issue
// Complete ACME validation and issue certificate.
// Must be called after /dns-challenge has been called and DNS records are set.
func (h *Handler) IssueCertificate(ctx *gin.Context) {
	domain := normalizeDomain(ctx.Param("domain"))

	info, err := h.store.GetDomain(domain)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		return
	}

	h.mu.Lock()
	client, hasClient := h.clients[domain]
	h.mu.Unlock()

	if !hasClient {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf(
				"no active ACME order found for %s. "+
					"Call POST /api/domains/%s/dns-challenge first, "+
					"set the required DNS record, then call this endpoint.",
				domain, domain,
			),
		})
		return
	}

	logger.Info.Printf("Completing ACME order for: %s", domain)
	fullchain, privkey, expiry, err := client.ProceedWithOrder(domain)

	h.mu.Lock()
	delete(h.clients, domain)
	h.mu.Unlock()

	if err != nil {
		logger.Error.Printf("Certificate issuance failed for %s: %v", domain, err)
		info.Status = models.StatusFailed
		info.Error = err.Error()
		_ = h.store.UpdateDomain(info)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error":  "certificate issuance failed: " + err.Error(),
			"domain": domain,
		})
		return
	}

	if err := h.store.SaveCert(domain, fullchain, privkey); err != nil {
		logger.Error.Printf("Failed to save cert for %s: %v", domain, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save certificate files"})
		return
	}

	info.Status = models.StatusIssued
	info.CertExpiry = &expiry
	info.Error = ""
	_ = h.store.UpdateDomain(info)

	logger.Info.Printf("Certificate issued for %s, expires: %s", domain, expiry.Format(time.RFC3339))
	ctx.JSON(http.StatusOK, gin.H{
		"domain":          domain,
		"status":          "issued",
		"cert_expiry":     expiry.Format(time.RFC3339),
		"domains_covered": []string{"*." + domain, domain},
		"message":         "Certificate issued successfully. Retrieve it via GET /api/domains/" + domain + "/cert",
	})
}

// GET /api/domains/:domain/cert
// Return certificate and private key content
func (h *Handler) GetCertificate(ctx *gin.Context) {
	domain := normalizeDomain(ctx.Param("domain"))

	info, err := h.store.GetDomain(domain)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		return
	}

	if info.Status != models.StatusIssued {
		ctx.JSON(http.StatusNotFound, gin.H{
			"error":  "no certificate issued for this domain",
			"status": info.Status,
		})
		return
	}

	fullchain, privkey, err := h.store.ReadCert(domain)
	if err != nil {
		logger.Error.Printf("Failed to read cert for %s: %v", domain, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read certificate: " + err.Error()})
		return
	}

	expiry := ""
	if info.CertExpiry != nil {
		expiry = info.CertExpiry.Format(time.RFC3339)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"domain":          domain,
		"status":          info.Status,
		"cert_expiry":     expiry,
		"fullchain_cer":   string(fullchain),
		"private_key":     string(privkey),
		"domains_covered": []string{"*." + domain, domain},
	})
}

// --- Helpers ---

func normalizeDomain(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	domain = strings.TrimPrefix(domain, "*.")
	domain = strings.Trim(domain, "/")
	return domain
}

func checkDNS(challengeDomain string) (bool, string) {
	// Check CNAME
	cname, err := net.LookupCNAME(challengeDomain)
	if err == nil && cname != "" && cname != challengeDomain+"." {
		return true, fmt.Sprintf("CNAME found: %s -> %s", challengeDomain, cname)
	}

	// Check TXT
	txts, err := net.LookupTXT(challengeDomain)
	if err == nil {
		for _, txt := range txts {
			if len(txt) >= 20 {
				return true, fmt.Sprintf("TXT record found: %s (length=%d)", txt[:20]+"...", len(txt))
			}
		}
		if len(txts) > 0 {
			return true, fmt.Sprintf("TXT record found: %s", txts[0])
		}
	}

	return false, fmt.Sprintf("DNS record not found for %s (err: %v). DNS may not have propagated yet.", challengeDomain, err)
}
