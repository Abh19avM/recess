import React, { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useWebSocket } from '../hooks/useWebSocket'
import { useAuthStore } from '../store/authStore'
import { PaperCard } from '../components/ui/PaperCard'
import { Button } from '../components/ui/Button'
import { Badge, Stamp } from '../components/ui/Badge'
import { Avatar } from '../components/ui/Avatar'
import { toast } from '../store/toastStore'
import {
  RotateCcw,
  Sparkles,
  Users,
  Send,
  ArrowLeft,
  Wifi,
  WifiOff,
  Zap,
} from 'lucide-react'
import confetti from 'canvas-confetti'

interface PaperFootballBoardState {
  phase: 'drive' | 'extra_point' | 'finished'
  possession_id: string
  ball_position: number
  down: number
  scores: Record<string, number>
  quarter: number
  last_flick_result?: string
  last_distance: number
  message: string
}

interface PaperFootballGameState {
  game_id: string
  game_type: string
  status: 'waiting' | 'active' | 'finished' | 'abandoned'
  current_turn: string
  move_count: number
  board_state: PaperFootballBoardState
  result?: {
    winner_id: string
    is_draw: boolean
    scores: Record<string, number>
    reason: string
  }
  version: number
}

export const PaperFootballArenaPage: React.FC = () => {
  const { roomId } = useParams<{ roomId?: string }>()
  const { user } = useAuthStore()

  const targetRoom = roomId || 'RECESS-PF-DESK'
  const {
    status,
    currentRoom,
    members,
    events,
    isReady,
    joinRoom,
    toggleReady,
    sendEvent,
    sendMessage,
  } = useWebSocket({ autoConnect: true, initialRoom: targetRoom })

  const [gameState, setGameState] = useState<PaperFootballGameState | null>(null)
  const [chatInput, setChatInput] = useState('')
  const [deskNotes, setDeskNotes] = useState<Array<{ sender: string; text: string; time: string }>>([])

  // Flick Controls
  const [flickPower, setFlickPower] = useState(65)
  const [flickAngle, setFlickAngle] = useState(0)

  // Solo Practice Bot State
  const [isBotMode, setIsBotMode] = useState(false)
  const [botBallPos, setBotBallPos] = useState(20)
  const [botDown, setBotDown] = useState(1)
  const [botPhase, setBotPhase] = useState<'drive' | 'extra_point'>('drive')
  const [botTurn, setBotTurn] = useState<'player' | 'bot'>('player')
  const [botPlayerScore, setBotPlayerScore] = useState(0)
  const [botScore, setBotScore] = useState(0)
  const [botMessage, setBotMessage] = useState('Flick the paper triangle across the classroom desk!')
  const [botWinnerMsg, setBotWinnerMsg] = useState<string | null>(null)

  // Listen to WebSocket game.state updates
  useEffect(() => {
    if (events.length > 0) {
      const latest = events[0]
      if (latest.type === 'game.state' && latest.payload) {
        const state = latest.payload as PaperFootballGameState
        setGameState(state)

        if (state.status === 'finished' && state.result?.winner_id === user?.id) {
          confetti({
            particleCount: 100,
            spread: 80,
            origin: { y: 0.6 },
            colors: ['#1A365D', '#15803D', '#B45309', '#F59E0B'],
          })
        }
      } else if (latest.type === 'room.message' && latest.payload) {
        const p = latest.payload as { sender: string; text: string }
        setDeskNotes((prev) => [
          ...prev,
          {
            sender: p.sender,
            text: p.text,
            time: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
          },
        ])
      }
    }
  }, [events, user])

  useEffect(() => {
    if (status === 'OPEN' && currentRoom !== targetRoom) {
      joinRoom(targetRoom)
    }
  }, [status, targetRoom, currentRoom, joinRoom])

  // Move Submissions
  const handleExecuteFlick = () => {
    if (isBotMode) {
      handleBotFlick()
      return
    }

    if (!gameState || gameState.status !== 'active') {
      if (members.length < 2) {
        toast.info('Waiting for Classmate', 'Invite a classmate to join this desk room to start.')
      } else if (!isReady) {
        toast.info('Ready Up to Play', 'Toggle "Ready Up" to kickoff Paper Football.')
      }
      return
    }

    if (gameState.current_turn !== user?.id) {
      toast.error('Not Your Turn!', 'Wait for opponent to finish their desk flick.')
      return
    }

    const action = gameState.board_state?.phase === 'extra_point' ? 'extra_point' : 'flick'
    sendEvent('game.move', {
      action,
      power: flickPower,
      angle: flickAngle,
    })
  }

  const handleRematch = () => {
    if (isBotMode) {
      resetBotGame()
      return
    }
    sendEvent('game.rematch', {})
    toast.success('Rematch Requested', 'Starting a fresh desk match!')
  }

  const handleSendNote = (e: React.FormEvent) => {
    e.preventDefault()
    if (!chatInput.trim()) return
    sendMessage(chatInput.trim())
    setChatInput('')
  }

  // Offline Bot Logic
  const handleBotFlick = () => {
    if (botWinnerMsg || botTurn !== 'player') return

    if (botPhase === 'extra_point') {
      // Player kicks extra point
      if (flickPower >= 35 && Math.abs(flickAngle) <= 20) {
        const nextScore = botPlayerScore + 1
        setBotPlayerScore(nextScore)
        setBotMessage('EXTRA POINT IS GOOD! (+1 Point)')
        confetti({ particleCount: 50, spread: 50 })
        if (nextScore >= 21) {
          setBotWinnerMsg('🏆 Victory! You reached 21 Points!')
          return
        }
      } else {
        setBotMessage('EXTRA POINT MISSED! Hit the uprights.')
      }
      setBotPhase('drive')
      setBotTurn('bot')
      setBotBallPos(20)
      setBotDown(1)
      triggerBotTurn(botPlayerScore + 1, botScore)
      return
    }

    // Drive Flick
    const rad = flickAngle * (Math.PI / 180)
    const distanceGained = flickPower * Math.cos(rad) * 0.75
    const newPos = botBallPos + distanceGained

    if (newPos >= 90 && newPos <= 100) {
      // Touchdown!
      const nextScore = botPlayerScore + 6
      setBotPlayerScore(nextScore)
      setBotBallPos(newPos)
      setBotPhase('extra_point')
      setBotMessage('TOUCHDOWN! Paper triangle hanging off edge! (+6 Points) Kick the extra point!')
      confetti({ particleCount: 70, spread: 60 })
      if (nextScore >= 21) {
        setBotWinnerMsg('🏆 Victory! You reached 21 Points!')
      }
      return
    }

    if (newPos > 100) {
      // Table Fall
      setBotMessage('OVER-FLICK! Triangle fell off table! Turnover to ClassBot.')
      setBotTurn('bot')
      setBotBallPos(20)
      setBotDown(1)
      triggerBotTurn(botPlayerScore, botScore)
      return
    }

    // Advance
    const nextDown = botDown + 1
    if (nextDown > 4) {
      setBotMessage('Turnover on downs! ClassBot takes over possession.')
      setBotTurn('bot')
      setBotBallPos(Math.max(10, 100 - newPos))
      setBotDown(1)
      triggerBotTurn(botPlayerScore, botScore)
    } else {
      setBotBallPos(newPos)
      setBotDown(nextDown)
      setBotMessage(`Advanced to ${newPos.toFixed(1)}% of desk. Down ${nextDown} of 4.`)
    }
  }

  const triggerBotTurn = (_pScore: number, bScore: number) => {
    setTimeout(() => {
      // Bot flicks with randomized power around 60-80%
      const botPow = 55 + Math.floor(Math.random() * 35)
      const dist = botPow * 0.75
      const botPos = 20 + dist

      if (botPos >= 90 && botPos <= 100) {
        const nextBScore = bScore + 6
        setBotScore(nextBScore)
        setBotMessage('ClassBot scored a TOUCHDOWN! (+6 Points)')
        if (nextBScore >= 21) {
          setBotWinnerMsg('ClassBot reached 21 Points! Better luck next period.')
          return
        }
        // Bot auto kicks extra point
        const kickGood = Math.random() > 0.3
        const finalBScore = nextBScore + (kickGood ? 1 : 0)
        setBotScore(finalBScore)
        setBotTurn('player')
        setBotBallPos(20)
        setBotDown(1)
      } else if (botPos > 100) {
        setBotMessage('ClassBot over-flicked off the desk! Your possession from 20% mark.')
        setBotTurn('player')
        setBotBallPos(20)
        setBotDown(1)
      } else {
        setBotTurn('player')
        setBotBallPos(Math.max(10, 100 - botPos))
        setBotDown(1)
        setBotMessage(`ClassBot advanced. Your possession from ${Math.max(10, 100 - botPos).toFixed(1)}%.`)
      }
    }, 800)
  }

  const resetBotGame = () => {
    setBotBallPos(20)
    setBotDown(1)
    setBotPhase('drive')
    setBotTurn('player')
    setBotPlayerScore(0)
    setBotScore(0)
    setBotMessage('Flick the paper triangle across the classroom desk!')
    setBotWinnerMsg(null)
  }

  // Active Board Data
  const board = gameState?.board_state
  const ballPos = isBotMode ? botBallPos : board?.ball_position || 20
  const down = isBotMode ? botDown : board?.down || 1
  const phase = isBotMode ? botPhase : board?.phase || 'drive'
  const otherMember = members.find((m) => m.user_id !== user?.id)
  const myScore = isBotMode ? botPlayerScore : (user && board?.scores?.[user.id]) || 0
  const oppScore = isBotMode ? botScore : (otherMember && board?.scores?.[otherMember.user_id]) || 0
  const isMyTurn = isBotMode
    ? botTurn === 'player'
    : gameState?.status === 'active' && gameState?.current_turn === user?.id

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16">
      {/* 1. Header Navigation */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b-2 border-[#1E242B] pb-4">
        <div className="flex items-center gap-3">
          <Link to="/games">
            <Button variant="secondary" size="sm" leftIcon={<ArrowLeft className="w-4 h-4" />}>
              Syllabus
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-extrabold text-[#1E242B]">Paper Football Arena</h1>
              <Stamp tone="red">DESK FLICK</Stamp>
            </div>
            <p className="font-hand text-sm text-[#475569]">
              Room: <span className="font-mono font-bold text-[#1A365D]">{targetRoom}</span> • First to 21 Points • Edge = Touchdown (6 Pts)!
            </p>
          </div>
        </div>

        {/* Solo Practice Mode Toggle */}
        <div className="flex items-center gap-2">
          <Button
            onClick={() => {
              setIsBotMode(!isBotMode)
              if (!isBotMode) resetBotGame()
            }}
            variant={isBotMode ? 'primary' : 'secondary'}
            size="sm"
            leftIcon={<Sparkles className="w-3.5 h-3.5" />}
          >
            {isBotMode ? 'Solo Practice (Active)' : 'Practice vs Bot'}
          </Button>

          {!isBotMode && (
            <div className="flex items-center gap-1 text-xs font-mono">
              {status === 'OPEN' ? (
                <span className="flex items-center gap-1 text-[#15803D] bg-[#DCFCE7] px-2 py-1 rounded border border-[#86EFAC]">
                  <Wifi className="w-3.5 h-3.5" /> LIVE
                </span>
              ) : (
                <span className="flex items-center gap-1 text-[#991B1B] bg-[#FFE4E6] px-2 py-1 rounded border border-[#FECDD3]">
                  <WifiOff className="w-3.5 h-3.5" /> OFFLINE
                </span>
              )}
            </div>
          )}
        </div>
      </div>

      {/* 2. Main Arena Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left: Desk Pitch & Flick Controls (7 cols) */}
        <div className="lg:col-span-7 space-y-4">
          {/* Action Message Banner */}
          <PaperCard variant="sticky" stickyColor="yellow" showTape className="p-3 text-center">
            <div className="font-hand text-lg font-bold text-[#1A365D]">
              {isBotMode
                ? botWinnerMsg || botMessage
                : gameState?.board_state?.message || 'Both players ready up to kickoff the desk drive!'}
            </div>
          </PaperCard>

          {/* Wooden Classroom Desk Surface Pitch */}
          <PaperCard variant="plain" className="p-6 relative select-none bg-[#D97706]/10 border-4 border-[#92400E] shadow-[6px_6px_0px_0px_#78350F] rounded-2xl overflow-hidden">
            {/* Score HUD */}
            <div className="flex items-center justify-between mb-4 bg-white/90 px-4 py-2 rounded-lg border border-[#CBD5E1] font-mono font-bold text-sm">
              <span className="text-[#1A365D]">You: {myScore} PTS</span>
              <Badge variant="pencil" size="sm">DOWN {down} OF 4</Badge>
              <span className="text-[#991B1B]">Opponent: {oppScore} PTS</span>
            </div>

            {/* Desk Surface Field Grid */}
            <div className="relative h-64 w-full bg-[#FEF3C7] rounded-xl border-2 border-dashed border-[#B45309] p-4 flex flex-col justify-between overflow-hidden">
              {/* Opponent End / Touchdown Edge */}
              <div className="w-full text-center pb-2 border-b-2 border-[#B45309] flex items-center justify-between font-mono text-xs font-bold text-[#92400E]">
                <span>OPPONENT GOAL (0%)</span>
                <span className="bg-[#DCFCE7] text-[#15803D] px-2 py-0.5 rounded border border-[#86EFAC]">
                  TOUCHDOWN ZONE (90% - 100%)
                </span>
                <span>DESK EDGE</span>
              </div>

              {/* Midfield Yard Markers */}
              <div className="w-full flex items-center justify-between font-mono text-[10px] text-[#B45309]/60 px-4">
                <span>| 25%</span>
                <span>| 50% MIDFIELD</span>
                <span>| 75%</span>
              </div>

              {/* Origami Paper Football Indicator */}
              <div className="relative w-full h-16 flex items-center">
                <div
                  className="absolute transition-all duration-300 transform -translate-y-1/2 flex flex-col items-center cursor-grab active:cursor-grabbing"
                  style={{ left: `${Math.min(95, Math.max(5, ballPos))}%` }}
                >
                  <span className="text-3xl drop-shadow-md animate-bounce">📐</span>
                  <span className="text-[10px] font-mono font-bold bg-[#1E242B] text-white px-1.5 py-0.5 rounded-xs mt-1">
                    {ballPos.toFixed(1)}%
                  </span>
                </div>
              </div>

              {/* Own Goal Line */}
              <div className="w-full text-center pt-2 border-t-2 border-[#B45309] font-mono text-xs font-bold text-[#92400E]">
                OWN GOAL LINE (KICKOFF 20%)
              </div>
            </div>

            {/* Extra Point Uprights Display (When in Extra Point phase) */}
            {phase === 'extra_point' && (
              <div className="mt-4 p-4 bg-[#FEF9C3] rounded-xl border-2 border-[#FDE047] text-center">
                <p className="font-hand text-lg font-bold text-[#854D0E] mb-2">
                  \___/ KICK EXTRA POINT THROUGH OPPONENT'S FINGER GOALPOSTS! \___/
                </p>
                <Badge variant="green" size="sm">+1 Point on Target</Badge>
              </div>
            )}

            {/* Flick Controls Slider Bar */}
            <div className="mt-6 space-y-4 bg-white/90 p-4 rounded-xl border border-[#CBD5E1]">
              <div className="space-y-2">
                <div className="flex items-center justify-between text-xs font-mono font-bold text-[#1E242B]">
                  <span>FLICK POWER: {flickPower}%</span>
                  <span className="text-[#B45309]">{flickPower > 85 ? '⚠️ High (Risk Over-flick)' : 'Optimal'}</span>
                </div>
                <input
                  type="range"
                  min="10"
                  max="100"
                  value={flickPower}
                  onChange={(e) => setFlickPower(Number(e.target.value))}
                  className="w-full accent-[#1A365D] cursor-pointer"
                />
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between text-xs font-mono font-bold text-[#1E242B]">
                  <span>AIM ANGLE: {flickAngle}°</span>
                  <span>{flickAngle === 0 ? 'Straight' : flickAngle < 0 ? 'Left' : 'Right'}</span>
                </div>
                <input
                  type="range"
                  min="-40"
                  max="40"
                  value={flickAngle}
                  onChange={(e) => setFlickAngle(Number(e.target.value))}
                  className="w-full accent-[#991B1B] cursor-pointer"
                />
              </div>

              <Button
                onClick={handleExecuteFlick}
                disabled={(!isBotMode && !isMyTurn) || (isBotMode && !!botWinnerMsg)}
                variant="primary"
                size="lg"
                className="w-full"
                leftIcon={<Zap className="w-5 h-5" />}
              >
                {phase === 'extra_point' ? 'Kick Extra Point Through Uprights' : 'Flick Paper Football 📐'}
              </Button>
            </div>

            {/* Bottom Rematch Controls */}
            <div className="flex items-center gap-3 mt-4 w-full">
              {!isBotMode ? (
                gameState?.status === 'finished' ? (
                  <Button
                    onClick={handleRematch}
                    variant="primary"
                    size="md"
                    className="w-full"
                    leftIcon={<RotateCcw className="w-4 h-4" />}
                  >
                    Play Rematch Round
                  </Button>
                ) : (
                  <Button
                    onClick={() => toggleReady()}
                    variant={isReady ? 'stamp' : 'primary'}
                    size="md"
                    className="w-full"
                  >
                    {isReady ? '✓ Ready (Waiting for Opponent)' : 'Ready Up for Kickoff'}
                  </Button>
                )
              ) : (
                <Button
                  onClick={resetBotGame}
                  variant="secondary"
                  size="md"
                  className="w-full"
                  leftIcon={<RotateCcw className="w-4 h-4" />}
                >
                  Restart Practice
                </Button>
              )}
            </div>
          </PaperCard>
        </div>

        {/* Right: Classroom Bench & Desk Notes (5 cols) */}
        <div className="lg:col-span-5 space-y-4">
          {/* Players Bench */}
          <PaperCard variant="ruled" className="p-5">
            <div className="flex items-center justify-between pb-3 border-b border-[#CBD5E1] mb-3">
              <div className="flex items-center gap-2">
                <Users className="w-4 h-4 text-[#1A365D]" />
                <h3 className="font-bold text-sm text-[#1E242B]">Classroom Bench</h3>
              </div>
              <Stamp tone="blue">{members.length}/2 SEATED</Stamp>
            </div>

            <div className="space-y-3">
              {/* Player 1 (You) */}
              <div className="flex items-center justify-between p-2.5 rounded bg-white border border-[#CBD5E1] shadow-2xs">
                <div className="flex items-center gap-2.5">
                  <Avatar username={user?.username || 'You'} size="sm" isOnline />
                  <div>
                    <p className="text-xs font-bold text-[#1E242B]">
                      @{user?.username || 'You'} <span className="text-[#1A365D]">(You)</span>
                    </p>
                    <span className="font-hand text-sm text-[#1A365D] font-bold">
                      Score: {myScore} PTS
                    </span>
                  </div>
                </div>
                <Badge variant={isReady || isBotMode ? 'green' : 'default'} size="sm">
                  {isReady || isBotMode ? 'READY' : 'NOT READY'}
                </Badge>
              </div>

              {/* Player 2 (Opponent) */}
              <div className="flex items-center justify-between p-2.5 rounded bg-white border border-[#CBD5E1] shadow-2xs">
                <div className="flex items-center gap-2.5">
                  <Avatar
                    username={otherMember?.username || (isBotMode ? 'ClassBot' : 'Waiting...')}
                    size="sm"
                    isOnline={!!otherMember || isBotMode}
                  />
                  <div>
                    <p className="text-xs font-bold text-[#1E242B]">
                      {isBotMode ? 'ClassBot (AI)' : otherMember ? `@${otherMember.username}` : 'Empty Desk...'}
                    </p>
                    <span className="font-hand text-sm text-[#991B1B] font-bold">
                      Score: {oppScore} PTS
                    </span>
                  </div>
                </div>
                <Badge variant={isBotMode || otherMember?.is_ready ? 'green' : 'default'} size="sm">
                  {isBotMode || otherMember?.is_ready ? 'READY' : 'WAITING'}
                </Badge>
              </div>
            </div>
          </PaperCard>

          {/* Desk Notes Chat */}
          <PaperCard variant="plain" className="p-4 flex flex-col h-64">
            <h4 className="font-bold text-xs uppercase tracking-wider text-[#1E242B] font-mono mb-2 pb-1 border-b border-[#CBD5E1]">
              Pass Desk Notes:
            </h4>

            {/* Note log */}
            <div className="flex-1 overflow-y-auto space-y-2 pr-1 text-xs">
              {deskNotes.length === 0 ? (
                <p className="text-[#94A3B8] font-hand text-sm text-center py-6">
                  No notes passed yet. Whisper something to your benchmate!
                </p>
              ) : (
                deskNotes.map((note, idx) => (
                  <div key={idx} className="bg-[#FEF9C3] p-2 rounded border border-[#FDE047] text-left">
                    <div className="flex items-center justify-between text-[10px] text-[#854D0E] font-mono font-bold">
                      <span>@{note.sender}</span>
                      <span>{note.time}</span>
                    </div>
                    <p className="font-hand text-sm text-[#1E242B] mt-0.5">{note.text}</p>
                  </div>
                ))
              )}
            </div>

            {/* Note Input */}
            <form onSubmit={handleSendNote} className="flex gap-2 pt-2 mt-2 border-t border-[#CBD5E1]">
              <input
                type="text"
                placeholder="Whisper a quick note..."
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                className="flex-1 px-2.5 py-1.5 text-xs bg-white rounded border border-[#94A3B8] outline-none"
              />
              <Button type="submit" variant="primary" size="sm">
                <Send className="w-3.5 h-3.5" />
              </Button>
            </form>
          </PaperCard>
        </div>
      </div>
    </div>
  )
}
