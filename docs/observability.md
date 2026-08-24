# Recess Observability & Telemetry Guide

## 1. Overview
Recess implements production-grade observability combining:
- **Metrics**: Standardized Prometheus exposition on `/metrics`.
- **Distributed Tracing**: OpenTelemetry (`go.opentelemetry.io/otel`) spans for HTTP handlers and operations.
- **Structured Logging**: `log/slog` structured logging with automatic sensitive information sanitization.
- **Dashboards & Visualization**: Auto-provisioned Grafana with real-time platform telemetry.

---

## 2. Core Metrics Reference

| Metric Name | Type | Labels | Description |
| :--- | :--- | :--- | :--- |
| `recess_active_players` | Gauge | — | Total number of connected students / players. |
| `recess_active_matches` | Gauge | — | Total number of multiplayer game matches currently active. |
| `recess_websocket_connections` | Gauge | — | Total open WebSocket client connections across all rooms. |
| `recess_websocket_messages_total` | Counter | `type`, `direction` | Real-time velocity of WebSocket envelopes (inbound/outbound). |
| `recess_matchmaking_queue_size` | Gauge | `game_type` | Count of tickets currently waiting in matchmaking queue. |
| `recess_matchmaking_latency_seconds` | Histogram | `game_type` | Time taken to find an opponent and form a match. |
| `recess_http_request_duration_seconds` | Histogram | `method`, `path`, `status` | HTTP API request duration (p50, p95, p99). |
| `recess_redis_latency_seconds` | Histogram | `operation` | Latency of Redis caching, pub/sub, and presence operations. |
| `recess_database_latency_seconds` | Histogram | `query`, `table` | PostgreSQL query execution time. |
| `recess_errors_total` | Counter | `component`, `type` | Total error count by subsystem (e.g. `http_handler`, `websocket`). |
| `recess_game_duration_seconds` | Histogram | `game_type` | Time elapsed from match start to game completion. |

---

## 3. Sensitive Data Redaction & Security
All log entries pass through `RedactingHandler` before output. Any attribute matching credentials or authentication secrets is masked:
- `password`, `passcode` $\to$ `[REDACTED]`
- `token`, `access_token`, `refresh_token`, `jwt` $\to$ `[REDACTED]`
- `authorization`, `secret` $\to$ `[REDACTED]`

---

## 4. Running the Observability Stack Locally

### Option A: Docker Compose (All-in-One)
To launch PostgreSQL, Redis, Recess Backend, Prometheus, and Grafana:
```bash
docker compose up -d
```

Access services:
- **Recess API**: `http://localhost:8080`
- **Prometheus Metrics Scraper**: `http://localhost:9090`
- **Grafana Dashboard**: `http://localhost:3000` (User: `admin`, Password: `admin`)
  - The dashboard **"Recess Platform Overview"** is auto-provisioned at `http://localhost:3000/d/recess-overview`.

### Option B: Local Backend + Docker Monitoring
If running the Go backend locally on your host (`go run ./cmd/server`):
```bash
# Start Prometheus and Grafana
docker compose up -d prometheus grafana
```
Prometheus will scrape the local backend at `http://localhost:8080/metrics`.

---

## 5. Verification & Testing
Run automated metrics, tracing, and redactor tests:
```bash
cd backend
go test -v -count=1 ./internal/metrics
```
Verify the live metrics endpoint:
```bash
curl -s http://localhost:8080/metrics | grep recess_
```
