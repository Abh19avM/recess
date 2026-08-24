# Connect 4 Game Rules & Protocol

## 1. Overview
Connect 4 is a classic two-player strategy game played on a vertical $6 \times 7$ grid (6 rows, 7 columns). Players take turns dropping colored discs into any of the 7 columns. Discs fall to the lowest unoccupied cell in that column via simulated gravity.

## 2. Objective
Be the first player to form a continuous straight line of **4 matching discs** horizontally, vertically, or diagonally.

## 3. Server-Authoritative Mechanics
- **Grid Dimensions**: 6 Rows $\times$ 7 Columns (42 total slots).
- **Player Colors**: Player 1 = Navy Ink Disc (`#0284C7`), Player 2 = Red Ink Disc (`#DC2626`).
- **Turn Alternation**: Player 1 drops first, followed by Player 2 until a terminal state is reached.
- **Move Validation**:
  - Valid column index $0 \le \text{col} < 7$.
  - Column must not be full (row index 0 of column must be unoccupied).
  - Move must originate from the player whose turn it is.
- **Victory Condition**: 4 consecutive discs in any of the 4 vectors:
  1. Horizontal (—)
  2. Vertical (|)
  3. Diagonal Down-Right (\)
  4. Diagonal Down-Left (/)
- **Draw Condition**: All 42 slots filled without a 4-disc connection.

## 4. WebSocket Move Envelope
```json
{
  "type": "game.move",
  "room_id": "RECESS-C4-DESK",
  "sequence": 12,
  "payload": {
    "col": 3
  }
}
```
