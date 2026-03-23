<div align="right" width="100%">
    <img src="./static/ico.png" width="40" alt="PulseGuard Logo" />
</div>

# PulseGuard

**Simple, self-hosted uptime monitoring. Check if your resources are up.**

![License](https://img.shields.io/badge/license-MIT-blue)
![Go](https://img.shields.io/badge/go-1.25%2B-00ADD8)
![Vue](https://img.shields.io/badge/vue-3.x-4FC08D)
[![GitHub Stars](https://img.shields.io/github/stars/denisakp/pulseguard?style=flat)](https://github.com/denisakp/pulseguard)

PulseGuard monitors your resources, including websites, APIs, and services. If something goes down, you get notified. That's it.

No complex setup. No overwhelming dashboards. Just pure uptime monitoring.

<img src="./static/dashboard.png" alt="PulseGuard Dashboard" width="100%" style="border-radius: 8px; margin-top: 20px;" />

---

## 🤔 Why PulseGuard?

I started exploring monitoring stacks like Prometheus, Grafana, Tempo, and AlertManager. But configuring dozens of config files just to check if my resources were up seemed crazy.

So I built this during my internship in 2023 with TypeScript and NestJS. Later, I rewrote it in Go while learning the language. Now it's a simple, straightforward monitoring tool that just works.

---

##  Get Started in 30 Seconds

```bash
git clone https://github.com/denisakp/pulseguard.git
cd pulseguard
docker compose -f docker-compose.community.yml up -d
```

This community path uses embedded SQLite and Redis only. Hosted PostgreSQL deployments still use `docker compose up -d`.

Open **http://localhost:8080** and log in with:
- Email: `admin@pulseguard.test`
- Password: `puls3gu@rd`

Change the password on first login.
<img src="./static/login.png" alt="PulseGuard Login Screen" width="100%" style="border-radius: 8px; margin-top: 20px;" />
---

## ✨ What You Get

- 🌐 **Monitor Websites** – HTTP/HTTPS checks
- 🔌 **Monitor Services** – TCP port checks
- 🔔 **Get Notified** – Email, Slack, Webhooks
- 📊 **Track Incidents** – See when things went wrong
- 🌍 **Status Page** – Share status with customers
- 🛠️ **Maintenance Windows** – Avoid false alarms during updates
- 🏷️ **Organize** – Tag and group monitors
- 🔐 **Secure** – 2FA support

<img src="./static/monitored-resource.png" alt="Create and Monitor Resources" width="100%" style="border-radius: 8px; margin-top: 20px;" />

---

## 📑 Table of Contents

- [Installation](#installation)
- [How It Works](#how-it-works)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [License](#license)

---

## Installation

### Docker (Recommended)

```bash
git clone https://github.com/denisakp/pulseguard.git
cd pulseguard
cp .env.example .env
docker compose -f docker-compose.community.yml up -d
```

Access at **http://localhost:8080**

For a hosted PostgreSQL stack, run `docker compose up -d` instead.

---

## How It Works

1. **Add Monitors** – Tell PulseGuard what to check (websites, APIs, services)
2. **Automatic Checks** – It checks every 5 minutes by default (customizable)
3. **Track Status** – See uptime history and incident timeline
4. **Get Alerts** – Email notifications when things go down
5. **Status Page** – Share public status with customers

That's it. No complexity.

<img src="./static/incident.png" alt="Incident Tracking and Timeline" width="100%" style="border-radius: 8px; margin-top: 20px;" />

<img src="./static/notification-configuration.png" alt="Notification Channels Setup" width="100%" style="border-radius: 8px; margin-top: 20px;" />

<img src="./static/maintenance.png" alt="Maintenance Windows" width="100%" style="border-radius: 8px; margin-top: 20px;" />

<img src="./static/public-status-page.png" alt="Public Status Page" width="100%" style="border-radius: 8px; margin-top: 20px;" />

---

## Configuration

### Environment Variables

```env
# Database
DB_DRIVER=sqlite
SQLITE_PATH=/data/pulseguard.db
DB_LOG_LEVEL=error
DATABASE_URL=postgres://user:password@host:5432/pulseguard
REDIS_URL=localhost:6379

```

All options in `.env.example`

SQLite removes the external database dependency for Community Edition, but Redis is still required for scheduling and worker execution.

Automatic PostgreSQL-to-SQLite data migration is out of scope. Switch to SQLite only for fresh community deployments.

---

---

## 💭 Feedback & Testing

We're actively developing PulseGuard and value your input! Help us improve by:

- **[Share Your Feedback](https://kawa-bunga.notion.site/2d1e5ad0a17d80dc8859e77817d901e3)** (Anonymous form) – Tell us what you think about the UI, features, and user experience
- **Report Bugs** – Found something broken? Open an [issue](https://github.com/denisakp/pulseguard/issues)
- **Suggest Features** – Have ideas? Start a [discussion](https://github.com/denisakp/pulseguard/discussions)

Your feedback helps shape the future of PulseGuard. The feedback form is completely anonymous and takes about 2 minutes.

## 💬 Contributing

Found a bug? Have an idea? Let us know!

- **[GitHub Issues](https://github.com/denisakp/pulseguard/issues)** – Report bugs or request features
- **[GitHub Discussions](https://github.com/denisakp/pulseguard/discussions)** – Ask questions

We welcome pull requests. Please read [CONTRIBUTING.md](./CONTRIBUTING.md) first.

---

## 📄 License

MIT License – See [LICENSE](./LICENSE) for details.

You can use PulseGuard for commercial or personal projects.

---

## 📚 More Info

- **[Quick Start Guide](./QUICKSTART.md)** – Detailed setup walkthrough
- **[Contributing Guidelines](./CONTRIBUTING.md)** – How to help
- **[Architecture Docs](./backend/ARCHITECTURE.md)** – How it works under the hood
- **[Security Policy](./SECURITY.md)** – Reporting security issues

---

<div align="center">

**[⬆ Back to top](#pulseguard)**

Built with ❤️ by [denisakp](https://github.com/denisakp)

</div>
