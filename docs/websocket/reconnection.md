# WebSocket Connection Recovery & Session Resumption Protocol

## 1. Overview
Recess features an authoritative, sequence-tracked connection recovery protocol that allows players experiencing network drops, browser refreshes, or socket disconnects to resume active multiplayer matches without losing match state.

---

## 2. Reconnection Lifecycle Flow

```
   Client                                                                  Server (Hub)
     │                                                                           │
     │ ──── (1) Match in Progress (Seq 14) ────────────────────────────────────► │
     │                                                                           │
     │ ✖ ── (2) Socket Disconnected (TCP Drop / Refresh) ──────────────────────► │
     │                                                                           │
     │                                     ┌───────────────────────────────────┐ │
     │                                     │  • Preserves Game State in Engine │ │
     │                                     │  • Registers DisconnectedSession  │ │
     │                                     │  • Starts 30s Grace Timer         │ │
     │                                     │  • Broadcasts player.reconnecting │ │
     │                                     └───────────────────────────────────┘ │
     │                                                                           │
     │ ──── (3) Re-establish WebSocket Connection ─────────────────────────────► │
     │ ◄─── (4) room.state (Current Members) ─────────────────────────────────── │
     │                                                                           │
     │ ──── (5) session.reconnect { session_id, room_id, last_sequence: 14 } ──► │
     │                                                                           │
     │                                     ┌───────────────────────────────────┐ │
     │                                     │  • Cancels Grace Period Timer     │ │
     │                                     │  • Restores Player to Room        │ │
     │                                     │  • Collects Missed Events (>14)   │ │
     │                                     │  • Generates State Snapshot       │ │
     │                                     └───────────────────────────────────┘ │
     │                                                                           │
     │ ◄─── (6) session.reconnected { current_sequence, missed_events, state } ─ │
     │ ◄─── (7) Broadcast player.reconnected to Opponent ─────────────────────── │
     │                                                                           │
     │ ──── (8) Player Resumes Play Immediately! ──────────────────────────────► │
```

---

## 3. Protocol Envelope Formats

### 3.1. Client Reconnection Request (`session.reconnect`)
Sent by the reconnecting client upon establishing a fresh socket:
```json
{
  "type": "session.reconnect",
  "room_id": "RECESS-HC-DESK",
  "sequence": 0,
  "timestamp": 1724500000000,
  "payload": {
    "session_id": "sess_usr_alice_8f92a",
    "room_id": "RECESS-HC-DESK",
    "last_sequence": 14
  }
}
```

### 3.2. Server Resumption Confirmation (`session.reconnected`)
Delivered exclusively to the resuming client with full snapshot and missed event queue:
```json
{
  "type": "session.reconnected",
  "room_id": "RECESS-HC-DESK",
  "sequence": 18,
  "timestamp": 1724500000500,
  "payload": {
    "session_id": "sess_usr_alice_8f92a",
    "room_id": "RECESS-HC-DESK",
    "current_sequence": 18,
    "missed_events": [
      {
        "type": "game.move",
        "room_id": "RECESS-HC-DESK",
        "sequence": 15,
        "payload": { "action": "choose_number", "value": 4 }
      }
    ],
    "game_state": {
      "game_id": "RECESS-HC-DESK",
      "game_type": "hand_cricket",
      "status": "active",
      "current_turn": "usr_alice",
      "move_count": 8,
      "board_state": { "current_inning": 2, "target": 45, "runs": 32 }
    },
    "members": [
      { "user_id": "usr_alice", "username": "Alice", "is_ready": true },
      { "user_id": "usr_bob", "username": "Bob", "is_ready": true }
    ]
  }
}
```

### 3.3. Opponent Notifications
- **`player.reconnecting`**: `{ "user_id": "...", "username": "...", "grace_period_seconds": 30 }`
- **`player.reconnected`**: `{ "user_id": "...", "username": "..." }`

---

## 4. Key Guarantees & Edge Cases Handled

1. **Deterministic State Preservation**: Engine instances are decoupled from transport sockets. A network drop never resets or mutates in-memory game state.
2. **30-Second Grace Period**: Prevents instant forfeits or phantom room closures during temporary mobile/Wi-Fi transitions.
3. **Missed Event Ring Buffer**: Rooms maintain a 100-event ring buffer so any events executed during disconnect are cleanly replayed to the client.
4. **Duplicate Sequence Deduplication**: Client and server sequence counters reject or ignore stale retransmitted envelopes.
5. **Heartbeat / Ping-Pong**: Ping frames sent every 25 seconds detect dead sockets before the OS TCP timeout.
