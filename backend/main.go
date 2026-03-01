package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"letsencrypt-manager/config"
	"letsencrypt-manager/handlers"
	"letsencrypt-manager/logger"
	"letsencrypt-manager/middleware"
	"letsencrypt-manager/models"
)

func main() {
	configPath := "config.json"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(cfg.DataDir + "/logs"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init logger: %v\n", err)
		os.Exit(1)
	}

	logger.Info.Printf("Starting letsencrypt-manager, config: %s", configPath)
	logger.Info.Printf("Data directory: %s", cfg.DataDir)

	// Validate ACME email at startup — fail fast before any ACME call is made
	if err := config.ValidateACMEEmail(cfg.ACMEEmail); err != nil {
		logger.Error.Fatalf("Configuration error: %v", err)
	}

	store, err := models.NewStore(cfg.DataDir)
	if err != nil {
		logger.Error.Fatalf("Failed to init store: %v", err)
	}

	if cfg.IsStaging() {
		gin.SetMode(gin.DebugMode)
		logger.Warn.Println("Running in STAGING mode (Let's Encrypt staging server)")
	} else {
		gin.SetMode(gin.ReleaseMode)
		logger.Info.Println("Running in PRODUCTION mode (Let's Encrypt production server)")
	}

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())

	h := handlers.New(cfg, store)

	// Public routes
	r.POST("/api/auth/login", h.Login)
	r.GET("/health", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "ok"})
	})

	// Protected routes
	api := r.Group("/api", middleware.AuthMiddleware(cfg.JWTSecret))
	{
		api.POST("/domains", h.AddDomain)
		api.GET("/domains", h.ListDomains)
		api.POST("/domains/:domain/dns-challenge", h.GetDNSChallenge)
		api.GET("/domains/:domain/dns-verify", h.VerifyDNS)
		api.POST("/domains/:domain/issue", h.IssueCertificate)
		api.GET("/domains/:domain/cert", h.GetCertificate)
	}

	logger.Info.Printf("Server listening on %s", cfg.ListenAddr)
	if err := r.Run(cfg.ListenAddr); err != nil {
		logger.Error.Fatalf("Server failed: %v", err)
	}
}
