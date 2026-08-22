import { describe, it, expect } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useWebSocket } from './useWebSocket'

describe('useWebSocket hook', () => {
  it('initializes with default closed state and room info', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false, initialRoom: 'RECESS-TEST' }))

    expect(result.current.status).toBe('CLOSED')
    expect(result.current.currentRoom).toBe('RECESS-TEST')
    expect(result.current.members).toEqual([])
    expect(result.current.isReady).toBe(false)
  })

  it('exposes joinRoom, leaveRoom, and toggleReady methods', () => {
    const { result } = renderHook(() => useWebSocket({ autoConnect: false }))

    act(() => {
      result.current.joinRoom('RECESS-BENCH-9')
    })

    expect(result.current.currentRoom).toBe('RECESS-BENCH-9')

    act(() => {
      result.current.leaveRoom()
    })

    expect(result.current.currentRoom).toBe('')
    expect(result.current.members).toEqual([])
  })
})
