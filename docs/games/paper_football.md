# Paper Football Game Rules & Protocol

## 1. Overview
Paper Football is a beloved classroom desk sport played with a folded triangular origami paper football. Two players face off on opposite ends of a wooden classroom desk.

## 2. Objective
Score points by flicking the paper football across the desk so that it stops with a portion hanging over the opponent's desk edge (**Touchdown: 6 Points**), followed by an extra point kick through the opponent's finger uprights (**1 Point**). First player to **21 Points** wins the match!

## 3. Server-Authoritative Mechanics
- **Desk Coordinates**: $0.0\%$ (own goal line) to $100.0\%$ (opponent edge).
- **Starting Position**: Possession starts at $20.0\%$ mark.
- **Downs**: 4 downs to score or advance.
- **Flick Mechanics**:
  - `power`: $1$ to $100$ percentage.
  - `angle`: $-45^\circ$ to $+45^\circ$ aim deviation.
  - Distance formula: $\Delta d = \text{power} \times \cos(\text{angle}) \times 0.75$.
- **Outcomes**:
  - **Touchdown (6 Pts)**: Ball lands within $90.0\% \le \text{pos} \le 100.0\%$. Triggers Extra Point phase.
  - **Table Fall (Turnover)**: Ball position exceeds $100.0\%$. Opponent takes over possession from their $20.0\%$ mark.
  - **Turnover on Downs**: Failing to reach Touchdown zone after 4 downs transfers possession to opponent from inverse desk position.
  - **Extra Point Kick (+1 Pt)**: Flicking with $\ge 35$ power and angle within $[-20^\circ, +20^\circ]$.

## 4. WebSocket Move Envelope
```json
{
  "type": "game.move",
  "room_id": "RECESS-PF-DESK",
  "sequence": 6,
  "payload": {
    "action": "flick",
    "power": 75,
    "angle": 0
  }
}
```
