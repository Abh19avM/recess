import { useState, useEffect, useRef, useCallback } from 'react'
import { useAuthStore } from '../store/authStore'
import { toast } from '../store/toastStore'

export type ConnectionStatus = 'CONNECTING' | 'OPEN' | 'CLOSING' | 'CLOSED'

export interface PlayerInfo {
  user_id: string
  username: string
  is_guest: boolean
  avatar_preset?: string
  is_ready: boolean
}

export interface EventEnvelope<T = any> {
  type: string
  room_id: string
  sequence?: number
  timestamp?: number
  payload?: T
}

export interface UseWebSocketOptions {
  autoConnect?: boolean
  initialRoom?: string
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const { autoConnect = true, initialRoom = '' } = options
  const { user, tokens } = useAuthStore()

  const [status, setStatus] = useState<ConnectionStatus>('CLOSED')
  const [currentRoom, setCurrentRoom] = useState<string>(initialRoom)
  const [members, setMembers] = useState<PlayerInfo[]>([])
  const [events, setEvents] = useState<EventEnvelope[]>([])
  const [isReady, setIsReady] = useState<boolean>(false)

  const socketRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<number | null>(null)
  const shouldReconnectRef = useRef<boolean>(true)

  const addEvent = useCallback((event: EventEnvelope) => {
    setEvents((prev) => [event, ...prev.slice(0, 49)]) // Keep last 50 events
  }, [])

  const connect = useCallback(() => {
    if (socketRef.current && (socketRef.current.readyState === WebSocket.OPEN || socketRef.current.readyState === WebSocket.CONNECTING)) {
      return
    }

    setStatus('CONNECTING')

    // Determine host and protocol
    const isSecure = window.location.protocol === 'https:'
    const protocol = isSecure ? 'wss:' : 'ws:'
    const host = window.location.host
    let wsUrl = `${protocol}//${host}/ws`

    // Add query params
    const params = new URLSearchParams()
    if (tokens?.access_token) {
      params.append('token', tokens.access_token)
    } else if (user) {
      params.append('user_id', user.id)
      params.append('username', user.username)
      if (user.avatar_preset) {
        params.append('avatar_preset', user.avatar_preset)
      }
    } else {
      params.append('guest_name', 'Guest_' + Math.random().toString(36).substring(2, 6))
    }

    if (currentRoom) {
      params.append('room', currentRoom)
    }

    const fullUrl = `${wsUrl}?${params.toString()}`

    try {
      const ws = new WebSocket(fullUrl)
      socketRef.current = ws

      ws.onopen = () => {
        setStatus('OPEN')
        addEvent({
          type: 'client.connected',
          room_id: currentRoom,
          timestamp: Date.now(),
          payload: { message: 'WebSocket connection opened' },
        })
      }

      ws.onmessage = (messageEvent) => {
        try {
          const envelope: EventEnvelope = JSON.parse(messageEvent.data)
          addEvent(envelope)

          switch (envelope.type) {
            case 'room.state': {
              const payload = envelope.payload as { room_id: string; members: PlayerInfo[] }
              setCurrentRoom(payload.room_id)
              setMembers(payload.members || [])
              const me = (payload.members || []).find((m) => m.user_id === user?.id || m.username === user?.username)
              if (me) {
                setIsReady(me.is_ready)
              }
              break
            }

            case 'player.joined': {
              const payload = envelope.payload as { player: PlayerInfo }
              if (payload?.player) {
                setMembers((prev) => {
                  const filtered = prev.filter((p) => p.user_id !== payload.player.user_id)
                  return [...filtered, payload.player]
                })
                toast.info('Classmate Arrived', `@${payload.player.username} joined the desk`)
              }
              break
            }

            case 'player.left': {
              const payload = envelope.payload as { user_id: string; username: string }
              if (payload) {
                setMembers((prev) => prev.filter((p) => p.user_id !== payload.user_id))
                toast.info('Classmate Departed', `@${payload.username} left the desk`)
              }
              break
            }

            case 'player.ready': {
              const payload = envelope.payload as { user_id: string; username: string; is_ready: boolean }
              if (payload) {
                setMembers((prev) =>
                  prev.map((p) =>
                    p.user_id === payload.user_id || p.username === payload.username
                      ? { ...p, is_ready: payload.is_ready }
                      : p
                  )
                )
                if (payload.user_id === user?.id || payload.username === user?.username) {
                  setIsReady(payload.is_ready)
                }
              }
              break
            }

            case 'player.disconnected': {
              const payload = envelope.payload as { user_id: string; username: string }
              if (payload) {
                setMembers((prev) => prev.filter((p) => p.user_id !== payload.user_id))
                toast.warning('Classmate Disconnected', `@${payload.username}'s connection dropped`)
              }
              break
            }

            case 'room.error': {
              const payload = envelope.payload as { code: string; message: string }
              toast.error(payload.code || 'Room Error', payload.message)
              break
            }

            case 'room.message': {
              const payload = envelope.payload as { sender: string; text: string }
              toast.info(`Note from @${payload.sender}`, payload.text)
              break
            }
          }
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err)
        }
      }

      ws.onclose = () => {
        setStatus('CLOSED')
        addEvent({
          type: 'client.disconnected',
          room_id: currentRoom,
          timestamp: Date.now(),
          payload: { message: 'WebSocket connection closed' },
        })

        if (shouldReconnectRef.current) {
          reconnectTimeoutRef.current = window.setTimeout(() => {
            connect()
          }, 3000)
        }
      }

      ws.onerror = (err) => {
        console.error('WebSocket encountered an error:', err)
        ws.close()
      }
    } catch (err) {
      console.error('Failed to instantiate WebSocket:', err)
      setStatus('CLOSED')
    }
  }, [user, tokens, currentRoom, addEvent])

  useEffect(() => {
    shouldReconnectRef.current = true
    if (autoConnect) {
      connect()
    }

    return () => {
      shouldReconnectRef.current = false
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      if (socketRef.current) {
        socketRef.current.close()
      }
    }
  }, [autoConnect, connect])

  const sendEvent = useCallback((type: string, payload?: any, roomIdOverride?: string) => {
    if (!socketRef.current || socketRef.current.readyState !== WebSocket.OPEN) {
      toast.error('Connection Offline', 'Cannot send event: WebSocket is not open')
      return false
    }

    const envelope: EventEnvelope = {
      type,
      room_id: roomIdOverride || currentRoom,
      timestamp: Date.now(),
      payload,
    }

    try {
      socketRef.current.send(JSON.stringify(envelope))
      return true
    } catch (err) {
      console.error('Failed to send WebSocket envelope:', err)
      return false
    }
  }, [currentRoom])

  const joinRoom = useCallback((roomId: string, passcode?: string) => {
    const clean = roomId.trim().toUpperCase()
    if (!clean) return
    setCurrentRoom(clean)
    sendEvent('room.join', { passcode }, clean)
  }, [sendEvent])

  const leaveRoom = useCallback(() => {
    if (!currentRoom) return
    sendEvent('room.leave', {}, currentRoom)
    setCurrentRoom('')
    setMembers([])
    setIsReady(false)
  }, [currentRoom, sendEvent])

  const toggleReady = useCallback((readyOverride?: boolean) => {
    const nextReady = readyOverride !== undefined ? readyOverride : !isReady
    sendEvent('player.ready', { is_ready: nextReady })
  }, [isReady, sendEvent])

  const sendMessage = useCallback((text: string) => {
    if (!text.trim()) return
    sendEvent('room.message', { text: text.trim() })
  }, [sendEvent])

  return {
    status,
    currentRoom,
    members,
    events,
    isReady,
    connect,
    joinRoom,
    leaveRoom,
    toggleReady,
    sendMessage,
    sendEvent,
  }
}
