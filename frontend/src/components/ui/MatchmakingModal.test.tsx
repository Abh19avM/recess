import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { MatchmakingModal } from './MatchmakingModal'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'

const queryClient = new QueryClient()

describe('MatchmakingModal', () => {
  const defaultProps = {
    isOpen: true,
    onClose: vi.fn(),
    gameId: 'hand_cricket',
    gameTitle: 'Hand Cricket',
    gameIcon: '🏏',
  }

  const renderWithProviders = (component: React.ReactNode) => {
    return render(
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>{component}</BrowserRouter>
      </QueryClientProvider>
    )
  }

  it('renders game title and play mode options', () => {
    renderWithProviders(<MatchmakingModal {...defaultProps} />)
    expect(screen.getByText(/Classroom Matchmaking: Hand Cricket/i)).toBeDefined()
    expect(screen.getByText(/Casual Match/i)).toBeDefined()
    expect(screen.getByText(/Ranked Duel/i)).toBeDefined()
  })

  it('allows toggling between casual and ranked modes', () => {
    renderWithProviders(<MatchmakingModal {...defaultProps} />)
    const rankedTab = screen.getByText(/Ranked Duel/i)
    fireEvent.click(rankedTab)
    expect(screen.getByText(/Climb the classroom leaderboard/i)).toBeDefined()
  })

  it('renders search for opponent button', () => {
    renderWithProviders(<MatchmakingModal {...defaultProps} />)
    expect(screen.getByText(/Search for Opponent/i)).toBeDefined()
  })
})
