import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { api } from '../lib/api'

export interface User {
  id: string
  username: string
  email?: string
  is_guest: boolean
  avatar_preset: string
  title: string
  rating: number
  created_at: string
  updated_at: string
}

export interface UserProfileResponse {
  user: User
  games_played: number
  games_won: number
  win_rate: number
  game_stats?: Array<{
    game_type: string
    rating: number
    played: number
    won: number
    lost: number
  }>
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  expires_in: number
  token_type: string
}

export interface AuthResponse {
  tokens: TokenPair
  user: User
}

interface AuthState {
  user: User | null
  tokens: TokenPair | null
  profileStats: UserProfileResponse | null
  isAuthenticated: boolean
  isLoading: boolean
  error: string | null

  login: (credentials: { username: string; password: string }) => Promise<void>
  register: (data: { username: string; email?: string; password: string }) => Promise<void>
  guestLogin: (nickname?: string) => Promise<void>
  logout: () => Promise<void>
  fetchMe: () => Promise<void>
  clearError: () => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      tokens: null,
      profileStats: null,
      isAuthenticated: false,
      isLoading: false,
      error: null,

      login: async (credentials) => {
        set({ isLoading: true, error: null })
        try {
          const resp = await api.post<AuthResponse>('/auth/login', credentials)
          set({
            user: resp.user,
            tokens: resp.tokens,
            isAuthenticated: true,
            isLoading: false,
          })
          // Fetch full profile stats in background
          get().fetchMe()
        } catch (err: any) {
          set({
            isLoading: false,
            error: err.message || 'Login failed. Please check your username and password.',
          })
          throw err
        }
      },

      register: async (data) => {
        set({ isLoading: true, error: null })
        try {
          const resp = await api.post<AuthResponse>('/auth/register', data)
          set({
            user: resp.user,
            tokens: resp.tokens,
            isAuthenticated: true,
            isLoading: false,
          })
          get().fetchMe()
        } catch (err: any) {
          set({
            isLoading: false,
            error: err.message || 'Registration failed. Please try a different handle.',
          })
          throw err
        }
      },

      guestLogin: async (nickname) => {
        set({ isLoading: true, error: null })
        try {
          const resp = await api.post<AuthResponse>('/auth/guest', {
            nickname: nickname || undefined,
            avatar_preset: 'pencil_sketch_guest',
          })
          set({
            user: resp.user,
            tokens: resp.tokens,
            isAuthenticated: true,
            isLoading: false,
          })
        } catch (err: any) {
          set({
            isLoading: false,
            error: err.message || 'Failed to create guest session.',
          })
          throw err
        }
      },

      logout: async () => {
        const { tokens } = get()
        if (tokens?.refresh_token) {
          try {
            await api.post('/auth/logout', { refresh_token: tokens.refresh_token })
          } catch {
            // Ignore logout API failures
          }
        }
        set({
          user: null,
          tokens: null,
          profileStats: null,
          isAuthenticated: false,
          error: null,
        })
      },

      fetchMe: async () => {
        if (!get().isAuthenticated) return
        try {
          const profile = await api.get<UserProfileResponse>('/users/me')
          set({
            profileStats: profile,
            user: profile.user || get().user,
          })
        } catch (err: any) {
          if (err.status === 401) {
            get().logout()
          }
        }
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'recess_auth',
      partialize: (state) => ({
        user: state.user,
        tokens: state.tokens,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
)
