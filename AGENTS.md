# AGENTS.md - Developer Guidelines for letsencrypt-manager

This document provides guidelines for agents working on this codebase.

---

## Project Overview

A full-stack SSL certificate management system with:
- **Backend**: Go + Gin web framework
- **Frontend**: Vue 3 + TypeScript + PrimeVue
- **ACME**: lego/v4 for Let's Encrypt integration

---

## Build & Run Commands

### From Root Directory

```bash
# Install all dependencies
make install

# Run both backend and frontend
make run

# Run backend only
make run-backend

# Run frontend only
make run-frontend

# Build both
make build

# Run backend tests
make test-backend

# Generate password hash (SHA256)
make hash-password PASSWORD=yourpassword

# Clean build artifacts
make clean
```

### Backend Only

```bash
cd backend
go run main.go config.json           # Run
go build -o letsencrypt-manager .   # Build binary
go test ./... -v                    # Run all tests
go test -v -run TestName ./...      # Run single test
```

### Frontend Only

```bash
cd frontend
npm run dev          # Development server (localhost:5173)
npm run build        # Production build
npm run preview      # Preview production build
npx vue-tsc --noEmit  # Type check only
```

---

## Backend (Go) Code Style

### Project Structure

```
backend/
├── main.go           # Entry point
├── handlers/         # HTTP handlers
├── middleware/       # Auth, logging
├── models/           # Data store
├── acme/             # ACME client
├── config/           # Configuration
└── logger/          # Logging
```

### Naming Conventions

- **Files**: `snake_case.go` (e.g., `auth_handler.go`)
- **Types/Structs**: `PascalCase` (e.g., `Handler`, `Config`)
- **Functions/Variables**: `camelCase`
- **Constants**: `PascalCase` or `UPPER_SNAKE_CASE`
- **Interfaces**: `PascalCase` with `er` suffix (e.g., `Provider`)

### Imports

Group imports in this order:
1. Standard library
2. Third-party packages
3. Internal packages

```go
import (
    "fmt"
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"

    "letsencrypt-manager/config"
    "letsencrypt-manager/models"
)
```

### Error Handling

- Return errors with context: `return nil, fmt.Errorf("failed to init store: %w", err)`
- Use named return values for error variables: `func foo() (string, error) {`
- Never suppress errors with `_`
- Log errors before returning: `logger.Error.Printf("failed: %v", err)`

### HTTP Handlers

- Use struct tags for request binding
- Return proper HTTP status codes
- Use `gin.H{}` for JSON responses

```go
type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

func (h *Handler) Login(ctx *gin.Context) {
    var req LoginRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // ...
}
```

### Testing

- Test files: `filename_test.go`
- Test functions: `func TestName(t *testing.T)`
- Use table-driven tests when appropriate

```go
func TestHandler_Login(t *testing.T) {
    tests := []struct {
        name     string
        username string
        wantCode int
    }{
        {"valid", "admin", http.StatusOK},
        {"invalid", "wrong", http.StatusUnauthorized},
    }
    // ...
}
```

---

## Frontend (Vue 3 + TypeScript) Code Style

### Project Structure

```
frontend/src/
├── main.ts           # Entry point
├── App.vue           # Root component
├── router/           # Vue Router
├── store/            # Pinia stores
└── views/            # Page components
```

### TypeScript Conventions

- **Interfaces**: `PascalCase` with descriptive names
- **Props/Events**: Use `defineProps` and `defineEmits`
- **Store**: Use composition API style with Pinia

```typescript
interface DomainInfo {
  domain: string
  status: string
  cert_expiry: string | null
  challenge: Challenge | null
}

export const useDomainStore = defineStore('domain', {
  state: () => ({
    domains: [] as DomainInfo[],
  }),
  // ...
})
```

### Vue Components

- Use `<script setup lang="ts">` syntax
- Use TypeScript for all props and emits
- Keep templates clean and readable

```vue
<script setup lang="ts">
import { ref, computed } from 'vue'
import type { DomainInfo } from '@/store/domain'

const props = defineProps<{
  domain: DomainInfo
}>()

const emit = defineEmits<{
  (e: 'select', domain: string): void
}>()
</script>

<template>
  <div class="domain-card" @click="emit('select', props.domain.domain)">
    {{ props.domain.domain }}
  </div>
</template>
```

### Imports

```typescript
// Vue/Framework imports first
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'

// Third-party
import axios from 'axios'
import { useToast } from 'primevue/usetoast'

// Internal
import { useAuthStore } from '@/store/auth'
import { useDomainStore, type DomainInfo } from '@/store/domain'
```

### File Naming

- **Components**: `PascalCase.vue` (e.g., `LoginForm.vue`)
- **Stores**: `camelCase.ts` (e.g., `authStore.ts`)
- **Types**: `camelCase.ts` or `types.ts`
- **Utilities**: `camelCase.ts` (e.g., `dateUtils.ts`)

### CSS/Styling

- Use Tailwind CSS for utility classes
- Use scoped styles in Vue components
- Custom CSS goes in `<style scoped>` blocks

---

## API Endpoints

### Authentication

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| POST | `/api/auth/login` | ❌ | Login, returns JWT |
| GET | `/health` | ❌ | Health check |

### Domains (require JWT)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/domains` | Add domain |
| GET | `/api/domains` | List domains |
| POST | `/api/domains/:domain/dns-challenge` | Start ACME order |
| GET | `/api/domains/:domain/dns-verify` | Verify DNS |
| POST | `/api/domains/:domain/issue` | Issue certificate |
| GET | `/api/domains/:domain/cert` | Get certificate |

---

## Configuration

Backend config file: `backend/config.json`

```json
{
  "listen_addr": ":8080",
  "data_dir": "./data",
  "acme_email": "admin@example.com",
  "acme_server": "staging",
  "jwt_secret": "your-secret",
  "token_expiry_hours": 24,
  "accounts": [
    { "username": "admin", "password": "sha256-hash" }
  ]
}
```

---

## General Guidelines

1. **Never commit secrets** - Use environment variables or config files
2. **Run tests before committing** - `make test-backend`
3. **Type check before building** - `npx vue-tsc --noEmit`
4. **Follow existing patterns** - Match the code style in the codebase
5. **Add comments for complex logic** - Explain WHY, not WHAT
6. **Handle errors gracefully** - Never leave users with cryptic errors

---

## Dependencies

### Backend
- Go 1.21+
- gin v1.9.1
- lego/v4 v4.14.2 (ACME)
- golang-jwt/jwt/v5

### Frontend
- Node.js 18+
- Vue 3.5
- TypeScript 5.7
- PrimeVue 4
- Pinia 3
- Vite 6
- Tailwind CSS 4
