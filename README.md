# Recess 🏫🎮

> A production-quality real-time multiplayer gaming platform inspired by classic school-time games.

---

## 🎮 The Games

1. **Hand Cricket** — Odd/Even toss, batting, bowling, synchronized finger choice reveals, target chasing.
2. **Dots & Boxes** — Graph paper grid, line draws, 1x1 box claiming, and bonus turns.
3. **XO / Tic-Tac-Toe** — 3x3 Classic and 4x4/5x5 Extended with countdown timers and line-strike animations.
4. **Connect 4** — 7x6 wooden desk grid, gravity drops, and 4-in-a-row detection.
5. **Paper Football** — Desk flick physics, table-edge overhang touchdowns (6 pts), and field goal kicks.
6. **Name–Place–Animal–Thing (NPAT)** — Letter generator, synchronized timer, STOP button, and dictionary/peer validation.

---

## 🏗️ Architecture & Technology Stack

- **Backend**: Go (Chi router, Gorilla WebSockets, `pgxpool`, `go-redis`, `log/slog` structured logging, Prometheus metrics).
- **Persistence**: PostgreSQL 16 (`pgx`/`sqlc`) + Redis 7 (Pub/Sub & session locks).
- **Frontend**: React, TypeScript, Vite, Tailwind CSS, Zustand, TanStack Query, Framer Motion.
- **Infrastructure**: Docker, Docker Compose, AWS ECS/Fargate, ALB, CloudFront, RDS, ElastiCache.

---

## 🚀 Quick Start: Local Development Infrastructure (Phase 1)

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) & Docker Compose
- [Go](https://golang.org/dl/) (version 1.22+)
- `curl` and `jq` (optional, for inspecting JSON responses)

### 1. Configuration Setup
Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Default local environment variables:
```dotenv
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=info
POSTGRES_USER=recess
POSTGRES_PASSWORD=recess_secret
POSTGRES_DB=recess_db
DATABASE_URL=postgres://recess:recess_secret@postgres:5432/recess_db?sslmode=disable
REDIS_URL=redis://redis:6379/0
```

### 2. Start Services with Docker Compose
Run the entire backend ecosystem with a single command:

```bash
docker compose up -d --build
```

Check the health and status of all containers:
```bash
docker compose ps
```

You should see 3 healthy containers:
- `recess-postgres` (PostgreSQL 16 on port `5432`)
- `recess-redis` (Redis 7 on port `6379`)
- `recess-backend` (Go API on port `8080`)

### 3. Verify Health & Readiness Endpoints

#### Liveness Probe (`GET /health`)
Verifies that the application process is running:
```bash
curl http://localhost:8080/health
```
Response:
```json
{
  "status": "ok",
  "environment": "development",
  "uptime_seconds": 12.4,
  "timestamp": "2026-08-22T06:50:00Z"
}
```

#### Readiness Probe (`GET /ready`)
Verifies active connectivity and response latencies for PostgreSQL and Redis:
```bash
curl http://localhost:8080/ready
```
Response (HTTP `200 OK` when all dependencies are healthy):
```json
{
  "status": "ready",
  "environment": "development",
  "uptime_seconds": 12.4,
  "timestamp": "2026-08-22T06:50:00Z",
  "checks": {
    "database": {
      "status": "up",
      "latency_ms": 1.15
    },
    "redis": {
      "status": "up",
      "latency_ms": 0.82
    }
  }
}
```

If any service is unreachable or degraded, `/ready` responds with HTTP `503 Service Unavailable` with details in the JSON body.

---

## 🛠️ Developer Commands (`Makefile`)

| Command | Description |
| :--- | :--- |
| `make run-backend` | Run Go backend locally on the host machine |
| `make test-backend` | Run all Go unit and integration tests |
| `make fmt` | Format Go source code with `gofmt` |
| `make lint` | Run `go vet ./...` static analysis |
| `make docker-up` | Build and start all containers in background |
| `make docker-down` | Stop and remove all containers |
| `make docker-logs` | Stream logs from all Docker containers |
| `make docker-ps` | List running Docker containers and health statuses |
| `make health` | Query `GET /health` |
| `make ready` | Query `GET /ready` |

---

## 📂 Project Structure

```
Recess/
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       ├── main.go             # Chi HTTP router, health endpoints, graceful shutdown
│   │       └── main_test.go        # Liveness & readiness test suite
│   ├── internal/
│   │   ├── config/                 # Environment & runtime configuration
│   │   ├── database/               # PostgreSQL pgxpool connection & health check
│   │   ├── redis/                  # Redis client connection & health check
│   │   ├── middleware/             # Structured logging & X-Request-ID propagation
│   │   ├── games/                  # Independent game engines (Phase 3)
│   │   ├── websocket/              # Real-time WebSocket hub (Phase 4)
│   │   ├── matchmaking/            # Matchmaking queues (Phase 4)
│   │   ├── rooms/                  # Custom room management (Phase 4)
│   │   ├── auth/                   # JWT & guest auth (Phase 5)
│   │   ├── users/                  # User accounts & Elo (Phase 5)
│   │   └── leaderboard/            # Honor roll rankings (Phase 5)
│   ├── Dockerfile                  # Multi-stage production container
│   ├── go.mod
│   └── go.sum
├── frontend/                       # React + TypeScript + Tailwind frontend
├── docs/
│   └── DESIGN.md                   # Complete UI/UX design system & board specs
├── infra/
│   └── docker/                     # Infrastructure Dockerfiles
├── .env.example
├── docker-compose.yml              # PostgreSQL 16 + Redis 7 + Go backend
├── Makefile                        # Convenient development targets
└── README.md
```
