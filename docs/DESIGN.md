# Recess — Design System & UI/UX Architecture

> **"The digital recreation of the golden age of school-time games."**  
> *Recess* is crafted to evoke the authentic, tactile nostalgia of classroom desks, ruled notebooks, graph paper, chalkboard scoreboards, and pencil scribbles — built as a modern, high-performance, responsive web application.

---

## 1. Design Philosophy & Aesthetic Identity

### 1.1 Core Aesthetic
Recess marries the warmth and tactile imperfections of **vintage school stationery** with the crisp ergonomics, accessibility, and speed of a modern web application.

- **Nostalgic, Not Childish**: The interface reflects genuine school memories (ballpoint pens, red teacher ink, graph notebooks, blackboard chalk dust, stamped library cards) rather than cartoonish or childish clip art.
- **Physicality & Tactility**: Elements possess subtle physical qualities — paper grain textures, slight rotations on stamps/stickers, subtle pencil sketch borders, ruled margin lines, and soft dropshadows reminiscent of layered paper sheets.
- **Intentional Restraint**: No generic purple "AI SaaS" gradients, no excessive glassmorphism, no random emoji spam, and no floating rounded mega-cards. Every component is grounded in schoolroom materials.

```
┌────────────────────────────────────────────────────────────────────────┐
│                        RECESS VISUAL ECOSYSTEM                         │
├───────────────────┬───────────────────┬────────────────────────────────┤
│  Stationery & Ink │  Classroom Fixtures│   Game Surfaces                │
├───────────────────┼───────────────────┼────────────────────────────────┤
│ • Ruled Notebook  │ • Notice Board    │ • Ruled Pitch (Hand Cricket)   │
│ • Graph Paper     │ • Chalkboard      │ • Graph Grid (Dots & Boxes)    │
│ • Rubber Stamps   │ • Wooden Desk     │ • Chalk Margin (XO)            │
│ • Red Margin Line │ • Report Card     │ • Wooden Frame (Connect 4)     │
│ • Graphite Pencil │ • Wall Clock      │ • Desk Oak Wood (Paper FB)     │
│ • Ballpoint Ink   │ • Push Pins/Tape  │ • Exam Sheet (NPAT)            │
└───────────────────┴───────────────────┴────────────────────────────────┘
```

---

## 2. Color Palette & Token Architecture

The color system is defined via CSS Custom Properties and integrated seamlessly into Tailwind CSS. It is split into **Materials**, **Inks**, **Accents**, and **Chalkboard** tiers.

### 2.1 The Material Tones (Backgrounds & Paper)
| Token | Hex | Name | Usage |
| :--- | :--- | :--- | :--- |
| `--color-paper-base` | `#FBF9F3` | Notebook Cream | Default page & light-theme background |
| `--color-paper-card` | `#FFFFFF` | Fresh Sheet White | Cards, sheets, dialog surfaces |
| `--color-paper-muted` | `#F2EDE0` | Aged Manila | Secondary cards, inactive tabs, disabled states |
| `--color-desk-wood` | `#4A3525` | Oak Classroom Desk | Outer frame trim, game board borders |
| `--color-desk-wood-light` | `#6E5039` | Polished Teak | Subtle highlights on wooden components |
| `--color-cork` | `#D8B17A` | Bulletin Board Cork | Notice board & tournament boards |

### 2.2 The Inks & Graphite (Text & Typography)
| Token | Hex | Name | Usage |
| :--- | :--- | :--- | :--- |
| `--color-ink-black` | `#1E242B` | Fountain Pen Charcoal | Primary body text, main headers, high-contrast UI |
| `--color-ink-blue` | `#1A365D` | Classic Ballpoint Blue | Links, primary buttons, Player 1 indicators |
| `--color-ink-red` | `#991B1B` | Teacher's Red Ink | Notebook margins, error alerts, Player 2 indicators |
| `--color-pencil-lead` | `#475569` | Graphite 2B | Secondary text, meta info, grid lines, disabled text |
| `--color-pencil-light` | `#94A3B8` | Light Graphite H | Subtle dividers, inactive borders, placeholder text |

### 2.3 The Ruled Lines & Grid Accent
| Token | Hex | Name | Usage |
| :--- | :--- | :--- | :--- |
| `--color-line-blue` | `#D9E4F0` | Ruled Line Blue | Horizontal notebook lines (`rgba(26, 54, 93, 0.12)`) |
| `--color-line-graph` | `#E2E8F0` | Graph Grid Line | Square grids for Dots & Boxes / XO |
| `--color-line-margin`| `#FCA5A5` | Left Margin Line | Vertical notebook margin guide (`#DC2626` at 30% alpha) |

### 2.4 Stamped Accents & Status Tones
| Token | Hex | Name | Usage |
| :--- | :--- | :--- | :--- |
| `--color-stamp-red` | `#B91C1C` | Detention / Out Stamp | Wicket alerts, defeats, warning badges, "OUT!" |
| `--color-stamp-green`| `#15803D` | Approved / A+ Stamp | Victory stamps, success notices, high scores |
| `--color-stamp-blue` | `#1D4ED8` | Principal's Seal | Official badges, tournament qualifiers, leaderboard top |
| `--color-stamp-amber`| `#B45309` | Brass Trophy Gold | Streak counters, gold trophies, coin balances |

### 2.5 Chalkboard Mode (Scoreboards & Night Games)
| Token | Hex | Name | Usage |
| :--- | :--- | :--- | :--- |
| `--color-chalkboard` | `#18231C` | Deep Slate Green | Chalkboard backgrounds, scoreboard panels |
| `--color-chalk-white`| `#F8FAFC` | Chalk Dust White | Chalkboard text, drawn lines, X & O strokes |
| `--color-chalk-yellow`| `#FEF08A` | Yellow Chalk | Chalkboard highlights, active turn glow |
| `--color-chalk-blue` | `#93C5FD` | Blue Chalk | Player 1 chalk scores & commentary |
| `--color-chalk-pink` | `#FCA5A5` | Pink Chalk | Player 2 chalk scores |

---

## 3. Typography & Font Hierarchy

A strict 3-tier font system ensures optimal legibility for core UI while injecting nostalgic handwritten character at appropriate accents.

```
┌────────────────────────────────────────────────────────────────────────┐
│                        TYPOGRAPHY SYSTEM                               │
├───────────────────┬──────────────────────────────┬─────────────────────┤
│ Purpose           │ Font Family                  │ Application         │
├───────────────────┼──────────────────────────────┼─────────────────────┤
│ 1. Primary UI     │ Plus Jakarta Sans / Inter    │ Buttons, Modals,    │
│    (Clean & Crisp)│ (Sans-Serif)                 │ Tables, Settings    │
├───────────────────┼──────────────────────────────┼─────────────────────┤
│ 2. Handwritten    │ Patrick Hand / Caveat        │ Notes, Doodles,     │
│    Accents        │ (Handwritten Script)         │ Stamped Labels, Chat│
├───────────────────┼──────────────────────────────┼─────────────────────┤
│ 3. Numerals & Tech│ JetBrains Mono               │ Timers, Elo, Scores,│
│    (Monospaced)   │ (Monospace)                  │ Room Codes, Coordinates │
└───────────────────┴──────────────────────────────┴─────────────────────┘
```

### 3.1 Type Scale & Rules
- **Display 1** (`text-4xl sm:text-5xl font-black tracking-tight`): Game Arena Titles, Victory Heads.
- **Header 1** (`text-2xl sm:text-3xl font-bold text-ink-black`): Page Titles, Notice Board Headers.
- **Header 2** (`text-xl font-semibold text-ink-black`): Section Headers, Card Titles.
- **Handwritten Callout** (`font-hand text-xl sm:text-2xl text-ink-blue`): Teacher notes, sticky notes, annotations.
- **Body Regular** (`text-sm sm:text-base text-ink-black leading-relaxed font-sans`): General reading, descriptions, rules.
- **Body Muted** (`text-xs sm:text-sm text-pencil-lead font-sans`): Meta tags, timestamps, subtitles.
- **Code / Mono** (`font-mono text-sm sm:text-base font-semibold tracking-wider`): Room codes (`RECESS-7X9P`), clock countdowns (`00:14.8`), score counts (`142/3`).

---

## 4. UI Components & Design Patterns

### 4.1 Paper Surfaces & Cards

#### A. Ruled Notebook Card
- Cream paper background (`#FBF9F3`) with faint blue horizontal ruled lines (spaced `28px` vertically).
- Vertical red margin line positioned `48px` from the left edge.
- Subtle drop shadow: `0 4px 6px -1px rgba(30, 36, 43, 0.08), 0 2px 4px -2px rgba(30, 36, 43, 0.05)`.
- Spiral binding punched holes or top tear perforations on active tabs.

#### B. Graph Paper Sheet (Dots & Boxes / Strategy)
- Fine `20px x 20px` grid in soft grey-blue (`#E2E8F0`).
- Crisp 1px graphite border with hand-drawn corner accents.

#### C. Notice Board Pinned Slip
- Manila paper or pastel sticky note tint (Cream, Pastel Yellow `#FEF9C3`, Pastel Blue `#E0F2FE`).
- Realistic 3D metallic push-pin at top-center with shadow, or a piece of semi-translucent masking tape.
- Slight random tilt (`-1.5deg` to `+1.5deg`) to break digital rigidity.

#### D. Classroom Chalkboard
- Deep slate green background (`#18231C`) with subtle chalk dust texture overlay.
- Beveled natural oak wooden frame (`border-4 border-[#5c4033] shadow-inner`).
- Chalk text rendered with soft outer blur (`text-shadow: 0 0 2px rgba(255,255,255,0.4)`).

---

### 4.2 Buttons & Interactive Elements

```
┌───────────────────────────┐   ┌───────────────────────────┐   ┌───────────────────────────┐
│   [ STAMPED INK BTN ]     │   │   [ PENCIL SKETCH BTN ]   │   │   [ NOTEBOOK TAB BTN ]    │
│   Solid ink fill with     │   │   1.5px graphite border,  │   │   Tab with folded top,    │
│   slight ink bleed shadow │   │   paper hover fill, press │   │   active ruled underline  │
└───────────────────────────┘   └───────────────────────────┘   └───────────────────────────┘
```

1. **Primary Action (Stamped Ballpoint Button)**:
   - Background: Deep Ink Blue (`#1A365D`) or Teacher Red (`#991B1B`).
   - Text: White, crisp font-bold.
   - Border: 2px solid matching ink with subtle `rounded-md` (not overly rounded pill).
   - Interaction: Slight translation on click (`active:translate-y-0.5`), sharp tactile feedback.

2. **Secondary Action (Graphite Sketch Button)**:
   - Background: Paper Card White (`#FFFFFF`).
   - Border: 1.5px solid `pencil-lead` (`#475569`).
   - Hover: Background transitions to `paper-muted` (`#F2EDE0`), border darkens to `ink-black`.

3. **Rubber Stamp Badges & Tags**:
   - Skewed `-3deg` to `+4deg` rotation.
   - Red or Green double-bordered rectangle with stamped vintage font.
   - Example stamps: `[ PASSED ]`, `[ OUT! ]`, `[ TOUCHDOWN ]`, `[ WINNER ]`, `[ DETENTION ]`.

---

### 4.3 Navigation & Header Layout

- **Classroom Header**:
  - Left: **Recess Logo** — hand-lettered bold notebook stamp with mini pencil icon.
  - Center: **Notebook Tabs** — `Classroom Hub`, `Notice Board (Leaderboards)`, `Report Card (Profile)`, `School Handbook (Rules)`.
  - Right: **Player Desk Badge** — Avatar preset (Pencil sketch / Doodle avatar), Username, Attendance Streak (🔥 Days), Sound FX Toggle (Pencil/Chalk/Mute).
- **Responsive Navigation**:
  - Bottom notebook tab bar for mobile screens (< 768px).

---

## 5. Game-Specific UI & Board Layouts

### 5.1 Hand Cricket
- **Arena Concept**: School notebook cricket pitch drawn during 5th-period break.
- **Top Scoreboard**: Chalkboard scoreboard displaying:
  - Batting Player vs Bowling Player.
  - Current Score, Target, Wickets Remaining, Over count (`1.4 / 5.0`).
  - Required Run Rate (RRR) and Run Log wagon strip (`[4] [1] [6] [W] [2]`).
- **Interactive Hand Chooser**:
  - 6 large tactile gesture cards (Numbers 1 to 6) with clear finger illustrations and bold digits.
  - Toss Phase: Odd or Even selector with coin flip animation.
  - Simultaneous reveal card flip: Both choices slam down together with a stamped umpire call ("4 RUNS!" or red stamped "OUT!").

### 5.2 Dots & Boxes
- **Arena Concept**: Quad-ruled graph paper exercise sheet.
- **Grid Options**: 3x3 (Quick Recess), 4x4 (Standard Period), 5x5 (Championship).
- **Board Styling**:
  - Hand-drawn circular graphite dots at vertices.
  - Hovering over an open edge previews a light pencil dashed line.
  - Claimed lines draw as crisp ink lines (Blue for P1, Red for P2).
  - Completed 1x1 boxes smoothly fill with light crayon wash and the player's bold initial stamp with a "+1 BONUS TURN" banner.

### 5.3 XO (Tic-Tac-Toe)
- **Arena Concept**: Corner of a notebook margin.
- **Grid Modes**: Classic 3x3, Extended 4x4, Master 5x5.
- **Visuals**:
  - Drawn with authentic sketchy pencil cross-hatches.
  - X marks: Blue graphite cross with dynamic draw animation.
  - O marks: Red crayon loop with smooth SVG stroke-dashoffset draw.
  - Win condition: A solid colored ruler strike-through line cuts through the winning line.

### 5.4 Connect 4
- **Arena Concept**: Vintage wooden desk game fixture.
- **Board Styling**:
  - Vertical 7-column x 6-row polished oak frame (`border-8 border-desk-wood rounded-lg shadow-2xl`).
  - Open circular slots showing dark inner chamber.
  - Dropping discs: Physics-tuned spring drop animation with subtle wood click audio.
  - Discs: Bright Sunshine Yellow vs Ruby Red stamped tokens.
  - Win indicator: Neon chalk connecting halo around the 4 winning discs.

### 5.5 Paper Football
- **Arena Concept**: Varnished wooden school desk with taped yard lines.
- **Mechanics & Visuals**:
  - Desktop view with notebook paper field lines, 50-yard center tape, and endzones.
  - Football: Realistic folded triangular paper football with realistic shadow.
  - Drag-and-flick arrow indicator: Distance and angle mapped from player's touch/mouse drag.
  - Table Edge Overhang: Zoomed-in camera view when football approaches desk edge. If 1/3 hangs over: **TOUCHDOWN! (6 PTS)**.
  - Field Goal Kick: Upright finger goalposts appear; player flicks through the posts for extra point.

### 5.6 Name–Place–Animal–Thing (NPAT)
- **Arena Concept**: Formal classroom pop-quiz answer sheet.
- **Game Layout**:
  - Top: Round Letter Generator (e.g. Letter **"M"**) with a spinning classroom wall clock timer (45s).
  - 4 Ruled Answer Columns: **Name**, **Place**, **Animal**, **Thing** (plus optional 5th **School Item** category).
  - "STOP!" Call Button: First to finish slams the red "STOP!" buzzer, giving opponents an urgent 10-second countdown.
  - **Grading & Review Phase**:
    - Both sheets placed side-by-side on the desk.
    - Automated dictionary check with teacher's red ink tick (`✓ 10 pts`) or cross (`✗ 0 pts`).
    - Match duplicate check: Shared answers get half points (`~ 5 pts`).
    - Interactive dispute/peer vote button ("Raise Hand to Challenge").

---

## 6. Motion Language & Micro-Interactions

Animations in Recess are purposeful, grounded, and physically delightful.

| Interaction | Animation Style | Duration / Easing |
| :--- | :--- | :--- |
| **Page Navigation** | Notebook page turn / slide with faint paper shadow | `300ms cubic-bezier(0.2, 0.8, 0.2, 1)` |
| **Stamp Drop** | Slam down with `1.15x -> 1.0x` bounce and `-3deg` tilt | `250ms cubic-bezier(0.34, 1.56, 0.64, 1)` |
| **Pencil Line Draw** | SVG `stroke-dashoffset` path draw | `180ms ease-out` |
| **Connect 4 Chip Drop**| Vertical gravity fall with 2 micro bounces | `400ms cubic-bezier(0.25, 0.46, 0.45, 0.94)` |
| **Paper Football Slide**| Smooth deceleration friction physics | `600ms-1200ms ease-out` |
| **Wicket / OUT Stamp** | Screen shake (2px, 3 cycles) + Red ink splash | `350ms ease-in-out` |
| **Match Victory** | Falling pencil shavings & gold paper confetti | Canvas particle burst (2.5s) |

---

## 7. Accessibility, Responsiveness & Ergonomics

1. **Color Contrast Compliance**:
   - All text maintains WCAG AAA (7:1) on paper and chalkboard backgrounds.
   - Red margin and blue grid lines are strictly decorative and do not obscure text.
2. **Touch Targets**:
   - Minimum tap target size: `48px x 48px` for all buttons, hand cricket choice cards, and grid cells.
3. **Keyboard Accessibility**:
   - Full keyboard navigation for all menus (`Tab`, `Enter`, `Escape`).
   - Game keys: Hand cricket numbers `1-6`, Connect 4 columns `1-7`, Tic-Tac-Toe keypad `1-9` or arrow keys.
4. **Responsive Breakpoints**:
   - **Mobile (< 640px)**: Compact single-column arenas, bottom sheet menus, larger touch cards.
   - **Tablet (640px - 1024px)**: Dual-pane layout (Scoreboard + Board side-by-side).
   - **Desktop (> 1024px)**: Full desk recreation with chat sidebar, spectator stands, and match logs.

---

## 8. Summary Checklist for Implementation

- [x] Defined complete 4-tier color palette with CSS variables and Tailwind configuration.
- [x] Defined 3-tier typography hierarchy (UI, Handwritten, Monospace).
- [x] Defined paper textures, ruled margins, graph grids, and chalkboard styles.
- [x] Detailed UI board architecture for all 6 classic school games.
- [x] Established motion design guidelines and accessibility standards.
- [x] Ready for Phase 1 frontend theme and layout implementation.
