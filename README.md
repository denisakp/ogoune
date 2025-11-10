# Pulseguard – Open-source Monitoring

An open-source monitoring tool for websites, APIs, TCP services, and cron jobs — with alerting via SMTP and webhooks, plus status pages.

![Build Status](https://img.shields.io/badge/build-status-grey)
![License](https://img.shields.io/badge/license-MIT-blue)
![Go](https://img.shields.io/badge/go-1.24%2B-00ADD8)

---

## ✨ Features

### Core Capabilities
- **HTTP/HTTPS Monitoring** – Health checks for websites and APIs
- **TCP Monitoring** – Port availability checks for services
- **Incident Management** – Automatic detection and resolution tracking
- **Notifications** – Email alerts and webhook notifications
- **Event Tracking** – Complete incident history with timeline
- **Public Status Page** – Share uptime and incident status with stakeholders
- **Activity Logging** – Detailed records of all monitoring activity
- **Resource Organization** – Tags to group and categorize monitors
- **Global Statistics** – System-wide uptime and incident metrics

### Planned Features
- [ ] Real-time status updates via Server-Sent Events (SSE)
- [ ] Advanced analytics and performance insights
- [ ] User authentication and multi-user support
- [ ] SSL certificate monitoring and renewal alerts
- [ ] Mobile app for iOS and Android
- [ ] Custom alert rules and workflows

---

## 🏗️ Architecture Overview

Pulseguard consists of two main components:

- **Backend** (Go) – API server, background job processor, and monitoring engine
- **Frontend** (Vue 3) – Interactive dashboard for monitor and incident management

The system automatically monitors your resources, detects failures, and sends notifications through your configured channels.

---

## 🚀 Getting Started

### Prerequisites
- Docker (for local databases) or PostgreSQL + Redis installed locally
- Go 1.24+ (for backend)
- Node.js 22+ and pnpm (for frontend)

### Quick Start

**1. Start the backend:**

```bash
cd backend
docker compose up -d          # Start PostgreSQL and Redis
cp .env.example .env          # Configure with your SMTP/webhook details
go run ./cmd/api
```

Backend runs at: `http://localhost:8080/api`

**2. Start the frontend:**

```bash
cd frontend
pnpm install
export VITE_API_BASE_URL=http://localhost:8080/api
pnpm dev
```

Dashboard available at: `http://localhost:5173`

**3. Create your first monitor:**

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

Your first monitor is now active. If it fails 3 times in a row, an incident will be created and notifications sent.

---

## 📚 Documentation

### For Users & Operators
- [Backend Setup Guide](./backend/README.md) – Installation, configuration, and operation
- [Frontend Setup Guide](./frontend/README.md) – Dashboard setup and usage

### For Developers & Architects
- [Backend Architecture](./backend/docs/ARCHITECTURE.md) – Technical design and implementation details
- [Backend Overview](./backend/docs/BACKEND_OVERVIEW.md) – System overview and data flow
- [Frontend Architecture](./frontend/ARCHITECTURE.md) – Frontend design patterns and conventions
- [Webhook Notification Flow](./WEBHOOK_NOTIFICATION_FLOW.md) – How webhooks are triggered and sent

### Reference
- [Refactoring Summary](./REFACTORING_SUMMARY.md) – Changes from previous versions

---

## 🎯 How It Works

### Monitoring Cycle
1. Each active resource is checked at its configured interval
2. Health checks run for HTTP/HTTPS or TCP connectivity
3. Results are recorded in the activity log
4. After 3 consecutive failures, an incident is created
5. Notifications are sent (SMTP email and/or webhook)
6. When the resource recovers, the incident is resolved

### Notifications
Pulseguard sends notifications through two channels:

- **Email (SMTP)** – System-wide notifications to a configured recipient
- **Webhooks** – HTTP POST requests to your endpoint for custom integration

Both are optional and configured via environment variables. See [Backend Setup Guide](./backend/README.md) for configuration details.

---

## 🤝 Contributing

Contributions are welcome! To get started:

1. Read [CONTRIBUTING.md](./CONTRIBUTING.md)
2. Follow the existing code and documentation conventions
3. Submit well-documented, focused pull requests

Areas where we'd love help:
- Bug fixes and performance improvements
- Documentation improvements
- Testing enhancements
- New monitoring strategies (DNS, SSL, gRPC, etc.)
- Frontend UI/UX improvements

---

## 📄 License

MIT – see [LICENSE](./LICENSE) for details.

---

## 🙏 Built With

- [Go](https://golang.org/) – Backend
- [Vue 3](https://vuejs.org/) – Frontend
- [Ant Design Vue](https://www.antdv.com/) – UI components
- [Asynq](https://github.com/hibiken/asynq) – Background job processing
- [GORM](https://gorm.io/) – Database ORM

---

**Status:** Active Development  
**Latest Release:** See [Releases](https://github.com/your-org/pulseguard/releases)  
**Issues & Feedback:** [GitHub Issues](https://github.com/your-org/pulseguard/issues)