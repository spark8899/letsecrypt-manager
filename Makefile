# Main Makefile for LetsEncrypt Manager

.PHONY: all install run run-backend run-frontend build build-backend build-frontend clean test test-backend help

# Default target
all: help

# Install all dependencies
install:
	@echo "Installing backend dependencies..."
	cd backend && go mod tidy && go mod download
	@echo "Installing frontend dependencies..."
	cd frontend && npm install

# Run backend in development
run-backend:
	cd backend && go run main.go config.json

# Run frontend in development
run-frontend:
	cd frontend && npm run dev

# Run both frontend and backend concurrently
run:
	@echo "Starting backend and frontend..."
	(cd backend && go run main.go config.json) & (cd frontend && npm run dev)

# Build both frontend and backend
build: build-frontend build-backend

build-backend:
	@echo "Building backend binary..."
	cd backend && go build -o ../letsencrypt-manager .

build-frontend:
	@echo "Building frontend assets..."
	cd frontend && npm run build

# Run backend tests
test-backend:
	cd backend && go test ./... -v

# Generate sha256 password hash
# Usage: make hash-password PASSWORD=yourpassword
hash-password:
	@echo -n "$(PASSWORD)" | sha256sum | awk '{print $$1}'

# Clean build artifacts
clean:
	rm -f letsencrypt-manager
	rm -rf backend/data/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/
	rm -rf backend/vendor/

# Help menu
help:
	@echo ""
	@echo "=== LetsEncrypt Manager Management ==="
	@echo ""
	@echo "  make install         Install both backend and frontend dependencies"
	@echo "  make run             Start backend and frontend in parallel"
	@echo "  make run-backend     Start backend server only"
	@echo "  make run-frontend    Start frontend dev server only"
	@echo "  make build           Build both backend binary and frontend assets"
	@echo "  make test-backend    Run all backend tests"
	@echo "  make hash-password   Generate SHA256 hash (Usage: make hash-password PASSWORD=xyz)"
	@echo "  make clean           Remove build artifacts and dependencies"
	@echo ""
