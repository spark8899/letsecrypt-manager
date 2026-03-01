# Let's Encrypt Manager - Frontend

A modern, responsive web interface for managing wildcard SSL certificates via the ACME protocol.

## Features

- **Dashboard**: Real-time overview of all managed domains and certificate status.
- **Guided Issuance**: Step-by-step workflow for generating DNS challenges, verifying propagation, and issuing certificates.
- **Certificate Viewer**: Securely view and copy fullchain certificates and private keys.
- **Responsive Design**: Built with Vue 3, PrimeVue, and Lucide icons for a clean, professional look.

## Tech Stack

- **Framework**: Vue 3 (Composition API + TypeScript)
- **UI Library**: PrimeVue 4 (Aura theme)
- **State Management**: Pinia
- **Icons**: Lucide Vue
- **Build Tool**: Vite

## Getting Started

### Prerequisites

- Node.js (v18+)
- npm

### Installation

```bash
cd frontend
npm install
```

### Development

Run the development server with hot-reload:

```bash
npm run dev
```

The frontend will be available at `http://localhost:5173`. It is configured to proxy API requests to `http://localhost:8080` (default backend port).

### Production Build

Build for production:

```bash
npm run build
```

The output will be in the `dist/` directory.

## Integration

The frontend connects to the following backend endpoints:

- `POST /api/auth/login`: Authentication
- `GET /api/domains`: List managed domains
- `POST /api/domains`: Add new domain
- `POST /api/domains/:domain/dns-challenge`: Start certificate order
- `GET /api/domains/:domain/dns-verify`: Check DNS propagation
- `POST /api/domains/:domain/issue`: Finalize order and save certificate
- `GET /api/domains/:domain/cert`: Retrieve certificate and private key
