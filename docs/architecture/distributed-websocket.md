# Horizontally Scalable WebSocket Architecture

## 1. Overview
Recess employs a horizontally scalable, multi-instance Go backend architecture where client WebSocket connections are terminated across any number of stateless game servers behind a standard Layer-4 or Layer-7 Load Balancer without requiring sticky sessions.

---

## 2. System Architecture

```
                            ┌────────────────────────────────────────┐
                            │             Load Balancer              │
                            │   (Round-Robin / Least Connections)    │
                            └───────────────────┬────────────────────┘
                                                │
                 ┌──────────────────────────────┼──────────────────────────────┐
                 ▼                              ▼                              ▼
     ┌───────────────────────┐      ┌───────────────────────┐      ┌───────────────────────┐
     │     Game Server A     │      │     Game Server B     │      │     Game Server C     │
     │  (Instance: srv-a-1)  │      │  (Instance: srv-b-1)  │      │  (Instance: srv-c-1)  │
     │                       │      │                       │      │                       │
     │  • Local Clients:     │      │  • Local Clients:     │      │  • Local Clients:     │
     │    [Player A (Alice)] │      │    [Player B (Bob)]   │      │    [Spectator C]      │
     │  • Active Engine      │      │  • Active Engine      │      │  • Active Engine      │
     │  • Room Subscription  │      │  • Room Subscription  │      │  • Room Subscription  │
     └───────────┬───────────┘      └───────────┬───────────┘      └───────────┬───────────┘
                 │                              │                              │
                 │     pubsub:room:{id}         │     pubsub:room:{id}         │
                 └──────────────────────────────┼──────────────────────────────┘
                                                │
                                                ▼
                               ┌─────────────────────────────────┐
                               │           Redis 7.0             │
                               │  • Pub/Sub Room Broadcasts      │
                               │  • Game State Snapshots         │
                               │  • Matchmaking Queues           │
                               │  • Presences & Leaderboards     │
                               └─────────────────────────────────┘
```

---

## 3. Core Mechanisms

### 3.1. Instance-Aware Connection Management
- Each running server process is assigned a unique `InstanceID` (e.g. `inst_{timestamp}_{pid}` or configurable `server-A`).
- When a client connects and joins room `RECESS-HC-DESK`, the local Hub instance creates a dedicated Redis subscription to `pubsub:room:RECESS-HC-DESK` if not already subscribed.

### 3.2. Redis Pub/Sub Event Routing (`DistributedEnvelope`)
When a local client sends a message or executes a game move:
1. The server executes move validation and state updates locally.
2. The envelope is dispatched directly to any local clients in that room.
3. The server wraps the event in a `DistributedEnvelope`:
   ```json
   {
     "origin_instance_id": "game-server-A",
     "room_id": "RECESS-HC-DESK",
     "envelope": {
       "type": "game.move",
       "room_id": "RECESS-HC-DESK",
       "sequence": 15,
       "payload": { "action": "choose_number", "value": 6 }
     }
   }
   ```
4. The envelope is published to Redis channel `pubsub:room:RECESS-HC-DESK`.
5. Other game servers (e.g. Server B) receive the message:
   - Compares `OriginInstanceID == h.instanceID` to prevent duplicate local echoing.
   - Updates local room sequence and event history ring buffer.
   - Forwards the event to Player B on Server B.

### 3.3. Cross-Instance Match Execution
- **Player A on Server A** vs **Player B on Server B**:
  - Both players receive synchronized `room.state`, `player.joined`, `player.ready`, and `game.state` events.
  - Game state snapshots are continuously cached in Redis key `game:{room_id}` for fast cross-instance recovery.
  - Chat notes (`room.message`) are relayed in real time across instances.

### 3.4. Safe Subscription Cleanup
- When the last client in a room on a given instance leaves or disconnects, that instance cleanly closes its Redis Pub/Sub subscription for that channel (`pubsub.Close()`) to eliminate memory leaks and dangling subscriptions.

---

## 4. Verification

The multi-instance architecture is verified end-to-end in `backend/internal/websocket/websocket_test.go`:
- `TestWebSocket_MultiInstance_DistributedMatch`: Spawns two independent HTTP game server instances (`Server A` and `Server B`) connected via Redis Pub/Sub. Player A connects to Server A and Player B connects to Server B, participating in the same match room with real-time bidirectional synchronization.
