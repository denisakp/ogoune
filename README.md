# Pulseguard – Open-source Monitoring

An open-source monitoring tool for websites, APIs, TCP services, SSL certificates, and cron jobs — with alerting and status pages.

![Build Status](https://img.shields.io/badge/build-status-grey)
![License](https://img.shields.io/badge/license-MIT-blue)
![Go](https://img.shields.io/badge/go-1.25%2B-00ADD8)

---

## ✨ Features

### Core Features
- HTTP/HTTPS and TCP monitoring
- SSL/TLS certificate expiry alerts
- Incident management with 3-failure threshold
- Notifications: SMTP, Slack, Discord, Google Chat
- Public status page (backend data + frontend UI)
- Real-time updates (WebSockets)

### Upcoming Features
- [ ] **SSE Real-time Updates** – Server-Sent Events for faster status propagation
- [ ] **Advanced Analytics Dashboard** – Deep insights into uptime trends and performance metrics
- [ ] **User Authentication** – Multi-user support with role-based access control
- [ ] **SSL Certificate Monitoring** – Enhanced certificate tracking and renewal alerts
- [ ] **Integration Marketplace** – Extensible notification and data pipeline integrations
- [ ] **Mobile App** – Native mobile monitoring and alerting
- [ ] **Custom Alert Workflows** – Flexible, rule-based alert orchestration

---

## 🚀 Project Structure

This repository is a monorepo with two main components:

- 📁 `backend/`: Core API, background worker, and monitoring engine in Go.
- 📁 `frontend/`: Vue 3 (Vite) dashboard using Ant Design Vue.

Please see the README in each directory for setup and contributing guidelines.

---

## ⚡ Quick Start

### Prerequisites
- Docker (for Postgres and Redis)
- Go 1.25+
- Node.js and pnpm (for frontend)

### Backend Setup

```bash
# Navigate to the backend directory
cd backend

# Install dependencies
go mod download

# Set up environment variables (see .env.example)
cp .env.example .env

# Start Postgres and Redis
docker compose up -d

# Run the API and worker
go run ./cmd/api
```

API base: http://localhost:8080/api

### Frontend Setup

```bash
# Navigate to the frontend directory
cd frontend

# Install dependencies
pnpm install

# Set environment variables
export VITE_API_BASE_URL=http://localhost:8080

# Run development server
pnpm dev
```

Dashboard: http://localhost:5173

### Create a Monitor

```bash
curl -X POST http://localhost:8080/api/resources \
  -H "Content-Type: application/json" \
  -d '{
    "name": "My Website",
    "type": "http",
    "target": "https://example.com",
    "interval": 60,
    "timeout": 10
  }'
```

---

## 🧭 Documentation

- Backend architecture: `backend/docs/ARCHITECTURE.md`
- Backend setup: `backend/README.md`
- Frontend setup: `frontend/README.md`

---

## 🤝 Contributing

Contributions are welcome! Please read `CONTRIBUTING.md` and follow the style and layering guidelines. Small, focused PRs are appreciated.

---

## 📄 License

MIT — see `LICENSE`.