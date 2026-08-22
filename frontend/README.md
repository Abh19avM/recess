# Recess Frontend

The client web application for Recess, built with React 19, TypeScript, Vite, Tailwind CSS v4, Zustand, and TanStack React Query.

---

## Architectural Highlights

- **School Stationery Design System**: Built with custom Tailwind CSS utility layers simulating ruled notebooks, graph paper, chalkboard themes, push-pins, and rubber stamp badges.
- **Authoritative Real-Time WebSocket Hook**: Typed `useWebSocket` hook handling automatic reconnection, room presence, readiness state, and structured event dispatching.
- **State Management**: Zustand stores for persisted authentication (`useAuthStore`) and global toast notifications (`useToastStore`).
- **REST Client**: Type-safe HTTP client with automatic Authorization header attachment and standardized JSON error formatting.

---

## Directory Structure

```
src/
├── components/
│   ├── layout/       # AppLayout, Navbar, Footer
│   └── ui/           # Avatar, Badge, Button, Input, Modal, PaperCard, ToastContainer, States
├── hooks/            # useWebSocket, useAuthStore hooks
├── lib/              # api.ts (HTTP client), utils.ts
├── pages/            # LandingPage, LoginPage, RegisterPage, DashboardPage, GamesPage, ProfilePage, WebSocketTestPage
├── routes/           # ProtectedRoute barrier
└── store/            # authStore.ts, toastStore.ts
```

---

## Available Scripts

| Command | Description |
| :--- | :--- |
| `npm run dev` | Start the Vite local development server with HMR |
| `npm run build` | Run TypeScript compiler check (`tsc -b`) and bundle for production (`vite build`) |
| `npm run test` | Execute unit and component tests with Vitest |
| `npm run lint` | Run ESLint across the codebase |
| `npm run preview` | Locally preview the production build |

---

## Testing

Run unit and component tests:
```bash
npm run test
```
