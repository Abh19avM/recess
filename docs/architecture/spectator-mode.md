# Spectator Mode Architecture & Playground Sideline Protocol

## 1. Overview
Spectator Mode allows classmates and bystanders to watch public multiplayer desk matches in real time from the "Schoolyard Sideline" without influencing game outcomes, submitting moves, or disrupting player reconnection grace periods.

---

## 2. Capabilities & Constraints

### 2.1. Spectator Capabilities
- **Live Match Observation**: Receive initial `room.state` and authoritative `game.state` upon joining.
- **Real-Time Stream**: Receive all subsequent moves, round transitions, innings, and timer events.
- **Sideline Notes**: Send chat notes and cheers (`room.message`) to support players.
- **Zero-Friction Departure**: Leave the sideline at any time without triggering game forfeits or room closure.

### 2.2. Spectator Constraints & Authoritative Protection
- **Move Rejection**: Any move submission (`game.move`, `player.ready`, `game.start`, `game.rematch`) from a spectator connection is authoritatively rejected server-side with `SPECTATOR_CANNOT_MOVE`.
- **Capacity Limits**: A hard limit of **50 spectators** per desk match prevents resource exhaustion. Excess spectator join requests receive `SPECTATOR_CAPACITY_REACHED`.
- **Session Isolation**: Spectators do not consume player seats or trigger disconnection grace timers.

---

## 3. Protocol Flow & Event Envelopes

```
   Spectator (Charlie)                                                     Hub (Server)
           │                                                                    │
           │ ─── (1) room.join { role: "spectator" } ────────────────────────► │
           │                                                                    │
           │                                 ┌────────────────────────────────┐ │
           │                                 │ • Validates spectator capacity │ │
           │                                 │ • Registers in hub.spectators  │ │
           │                                 │ • Increments spectatorCount    │ │
           │                                 └────────────────────────────────┘ │
           │                                                                    │
           │ ◄── (2) room.state { members, spectator_count: 1 } ─────────────── │
           │ ◄── (3) game.state { status: "active", ... } ───────────────────── │
           │ ◄── (4) Broadcast spectator.joined { user_id, count: 1 } ───────── │
           │                                                                    │
           │ ─── (5) Attempt Move (game.move) ────────────────────────────────► │
           │ ◄── (6) room.error { code: "SPECTATOR_CANNOT_MOVE" } ───────────── │
           │                                                                    │
           │ ◄── (7) Live Game Event (game.state) from Active Player ─────────── │
           │                                                                    │
           │ ─── (8) room.leave ──────────────────────────────────────────────► │
           │ ◄── (9) Broadcast spectator.left { user_id, count: 0 } ─────────── │
```

---

## 4. Spectator UI Aesthetic

The spectator interface (`frontend/src/pages/SpectatorArenaPage.tsx` at `/spectate/:roomId`) embodies a nostalgic classroom sideline:
- **Pulsing Indicator**: `🔴 SIDE-BENCH LIVE` status banner.
- **Desk Head-to-Head Card**: Player 1 (Blue ink) vs Player 2 (Red ink) with live turn indicators.
- **Classmate Bystander Counter**: `👀 {count} Watching on Sideline`.
- **Sideline Notes & Cheers**: Chalkboard message feed for spectators to pass cheering notes.
- **Clean Leave Control**: Dedicated "Leave Sideline" button returning safely to the syllabus.
