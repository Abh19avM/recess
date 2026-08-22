import React, { useState } from 'react'
import { useWebSocket } from '../hooks/useWebSocket'
import { useAuthStore } from '../store/authStore'
import { PaperCard } from '../components/ui/PaperCard'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Badge, Stamp } from '../components/ui/Badge'
import { Avatar } from '../components/ui/Avatar'
import {
  Wifi,
  WifiOff,
  Users,
  Send,
  CheckCircle,
  RefreshCw,
  LogOut,
  Sparkles,
  Terminal,
  Activity,
} from 'lucide-react'

export const WebSocketTestPage: React.FC = () => {
  const { user } = useAuthStore()
  const {
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
  } = useWebSocket({ autoConnect: true })

  const [inputRoom, setInputRoom] = useState('RECESS-BENCH-1')
  const [chatMessage, setChatMessage] = useState('')
  const [customEventType, setCustomEventType] = useState('game.move')
  const [customPayload, setCustomPayload] = useState('{"move": "strike", "runs": 4}')

  const handleJoin = (e: React.FormEvent) => {
    e.preventDefault()
    if (inputRoom.trim()) {
      joinRoom(inputRoom.trim())
    }
  }

  const handleSendChat = (e: React.FormEvent) => {
    e.preventDefault()
    if (chatMessage.trim()) {
      sendMessage(chatMessage)
      setChatMessage('')
    }
  }

  const handleSendCustom = (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const parsed = JSON.parse(customPayload)
      sendEvent(customEventType, parsed)
    } catch {
      alert('Invalid JSON in custom payload')
    }
  }

  const getStatusBadge = () => {
    switch (status) {
      case 'OPEN':
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold font-mono bg-[#DCFCE7] text-[#15803D] border border-[#86EFAC] shadow-xs">
            <span className="w-2 h-2 rounded-full bg-[#15803D] animate-pulse" />
            WSS ONLINE (CONNECTED)
          </span>
        )
      case 'CONNECTING':
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold font-mono bg-[#FEF9C3] text-[#854D0E] border border-[#FDE047]">
            <RefreshCw className="w-3 h-3 animate-spin" />
            CONNECTING...
          </span>
        )
      default:
        return (
          <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold font-mono bg-[#FFE4E6] text-[#991B1B] border border-[#FECDD3]">
            <WifiOff className="w-3 h-3" />
            OFFLINE (DISCONNECTED)
          </span>
        )
    }
  }

  return (
    <div className="space-y-8 max-w-6xl mx-auto pb-16">
      {/* 1. Header Banner */}
      <PaperCard variant="ruled" className="p-6 relative">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-2xl sm:text-3xl font-extrabold text-[#1E242B]">
                Real-Time WebSocket Test Bench
              </h1>
              <Stamp tone="red">DEV ARENA</Stamp>
            </div>
            <p className="font-hand text-lg text-[#1A365D] mt-0.5">
              Live bidirectional testing for rooms, presence, readiness, and structured event envelopes.
            </p>
          </div>

          <div className="flex items-center gap-3">
            {getStatusBadge()}
            {status === 'CLOSED' && (
              <Button onClick={connect} size="sm" variant="primary" leftIcon={<Wifi className="w-4 h-4" />}>
                Reconnect
              </Button>
            )}
          </div>
        </div>

        {/* Multi-Tab Testing Tip */}
        <div className="mt-4 p-3 rounded bg-[#E0F2FE] border border-[#BAE6FD] flex items-start gap-2.5 text-xs text-[#0369A1]">
          <Sparkles className="w-4 h-4 text-[#0284C7] shrink-0 mt-0.5" />
          <div>
            <strong className="font-bold">Multi-Client Testing Tip:</strong> Open this page in a second browser window (or Incognito tab) with a different guest handle. Join room <code className="font-mono bg-white/80 px-1 py-0.5 rounded text-[#0C4A6E] font-bold">RECESS-BENCH-1</code> in both tabs to verify instant presence, readiness sync, and message broadcasts.
          </div>
        </div>
      </PaperCard>

      {/* 2. Main Two-Column Control & Inspector Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
        {/* Left Column: Room Controls & Actions (5 cols) */}
        <div className="lg:col-span-5 space-y-6">
          {/* Room Selection Card */}
          <PaperCard variant="plain" className="p-6">
            <div className="flex items-center justify-between mb-4 pb-2 border-b border-[#CBD5E1]">
              <div className="flex items-center gap-2">
                <Users className="w-4 h-4 text-[#1A365D]" />
                <h3 className="font-bold text-base text-[#1E242B]">Desk Room Control</h3>
              </div>
              {currentRoom && <Stamp tone="green">IN ROOM: {currentRoom}</Stamp>}
            </div>

            <form onSubmit={handleJoin} className="space-y-3">
              <Input
                label="Room Code"
                placeholder="e.g. RECESS-BENCH-1"
                value={inputRoom}
                onChange={(e) => setInputRoom(e.target.value.toUpperCase())}
                className="font-mono font-bold tracking-wider"
              />

              <div className="flex items-center gap-2">
                <Button type="submit" variant="primary" size="md" className="flex-1">
                  Join Room →
                </Button>
                {currentRoom && (
                  <Button
                    type="button"
                    onClick={leaveRoom}
                    variant="ghost"
                    size="md"
                    className="text-[#991B1B]"
                    leftIcon={<LogOut className="w-4 h-4" />}
                  >
                    Leave
                  </Button>
                )}
              </div>
            </form>

            {/* Quick Preset Buttons */}
            <div className="mt-4 pt-3 border-t border-[#E2E8F0]">
              <span className="text-[11px] font-mono text-[#64748B] uppercase block mb-2">
                Quick Test Rooms:
              </span>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() => {
                    setInputRoom('RECESS-BENCH-1')
                    joinRoom('RECESS-BENCH-1')
                  }}
                >
                  Bench 1
                </Button>
                <Button
                  size="sm"
                  variant="secondary"
                  onClick={() => {
                    setInputRoom('RECESS-BENCH-2')
                    joinRoom('RECESS-BENCH-2')
                  }}
                >
                  Bench 2
                </Button>
                <Button
                  size="sm"
                  variant="ghost"
                  onClick={() => {
                    const random = `RECESS-${Math.random().toString(36).substring(2, 6).toUpperCase()}`
                    setInputRoom(random)
                    joinRoom(random)
                  }}
                >
                  + Random
                </Button>
              </div>
            </div>
          </PaperCard>

          {/* Player Readiness Card */}
          <PaperCard variant="sticky" stickyColor="yellow" showPushPin className="p-6">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <Activity className="w-4 h-4 text-[#B45309]" />
                <h3 className="font-bold text-base text-[#1E242B]">Classroom Readiness</h3>
              </div>
              <Stamp tone={isReady ? 'green' : 'amber'}>
                {isReady ? 'READY TO PLAY' : 'NOT READY'}
              </Stamp>
            </div>

            <p className="font-hand text-base text-[#1A365D] mb-4">
              Toggle your state to broadcast <code className="font-mono text-xs">player.ready</code> to all peers in the room.
            </p>

            <Button
              onClick={() => toggleReady()}
              variant={isReady ? 'secondary' : 'primary'}
              size="md"
              className="w-full"
              leftIcon={<CheckCircle className="w-4 h-4" />}
            >
              {isReady ? 'Cancel Ready (Set Unready)' : 'Ring Bell & Set Ready!'}
            </Button>
          </PaperCard>

          {/* Broadcast Note Form */}
          <PaperCard variant="plain" className="p-6">
            <h3 className="font-bold text-base text-[#1E242B] mb-2 flex items-center gap-2">
              <Send className="w-4 h-4 text-[#1A365D]" />
              Pass Classroom Note
            </h3>
            <form onSubmit={handleSendChat} className="space-y-3">
              <Input
                placeholder="Type a secret message..."
                value={chatMessage}
                onChange={(e) => setChatMessage(e.target.value)}
                disabled={!currentRoom}
              />
              <Button
                type="submit"
                variant="primary"
                size="sm"
                className="w-full"
                disabled={!currentRoom || !chatMessage.trim()}
              >
                Send Note to Room Peers
              </Button>
            </form>
          </PaperCard>

          {/* Custom Event Emitter */}
          <PaperCard variant="plain" className="p-6">
            <h3 className="font-bold text-base text-[#1E242B] mb-2 flex items-center gap-2">
              <Terminal className="w-4 h-4 text-[#15803D]" />
              Emit Custom Game Envelope
            </h3>
            <form onSubmit={handleSendCustom} className="space-y-3">
              <Input
                label="Event Type"
                value={customEventType}
                onChange={(e) => setCustomEventType(e.target.value)}
                placeholder="e.g. game.action"
                disabled={!currentRoom}
              />
              <div>
                <label className="block text-xs font-semibold text-[#1E242B] mb-1">
                  JSON Payload
                </label>
                <textarea
                  value={customPayload}
                  onChange={(e) => setCustomPayload(e.target.value)}
                  rows={2}
                  disabled={!currentRoom}
                  className="w-full text-xs font-mono bg-white p-2 rounded-md border-2 border-[#CBD5E1] outline-none"
                />
              </div>
              <Button
                type="submit"
                variant="secondary"
                size="sm"
                className="w-full"
                disabled={!currentRoom}
              >
                Emit Custom Envelope
              </Button>
            </form>
          </PaperCard>
        </div>

        {/* Right Column: Room Roster & Live Event Stream (7 cols) */}
        <div className="lg:col-span-7 space-y-6">
          {/* Active Desk Roster */}
          <PaperCard variant="ruled" className="p-6">
            <div className="flex items-center justify-between mb-4 pb-2 border-b border-[#CBD5E1]">
              <div className="flex items-center gap-2">
                <Users className="w-5 h-5 text-[#1A365D]" />
                <h3 className="font-bold text-base text-[#1E242B]">
                  Desk Roster ({members.length} {members.length === 1 ? 'Student' : 'Students'})
                </h3>
              </div>
              <Badge variant="ink-blue" size="sm">
                Room: {currentRoom || 'None (In Lobby)'}
              </Badge>
            </div>

            {members.length === 0 ? (
              <div className="text-center py-6 text-[#64748B] font-hand text-lg">
                No students at this desk yet. Join a room above!
              </div>
            ) : (
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                {members.map((member) => {
                  const isMe = member.user_id === user?.id || member.username === user?.username
                  return (
                    <div
                      key={member.user_id}
                      className={`p-3 rounded-md border-2 transition-all flex items-center justify-between ${
                        isMe
                          ? 'bg-[#F0FDF4] border-[#15803D] shadow-[2px_2px_0px_0px_#15803D]'
                          : 'bg-white border-[#CBD5E1]'
                      }`}
                    >
                      <div className="flex items-center gap-2.5">
                        <Avatar username={member.username} preset={member.avatar_preset} size="sm" />
                        <div>
                          <div className="flex items-center gap-1.5">
                            <span className="font-bold text-xs text-[#1E242B]">
                              @{member.username}
                            </span>
                            {isMe && (
                              <span className="text-[9px] font-mono font-bold bg-[#DCFCE7] text-[#15803D] px-1 rounded">
                                YOU
                              </span>
                            )}
                          </div>
                          <span className="text-[10px] font-mono text-[#64748B] block">
                            {member.is_guest ? 'Guest Student' : 'Enrolled'}
                          </span>
                        </div>
                      </div>

                      <Stamp tone={member.is_ready ? 'green' : 'amber'} className="text-[9px] py-0 px-1.5">
                        {member.is_ready ? 'READY' : 'WAITING'}
                      </Stamp>
                    </div>
                  )
                })}
              </div>
            )}
          </PaperCard>

          {/* Live WebSocket Event Ledger */}
          <PaperCard variant="plain" className="p-6">
            <div className="flex items-center justify-between mb-4 pb-2 border-b border-[#CBD5E1]">
              <div className="flex items-center gap-2">
                <Terminal className="w-5 h-5 text-[#1A365D]" />
                <h3 className="font-bold text-base text-[#1E242B]">Live WebSocket Event Stream</h3>
              </div>
              <span className="text-xs font-mono text-[#64748B]">
                {events.length} Envelopes
              </span>
            </div>

            <div className="bg-[#18231C] rounded-md p-4 text-[#F8FAFC] font-mono text-xs max-h-96 overflow-y-auto space-y-3 shadow-inner">
              {events.length === 0 ? (
                <div className="text-[#94A3B8] text-center py-8">
                  Awaiting real-time WebSocket events...
                </div>
              ) : (
                events.map((evt, idx) => (
                  <div key={idx} className="border-b border-[#334155]/60 pb-2 last:border-0 last:pb-0">
                    <div className="flex items-center justify-between text-[11px] mb-1">
                      <span className="text-[#38BDF8] font-bold">
                        [{evt.type}]
                      </span>
                      <span className="text-[#94A3B8]">
                        seq: {evt.sequence || 0} • {evt.timestamp ? new Date(evt.timestamp).toLocaleTimeString() : 'now'}
                      </span>
                    </div>
                    {evt.room_id && (
                      <div className="text-[10px] text-[#A7F3D0] mb-1">
                        room: {evt.room_id}
                      </div>
                    )}
                    {evt.payload && (
                      <pre className="text-[11px] text-[#E2E8F0] overflow-x-auto bg-[#0F172A]/80 p-2 rounded">
                        {JSON.stringify(evt.payload, null, 2)}
                      </pre>
                    )}
                  </div>
                ))
              )}
            </div>
          </PaperCard>
        </div>
      </div>
    </div>
  )
}
