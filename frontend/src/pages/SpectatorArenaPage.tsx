import React, { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Eye, ArrowLeft, Send, Sparkles, Trophy, Radio, MessageSquare, Volume2, ShieldAlert } from 'lucide-react'
import { useWebSocket } from '../hooks/useWebSocket'

export const SpectatorArenaPage: React.FC = () => {
  const { gameType = 'xo', roomId = '' } = useParams<{ gameType: string; roomId: string }>()
  const navigate = useNavigate()

  const {
    status,
    members,
    spectatorCount,
    events,
    sendMessage,
    leaveRoom,
  } = useWebSocket({
    autoConnect: true,
    initialRoom: roomId,
    role: 'spectator',
  })

  const [gameState, setGameState] = useState<any>(null)
  const [chatInput, setChatInput] = useState('')
  const [chatMessages, setChatMessages] = useState<Array<{ sender: string; text: string; time: string }>>([])

  // Listen for game.state and room.message events
  useEffect(() => {
    if (events.length === 0) return
    const latest = events[0]
    if (latest.type === 'game.state' && latest.payload) {
      setGameState(latest.payload)
    } else if (latest.type === 'room.message' && latest.payload) {
      setChatMessages((prev) => [
        ...prev.slice(-20),
        {
          sender: latest.payload.sender || 'Classmate',
          text: latest.payload.text || '',
          time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        },
      ])
    }
  }, [events])

  const handleSendChat = (e: React.FormEvent) => {
    e.preventDefault()
    if (!chatInput.trim()) return
    sendMessage(chatInput.trim())
    setChatInput('')
  }

  const handleLeave = () => {
    leaveRoom()
    navigate('/games')
  }

  const p1 = members[0] || { username: 'Player 1', user_id: 'p1' }
  const p2 = members[1] || { username: 'Player 2', user_id: 'p2' }
  const currentTurn = gameState?.current_turn
  const isP1Turn = currentTurn === p1.user_id
  const isFinished = gameState?.status === 'finished'

  const gameTitles: Record<string, string> = {
    xo: 'Tic-Tac-Toe / XO',
    hand_cricket: 'Hand Cricket (Even-Odd)',
    dots_boxes: 'Dots & Boxes',
    connect4: 'Connect 4',
    paper_football: 'Paper Football (Origami Tabletop)',
    npat: 'Name–Place–Animal–Thing',
  }

  return (
    <div className="min-h-screen bg-[#FBF9F1] text-[#2C3E50] p-4 md:p-8 font-sans selection:bg-[#EAE0D5]">
      {/* Top Banner & Sideline Bar */}
      <div className="max-w-5xl mx-auto mb-6">
        <div className="flex flex-wrap items-center justify-between gap-4 bg-white/90 backdrop-blur border-2 border-[#4A6B82]/30 rounded-2xl p-4 shadow-md">
          <div className="flex items-center gap-3">
            <button
              onClick={handleLeave}
              className="flex items-center gap-1.5 px-3 py-1.5 bg-[#EAE0D5] hover:bg-[#D8C7B5] text-[#2C3E50] text-sm font-bold rounded-xl transition border border-[#4A6B82]/20"
            >
              <ArrowLeft className="w-4 h-4" />
              Leave Sideline
            </button>
            <div className="flex items-center gap-2 px-3 py-1 bg-red-100 text-red-700 font-extrabold text-xs rounded-full border border-red-300">
              <span className="w-2.5 h-2.5 rounded-full bg-red-600 animate-ping inline-block" />
              <Radio className="w-3.5 h-3.5" />
              SIDE-BENCH LIVE
            </div>
          </div>

          <div className="flex items-center gap-3 text-xs md:text-sm font-semibold">
            <div className={`px-2.5 py-1 rounded-lg text-xs font-bold border ${
              status === 'OPEN'
                ? 'bg-emerald-100 text-emerald-800 border-emerald-300'
                : 'bg-amber-100 text-amber-800 border-amber-300'
            }`}>
              {status}
            </div>
            <div className="flex items-center gap-1.5 px-3 py-1 bg-[#4A6B82]/10 text-[#4A6B82] rounded-lg border border-[#4A6B82]/20">
              <Eye className="w-4 h-4" />
              <span>{spectatorCount > 0 ? spectatorCount : 1} Watching on Sideline</span>
            </div>
            <div className="px-3 py-1 bg-[#8B5A2B]/10 text-[#8B5A2B] font-mono rounded-lg border border-[#8B5A2B]/20">
              Desk Room: #{roomId}
            </div>
          </div>
        </div>
      </div>

      <div className="max-w-5xl mx-auto grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Main Arena Display */}
        <div className="lg:col-span-2 space-y-6">
          {/* Match Scorecard Header */}
          <div className="bg-white border-2 border-[#4A6B82]/30 rounded-2xl p-5 shadow-md relative overflow-hidden">
            <div className="absolute top-0 right-0 px-4 py-1 bg-[#4A6B82] text-white text-xs font-bold rounded-bl-xl tracking-wider uppercase">
              {gameTitles[gameType] || gameType.toUpperCase()}
            </div>

            <div className="flex items-center justify-between gap-4 mt-2">
              {/* Player 1 Card */}
              <div
                className={`flex-1 p-3.5 rounded-xl border-2 transition ${
                  isP1Turn && !isFinished
                    ? 'border-blue-500 bg-blue-50/70 shadow-sm ring-2 ring-blue-300'
                    : 'border-gray-200 bg-gray-50/50'
                }`}
              >
                <div className="flex items-center gap-2">
                  <div className="w-9 h-9 rounded-full bg-blue-600 text-white font-bold flex items-center justify-center text-sm shadow">
                    {p1.username?.[0]?.toUpperCase() || '1'}
                  </div>
                  <div className="overflow-hidden">
                    <p className="font-bold text-sm truncate text-[#2C3E50]">@{p1.username}</p>
                    <p className="text-[11px] text-blue-700 font-semibold uppercase tracking-wide">Player 1</p>
                  </div>
                </div>
                {isP1Turn && !isFinished && (
                  <div className="mt-2 text-[11px] font-bold text-blue-600 flex items-center gap-1 animate-pulse">
                    <Sparkles className="w-3 h-3" /> Thinking...
                  </div>
                )}
              </div>

              {/* VS Center Marker */}
              <div className="text-center px-2">
                <span className="font-extrabold text-sm text-[#8B5A2B] bg-[#EAE0D5] px-2.5 py-1 rounded-full border border-[#8B5A2B]/20">
                  VS
                </span>
                {gameState?.move_count !== undefined && (
                  <p className="text-[10px] text-gray-500 font-mono mt-1">Move #{gameState.move_count}</p>
                )}
              </div>

              {/* Player 2 Card */}
              <div
                className={`flex-1 p-3.5 rounded-xl border-2 transition text-right ${
                  !isP1Turn && currentTurn && !isFinished
                    ? 'border-red-500 bg-red-50/70 shadow-sm ring-2 ring-red-300'
                    : 'border-gray-200 bg-gray-50/50'
                }`}
              >
                <div className="flex items-center justify-end gap-2">
                  <div className="overflow-hidden">
                    <p className="font-bold text-sm truncate text-[#2C3E50]">@{p2.username}</p>
                    <p className="text-[11px] text-red-700 font-semibold uppercase tracking-wide">Player 2</p>
                  </div>
                  <div className="w-9 h-9 rounded-full bg-red-600 text-white font-bold flex items-center justify-center text-sm shadow">
                    {p2.username?.[0]?.toUpperCase() || '2'}
                  </div>
                </div>
                {!isP1Turn && currentTurn && !isFinished && (
                  <div className="mt-2 text-[11px] font-bold text-red-600 flex items-center justify-end gap-1 animate-pulse">
                    <Sparkles className="w-3 h-3" /> Thinking...
                  </div>
                )}
              </div>
            </div>

            {/* Match Status Banner */}
            {isFinished && (
              <div className="mt-4 p-3 bg-amber-50 border border-amber-300 rounded-xl text-center flex items-center justify-center gap-2 text-amber-900 font-bold">
                <Trophy className="w-5 h-5 text-amber-600" />
                <span>
                  Match Concluded! {gameState.result?.winner_id ? `Winner: @${gameState.result.winner_id}` : 'Game Tied!'}
                </span>
              </div>
            )}
          </div>

          {/* Live Game Visualizer */}
          <div className="bg-white border-2 border-[#4A6B82]/30 rounded-2xl p-6 shadow-md min-h-[340px] flex flex-col items-center justify-center relative">
            <div className="absolute top-3 left-4 flex items-center gap-1.5 text-xs text-[#8B5A2B] font-bold">
              <Volume2 className="w-3.5 h-3.5" />
              Live Sideline Broadcast
            </div>

            {/* Visualizer Renderer based on game state */}
            {gameState ? (
              <div className="w-full max-w-sm text-center">
                <div className="p-4 bg-[#FBF9F1] rounded-xl border border-[#EAE0D5] font-mono text-sm shadow-inner">
                  <p className="text-xs font-bold text-gray-500 uppercase tracking-wide mb-2">Live Board State</p>
                  <pre className="text-xs text-left bg-white p-3 rounded-lg border border-gray-200 overflow-x-auto text-[#2C3E50]">
                    {JSON.stringify(gameState.board_state || gameState, null, 2)}
                  </pre>
                </div>
              </div>
            ) : (
              <div className="text-center py-12 text-gray-400">
                <Radio className="w-8 h-8 mx-auto mb-2 animate-spin text-[#4A6B82]" />
                <p className="font-semibold text-sm">Tuning into classroom desk feed...</p>
                <p className="text-xs text-gray-400 mt-1">Waiting for players to commence duel.</p>
              </div>
            )}

            {/* Spectator notice at bottom */}
            <div className="mt-6 flex items-center gap-2 text-xs text-[#8B5A2B] bg-[#EAE0D5]/50 px-4 py-2 rounded-xl border border-[#8B5A2B]/20">
              <ShieldAlert className="w-4 h-4 text-[#8B5A2B] flex-shrink-0" />
              <span>Sideline Mode: Live actions sync automatically. Move submissions are disabled for bystanders.</span>
            </div>
          </div>
        </div>

        {/* Sideline Cheering & Notes Feed */}
        <div className="space-y-4">
          <div className="bg-white border-2 border-[#4A6B82]/30 rounded-2xl p-4 shadow-md flex flex-col h-[500px]">
            <div className="flex items-center gap-2 pb-3 border-b border-gray-100 font-bold text-sm text-[#2C3E50]">
              <MessageSquare className="w-4 h-4 text-[#4A6B82]" />
              Sideline Notes & Cheers
            </div>

            {/* Messages Scroll Area */}
            <div className="flex-1 overflow-y-auto py-3 space-y-2.5 pr-1">
              {chatMessages.length === 0 ? (
                <div className="text-center py-16 text-gray-400 text-xs italic">
                  No sideline notes yet.<br />Pass a note to cheer on the players!
                </div>
              ) : (
                chatMessages.map((msg, idx) => (
                  <div key={idx} className="bg-[#FBF9F1] p-2.5 rounded-xl border border-[#EAE0D5] text-xs">
                    <div className="flex items-center justify-between text-[10px] text-gray-400 font-semibold mb-1">
                      <span className="text-[#4A6B82] font-bold">@{msg.sender}</span>
                      <span>{msg.time}</span>
                    </div>
                    <p className="text-[#2C3E50]">{msg.text}</p>
                  </div>
                ))
              )}
            </div>

            {/* Pass Note Input Form */}
            <form onSubmit={handleSendChat} className="pt-3 border-t border-gray-100 flex gap-2">
              <input
                type="text"
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                placeholder="Pass a cheer or note..."
                className="flex-1 px-3 py-2 text-xs bg-[#FBF9F1] border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-[#4A6B82]/40"
              />
              <button
                type="submit"
                disabled={!chatInput.trim()}
                className="p-2 bg-[#4A6B82] hover:bg-[#3D586B] disabled:opacity-50 text-white rounded-xl transition"
              >
                <Send className="w-4 h-4" />
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  )
}
