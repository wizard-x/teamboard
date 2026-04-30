<div align="center">

# 🗂️ TeamBoard

**A little, self-hosted project management board — like Trello, but yours.**

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go)](https://go.dev/)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.x-4FC08D?style=flat-square&logo=vue.js)](https://vuejs.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat-square&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat-square&logo=redis)](https://redis.io/)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?style=flat-square&logo=docker)](https://www.docker.com/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.x-3178C6?style=flat-square&logo=typescript)](https://www.typescriptlang.org/)

[Features](#-features) · [Quick Start](#-quick-start) · [API](#-api) · [Tech Stack](#-tech-stack) · [Testing](#-testing)

</div>

---

## ✨ Features

- 📋 **Kanban Boards** — create boards with customizable columns and drag-and-drop task management
- 👥 **Team Collaboration** — add members, assign tasks, and manage teams via API keys
- 📝 **Task Management** — create, assign, comment, set status, move tasks between columns
- 🔐 **API Key Authentication** — every team member gets a unique API key for secure access
- ⚡ **Redis Caching** — auth context caching and rate limiting for blazing-fast responses
- 🐳 **One-Command Deploy** — `docker compose up` and you're running
- 🎨 **Embedded SPA** — Vue.js frontend is compiled into the Go binary, no separate deploy needed
- 🧪 **Fully Tested** — backend unit + integration tests, frontend component tests

## 🏗️ Architecture

TeamBoard follows **Clean Architecture** with strict layer separation and SOLID principles:

```
┌─────────────────────────────────────────────────────┐
│                    Vue.js SPA                       │
│         (embedded into Go binary via embed.FS)      │
├─────────────────────────────────────────────────────┤
│  Handler Layer (HTTP)  ─── Echo v4 + Middleware     │
│    ├── Auth (API Key)  ├── CORS  ├── Rate Limiter   │
│    └── Logger          └── Security Headers         │
├─────────────────────────────────────────────────────┤
│  Service Layer (Business Logic)                     │
│    └── DTOs between layers, validation              │
├─────────────────────────────────────────────────────┤
│  Repository Layer (Data Access)                     │
│    ├── PostgreSQL (pgxpool) ── persistent storage   │
│    └── Redis (go-redis) ── cache & sessions         │
├─────────────────────────────────────────────────────┤
│  Domain Layer (Entities)                            │
│    Board, Column, Task, Comment, Team, Member       │
└─────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/yourname/teamboard.git
cd teamboard

# Start everything with one command
docker compose up -d --build
```

The app will be available at **http://localhost:8080**

### Local Development

```bash
# 1. Start infrastructure (PostgreSQL + Redis)
docker compose up -d postgres redis

# 2. Run migrations
make migrate-up

# 3. Start backend + frontend in dev mode
make dev
```

### Environment Variables

Copy the example env file and adjust:

```bash
cp src/backend/.env.example src/backend/.env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `trello` | Database user |
| `DB_PASSWORD` | `trello_secret` | Database password |
| `DB_NAME` | `trello` | Database name |
| `DB_SSL_MODE` | `disable` | SSL mode |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `SERVER_PORT` | `8080` | HTTP server port |

## 📡 API

RESTful API with JSON payloads, versioned under `/api/v1`. All endpoints (except registration and health) require an API key via the `X-API-Key` header.

### Endpoint Overview

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/health` | Health check |
| `POST` | `/api/v1/teams` | Create a team |
| `GET` | `/api/v1/teams/:id` | Get team details |
| `POST` | `/api/v1/teams/:id/members` | Add team member |
| `DELETE` | `/api/v1/teams/:id/members/:mid` | Remove member |
| `GET` | `/api/v1/boards` | List boards |
| `POST` | `/api/v1/boards` | Create board |
| `GET` | `/api/v1/boards/:id` | Get board with columns & tasks |
| `PATCH` | `/api/v1/boards/:id` | Update board |
| `DELETE` | `/api/v1/boards/:id` | Delete board (soft) |
| `POST` | `/api/v1/boards/:id/columns` | Add column |
| `PATCH` | `/api/v1/boards/:id/columns/:cid` | Update column |
| `DELETE` | `/api/v1/boards/:id/columns/:cid` | Delete column |
| `POST` | `/api/v1/columns/:cid/tasks` | Create task |
| `PATCH` | `/api/v1/tasks/:id` | Update task |
| `POST` | `/api/v1/tasks/:id/move` | Move task between columns |
| `POST` | `/api/v1/tasks/:id/comments` | Add comment |
| `GET` | `/api/v1/tasks/:id/comments` | List comments |
| `GET` | `/api/v1/members/me` | Get current member |
| `PATCH` | `/api/v1/members/me` | Update profile |

### Example

```bash
# Create a team
curl -X POST http://localhost:8080/api/v1/teams \
  -H "Content-Type: application/json" \
  -d '{"name": "My Team"}'

# Response contains API keys for team members
# Use the key in subsequent requests:
curl http://localhost:8080/api/v1/boards \
  -H "X-API-Key: tb_your_api_key_here"
```

## 🛠️ Tech Stack

### Backend

| Technology | Purpose |
|------------|---------|
| **Go 1.22** | Backend language — fast, typed, great concurrency |
| **Echo v4** | HTTP framework — routing, middleware, static serving |
| **PostgreSQL 16** | Primary data store — ACID, relations, JSON support |
| **pgxpool** | PostgreSQL driver — connection pooling, native types |
| **Redis 7** | Cache — auth context, rate limiting |
| **go-redis** | Redis client for Go |
| **cleanenv** | Config management via struct tags |
| **godotenv** | `.env` file loading |
| **goose** | SQL migrations with up/down support |
| **ulid** | Lexicographically sortable unique IDs |

### Frontend

| Technology | Purpose |
|------------|---------|
| **Vue.js 3** | UI framework — Composition API, SFC |
| **TypeScript** | Type safety |
| **Vite** | Build tool — HMR, tree-shaking |
| **Pinia** | State management |
| **VueDraggable Plus** | Kanban drag-and-drop |

### Infrastructure

| Technology | Purpose |
|------------|---------|
| **Docker + Compose** | Single-command deployment |
| **Multi-stage Dockerfile** | Frontend build → Go build → Alpine runtime |
| **Makefile** | Standardized build/test/run commands |

## 🧪 Testing

```bash
# Run all tests
make test

# Run backend tests only
make test-backend

# Run frontend tests only
make test-frontend

# Lint everything
make lint
```

### Coverage Targets

| Layer       | Minimum Coverage |
| -------------| ------------------|
| Service     | 80%              |
| Repository  | 70%              |
| Handler     | 75%              |
| **Overall** | **70%**          |

### What's Tested

- ✅ **Backend** — unit tests for services, handlers, entities; integration tests with real PostgreSQL + Redis
- ✅ **Frontend** — component tests (AppHeader, TaskCard), composable tests, API client tests, type tests
- ✅ **E2E** — full integration test suite with Docker Compose

## 📁 Project Structure

```
teamboard/
├── docker-compose.yml          # Service orchestration
├── Dockerfile                  # Multi-stage build
├── Makefile                    # Build automation
└── src/
    ├── backend/
    │   ├── cmd/server/         # Entry point + embedded frontend
    │   └── internal/
    │       ├── config/         # Environment config
    │       ├── domain/         # Entities + domain errors
    │       ├── dto/            # Request/Response DTOs
    │       ├── handler/        # HTTP handlers
    │       ├── middleware/     # Auth, CORS, rate limit, logging
    │       ├── repository/     # PostgreSQL + Redis repos
    │       └── service/        # Business logic
    └── frontend/
        └── src/
            ├── api/            # API client
            ├── components/     # Vue components
            ├── composables/    # Vue composables (useAuth, etc.)
            ├── pages/          # Page components
            ├── router/         # Vue Router config
            └── types/          # TypeScript types
```

## 📋 Makefile Commands

```
make help              Show all available commands
make dev               Start backend + frontend in dev mode
make build             Build everything (frontend + backend)
make test              Run all tests
make test-backend      Run backend tests
make test-frontend     Run frontend tests
make migrate-up        Run database migrations
make migrate-down      Rollback last migration
make migrate-create    Create new migration (name=...)
make docker-up         Start all services
make docker-down       Stop all services
make docker-build      Build Docker image
make lint              Lint everything
make clean             Clean build artifacts
```

## 📄 License

This project is licensed under the MIT License.

---

<div align="center">

**Built with ❤️ using Go + Vue.js**

</div>
