package acme

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"letsencrypt-manager/logger"
	"letsencrypt-manager/models"
)

// ---- ACME account ----

type ACMEUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *ACMEUser) GetEmail() string                        { return u.Email }
func (u *ACMEUser) GetRegistration() *registration.Resource { return u.Registration }
func (u *ACMEUser) GetPrivateKey() crypto.PrivateKey        { return u.key }

type savedAccount struct {
	Email        string                 `json:"email"`
	Registration *registration.Resource `json:"registration"`
}

// ---- orderResult carries the outcome of a background Obtain() call ----

type orderResult struct {
	resource *certificate.Resource
	err      error
}

// ---- waitingDNSProvider ----
// Present() captures challenge info, signals ready, then BLOCKS until
// the caller closes proceed. The certificate resource is returned via resultCh.

type waitingDNSProvider struct {
	mu         sync.Mutex
	challenges map[string]*models.DNSChallenge

	readyOnce sync.Once    // ensures ready is closed exactly once
	ready     chan struct{} // closed when first Present() is called

	proceed  chan struct{}    // closed by caller to unblock ALL Present() goroutines
	resultCh chan orderResult // receives the Obtain() result
}

func newWaitingProvider() *waitingDNSProvider {
	return &waitingDNSProvider{
		challenges: make(map[string]*models.DNSChallenge),
		ready:      make(chan struct{}),
		proceed:    make(chan struct{}),
		resultCh:   make(chan orderResult, 1),
	}
}

func (p *waitingDNSProvider) Present(domain, token, keyAuth string) error {
	p.mu.Lock()
	p.challenges[domain] = &models.DNSChallenge{
		Type:        "dns-01",
		Domain:      domain,
		CNAMETarget: domain + ".acme-dns.io.",
		Token:       token,
		KeyAuth:     keyAuth,
	}
	p.mu.Unlock()

	logger.Info.Printf("ACME DNS challenge present: domain=%s token=%s", domain, token)

	// Signal ready exactly once — lego calls Present() in parallel for wildcard
	// (*.domain) + base (domain), so two goroutines hit this concurrently.
	// sync.Once guarantees only the first close() runs; the second is a no-op.
	p.readyOnce.Do(func() { close(p.ready) })

	// Block all Present() goroutines until caller confirms DNS records are set.
	// Reading from a closed channel returns immediately, so closing proceed
	// unblocks every waiting goroutine at once — no double-close panic.
	<-p.proceed
	return nil
}

func (p *waitingDNSProvider) CleanUp(domain, token, keyAuth string) error {
	logger.Info.Printf("ACME DNS challenge cleanup: %s", domain)
	return nil
}

func (p *waitingDNSProvider) Timeout() (time.Duration, time.Duration) {
	return 24 * time.Hour, 30 * time.Second
}

func (p *waitingDNSProvider) getChallenge() *models.DNSChallenge {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, v := range p.challenges {
		return v
	}
	return nil
}

var _ challenge.Provider = (*waitingDNSProvider)(nil)

// ---- activeOrder tracks an in-flight certificate order ----

type activeOrder struct {
	provider *waitingDNSProvider
	// challenge info returned to the caller
	challenge *models.DNSChallenge
}

// ---- Client ----

type Client struct {
	store      *models.Store
	email      string
	legoClient *lego.Client
	user       *ACMEUser
	staging    bool

	mu     sync.Mutex
	orders map[string]*activeOrder // keyed by domain
}

func NewClient(store *models.Store, email string, staging bool) (*Client, error) {
	c := &Client{
		store:   store,
		email:   email,
		staging: staging,
		orders:  make(map[string]*activeOrder),
	}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) init() error {
	key, err := c.loadOrGenerateKey()
	if err != nil {
		return fmt.Errorf("account key: %w", err)
	}
	c.user = &ACMEUser{Email: c.email, key: key}

	cfg := lego.NewConfig(c.user)
	if c.staging {
		cfg.CADirURL = lego.LEDirectoryStaging
		logger.Info.Println("ACME: Using Let's Encrypt STAGING environment")
	} else {
		cfg.CADirURL = lego.LEDirectoryProduction
		logger.Info.Println("ACME: Using Let's Encrypt PRODUCTION environment")
	}

	legoClient, err := lego.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("lego client: %w", err)
	}
	if err := c.loadOrRegister(legoClient); err != nil {
		return fmt.Errorf("registration: %w", err)
	}
	c.legoClient = legoClient
	return nil
}

func (c *Client) loadOrGenerateKey() (crypto.PrivateKey, error) {
	keyPath := c.store.AccountKeyPath()
	data, err := os.ReadFile(keyPath)
	if err == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
				logger.Info.Println("ACME: Loaded existing account key")
				return key, nil
			}
		}
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	keyBytes, _ := x509.MarshalECPrivateKey(key)
	pemData := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	if err := os.WriteFile(keyPath, pemData, 0600); err != nil {
		return nil, err
	}
	logger.Info.Println("ACME: Generated new account key")
	return key, nil
}

func (c *Client) loadOrRegister(legoClient *lego.Client) error {
	infoPath := c.store.AccountInfoPath()
	data, err := os.ReadFile(infoPath)
	if err == nil {
		var saved savedAccount
		if json.Unmarshal(data, &saved) == nil && saved.Registration != nil {
			c.user.Registration = saved.Registration
			logger.Info.Printf("ACME: Loaded registration: %s", saved.Registration.URI)
			return nil
		}
	}

	reg, err := legoClient.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}
	c.user.Registration = reg

	saved := savedAccount{Email: c.email, Registration: reg}
	saveData, _ := json.MarshalIndent(saved, "", "  ")
	_ = os.WriteFile(infoPath, saveData, 0600)
	logger.Info.Printf("ACME: Registered new account: %s", reg.URI)
	return nil
}

// StartOrder begins a certificate order for *.domain + domain.
// It returns the DNS challenge info that the caller must publish as a DNS TXT record.
// The order is paused at the DNS validation step — call ProceedWithOrder() after
// publishing DNS to complete the order and receive the certificate.
func (c *Client) StartOrder(domain string) (*models.DNSChallenge, error) {
	provider := newWaitingProvider()

	if err := c.legoClient.Challenge.SetDNS01Provider(provider); err != nil {
		return nil, fmt.Errorf("set DNS provider: %w", err)
	}

	domains := []string{"*." + domain, domain}
	logger.Info.Printf("ACME: Starting order for domains: %v", domains)

	// Launch Obtain in background; it will call provider.Present() and block there.
	// When we later close(provider.proceed), Obtain resumes and completes.
	// The result (cert or error) is sent to provider.resultCh.
	go func() {
		res, err := c.legoClient.Certificate.Obtain(certificate.ObtainRequest{
			Domains: domains,
			Bundle:  true,
		})
		provider.resultCh <- orderResult{resource: res, err: err}
	}()

	// Wait until Present() has been called and challenge info is available
	select {
	case <-provider.ready:
	case <-time.After(60 * time.Second):
		return nil, fmt.Errorf("timeout waiting for ACME challenge info (60s)")
	}

	ch := provider.getChallenge()
	if ch == nil {
		return nil, fmt.Errorf("no challenge info captured from ACME")
	}

	// Persist the active order so ProceedWithOrder can find it
	c.mu.Lock()
	c.orders[domain] = &activeOrder{provider: provider, challenge: ch}
	c.mu.Unlock()

	logger.Info.Printf("ACME: Challenge ready for %s — DNS record: %s", domain, ch.Domain)
	return ch, nil
}

// ProceedWithOrder signals the paused order that DNS records are in place,
// waits for lego to complete validation + issuance, and returns the certificate.
func (c *Client) ProceedWithOrder(domain string) (fullchain, privkey []byte, expiry time.Time, err error) {
	c.mu.Lock()
	order, ok := c.orders[domain]
	if ok {
		delete(c.orders, domain)
	}
	c.mu.Unlock()

	if !ok {
		return nil, nil, time.Time{}, fmt.Errorf("no active order for %s — call /dns-challenge first", domain)
	}

	logger.Info.Printf("ACME: Proceeding with order for %s (unblocking DNS provider)", domain)

	// Unblock all Present() goroutines (wildcard cert has 2 challenges)
	close(order.provider.proceed)

	// Wait for Obtain() to complete
	select {
	case result := <-order.provider.resultCh:
		if result.err != nil {
			return nil, nil, time.Time{}, fmt.Errorf("certificate obtain failed: %w", result.err)
		}

		expiry = parseCertExpiry(result.resource.Certificate)
		logger.Info.Printf("ACME: Certificate issued for %s, expires: %s", domain, expiry.Format(time.RFC3339))
		return result.resource.Certificate, result.resource.PrivateKey, expiry, nil

	case <-time.After(10 * time.Minute):
		return nil, nil, time.Time{}, fmt.Errorf("timeout waiting for certificate issuance (10m)")
	}
}

// parseCertExpiry extracts the NotAfter field from the first PEM certificate block.
func parseCertExpiry(pemData []byte) time.Time {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return time.Time{}
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}
	}
	return cert.NotAfter
}
