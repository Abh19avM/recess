import { useState, useEffect, useRef, useCallback } from 'react'
import { useAuthStore } from '../store/authStore'
import { toast } from '../store/toastStore'

export type ConnectionStatus = 'CONNECTING' | 'OPEN' | 'RECONNECTING' | 'CLOSING' | 'CLOSED'

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
  role?: 'player' | 'spectator'
}

export function useWebSocket(options: UseWebSocketOptions = {}) {
  const { autoConnect = true, initialRoom = '', role = 'player' } = options
  const { user, tokens } = useAuthStore()

  const [status, setStatus] = useState<ConnectionStatus>('CLOSED')
  const [currentRoom, setCurrentRoom] = useState<string>(initialRoom)
  const [members, setMembers] = useState<PlayerInfo[]>([])
  const [spectatorCount, setSpectatorCount] = useState<number>(0)
  const [events, setEvents] = useState<EventEnvelope[]>([])
  const [isReady, setIsReady] = useState<boolean>(false)

  const socketRef = useRef<WebSocket | null>(null)
  const reconnectTimeoutRef = useRef<number | null>(null)
  const pingIntervalRef = useRef<number | null>(null)
  const shouldReconnectRef = useRef<boolean>(true)
  const lastSequenceRef = useRef<number>(0)
  const sessionIdRef = useRef<string>(
    'sess_' + (user?.id || 'guest') + '_' + Math.random().toString(36).substring(2, 9)
  )
  const reconnectAttemptsRef = useRef<number>(0)

  const addEvent = useCallback((event: EventEnvelope) => {
    if (event.sequence && event.sequence > lastSequenceRef.current) {
      lastSequenceRef.current = event.sequence
    }
    setEvents((prev) => [event, ...prev.slice(0, 49)]) // Keep last 50 events
  }, [])

  const connect = useCallback(() => {
    if (socketRef.current && (socketRef.current.readyState === WebSocket.OPEN || socketRef.current.readyState === WebSocket.CONNECTING)) {
      return
    }

    setStatus((prev) => (prev === 'CLOSED' ? 'CONNECTING' : 'RECONNECTING'))

    // Determine host and protocol
    const isSecure = window.location.protocol === 'https:'
    const protocol = isSecure ? 'wss:' : 'ws:'
    const isDev = window.location.port === '5173'
    const targetHost = isDev ? `${window.location.hostname}:8080` : window.location.host
    let wsUrl = import.meta.env.VITE_WS_URL || `${protocol}//${targetHost}/ws`

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

    if (role === 'spectator') {
      params.append('role', 'spectator')
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
        reconnectAttemptsRef.current = 0

        // If reconnecting to an active room, issue session.reconnect
        if (currentRoom && lastSequenceRef.current > 0) {
          const reconnEnv: EventEnvelope = {
            type: 'session.reconnect',
            room_id: currentRoom,
            timestamp: Date.now(),
            payload: {
              session_id: sessionIdRef.current,
              room_id: currentRoom,
              last_sequence: lastSequenceRef.current,
            },
          }
          ws.send(JSON.stringify(reconnEnv))
        }

        addEvent({
          type: 'client.connected',
          room_id: currentRoom,
          timestamp: Date.now(),
          payload: { message: 'WebSocket connection opened' },
        })

        // Setup ping heartbeat (every 25 seconds)
        if (pingIntervalRef.current) clearInterval(pingIntervalRef.current)
        pingIntervalRef.current = window.setInterval(() => {
          if (ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'room.ping', room_id: currentRoom, timestamp: Date.now() }))
          }
        }, 25000)
      }

      ws.onmessage = (messageEvent) => {
        try {
          const envelope: EventEnvelope = JSON.parse(messageEvent.data)
          addEvent(envelope)

          switch (envelope.type) {
            case 'session.reconnected': {
              const payload = envelope.payload as {
                session_id: string
                room_id: string
                current_sequence: number
                missed_events: EventEnvelope[]
                game_state?: any
                members: PlayerInfo[]
              }
              if (payload) {
                if (payload.members) setMembers(payload.members)
                if (payload.current_sequence) lastSequenceRef.current = payload.current_sequence

                // Deliver authoritative game state
                if (payload.game_state) {
                  addEvent({
                    type: 'game.state',
                    room_id: payload.room_id,
                    sequence: payload.current_sequence,
                    timestamp: Date.now(),
                    payload: payload.game_state,
                  })
                }

                // Replay missed events
                if (payload.missed_events?.length > 0) {
                  for (const missed of payload.missed_events) {
                    addEvent(missed)
                  }
                }

                toast.success('Connection Restored!', 'Rejoined match desk without losing game state.')
              }
              break
            }

            case 'player.reconnecting': {
              const payload = envelope.payload as { user_id: string; username: string; grace_period_seconds: number }
              if (payload) {
                toast.warning('Classmate Dropped', `@${payload.username} disconnected. Waiting ${payload.grace_period_seconds}s for reconnection...`)
              }
              break
            }

            case 'player.reconnected': {
              const payload = envelope.payload as { user_id: string; username: string }
              if (payload) {
                toast.success('Classmate Reconnected', `@${payload.username} resumed the match!`)
              }
              break
            }

            case 'spectator.joined': {
              const payload = envelope.payload as { user_id: string; username: string; count: number }
              if (payload) {
                setSpectatorCount(payload.count)
                if (role !== 'spectator') {
                  toast.info('Bystander on Sideline', `@${payload.username} joined to watch!`)
                }
              }
              break
            }

            case 'spectator.left': {
              const payload = envelope.payload as { user_id: string; username: string; count: number }
              if (payload) {
                setSpectatorCount(payload.count)
              }
              break
            }

            case 'spectator.count': {
              const payload = envelope.payload as { count: number }
              if (payload) {
                setSpectatorCount(payload.count)
              }
              break
            }

            case 'room.state': {
              const payload = envelope.payload as { room_id: string; members: PlayerInfo[]; spectator_count?: number }
              setCurrentRoom(payload.room_id)
              const rawMembers = payload.members || []
              const uniqueMap = new Map<string, PlayerInfo>()
              rawMembers.forEach((m) => {
                if (m && m.user_id) uniqueMap.set(m.user_id, m)
              })
              setMembers(Array.from(uniqueMap.values()))
              if (payload.spectator_count !== undefined) {
                setSpectatorCount(payload.spectator_count)
              }
              const me = rawMembers.find((m) => m.user_id === user?.id || m.username === user?.username)
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
        setStatus('RECONNECTING')
        if (pingIntervalRef.current) clearInterval(pingIntervalRef.current)

        addEvent({
          type: 'client.disconnected',
          room_id: currentRoom,
          timestamp: Date.now(),
          payload: { message: 'WebSocket connection closed' },
        })

        if (shouldReconnectRef.current) {
          reconnectAttemptsRef.current++
          const delay = Math.min(10000, 1000 * Math.pow(1.5, reconnectAttemptsRef.current))
          reconnectTimeoutRef.current = window.setTimeout(() => {
            connect()
          }, delay)
        } else {
          setStatus('CLOSED')
        }
      }

      ws.onerror = (err) => {
        if (shouldReconnectRef.current) {
          console.warn('WebSocket reconnect pending...', err)
        }
      }
    } catch (err) {
      console.error('Failed to instantiate WebSocket:', err)
      setStatus('CLOSED')
    }
  }, [user, tokens, currentRoom, role, addEvent])

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
      if (pingIntervalRef.current) {
        clearInterval(pingIntervalRef.current)
      }
      if (socketRef.current) {
        const ws = socketRef.current
        socketRef.current = null
        if (ws.readyState === WebSocket.OPEN) {
          ws.close(1000, 'Component unmounted')
        } else if (ws.readyState === WebSocket.CONNECTING) {
          ws.onopen = () => ws.close(1000, 'Component unmounted')
        }
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

  const joinRoom = useCallback((roomId: string, passcode?: string, joinRole?: 'player' | 'spectator') => {
    const clean = roomId.trim().toUpperCase()
    if (!clean) return
    setCurrentRoom(clean)
    sendEvent('room.join', { passcode, role: joinRole || role }, clean)
  }, [role, sendEvent])

  const leaveRoom = useCallback(() => {
    if (!currentRoom) return
    sendEvent('room.leave', {}, currentRoom)
    setCurrentRoom('')
    setMembers([])
    setSpectatorCount(0)
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
    spectatorCount,
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
