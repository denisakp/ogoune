# Pulseguard

🔐 **Pulseguard** is an open-source monitoring tool for SSL certificates, domain expirations, TCP services, and cronjobs — with alerting and public status pages.

It is designed with a cloud-native, scalable architecture to ensure reliability and maintainability, whether you're monitoring one service or thousands.

## 🌟 Features

-   🔐 Monitor **SSL/TLS certificate** expiration and algorithms   
-   🌐 Track **domain name** expiration (WHOIS-based)
-   📡 Check and monitor **TCP port** (HTTP, PostgresSQL, MySQL, etc.)
-   ⏱️ Monitor **cronjobs** via ping endpoints ("heartbeat" monitoring)
-   📈 Generate **public status pages** with SLA/SLO insights
-   📬 Get notified via **Email**, **Google Chat**, and **Slack**
-   🔄 Support for **webhooks** to integrate with other services
-    
----------

## 🚀 Tech Stack & Rationale

The technology stack for Pulseguard was chosen to prioritize performance, scalability, and ease of deployment in modern cloud environments.

### Backend: Golang + Chi

-   **Golang**: Go is the ideal choice for a high-performance monitoring tool. Its powerful concurrency model (goroutines) allows Pulseguard to perform thousands of checks simultaneously with a minimal memory footprint. The compilation to a single static binary makes containerization with Docker trivial and highly efficient.
    
-   **Chi (`go-chi/chi`)**: While the standard `http.ServeMux` is capable, **Chi** is used as the HTTP router. It provides a more powerful and flexible API for building structured web applications, offering excellent middleware support (for logging, metrics, authentication) and clean context-aware request handling, which is essential as the application grows.
    

### Database: PostgreSQL + Redis

-   **PostgreSQL**: Serves as the primary data store for persistent, relational data. This includes monitor configurations, historical check results, user accounts, and data for calculating uptime and SLAs.
    
-   **Redis**: Plays a crucial dual role, making the architecture asynchronous and scalable:
    
    1.  **Job Queue**: It acts as a message broker for all monitoring tasks. This decouples task scheduling from execution, meaning the API can remain fast and responsive while background workers handle the actual checks.
        
    2.  **Cache**: Used for caching recent check results to serve public status pages quickly without querying the main database on every page load.
        

### Frontend: HTMX + Daisy UI

-   This choice intentionally avoids the complexity of a full-fledged Single-Page Application (SPA) framework. **HTMX** allows us to build a modern, dynamic user interface while keeping all rendering logic on the Go backend. This simplifies the codebase, reduces the frontend build complexity, and aligns perfectly with Go's strengths. **Daisy UI** (on top of TailwindCSS) provides a clean component library for rapid UI development.
    
----------

## 🏗️ Architecture Overview

Pulseguard is built on a **scalable, asynchronous, layered architecture**. The core principle is the separation of concerns, which allows different parts of the application to be developed, tested, and scaled independently.

The system is composed of two main deployable units: the **API Server** and the **Worker**.

1.  **API Server (`cmd/api`)**: This is the user-facing component. It handles all HTTP requests, serves the web interface (HTMX), and provides an API for managing monitors. When a check needs to be performed (either scheduled or manually triggered), the API server **does not perform the check itself**. Instead, it pushes a "job" onto the **Redis job queue**.
    
2.  **Worker (`cmd/worker`)**: This is a background process that is completely separate from the API server. It continuously polls the Redis job queue for new tasks. When it picks up a job (e.g., "check SSL for `example.com`"), it executes the necessary logic, and stores the result in the PostgreSQL database.
    
3.  **Decoupling**: This asynchronous pattern is key to scalability. If you need to handle more web traffic, you can scale up the number of `api` server instances. If you need to perform more checks per second, you can scale up the number of `worker` instances, all without impacting each other.
    
----------

## 📁 Project Structure

The project follows the standard Go project layout to ensure maintainability and a clear separation of concerns.


``` plain
pulseguard/
├── cmd/                # Application entrypoints
│   ├── api/            # The main web server (API and HTMX frontend)
│   │   └── main.go
│   └── worker/         # The background job processor
│       └── main.go
│
├── internal/           # All private application logic
│   ├── api/            # HTTP handlers, routing, and middleware
│   │   ├── handler/    # Functions that handle specific HTTP requests
│   │   └── router.go   # Route definitions using Chi
│   │
│   ├── config/         # Configuration loading (from env vars or files)
│   │   └── config.go
│   │
│   ├── domain/         # Core business logic and data models. No knowledge of HTTP or SQL.
│   │   ├── monitor/    # Business logic for monitors
│   │   ├── check/      # The actual logic for performing SSL, TCP, etc., checks
│   │   └── models.go   # Primary data structures (Monitor, CheckResult, etc.)
│   │
│   ├── repository/     # Data access layer (interfaces and implementations)
│   │   ├── postgres/   # PostgreSQL-specific implementation
│   │   └── redis/      # Redis-specific implementation (job queueing, caching)
│   │
│   └── worker/         # The business logic for the job processor itself
│       └── processor.go
│
├── pkg/                # Sharable, public libraries (e.g., notification clients)
│   └── notifier/       # Clients for sending Slack, Email, etc., notifications
│
├── web/                # Frontend assets
│   ├── static/         # CSS, JS, images
│   └── template/       # HTML templates rendered by Go
│
├── go.mod
├── go.sum
└── Dockerfile

```

## 📦 Installation (soon)

Self-hosted setup via Docker Compose and Kubernetes Helm chart will be available soon.

## 📅 Roadmap

Check the [Pulseguard Roadmap](https://github.com/denisakp/pulseguard/projects) to see what's planned.

## 🤝 Contributing

Contributions are welcome! Please open issues or pull requests.

## 📄 License

MIT

----------
