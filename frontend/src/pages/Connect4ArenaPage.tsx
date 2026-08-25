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
  ChevronDown,
} from 'lucide-react'
import confetti from 'canvas-confetti'

interface CellPos {
  row: number
  col: number
}

interface Connect4BoardState {
  rows: number
  cols: number
  grid: string[][]
  player_colors: Record<string, string>
  last_move?: CellPos
  winning_cells?: CellPos[]
}

interface Connect4GameState {
  game_id: string
  game_type: string
  status: 'waiting' | 'active' | 'finished' | 'abandoned'
  current_turn: string
  move_count: number
  board_state: Connect4BoardState
  result?: {
    winner_id: string
    is_draw: boolean
    scores: Record<string, number>
    reason: string
  }
  version: number
}

export const Connect4ArenaPage: React.FC = () => {
  const { roomId } = useParams<{ roomId?: string }>()
  const { user } = useAuthStore()

  const targetRoom = roomId || 'RECESS-C4-DESK'
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

  const [gameState, setGameState] = useState<Connect4GameState | null>(null)
  const [chatInput, setChatInput] = useState('')
  const [deskNotes, setDeskNotes] = useState<Array<{ sender: string; text: string; time: string }>>([])
  const [hoveredCol, setHoveredCol] = useState<number | null>(null)

  // Solo Practice vs Bot State
  const [isBotMode, setIsBotMode] = useState(false)
  const [botGrid, setBotGrid] = useState<string[][]>(() =>
    Array(6).fill(null).map(() => Array(7).fill(''))
  )
  const [botTurn, setBotTurn] = useState<'player' | 'bot'>('player')
  const [botWinningCells, setBotWinningCells] = useState<CellPos[] | null>(null)
  const [botWinnerMsg, setBotWinnerMsg] = useState<string | null>(null)

  // Listen to WebSocket game.state updates
  useEffect(() => {
    if (events.length > 0) {
      const latest = events[0]
      if (latest.type === 'game.state' && latest.payload) {
        const state = latest.payload as Connect4GameState
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
  const handleDropDisc = (col: number) => {
    if (isBotMode) {
      handleBotMove(col)
      return
    }

    if (!gameState || gameState.status !== 'active') {
      if (members.length < 2) {
        toast.info('Waiting for Classmate', 'Invite a classmate to join this desk room to start.')
      } else if (!isReady) {
        toast.info('Ready Up to Play', 'Toggle "Ready Up" to begin the Connect 4 match.')
      }
      return
    }

    if (gameState.current_turn !== user?.id) {
      toast.error('Not Your Turn!', 'Wait for your classmate to drop their disc.')
      return
    }

    if (gameState.board_state?.grid?.[0]?.[col] !== '') {
      toast.error('Column Full!', 'Select a different column.')
      return
    }

    sendEvent('game.move', { col })
  }

  const handleRematch = () => {
    if (isBotMode) {
      resetBotGame()
      return
    }
    sendEvent('game.rematch', {})
    toast.success('Rematch Requested', 'Starting a fresh grid match!')
  }

  const handleSendNote = (e: React.FormEvent) => {
    e.preventDefault()
    if (!chatInput.trim()) return
    sendMessage(chatInput.trim())
    setChatInput('')
  }

  // Solo Bot Mechanics
  const handleBotMove = (col: number) => {
    if (botWinnerMsg || botTurn !== 'player') return
    if (botGrid[0][col] !== '') {
      toast.error('Column Full!', 'Pick an open column.')
      return
    }

    // Drop player disc
    const nextGrid = botGrid.map((r) => [...r])
    let targetRow = -1
    for (let r = 5; r >= 0; r--) {
      if (nextGrid[r][col] === '') {
        targetRow = r
        break
      }
    }
    if (targetRow === -1) return

    nextGrid[targetRow][col] = 'player'
    setBotGrid(nextGrid)

    // Check Win
    const winCells = checkConnect4Win(nextGrid, targetRow, col, 'player')
    if (winCells) {
      setBotWinningCells(winCells)
      setBotWinnerMsg('🏆 Victory! You connected 4 discs in a row!')
      confetti({ particleCount: 80, spread: 70 })
      return
    }

    // Check Draw
    if (isGridFull(nextGrid)) {
      setBotWinnerMsg('🤝 Board is Full! Match ended in a draw.')
      return
    }

    // Trigger Bot Turn
    setBotTurn('bot')
    setTimeout(() => {
      // Pick best column for bot
      const validCols: number[] = []
      for (let c = 0; c < 7; c++) {
        if (nextGrid[0][c] === '') validCols.push(c)
      }
      if (validCols.length === 0) return

      const chosenCol = validCols[Math.floor(Math.random() * validCols.length)]
      const botNext = nextGrid.map((r) => [...r])
      let botRow = -1
      for (let r = 5; r >= 0; r--) {
        if (botNext[r][chosenCol] === '') {
          botRow = r
          break
        }
      }
      if (botRow === -1) return

      botNext[botRow][chosenCol] = 'bot'
      setBotGrid(botNext)

      const botWin = checkConnect4Win(botNext, botRow, chosenCol, 'bot')
      if (botWin) {
        setBotWinningCells(botWin)
        setBotWinnerMsg('ClassBot connected 4 discs! Better luck next period.')
      } else if (isGridFull(botNext)) {
        setBotWinnerMsg('🤝 Board is Full! Match ended in a draw.')
      } else {
        setBotTurn('player')
      }
    }, 600)
  }

  const checkConnect4Win = (grid: string[][], r: number, c: number, id: string): CellPos[] | null => {
    const directions = [
      [0, 1],  // Horizontal (-)
      [1, 0],  // Vertical (|)
      [1, 1],  // Diagonal down-right (\)
      [1, -1], // Diagonal down-left (/)
    ]

    for (const [dr, dc] of directions) {
      const cells: CellPos[] = [{ row: r, col: c }]

      // Forward
      let step = 1
      while (true) {
        const nr = r + dr * step
        const nc = c + dc * step
        if (nr < 0 || nr >= 6 || nc < 0 || nc >= 7 || grid[nr][nc] !== id) break
        cells.push({ row: nr, col: nc })
        step++
      }

      // Reverse
      step = 1
      while (true) {
        const nr = r - dr * step
        const nc = c - dc * step
        if (nr < 0 || nr >= 6 || nc < 0 || nc >= 7 || grid[nr][nc] !== id) break
        cells.unshift({ row: nr, col: nc })
        step++
      }

      if (cells.length >= 4) return cells
    }
    return null
  }

  const isGridFull = (grid: string[][]): boolean => {
    for (let c = 0; c < 7; c++) {
      if (grid[0][c] === '') return false
    }
    return true
  }

  const resetBotGame = () => {
    setBotGrid(Array(6).fill(null).map(() => Array(7).fill('')))
    setBotTurn('player')
    setBotWinningCells(null)
    setBotWinnerMsg(null)
  }

  // Active Board Data
  const board = gameState?.board_state
  const grid = isBotMode ? botGrid : board?.grid || Array(6).fill(null).map(() => Array(7).fill(''))
  const winningCells = isBotMode ? botWinningCells : board?.winning_cells || null
  const otherMember = members.find((m) => m.user_id !== user?.id)
  const isMyTurn = isBotMode
    ? botTurn === 'player'
    : gameState?.status === 'active' && gameState?.current_turn === user?.id

  const isWinningCell = (r: number, c: number) => {
    return winningCells?.some((cell) => cell.row === r && cell.col === c)
  }

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16">
      {/* 1. Header Navigation */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b-2 border-[#1E242B] dark:border-slate-700 pb-4">
        <div className="flex items-center gap-3">
          <Link to="/games">
            <Button variant="secondary" size="sm" leftIcon={<ArrowLeft className="w-4 h-4" />}>
              Syllabus
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-extrabold text-[#1E242B] dark:text-slate-100">Connect 4 Arena</h1>
              <Stamp tone="amber">6x7 VERTICAL GRID</Stamp>
            </div>
            <p className="font-hand text-sm text-[#475569] dark:text-slate-300">
              Room: <span className="font-mono font-bold text-[#1A365D] dark:text-sky-300">{targetRoom}</span> • Drop Discs to Connect 4 in a Row!
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
                <span className="flex items-center gap-1 text-[#15803D] dark:text-emerald-300 bg-[#DCFCE7] dark:bg-emerald-950/60 px-2 py-1 rounded border border-[#86EFAC] dark:border-emerald-700/60">
                  <Wifi className="w-3.5 h-3.5" /> LIVE
                </span>
              ) : (
                <span className="flex items-center gap-1 text-[#991B1B] dark:text-red-300 bg-[#FFE4E6] dark:bg-red-950/60 px-2 py-1 rounded border border-[#FECDD3] dark:border-red-700/60">
                  <WifiOff className="w-3.5 h-3.5" /> OFFLINE
                </span>
              )}
            </div>
          )}
        </div>
      </div>

      {/* 2. Main Arena Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left: Connect 4 Board Canvas (7 cols) */}
        <div className="lg:col-span-7 space-y-4">
          {/* Turn Banner */}
          <PaperCard variant="sticky" stickyColor="yellow" showTape className="p-3 text-center">
            {isBotMode ? (
              <div className="font-hand text-lg font-bold text-[#1A365D] dark:text-amber-200">
                {botWinnerMsg
                  ? botWinnerMsg
                  : botTurn === 'player'
                  ? '🔵 Your Turn! Click a column to drop your Navy disc.'
                  : '⏳ ClassBot is planning its move...'}
              </div>
            ) : gameState?.status === 'active' ? (
              <div className="flex items-center justify-center gap-2">
                <span className="w-2.5 h-2.5 rounded-full bg-[#15803D] animate-ping" />
                <span className="font-hand text-xl font-bold text-[#1A365D] dark:text-amber-200">
                  {isMyTurn ? '🔵 Your Turn! Click any column arrow to drop disc.' : `⏳ Waiting for @${otherMember?.username || 'Opponent'}...`}
                </span>
              </div>
            ) : gameState?.status === 'finished' ? (
              <div className="font-hand text-xl font-bold text-[#991B1B] dark:text-red-300">
                {gameState.result?.is_draw
                  ? '🤝 Match Ended in a Draw (Board Full)!'
                  : gameState.result?.winner_id === user?.id
                  ? '🏆 VICTORY! 4 Discs Connected!'
                  : '🔴 Opponent Connected 4! Good game.'}
              </div>
            ) : (
              <div className="font-hand text-base text-[#475569] dark:text-amber-100">
                {members.length < 2
                  ? `Invite a classmate to join code "${targetRoom}"`
                  : 'Both players must toggle "Ready Up" to start.'}
              </div>
            )}
          </PaperCard>

          {/* Connect 4 Vertical Board Frame */}
          <PaperCard variant="plain" className="p-6 sm:p-8 flex flex-col items-center justify-center relative select-none bg-[#FDFBF7] dark:bg-slate-900/90">
            {/* Top Column Drop Indicators */}
            <div className="grid grid-cols-7 gap-2 sm:gap-3 w-full max-w-md mb-2">
              {Array.from({ length: 7 }).map((_, c) => (
                <button
                  key={`drop-btn-${c}`}
                  onClick={() => handleDropDisc(c)}
                  onMouseEnter={() => setHoveredCol(c)}
                  onMouseLeave={() => setHoveredCol(null)}
                  disabled={
                    grid[0]?.[c] !== '' ||
                    (!isBotMode && gameState?.status !== 'active') ||
                    (isBotMode && !!botWinnerMsg)
                  }
                  className={`h-8 rounded-t-lg flex items-center justify-center transition-all cursor-pointer ${
                    hoveredCol === c
                      ? 'bg-[#1A365D] dark:bg-sky-600 text-white shadow-xs -translate-y-0.5'
                      : 'bg-transparent text-[#94A3B8] hover:text-[#1A365D] dark:hover:text-sky-400'
                  }`}
                >
                  <ChevronDown className="w-5 h-5 animate-bounce" />
                </button>
              ))}
            </div>

            {/* 6x7 Grid Board Box */}
            <div className="bg-[#1A365D] dark:bg-[#0F2238] p-3 sm:p-4 rounded-2xl shadow-[6px_6px_0px_0px_#0F172A] dark:shadow-[6px_6px_0px_0px_#020617] border-4 border-[#0F172A] dark:border-slate-800 inline-block">
              <div className="grid grid-cols-7 gap-2 sm:gap-3">
                {Array.from({ length: 6 }).map((_, r) =>
                  Array.from({ length: 7 }).map((_, c) => {
                    const occupant = grid[r]?.[c]
                    const isPlayer = occupant === (isBotMode ? 'player' : user?.id)
                    const isOccupied = !!occupant
                    const isWin = isWinningCell(r, c)

                    return (
                      <div
                        key={`cell-${r}-${c}`}
                        onClick={() => handleDropDisc(c)}
                        onMouseEnter={() => setHoveredCol(c)}
                        onMouseLeave={() => setHoveredCol(null)}
                        className={`w-10 h-10 sm:w-12 sm:h-12 rounded-full cursor-pointer transition-all flex items-center justify-center relative overflow-hidden ${
                          isOccupied
                            ? isPlayer
                              ? 'bg-[#0284C7] shadow-[inset_0_3px_6px_rgba(255,255,255,0.4),0_3px_6px_rgba(0,0,0,0.3)]'
                              : 'bg-[#DC2626] shadow-[inset_0_3px_6px_rgba(255,255,255,0.4),0_3px_6px_rgba(0,0,0,0.3)]'
                            : 'bg-[#F8FAFC] dark:bg-slate-950 shadow-[inset_0_4px_6px_rgba(0,0,0,0.5)] hover:bg-[#E2E8F0] dark:hover:bg-slate-900'
                        } ${isWin ? 'ring-4 ring-[#FDE047] ring-offset-2 ring-offset-[#1A365D] animate-pulse scale-105' : ''}`}
                      >
                        {isOccupied && (
                          <div className="w-4 h-4 rounded-full border border-white/40 opacity-60" />
                        )}
                      </div>
                    )
                  })
                )}
              </div>
            </div>

            {/* Bottom Controls */}
            <div className="flex items-center gap-3 mt-6 w-full max-w-xs">
              {!isBotMode ? (
                gameState?.status === 'finished' ? (
                  <Button
                    onClick={handleRematch}
                    variant="primary"
                    size="md"
                    className="w-full"
                    leftIcon={<RotateCcw className="w-4 h-4" />}
                  >
                    Play Rematch
                  </Button>
                ) : (
                  <Button
                    onClick={() => toggleReady()}
                    variant={isReady ? 'stamp' : 'primary'}
                    size="md"
                    className="w-full"
                  >
                    {isReady ? '✓ Ready (Waiting)' : 'Ready Up'}
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
            <div className="flex items-center justify-between pb-3 border-b border-[#CBD5E1] dark:border-slate-700/60 mb-3">
              <div className="flex items-center gap-2">
                <Users className="w-4 h-4 text-[#1A365D] dark:text-sky-400" />
                <h3 className="font-bold text-sm text-[#1E242B] dark:text-slate-100">Classroom Bench</h3>
              </div>
              <Stamp tone="blue">{members.length}/2 SEATED</Stamp>
            </div>

            <div className="space-y-3">
              {/* Player 1 (You) */}
              <div className="flex items-center justify-between p-2.5 rounded bg-white dark:bg-slate-900/80 border border-[#CBD5E1] dark:border-slate-700 shadow-2xs">
                <div className="flex items-center gap-2.5">
                  <Avatar username={user?.username || 'You'} size="sm" isOnline />
                  <div>
                    <p className="text-xs font-bold text-[#1E242B] dark:text-slate-100">
                      @{user?.username || 'You'} <span className="text-[#1A365D] dark:text-sky-300">(You)</span>
                    </p>
                    <span className="font-hand text-sm text-[#0284C7] dark:text-sky-400 font-bold">
                      Discs: Blue Ink Disc
                    </span>
                  </div>
                </div>
                <Badge variant={isReady || isBotMode ? 'green' : 'default'} size="sm">
                  {isReady || isBotMode ? 'READY' : 'NOT READY'}
                </Badge>
              </div>

              {/* Player 2 (Opponent) */}
              <div className="flex items-center justify-between p-2.5 rounded bg-white dark:bg-slate-900/80 border border-[#CBD5E1] dark:border-slate-700 shadow-2xs">
                <div className="flex items-center gap-2.5">
                  <Avatar
                    username={otherMember?.username || (isBotMode ? 'ClassBot' : 'Waiting...')}
                    size="sm"
                    isOnline={!!otherMember || isBotMode}
                  />
                  <div>
                    <p className="text-xs font-bold text-[#1E242B] dark:text-slate-100">
                      {isBotMode ? 'ClassBot (AI)' : otherMember ? `@${otherMember.username}` : 'Empty Desk...'}
                    </p>
                    <span className="font-hand text-sm text-[#DC2626] dark:text-red-400 font-bold">
                      Discs: Red Ink Disc
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
            <h4 className="font-bold text-xs uppercase tracking-wider text-[#1E242B] dark:text-slate-100 font-mono mb-2 pb-1 border-b border-[#CBD5E1] dark:border-slate-700">
              Pass Desk Notes:
            </h4>

            {/* Note log */}
            <div className="flex-1 overflow-y-auto space-y-2 pr-1 text-xs">
              {deskNotes.length === 0 ? (
                <p className="text-[#94A3B8] dark:text-slate-400 font-hand text-sm text-center py-6">
                  No notes passed yet. Whisper something to your benchmate!
                </p>
              ) : (
                deskNotes.map((note, idx) => (
                  <div key={idx} className="bg-[#FEF9C3] dark:bg-amber-950/40 p-2 rounded border border-[#FDE047] dark:border-amber-700/60 text-left">
                    <div className="flex items-center justify-between text-[10px] text-[#854D0E] dark:text-amber-300 font-mono font-bold">
                      <span>@{note.sender}</span>
                      <span>{note.time}</span>
                    </div>
                    <p className="font-hand text-sm text-[#1E242B] dark:text-amber-100 mt-0.5">{note.text}</p>
                  </div>
                ))
              )}
            </div>

            {/* Note Input */}
            <form onSubmit={handleSendNote} className="flex gap-2 pt-2 mt-2 border-t border-[#CBD5E1] dark:border-slate-700">
              <input
                type="text"
                placeholder="Whisper a quick note..."
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                className="flex-1 px-2.5 py-1.5 text-xs bg-white dark:bg-slate-900 text-[#1E242B] dark:text-slate-100 rounded border border-[#94A3B8] dark:border-slate-700 outline-none"
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
