# Architecture Decision Record: Horizontally Scalable WebSocket Architecture

- **Status**: Accepted
- **Date**: 2026-08-24
- **Decision Makers**: Recess Engineering Team

---

## 1. Context and Problem Statement
Recess requires real-time bidirectional multiplayer gameplay across 6 games. As player concurrency grows, a single Go backend process will encounter socket and CPU bottlenecks. We need a horizontally scalable architecture where multiple Go server instances can concurrently host players participating in the same match room without requiring sticky sessions.

---

## 2. Decision: Redis Pub/Sub Distributed Routing

We chose **Stateless Multi-Instance Go Backends with Redis Pub/Sub Event Mesh** over Sticky Sessions or distributed actor frameworks:

1. **Decoupled Transport and Game State**:
   - WebSockets terminate on whichever server the Load Balancer routes the HTTP/WS upgrade request to.
   - Redis Pub/Sub (`pubsub:room:{room_id}`) acts as the event bus between backend instances.
2. **`DistributedEnvelope` Pattern**:
   - Envelopes carry `origin_instance_id` to prevent echo loops when broadcasting.
3. **Dynamic Subscription Lifecycle**:
   - Subscriptions are opened on-demand when the first local client joins a room and closed when the last local client departs.
4. **Resilience without Sticky Sessions**:
   - If Server A restarts or fails, Player A reconnects to Server B via the Load Balancer, issues `session.reconnect`, and resumes their match instantly without state loss.

---

## 3. Consequences

### Positive
- **Linear Horizontal Scaling**: Add or remove backend instances behind a standard Round-Robin load balancer at will.
- **Zero Sticky Session Lock-In**: Enables rolling deployments without dropping active games.
- **Cross-Instance Isolation**: Room channels are isolated by room ID (`pubsub:room:{id}`), avoiding global broadcast bottlenecks.

### Negative / Trade-offs
- Requires Redis as a shared dependency for multi-node deployments (with graceful fallback to in-memory mode when running single-node locally).
