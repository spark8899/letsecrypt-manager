# Let's Encrypt Manager

A comprehensive, automated wildcard SSL certificate management system built with Go (Backend) and Vue 3 (Frontend). It simplifies the ACME DNS-01 challenge workflow, providing a modern web interface to manage domains, verify DNS records, and issue certificates.

## 🚀 Features

### Backend (Go)
- **JWT Authentication**: Secure multi-account support with SHA256 hashed passwords.
- **Wildcard Support**: Issue certificates for both `*.example.com` and `example.com` simultaneously.
- **ACME DNS-01 Workflow**: Supports CNAME delegation and TXT record verification.
- **Persistence**: File-based storage for domains, ACME accounts, and certificates.
- **Logging**: Detailed request logging and file-based rotation.

### Frontend (Vue 3)
- **Modern Dashboard**: Real-time status overview of all managed domains.
- **Interactive Workflow**: Step-by-step guidance for DNS challenge, verification, and issuance.
- **Certificate Viewer**: Built-in viewer for fullchain certificates and private keys with one-click copy.
- **Responsive UI**: Clean and professional interface built with PrimeVue and Lucide icons.

## 🏗️ Project Structure

```text
letsencrypt-manager/
├── backend/                # Go Backend Service
│   ├── acme/               # ACME/lego client logic
│   ├── config/             # Configuration loader
│   ├── handlers/           # HTTP API handlers
│   ├── logger/             # Logging system
│   ├── middleware/         # JWT and logging middleware
│   ├── models/             # File-based data store
│   ├── config.json         # Backend configuration
│   └── main.go             # Application entry point
├── frontend/               # Vue 3 Frontend App
│   ├── src/                # Frontend source code
│   │   ├── store/          # Pinia state management
│   │   ├── views/          # Dashboard and Login pages
│   │   └── router/         # Vue Router configuration
│   └── vite.config.ts      # Vite configuration with API proxy
└── Makefile                # Unified management commands
```

## 🛠️ Getting Started

### Prerequisites
- **Go**: v1.21 or higher
- **Node.js**: v18 or higher
- **npm**: v9 or higher

### Installation
Install dependencies for both backend and frontend:
```bash
make install
```

### Running the Project
To start both the backend and frontend in development mode:
```bash
make run
```
- **Backend API**: http://localhost:8080
- **Frontend UI**: http://localhost:5173 (Proxies `/api` to backend)

### Building for Production
To build the backend binary and frontend static assets:
```bash
make build
```
This generates the `letsencrypt-manager` binary in the root directory and the frontend build in `frontend/dist/`.

## ⚙️ Configuration

Edit `backend/config.json` before running the production environment:

```json
{
  "listen_addr": ":8080",
  "data_dir": "./data",
  "acme_email": "admin@example.com",
  "acme_server": "staging",
  "jwt_secret": "your-secure-random-secret",
  "token_expiry_hours": 24,
  "accounts": [
    {
      "username": "admin",
      "password": "sha256-hashed-password"
    }
  ]
}
```

**Generate Password Hash:**
```bash
make hash-password PASSWORD=yourpassword
```

**ACME Server Modes:**
- `"staging"`: Uses Let's Encrypt staging environment (No rate limits, untrusted certs). **Recommended for testing.**
- `"production"`: Uses Let's Encrypt production environment (Trusted certs).

## 🧪 Testing

Run backend unit and integration tests:
```bash
make test-backend
```

## ⚠️ Important Notes
- **Rate Limits**: Let's Encrypt production environment has strict rate limits. Always test with `staging` first.
- **Security**: The `jwt_secret` should be changed to a long, random string in production.
- **Permissions**: Private keys are stored with `0600` permissions. Ensure the user running the service has appropriate data directory access.

---

Built with ❤️ using Go, Vue 3, and PrimeVue.
