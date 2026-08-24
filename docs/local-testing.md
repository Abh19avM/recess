# Recess Local Testing & Developer Guide

This document explains how to set up, run, and test the complete Recess platform locally from a clean state.

---

## 1. Prerequisites

- **Go**: Version 1.23 or 1.24+
- **Node.js**: Version 20+ or 22+ (with `npm`)
- **Docker & Docker Compose**: Version 24+ (optional for full-stack containerization)
- **k6**: Version 0.50+ (for load testing)

---

## 2. Fast Local Development Startup

### A. Run Entire Stack via Docker Compose (Recommended)
```bash
# Clone and enter directory
cd Recess

# Start PostgreSQL, Redis, Backend, Frontend, Prometheus, and Grafana
docker compose up --build -d

# Verify services health
docker compose ps
```

- **Frontend Application**: `http://localhost:5173`
- **Go Backend API**: `http://localhost:8080`
- **Prometheus Metrics**: `http://localhost:9090`
- **Grafana Dashboards**: `http://localhost:3000` (User: `admin`, Pass: `admin`)

---

### B. Run Backend & Frontend Independently

#### 1. Backend Server
```bash
cd backend

# Compile and run
go run ./cmd/server
# Listening on http://0.0.0.0:8080
```

Environment variables (optional overrides in `.env`):
```bash
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=info
DATABASE_URL=postgres://recess:recess_secret@localhost:5432/recess_db?sslmode=disable
REDIS_URL=redis://localhost:6379/0
```

#### 2. Frontend Development Server
```bash
cd frontend

# Install dependencies
npm install

# Start Vite dev server
npm run dev
# Ready at http://localhost:5173
```

---

## 3. Test Accounts & Guest Access

1. **Instant Guest Play**:
   - Click **"Play as Guest"** in the navigation bar or landing page.
   - Automatically provisions a temporary student identity with JWT session token.
2. **Standard Student Registration**:
   - Navigate to `/register`.
   - Sample credentials:
     - **Username**: `rahul_dravid`
     - **Email**: `rahul@recess.school`
     - **Password**: `SecretPassword123!`

---

## 4. Running Automated Test Suites

### A. Backend Unit, Integration & Engine Tests
```bash
cd backend
go test -count=1 ./...
go vet ./...
```
*Expected: All 23 packages passing in < 1.0s.*

### B. Frontend Vitest & React Testing Library
```bash
cd frontend
npm test
npm run build
```
*Expected: 6 test files, 27 tests passing, TypeScript bundle compiled cleanly.*

### C. Load & Concurrency Benchmarks (k6)
```bash
# HTTP REST API Benchmark
k6 run load-testing/http_load.js

# WebSocket Concurrent Sessions & Moves
k6 run load-testing/websocket_load.js

# Real-Time Matchmaking Queue & Pairing
k6 run load-testing/matchmaking_load.js

# Comprehensive Multi-Stage Load Scenario
k6 run load-testing/full_scenario.js
```

---

## 5. Known Limitations & Architecture Boundaries

1. **In-Memory Fallback vs Distributed Redis**:
   - When running in single-instance offline mode without Redis (`REDIS_URL=""`), matchmaking and session presence run through high-speed in-memory adapters.
   - In production (AWS ECS / multi-instance), Redis Pub/Sub provides cross-server event routing.
2. **Reconnection Window**:
   - Disconnected players have a **30-second grace period** to reconnect with their session token and sequence number before a forfeit or match termination occurs.
3. **Authentication Rate Limit**:
   - The auth routes (`/api/v1/auth/guest`, `/api/v1/auth/login`, `/api/v1/auth/register`) enforce a 30 requests/min per IP rate limiter to protect against credential stuffing.
