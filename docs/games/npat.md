# Name–Place–Animal–Thing (NPAT) Game Rules & Protocol

## 1. Overview
Name–Place–Animal–Thing is a classic multiplayer word game played across notebooks during recess and study halls.

## 2. Objective
Score the highest cumulative points across 3 rounds by writing unique words starting with the round's designated alphabet letter for all four categories: **Name**, **Place**, **Animal**, and **Thing**.

## 3. Server-Authoritative Mechanics
- **Letter Selection**: A designated letter is chosen per round (e.g. Round 1: `S`, Round 2: `A`, Round 3: `M`).
- **Simultaneous Turns**: All players write answers concurrently.
- **The "STOP!" Call**: The first player to finish all 4 categories calls **STOP!**, locking in submissions for evaluation.
- **Category Scoring Breakdown**:
  - **Unique Word (10 Points)**: Word starts with the target letter and no other player wrote the same word.
  - **Shared Word (5 Points)**: Word is valid but matches another classmate's answer.
  - **Invalid / Blank / Wrong Letter (0 Points)**: Empty, less than 2 characters, or does not start with the target letter.
  - **Max Round Score**: 40 Points (4 categories $\times$ 10 Points).
- **Match Conclusion**: Highest total accumulated points after 3 rounds is crowned the classroom vocabulary champion.

## 4. WebSocket Move Envelope
```json
{
  "type": "game.move",
  "room_id": "RECESS-NPAT-DESK",
  "sequence": 4,
  "payload": {
    "action": "call_stop",
    "name": "Sarah",
    "place": "Seattle",
    "animal": "Snake",
    "thing": "Scissors"
  }
}
```
