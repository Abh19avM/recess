# ADR 0011: Redis Infrastructure, Key Conventions & Data Ownership

## Status
**Accepted**

## Context
Recess is a real-time multiplayer gaming platform requiring sub-millisecond presence tracking, live game state caching, sliding-window rate limiting, and cross-instance Pub/Sub messaging. 

A clear boundary is required between persistent relational data and ephemeral in-memory state to prevent data loss while maximizing throughput.

---

## 1. Key Naming Conventions

All Redis keys in Recess follow strict namespace conventions using colon (`:`) separators:

| Namespace | Format | Data Structure | Description |
| :--- | :--- | :--- | :--- |
| **Presence** | `online:user:{user_id}` | JSON / String | Stores current student online presence, avatar, active room, and last seen timestamp. |
| **Session / Token** | `session:jwt:{token_hash}` | JSON / String | Tracks active refresh tokens and fast revocation status on logout. |
| **Game State** | `game:{game_id}` | JSON / String | Authoritative serialized snapshot of in-progress match (`GameState`). |
| **Room State** | `room:{room_id}` | JSON / String | Ephemeral metadata for active classroom desks. |
| **Queue** | `queue:{game_type}` | List / ZSET | Matchmaking waiting queue tickets per game. |
| **Rate Limiting** | `rate:user:{user_id}` / `rate:ip:{ip}` | ZSET (Sliding Window) | Sliding timestamp records for request throttling. |
| **Leaderboard** | `leaderboard:{game_type}` | Sorted Set (ZSET) | High-performance ranking of `user_id` $\to$ rating score. |
| **Pub/Sub Room** | `pubsub:room:{room_id}` | Channel | Real-time broadcast channel for multi-instance room synchronization. |

---

## 2. TTL (Time-To-Live) Strategy

Every non-permanent key in Redis has an explicit expiration policy to prevent unbounded memory growth:

| Key Category | Default TTL | Refresh Mechanism |
| :--- | :--- | :--- |
| `online:user:{id}` | **90 seconds** | Extended automatically on every 30s WebSocket ping/heartbeat. |
| `game:{id}` | **2 hours** | Cleaned up on match conclusion; auto-expires if match is abandoned. |
| `room:{id}` | **24 hours** | Extended when room members interact; auto-cleared when empty. |
| `session:jwt:{hash}` | **7 days** | Matches refresh token lifetime; deleted immediately on logout. |
| `rate:ip:{ip}` | **2 minutes** | Retains only recent timestamps within the 1-minute window. |

---

## 3. Data Ownership: Redis vs. PostgreSQL

```
                ┌────────────────────────────────────────────────────────┐
                │                  Client Application                    │
                └──────────────────────────┬─────────────────────────────┘
                                           │
                    ┌──────────────────────┴──────────────────────┐
                    ▼                                             ▼
       ┌─────────────────────────┐                   ┌─────────────────────────┐
       │   Redis (Speed Layer)   │                   │ PostgreSQL (Truth Layer)│
       ├─────────────────────────┤                   ├─────────────────────────┤
       │ • Presence & Heartbeats │                   │ • User Accounts & Auth  │
       │ • Live Game Caches      │                   │ • Player Profiles       │
       │ • Matchmaking Queues    │                   │ • Permanent Match History│
       │ • Sliding Rate Limits   │                   │ • Historical Match Scores│
       │ • Leaderboard Rankings  │                   │ • Canonical ELO Ratings │
       │ • Pub/Sub Broadcasts    │                   │ • Room Audit Log        │
       └─────────────────────────┘                   └─────────────────────────┘
```

- **PostgreSQL is the single source of truth** for all durable domain data: users, hashed credentials, profiles, completed match outcomes, player statistics, and permanent ratings.
- **Redis is the high-speed operational layer**: used exclusively for ephemeral, high-frequency, or transient state. If Redis restarts or crashes, no user accounts or match records are lost.

---

## 4. Failure & Degraded Mode Behavior

Recess implements **graceful degradation** when Redis is unavailable:

1. **Connection Failure on Startup**:
   - The server boots in degraded mode with structured warning logs rather than crashing.
2. **Auth & Sessions**:
   - Falls back to in-memory token revocation cache (`InMemoryTokenStore`).
3. **Rate Limiting**:
   - Rate limiters **fail open** on Redis errors, preventing legitimate players from being locked out during cache blips.
4. **Presence**:
   - Presence falls back to active in-memory WebSocket Hub connections.
5. **Health Checks**:
   - `/health` remains `200 OK` (liveness).
   - `/ready` reports degraded readiness with diagnostic error detail until Redis recovers.
