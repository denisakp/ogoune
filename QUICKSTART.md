# Pulseguard - Complete Implementation Guide

Welcome to **Pulseguard**, an open-source monitoring platform for websites, APIs, TCP services, SSL certificates, and cron jobs.

## 📚 Documentation Structure

### Getting Started

- **[QUICKSTART.md](./QUICKSTART.md)** ⭐ **START HERE**

  - 5-minute setup guide
  - Fastest way to get the application running
  - Common issues and solutions

- **[INTEGRATION.md](./INTEGRATION.md)** - Backend + Frontend Integration
  - How to run both backend API and frontend together
  - Workflow examples
  - Testing integration
  - Debugging tips

### Backend (Go API + Worker)

- **[backend/README.md](./backend/README.md)** - Backend Setup

  - Technology stack
  - Configuration guide
  - Running instructions
  - Developer notes

- **[backend/docs/ARCHITECTURE.md](./backend/ARCHITECTURE.md)** - Backend Architecture

  - Core architecture (API + Worker)
  - Scheduling engine
  - Real-time system (WebSockets)
  - Database models
  - Incident management
  - Notification system

- **[.github/copilot-instructions.md](./.github/copilot-instructions.md)** - Backend AI Instructions
  - For AI coding agents
  - Project patterns and conventions
  - Architecture boundaries
  - When adding features

### Frontend (Vue 3 + TypeScript)

- **[frontend/README.md](./frontend/README.md)** - Frontend Overview

  - Project structure
  - Technology stack
  - Getting started
  - Features overview
  - Styling with Daisy UI

- **[frontend/DEVELOPMENT.md](./frontend/DEVELOPMENT.md)** - Frontend Developer Guide

  - Architecture principles
  - Best practices
  - Adding new features (step-by-step)
  - Component patterns
  - Testing strategies
  - Common mistakes to avoid

- **[frontend/QUICKSTART.md](./frontend/QUICKSTART.md)** - Frontend Quick Start
  - 5-minute setup
  - Project summary
  - Feature overview
  - Development workflow
  - Troubleshooting

### Implementation Summaries

- **[FRONTEND_IMPLEMENTATION.md](./FRONTEND_IMPLEMENTATION.md)** - What Was Built
  - Complete deliverables checklist
  - Architecture overview
  - Features implemented
  - Technical details
  - Future enhancements

## 🚀 Quick Start (5 Minutes)

### Option A: Community Mode (Recommended for Local Testing)

No Redis required – everything runs in a single process with embedded SQLite:

**Terminal 1: Backend (with embedded scheduler)**

```bash
# Community mode: SQLite + in-process scheduler
DB_DRIVER=sqlite SQLITE_PATH=./pulseguard.db SCHEDULER_MODE=timingwheel make run
```

Or use Docker Compose Community edition:

```bash
docker compose -f docker-compose.community.yml up
```

**Terminal 2: Frontend SPA**

```bash
cd frontend
pnpm install
pnpm run dev      # http://localhost:5173
```

✅ **Done!** Open http://localhost:5173

---

### Option B: Hosted Mode (Full PostgreSQL + Redis Setup)

For production-like multi-container deployments:

**Terminal 1: Hosted/PostgreSQL Backend Services**

```bash
make docker-up    # Start PostgreSQL + Redis
```

**Terminal 2: Backend API + Worker**

```bash
make run          # Start Go API and background worker
```

**Terminal 3: Frontend SPA**

```bash
cd frontend
pnpm install
pnpm run dev      # http://localhost:5173
```

---

## Important: Scheduler Modes

- **Community/MVPMode (_timingwheel_)**: In-process scheduler, no Redis required
  - ✅ Perfect for single-server deployments
  - ✅ No external dependencies
  - ✅ All checks run deterministically in-process

- **Hosted Mode (_asynq_)**: Redis-based distributed scheduler
  - ✅ Scales across multiple instances
  - ✅ Requires Redis 6+
  - ✅ Recommended for production SaaS deployments

**Never use `SCHEDULER_MODE=asynq` without Redis**, or the process will fail to start.

---

### Terminal 1: Community Backend Services

```bash
docker compose -f docker-compose.community.yml up -d
```

This starts PulseGuard with embedded SQLite and Redis only.

### Terminal 2: Hosted/PostgreSQL Backend Services

```bash
make docker-up    # Start PostgreSQL + Redis
```

### Terminal 2: Backend API + Worker

```bash
make run          # Start Go API and background worker
```

### Terminal 3: Frontend SPA

```bash
cd frontend
pnpm install
pnpm run dev      # http://localhost:5173
```

**That's it!** 🎉 Open http://localhost:5173 in your browser.

## 📊 What You Can Do

### Monitor Your Services

✅ **HTTP/HTTPS Monitoring**

- Monitor websites and APIs
- Customizable check intervals
- Response time tracking

✅ **TCP Monitoring**

- Monitor server ports
- Connection testing
- Network availability

✅ **Status Dashboard**

- Real-time status display
- Uptime percentage
- Last check information

### Organize & Configure

✅ **Tags**

- Organize monitors with labels
- Filter by tags
- Better resource management

✅ **Integrations**

- Send notifications to Slack
- Receive alerts via Webhooks
- Email notifications via SMTP
- Webhook support

### View Activity & History

✅ **Monitoring Activities**

- View all check results
- Response time metrics
- Success/failure status
- Real-time updates (coming soon)

✅ **Incident Management**

- Automatic incident creation (3 consecutive failures)
- Incident resolution tracking
- Complete incident history

### Monitor Lifecycle Operations

✅ **Pause Monitoring**

Temporarily stop checks for a monitor without deleting it:

```bash
curl -X POST http://localhost:8080/api/resources/{resourceId}/pause
```

**Effect**: 
- Stops all scheduled checks immediately
- Monitor remains in the database
- Can be resumed later at any time
- No incidents or notifications while paused

**Use case**: Maintenance windows, temporary service downtime, testing

✅ **Resume Monitoring**

Resume checks for a paused monitor:

```bash
curl -X POST http://localhost:8080/api/resources/{resourceId}/resume
```

**Effect**: 
- Reactivates monitoring checks
- Uses the configured interval
- Checks resume immediately on next scheduler tick
- Incidents can be created again if threshold met

**Use case**: Resume after maintenance, restart monitoring for troubleshooting

✅ **Update Check Interval**

Change how frequently a monitor runs:

```bash
curl -X PATCH http://localhost:8080/api/resources/{resourceId} \
  -H "Content-Type: application/json" \
  -d '{"interval": 600}'  # 600 seconds = 10 minutes
```

**Interval** is specified in seconds. Changes take effect on the next scheduler tick.

## 🏗️ Architecture Overview

```
Frontend (Vue 3 + TypeScript)
   ↓ (JSON API calls)
Backend API (Go + Chi Router)
   ↓ (enqueues jobs)
Background Worker (Asynq)
   ↓ (executes checks)
SQLite or PostgreSQL Database
   ↓ (stores results)
Frontend (real-time updates via WebSocket - coming soon)
```

### Technology Stack

**Backend**:

- Language: Go 1.25+
- HTTP Router: Chi
- Database: driver-aware GORM runtime with PostgreSQL or embedded SQLite
- Queue: Redis + Asynq
- Real-time: WebSockets (nhooyr.io/websocket)

**Frontend**:

- Framework: Vue 3
- Language: TypeScript 5.9
- Build Tool: Vite 7
- Styling: Tailwind CSS + Daisy UI
- HTTP Client: Axios
- Routing: Vue Router 4

## 📖 Documentation by Role

### I'm a User

1. Start with: **[QUICKSTART.md](./QUICKSTART.md)**
2. Learn: **[frontend/README.md](./frontend/README.md)**
3. Refer: Feature list and troubleshooting in **[frontend/QUICKSTART.md](./frontend/QUICKSTART.md)**

### I'm a Frontend Developer

1. Start with: **[frontend/README.md](./frontend/README.md)**
2. Deep dive: **[frontend/DEVELOPMENT.md](./frontend/DEVELOPMENT.md)**
3. Reference: **[FRONTEND_IMPLEMENTATION.md](./FRONTEND_IMPLEMENTATION.md)**

### I'm a Backend Developer

1. Start with: **[backend/README.md](./backend/README.md)**
2. Architecture: **[backend/docs/ARCHITECTURE.md](./backend/docs/ARCHITECTURE.md)**
3. AI Instructions: **[.github/copilot-instructions.md](./.github/copilot-instructions.md)**

### I'm an AI Coding Agent

1. Start with: **[.github/copilot-instructions.md](./.github/copilot-instructions.md)** (Backend)
2. Frontend standards: **[frontend/DEVELOPMENT.md](./frontend/DEVELOPMENT.md)**
3. Integration points: **[INTEGRATION.md](./INTEGRATION.md)**

### I'm Contributing

1. Read: **[CONTRIBUTING.md](./CONTRIBUTING.md)**
2. Backend guide: **[.github/copilot-instructions.md](./.github/copilot-instructions.md)**
3. Frontend guide: **[frontend/DEVELOPMENT.md](./frontend/DEVELOPMENT.md)**
4. Test everything locally using **[INTEGRATION.md](./INTEGRATION.md)**

## ✨ Key Features

- ✅ Multiple monitor types (HTTP, TCP)
- ✅ Real-time status updates
- ✅ Flexible notification integrations
- ✅ Incident management
- ✅ Activity logging
- ✅ Tag-based organization
- ✅ Responsive web interface
- ✅ Type-safe API integration

## 🔮 Future Roadmap

- [ ] Write the test cases

## 🐛 Troubleshooting

### Can't connect to backend?

- Check: `curl http://localhost:8080/health`
- See: **[INTEGRATION.md](./INTEGRATION.md)** "Troubleshooting" section

### Frontend not loading?

- Check: Terminal 3 output
- See: **[frontend/DEVELOPMENT.md](./frontend/DEVELOPMENT.md)** "Debugging" section

### Database connection failed?

- Check: Docker containers running
- See: **[INTEGRATION.md](./INTEGRATION.md)** "Troubleshooting" section

### Need help with development?

- See: **[frontend/DEVELOPMENT.md](./frontend/DEVELOPMENT.md)** for patterns
- See: **[.github/copilot-instructions.md](./.github/copilot-instructions.md)** for backend patterns

## 🎯 Next Steps

1. **Get it running**: Follow [QUICKSTART.md](./QUICKSTART.md)
2. **Create a monitor**: Add your first website/API to monitor
3. **Configure notifications**: Set up Slack or email alerts
4. **Explore the code**: Understand the architecture
5. **Contribute**: Add features and improvements!

## 📞 Support

- **Setup Help**: Check relevant README files
- **How-To Guides**: See DEVELOPMENT guides
- **Architecture Questions**: See ARCHITECTURE documents
- **Issues**: Check GitHub Issues
- **Contributions**: See CONTRIBUTING.md

## 📄 License

MIT License - See [LICENSE](./LICENSE) file

## 🙏 Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

---

## File Index

```
/
├── backend/
│   ├── README.md                          # Backend setup
│   ├── docs/ARCHITECTURE.md               # Backend architecture
│   ├── cmd/api/main.go                    # Entry point
│   ├── internal/                          # Core implementation
│   └── pkg/notifier/                      # Notification providers
├── frontend/
│   ├── README.md                          # Frontend overview
│   ├── DEVELOPMENT.md                     # Frontend patterns
│   ├── QUICKSTART.md                      # Frontend quick start
│   ├── src/
│   │   ├── services/                      # API client layer
│   │   ├── composables/                   # State management
│   │   ├── views/                         # Page components
│   │   ├── components/                    # Reusable components
│   │   ├── router/                        # Routing
│   │   ├── types/                         # TypeScript interfaces
│   │   └── App.vue                        # Root component
│   └── tailwind.config.ts                 # Styling config
├── .github/
│   └── copilot-instructions.md            # Backend AI guidelines
├── CONTRIBUTING.md                        # Contribution guidelines
├── FRONTEND_IMPLEMENTATION.md             # What was built
├── INTEGRATION.md                         # Backend-Frontend integration
├── QUICKSTART.md                          # This document
└── README.md                              # Project overview
```

---

**Ready to get started? Open [QUICKSTART.md](./QUICKSTART.md)! 🚀**

_Last Updated: October 19, 2025_
_Pulseguard - Open-source Monitoring Platform_
