# E-commerce Go Microservices

A modular e-commerce backend built with **Go** and three independent microservices: authentication & email, product catalog, and search. Each service can be developed, deployed, and scaled independently.

---

## Architecture Overview

```
                    ┌─────────────────────────────────────────────────────────┐
                    │                     Clients / API Gateway               │
                    └───────────────────────────┬─────────────────────────────┘
                                                │
         ┌──────────────────────────────────────┼───────────────────────────────────────┐
         │                                      │                                       │
         ▼                                      ▼                                       ▼
┌─────────────────────┐              ┌─────────────────────┐              ┌─────────────────────┐
│  auth-email         │              │  product            │              │  search             │
│  microservice       │              │  microservice       │              │  microservice       │
│  (gRPC)             │              │  (gRPC)             │              │  (HTTP REST)        │
│  • Auth / users     │              │  • Product CRUD     │              │  • Product searc    │
│  • Email (RabbitMQ) │              │  • Kafka events     │              │  • RabbitMQ consuer │
│  • PostgreSQL       │              │  • MongoDB          │              │  • Elasticsearch    │
│  • Redis            │              │  • Redis            │              │  • Bulk indexing    │
└─────────────────────┘              └─────────────────────┘              └─────────────────────┘
```

---

## Microservices

### 1. Auth-Email Microservice (`auth-email-microservice`)

Handles **user authentication** and **email delivery** via message queues.

| Aspect | Details |
|--------|---------|
| **API** | gRPC (UserService: Register, Login, FindByEmail, FindByID, GetMe, Logout) |
| **Responsibilities** | User registration/login, sessions, sending emails (e.g. verification, notifications) |
| **Datastores** | **PostgreSQL** (users, emails), **Redis** (sessions/cache) |
| **Messaging** | **RabbitMQ** (consume email jobs, publish events) |
| **Email** | **Resend** (or SMTP) for outbound email |
| **Observability** | Jaeger (tracing), Prometheus, Grafana, Zap (logging) |

**Ports:** 5001 (gRPC), 5555, 7070 (metrics/debug)

---

### 2. Product Microservice (`product-microservice`)

Manages the **product catalog** and publishes product lifecycle events for other services.

| Aspect | Details |
|--------|---------|
| **API** | gRPC (product service: Create, Update, Get, List, etc.) |
| **Responsibilities** | Product CRUD, persistence, publishing create/update events |
| **Datastores** | **MongoDB** (products), **Redis** (cache) |
| **Messaging** | **Apache Kafka** (publish product created/updated events) |
| **Observability** | Jaeger (tracing), Prometheus, Grafana, Zap (logging) |

**Ports:** 5555 (gRPC), 5007 (metrics/debug)

---

### 3. Search Microservice (`search-microservice`)

Provides **full-text product search** and keeps the search index in sync with product events.

| Aspect | Details |
|--------|---------|
| **API** | HTTP REST (e.g. `GET /api/v1/products` with query params for search) |
| **Responsibilities** | Consume product index events, bulk-index into Elasticsearch, serve search and health |
| **Datastores** | **Elasticsearch** (product search index) |
| **Messaging** | **RabbitMQ** (consume product index messages, bulk indexer) |
| **Observability** | Jaeger (tracing), Prometheus, health endpoints (`/health/ready`, `/health/live`), Zap (logging) |

**Ports:** 8000 (HTTP API), 8001 (health), 9090 (metrics)

---

## Technologies Used

| Category | Technologies |
|----------|---------------|
| **Language** | Go 1.25.x |
| **APIs** | gRPC (auth, product), HTTP/REST (search) |
| **Databases** | PostgreSQL, MongoDB, Redis, Elasticsearch |
| **Messaging** | RabbitMQ, Apache Kafka |
| **Config** | Viper, YAML |
| **Logging** | Uber Zap |
| **Tracing** | OpenTracing, Jaeger |
| **Metrics** | Prometheus, Grafana |
| **Validation** | go-playground/validator |
| **Email** | Resend, gomail |

---

## Repository Structure

```
ecom-go-microservices/
├── auth-email-microservice/   # Auth + email (gRPC, PostgreSQL, Redis, RabbitMQ)
├── product-microservice/      # Products (gRPC, MongoDB, Redis, Kafka)
├── search-microservice/       # Search (HTTP, Elasticsearch, RabbitMQ)
└── README.md
```

Each microservice has its own:

- `go.mod` / `go.sum`
- `config` (YAML / env)
- `docker-compose.yaml` for local run
- `Dockerfile` (and optionally Kubernetes/Helm under `k8s/`)

---

## Quick Start (per service)

Each service can be run standalone with Docker Compose from its directory:

```bash
# Auth-email (PostgreSQL, Redis, RabbitMQ, Jaeger, Prometheus, Grafana)
cd auth-email-microservice && docker compose -f docker-compose.yaml up --build

# Product (MongoDB, Redis, Kafka, Jaeger, Prometheus, Grafana)
cd product-microservice && docker compose -f docker-compose.yaml up --build

# Search (Elasticsearch, Kibana, RabbitMQ, Jaeger, Prometheus)
cd search-microservice && docker compose -f docker-compose.yaml up --build
```

For **Kubernetes** (e.g. Minikube), the search microservice includes a Helm chart under `search-microservice/k8s/microservice/` (Elasticsearch, Kibana, RabbitMQ, Jaeger, and the app).

---
