import { describe, it, expect, beforeEach } from 'vitest'
import { useAuthStore } from './authStore'

describe('authStore', () => {
  beforeEach(() => {
    useAuthStore.setState({
      user: null,
      tokens: null,
      profileStats: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,
    })
  })

  it('initializes with unauthenticated state', () => {
    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(false)
    expect(state.user).toBeNull()
    expect(state.tokens).toBeNull()
  })

  it('allows manually setting auth user and tokens in state', () => {
    const mockUser = {
      id: 'usr_test_123',
      username: 'ClassMonitor',
      email: 'monitor@recess.school',
      role: 'student' as const,
      is_guest: false,
      avatar_preset: 'avatar_1',
      rating: 1200,
      title: 'Class Monitor',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }
    const mockTokens = {
      access_token: 'valid.mock.jwt',
      refresh_token: 'valid.mock.refresh',
      expires_in: 3600,
      token_type: 'Bearer',
    }

    useAuthStore.setState({
      user: mockUser,
      tokens: mockTokens,
      isAuthenticated: true,
    })

    const state = useAuthStore.getState()
    expect(state.isAuthenticated).toBe(true)
    expect(state.user?.username).toBe('ClassMonitor')
    expect(state.tokens?.access_token).toBe('valid.mock.jwt')
  })

  it('clears state upon logout', async () => {
    const mockUser = {
      id: 'usr_guest',
      username: 'Guest_123',
      role: 'guest' as const,
      is_guest: true,
      avatar_preset: 'avatar_1',
      rating: 1000,
      title: 'Guest Student',
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
    }
    useAuthStore.setState({
      user: mockUser,
      isAuthenticated: true,
    })
    expect(useAuthStore.getState().isAuthenticated).toBe(true)

    await useAuthStore.getState().logout()
    expect(useAuthStore.getState().isAuthenticated).toBe(false)
    expect(useAuthStore.getState().user).toBeNull()
  })
})
