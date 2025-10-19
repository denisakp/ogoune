# Pulseguard – Open-source Monitoring

An open-source monitoring tool for websites, APIs, TCP services, SSL certificates, and cron jobs — with alerting and status pages.

![Build Status](https://img.shields.io/badge/build-status-grey)
![License](https://img.shields.io/badge/license-MIT-blue)
![Go](https://img.shields.io/badge/go-1.25%2B-00ADD8)

---

## ✨ Features

- HTTP/HTTPS and TCP monitoring
- SSL/TLS certificate expiry alerts
- Incident management with 3-failure threshold
- Notifications: SMTP, Slack, Discord, Google Chat
- Public status page (backend data + frontend UI)
- Real-time updates (WebSockets)

---

## 🚀 Project Structure

This repository is a monorepo with two main components:

- 📁 `backend/`: Core API, background worker, and monitoring engine in Go.
- 📁 `frontend/`: React (Vite) dashboard using Shadcn/ui.

Please see the README in each directory for setup and contributing guidelines.

---

## ⚡ Quick Start

Using the provided Makefile (requires Docker for Postgres/Redis):

```bash
# From the repo root
make install      # install go deps
make docker-up    # start postgres and redis (local containers)
make run          # run API + worker
```

API: http://localhost:8080

Create a monitor:

```bash
curl -X POST http://localhost:8080/resources \
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
