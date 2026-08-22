# ADR 0007: Generic Game Engine Abstraction & Registry

## Status
**Accepted**

## Context
Recess is a real-time multiplayer gaming platform supporting 6 diverse school-time games:
1. **Hand Cricket** (Turn-based simultaneous choice reveals with chase/innings mechanics)
2. **Dots & Boxes** (Graph grid line claims with bonus turn chains)
3. **XO / Tic-Tac-Toe** (Grid coordinate placement with fast turn timers)
4. **Connect 4** (Gravity-dropped column tokens)
5. **Paper Football** (Desk physics vector drag-and-flick)
6. **Name–Place–Animal–Thing (NPAT)** (Multiplayer alphabet buzzer and dictionary validation)

A naive approach would tightly couple WebSocket event handlers or room managers with game-specific rules (e.g. handling cricket wickets inside the socket message reader). This leads to untestable monolithic handlers, fragile state transitions, and inability to simulate AI bots or run headless test scenarios.

## Decision
We establish a clean, transport-agnostic, server-authoritative `Engine` interface and a `Registry` under `backend/internal/games/engine`:

```
Game
├── Initialize(gameID, players, config) -> (*GameState, error)
├── ValidateMove(move) -> error
├── ApplyMove(move) -> (*GameState, error)
├── State() -> *GameState
├── IsFinished() -> bool
├── Result() -> *GameResult
├── NextTurn() -> string
└── GameType() -> GameType
```

### Core Tenets
1. **Separation of Transport from Rules**:
   - The WebSocket infrastructure (`internal/websocket`) and Room management (`internal/rooms`) only orchestrate connections, player presence, and event delivery.
   - All rules, move legality, state transitions, win conditions, and scoring are encapsulated strictly inside pure Go `Engine` implementations.

2. **Deterministic State Evolution**:
   - Replaying an identical ordered sequence of `Move` structs against an `Engine` will always produce the identical `GameState`, `BoardState`, and `Result`.

3. **Sequential State Versioning**:
   - Every `GameState` snapshot carries a monotonically increasing `Version` integer. This enables clients to detect missed packets, reconcile optimistic local updates, and re-sync on reconnects.

4. **Dynamic Extensibility via Registry**:
   - Engines register their factory constructors with `engine.Registry` (e.g., `engine.MustRegister(engine.GameTypeXO, NewXOEngine)`).
   - Rooms and Matchmakers can dynamically instantiate any supported game engine without compile-time hardcoding.

## Supported Game Types
- `xo` (XO / Tic-Tac-Toe) — Implemented as reference engine.
- `hand_cricket` (Hand Cricket)
- `dots_boxes` (Dots & Boxes)
- `connect4` (Connect 4)
- `paper_football` (Paper Football)
- `npat` (Name–Place–Animal–Thing)

## Consequences
- **Positive**:
  - Pure unit testing of complex school game mechanics without spinning up HTTP servers or WebSockets.
  - Zero risk of client-side game state tampering (all move validation is authoritative on the server).
  - Trivial integration with solo AI bots (the bot simply calls `Engine.ApplyMove()`).
- **Considerations**:
  - Game-specific board states are serialized as `json.RawMessage` within `GameState.BoardState`, allowing clients to deserialize domain-specific UI states without breaking the uniform envelope structure.
