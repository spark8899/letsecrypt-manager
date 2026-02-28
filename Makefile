.PHONY: build run deps clean hash-password test test-verbose test-cover

# Build the binary
build:
	go build -o letsencrypt-manager .

# Download dependencies
deps:
	go mod tidy
	go mod download

# Run in development mode
run:
	go run . config.json

# Clean build artifacts
clean:
	rm -f letsencrypt-manager
	rm -rf data/

# Run all tests
test:
	go test ./... -count=1 -timeout 30s

# Run tests with verbose output
test-verbose:
	go test ./... -v -count=1 -timeout 30s

# Run tests with coverage report
test-cover:
	go test ./... -count=1 -timeout 30s -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Run tests for a specific package
# Usage: make test-pkg PKG=./handlers
test-pkg:
	go test $(PKG) -v -count=1 -timeout 30s

# Run tests with race detector
test-race:
	go test ./... -race -count=1 -timeout 60s

# Generate sha256 password hash
# Usage: make hash-password PASSWORD=yourpassword
hash-password:
	@echo -n "$(PASSWORD)" | sha256sum | awk '{print $$1}'

# Show API usage
help:
	@echo ""
	@echo "=== LetsEncrypt Manager API ==="
	@echo ""
	@echo "1. Login:"
	@echo '   curl -X POST http://localhost:8080/api/auth/login \'
	@echo '     -H "Content-Type: application/json" \'
	@echo '     -d '"'"'{"username":"admin","password":"admin123"}'"'"''
	@echo ""
	@echo "2. Add domain:"
	@echo '   curl -X POST http://localhost:8080/api/domains \'
	@echo '     -H "Authorization: Bearer TOKEN" \'
	@echo '     -H "Content-Type: application/json" \'
	@echo '     -d '"'"'{"domain":"example.com"}'"'"''
	@echo ""
	@echo "3. Get DNS challenge:"
	@echo '   curl -X POST http://localhost:8080/api/domains/example.com/dns-challenge \'
	@echo '     -H "Authorization: Bearer TOKEN"'
	@echo ""
	@echo "4. Verify DNS:"
	@echo '   curl http://localhost:8080/api/domains/example.com/dns-verify \'
	@echo '     -H "Authorization: Bearer TOKEN"'
	@echo ""
	@echo "5. Issue certificate:"
	@echo '   curl -X POST http://localhost:8080/api/domains/example.com/issue \'
	@echo '     -H "Authorization: Bearer TOKEN"'
	@echo ""
	@echo "6. Get certificate:"
	@echo '   curl http://localhost:8080/api/domains/example.com/cert \'
	@echo '     -H "Authorization: Bearer TOKEN"'
