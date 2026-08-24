import { describe, it, expect, vi } from 'vitest'
import { render, screen } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { XOArenaPage } from './XOArenaPage'
import { HandCricketArenaPage } from './HandCricketArenaPage'
import { DotsBoxesArenaPage } from './DotsBoxesArenaPage'
import { Connect4ArenaPage } from './Connect4ArenaPage'
import { PaperFootballArenaPage } from './PaperFootballArenaPage'
import { NPATArenaPage } from './NPATArenaPage'
import { SpectatorArenaPage } from './SpectatorArenaPage'

vi.mock('../hooks/useWebSocket', () => ({
  useWebSocket: () => ({
    status: 'OPEN',
    currentRoom: 'room_test_123',
    members: [
      { user_id: 'usr_1', username: 'ClassCaptain', is_ready: true },
      { user_id: 'usr_2', username: 'Backbencher', is_ready: true },
    ],
    spectatorCount: 3,
    events: [],
    isReady: true,
    joinRoom: vi.fn(),
    leaveRoom: vi.fn(),
    sendMove: vi.fn(),
    setPlayerReady: vi.fn(),
    startGame: vi.fn(),
    sendChatMessage: vi.fn(),
  }),
}))

const queryClient = new QueryClient()

describe('Game Arena & Spectator Pages', () => {
  const renderArena = (component: React.ReactNode, path = '/games/xo/room_test_123', routePattern = '/games/xo/:roomId') => {
    return render(
      <QueryClientProvider client={queryClient}>
        <MemoryRouter initialEntries={[path]}>
          <Routes>
            <Route path={routePattern} element={component} />
          </Routes>
        </MemoryRouter>
      </QueryClientProvider>
    )
  }

  it('renders XO Arena with desk visual elements', () => {
    renderArena(<XOArenaPage />, '/games/xo/room_test_123', '/games/xo/:roomId')
    expect(screen.getAllByText(/XO/i).length).toBeGreaterThan(0)
  })

  it('renders Hand Cricket Arena with title', () => {
    renderArena(<HandCricketArenaPage />, '/games/hand-cricket/room_test_123', '/games/hand-cricket/:roomId')
    expect(screen.getAllByText(/Hand Cricket/i).length).toBeGreaterThan(0)
  })

  it('renders Dots & Boxes Arena with title', () => {
    renderArena(<DotsBoxesArenaPage />, '/games/dots-and-boxes/room_test_123', '/games/dots-and-boxes/:roomId')
    expect(screen.getAllByText(/Dots & Boxes/i).length).toBeGreaterThan(0)
  })

  it('renders Connect 4 Arena with title', () => {
    renderArena(<Connect4ArenaPage />, '/games/connect-4/room_test_123', '/games/connect-4/:roomId')
    expect(screen.getAllByText(/Connect 4/i).length).toBeGreaterThan(0)
  })

  it('renders Paper Football Arena with title', () => {
    renderArena(<PaperFootballArenaPage />, '/games/paper-football/room_test_123', '/games/paper-football/:roomId')
    expect(screen.getAllByText(/Paper Football/i).length).toBeGreaterThan(0)
  })

  it('renders NPAT Arena with title', () => {
    renderArena(<NPATArenaPage />, '/games/npat/room_test_123', '/games/npat/:roomId')
    expect(screen.getAllByText(/Name–Place–Animal–Thing/i).length).toBeGreaterThan(0)
  })

  it('renders Spectator Arena with sideline badge', () => {
    renderArena(<SpectatorArenaPage />, '/spectate/room_test_123', '/spectate/:roomId')
    expect(screen.getByText(/SIDE-BENCH LIVE/i)).toBeDefined()
    expect(screen.getByText(/Leave Sideline/i)).toBeDefined()
  })
})
