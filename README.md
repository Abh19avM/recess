# Recess

A production-grade, real-time multiplayer gaming platform inspired by classic school-time paper and desk games. Built with a server-authoritative Go backend, PostgreSQL, Redis, real-time WebSockets, and a React + TypeScript frontend adhering to a nostalgic classroom stationery design system.

---

## Overview

Recess brings classic schoolroom games online with low-latency multiplayer mechanics, deterministic server-side rule validation, rated Elo matchmaking, and room-based desk lobbies. The platform is designed from the ground up to prevent client-side tampering, handle abrupt network drops gracefully with session state recovery, and scale horizontally across nodes.

### Supported Games
1. **Hand Cricket** — Turn-based simultaneous reveals (1–6 fingers), odd/even toss, batting/bowling turns, sudden-death wicket detection, and target chasing.
2. **Dots & Boxes** — Graph paper grid coordinate claims, 1x1 box completion detection, continuous bonus turns, and territory tracking.
3. **XO / Tic-Tac-Toe** — 3x3 Classic and extended tactical grids with strict turn countdown timers, win vector analysis, and streak tracking.
4. **Connect 4** — 7x6 wooden desk grid with column gravity drops and four-in-a-row detection across horizontals, verticals, and diagonals.
5. **Paper Football** — Tabletop drag-and-flick physics with boundary overhang touchdown detection (6 pts) and field goal kick phases.
6. **Name–Place–Animal–Thing (NPAT)** — Multiplayer alphabet buzzer, synchronized round clocks, automated dictionary validation, and duplicate scoring logic.

---

## Architecture & Design Principles

```
                  +-----------------------------------+
                  |         Client Browsers           |
                  +-----------------+-----------------+
                                    |
                         HTTPS / WSS (JSON Envelope)
                                    |
                                    v
                  +-----------------------------------+
                  |         Go HTTP / WSS API         |
                  |     (Chi Router, Gorilla WS)      |
                  +--------+-----------------+--------+
                           |                 |
                +----------v-------+ +-------v----------+
                |   Game Engine    | |  WebSocket Hub   |
                | (Deterministic)  | |  (Room Presence) |
                +------------------+ +-------+----------+
                           |                 |
                +----------v-------+ +-------v----------+
                | PostgreSQL (pgx) | | Redis (Pub/Sub)  |
                +------------------+ +------------------+
```

### Core Tenets
- **Server-Authoritative State**: Clients submit moves and intentions; the Go backend validates legality, executes state transitions, and broadcasts deterministic state updates.
- **Transport Separation**: The game rule engines implement an isolated `Engine` interface with zero dependencies on HTTP or WebSocket transport layers.
- **Deterministic State Evolution**: Replaying an ordered series of moves produces identical board states and sequential version numbers across all nodes.
- **Resilient WebSocket Protocol**: Standard JSON event envelopes, protocol and application heartbeats, connection lifecycle tracking, and graceful disconnect detection.
- **Production Security**: Argon2id password hashing, short-lived JWT access tokens, cryptographically secure refresh tokens, and rate-limited auth endpoints.

---

## Technology Stack

### Backend
- **Language**: Go 1.24+
- **HTTP Routing**: Chi Router (`github.com/go-chi/chi/v5`)
- **WebSockets**: Gorilla WebSocket (`github.com/gorilla/websocket`)
- **Database & Persistence**: PostgreSQL 16 via `pgxpool` (`github.com/jackc/pgx/v5`) and `sqlc`
- **Cache & Pub/Sub**: Redis 7 via `go-redis` (`github.com/redis/go-redis/v9`)
- **Authentication**: Argon2id password hashing + HMAC-SHA256 JWT tokens
- **Logging & Metrics**: `log/slog` structured JSON logging, Prometheus instrumentation

### Frontend
- **Framework**: React 19, TypeScript, Vite
- **Styling**: Vanilla Tailwind CSS v4 with custom paper, ruled notebook, and rubber stamp utilities
- **State Management**: Zustand (persisted auth store, toast notifications)
- **Data Fetching**: TanStack React Query v5
- **Routing**: React Router v7
- **Icons**: Lucide React
- **Testing**: Vitest, React Testing Library, JSDOM

---

## Generic Game Engine Specification

All games conform to the `engine.Engine` interface defined in `backend/internal/games/engine`:

```go
type Engine interface {
    Initialize(gameID string, players []Player, config json.RawMessage) (*GameState, error)
    ValidateMove(move Move) error
    ApplyMove(move Move) (*GameState, error)
    State() *GameState
    IsFinished() bool
    Result() *GameResult
    NextTurn() string
    GameType() GameType
}
```

Engines register dynamically with a thread-safe `Registry` using unique identifiers (`xo`, `hand_cricket`, `dots_boxes`, `connect4`, `paper_football`, `npat`), enabling room managers to instantiate any supported game engine at runtime.

---

## Real-Time WebSocket Protocol

WebSocket communications utilize structured JSON event envelopes:

```json
{
  "type": "room.join",
  "room_id": "RECESS-BENCH-1",
  "sequence": 1,
  "timestamp": 1771665420000,
  "payload": {}
}
```

### Core Lifecycle Events
| Event Type | Direction | Description |
| :--- | :--- | :--- |
| `room.join` | Client -> Server | Request to join a desk room |
| `room.leave` | Client -> Server | Request to depart a desk room |
| `room.state` | Server -> Client | Initial room snapshot with active member roster |
| `player.joined` | Server -> Clients | Broadcast when a new student joins the room |
| `player.left` | Server -> Clients | Broadcast when a student leaves the room |
| `player.ready` | Bidirectional | Toggle and broadcast player ready status |
| `player.disconnected` | Server -> Clients | Broadcast upon unexpected connection termination |
| `room.message` | Bidirectional | Arbitrary broadcast messages / classroom notes |
| `room.error` | Server -> Client | Structured error frame on validation failures |

---

## Project Structure

```
Recess/
├── backend/
│   ├── cmd/server/             # Application entrypoint and bootstrapping
│   ├── internal/
│   │   ├── app/                # Server lifecycle, dependency injection, and router wiring
│   │   ├── auth/               # Argon2id password hashing, JWT tokens, session store
│   │   ├── config/             # Environment variables and runtime configuration
│   │   ├── database/           # PostgreSQL connection pool and sqlc models
│   │   ├── games/              # Game engines, registry, and metadata service
│   │   │   ├── engine/         # Generic Engine interface, types, and Registry
│   │   │   └── xo/             # Reference XO / Tic-Tac-Toe engine implementation
│   │   ├── httputil/           # Standard JSON response envelopes and error handlers
│   │   ├── middleware/         # Structured logger, recoverer, auth barrier, rate limiter
│   │   ├── redis/              # Redis client connection and health checks
│   │   ├── rooms/              # Room entity repositories and lifecycle management
│   │   ├── users/              # User domain models, repositories, and handlers
│   │   └── websocket/          # WebSocket connection manager, client pumps, and hub
│   ├── migrations/             # SQL schema migrations (PostgreSQL)
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── src/
│   │   ├── components/         # Reusable design system components (Button, Input, PaperCard, etc.)
│   │   ├── hooks/              # Custom hooks including typed useWebSocket
│   │   ├── lib/                # API client, utility functions
│   │   ├── pages/              # Primary routes (Landing, Login, Register, Dashboard, Games, Profile, WebSocket test)
│   │   ├── routes/             # ProtectedRoute barrier
│   │   └── store/              # Zustand auth and toast state stores
│   ├── index.html
│   ├── package.json
│   └── vite.config.ts
├── docs/
│   ├── API.md                  # Comprehensive REST and WebSocket API documentation
│   ├── DATABASE.md             # Schema documentation, tables, and relationships
│   ├── DESIGN.md               # UI/UX design specifications and color tokens
│   └── decisions/              # Architecture Decision Records (ADRs)
├── docker-compose.yml          # Local container orchestration (PostgreSQL 16, Redis 7, Backend)
├── Makefile                    # Standardized development and testing targets
└── README.md
```

---

## Quick Start & Local Setup

### Prerequisites
- [Go](https://golang.org/dl/) (1.24+)
- [Node.js](https://nodejs.org/) (20+) & `npm`
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose

### 1. Clone & Configure Environment
```bash
git clone https://github.com/Abh19avM/recess.git
cd recess
cp .env.example .env
```

### 2. Start Infrastructure Dependencies
```bash
docker compose up -d
```

Verify that containers are healthy:
```bash
docker compose ps
```

### 3. Run Backend Service
```bash
cd backend
go run ./cmd/server
```
The API server will listen on `http://localhost:8080`.

### 4. Run Frontend Development Server
```bash
cd frontend
npm install
npm run dev
```
The Vite development server will open at `http://localhost:5173`.

---

## Testing & Verification

### Run Backend Unit and Integration Tests
```bash
cd backend
go test -v ./...
```

### Run Frontend Unit Tests
```bash
cd frontend
npm run test
```

### Run Frontend Production Build Validation
```bash
cd frontend
npm run build
```

---

## Makefile Targets

| Target | Description |
| :--- | :--- |
| `make run-backend` | Start the Go backend application locally |
| `make test-backend` | Run all Go test packages with coverage reporting |
| `make fmt` | Format all Go code using `gofmt` |
| `make lint` | Run static analysis with `go vet ./...` |
| `make docker-up` | Spin up PostgreSQL, Redis, and backend containers |
| `make docker-down` | Tear down Docker Compose containers |
| `make docker-logs` | Stream unified container logs |
| `make health` | Query liveness probe `GET /health` |
| `make ready` | Query readiness probe `GET /ready` |
