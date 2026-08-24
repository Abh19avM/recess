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
} from 'lucide-react'
import confetti from 'canvas-confetti'

interface XOBoradState {
  grid: string[][]
  player_symbols: Record<string, string> // PlayerID -> "X" | "O"
  winning_line?: number[][]
}

interface XOGameState {
  game_id: string
  game_type: string
  status: 'waiting' | 'active' | 'finished' | 'abandoned'
  current_turn: string
  move_count: number
  board_state: XOBoradState
  result?: {
    winner_id: string
    is_draw: boolean
    scores: Record<string, number>
    reason: string
  }
  version: number
}

export const XOArenaPage: React.FC = () => {
  const { roomId } = useParams<{ roomId?: string }>()
  const { user } = useAuthStore()

  const targetRoom = roomId || 'RECESS-XO-DESK'
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

  const [gameState, setGameState] = useState<XOGameState | null>(null)
  const [chatInput, setChatInput] = useState('')
  const [deskNotes, setDeskNotes] = useState<Array<{ sender: string; text: string; time: string }>>([])
  const [isBotMode, setIsBotMode] = useState(false)
  const [botGrid, setBotGrid] = useState<string[][]>([
    ['', '', ''],
    ['', '', ''],
    ['', '', ''],
  ])
  const [botTurn, setBotTurn] = useState<'X' | 'O'>('X')
  const [botWinner, setBotWinner] = useState<string | null>(null)
  const [botWinLine, setBotWinLine] = useState<number[][] | null>(null)

  // Parse game.state events from WebSocket stream
  useEffect(() => {
    if (events.length > 0) {
      const latest = events[0]
      if (latest.type === 'game.state' && latest.payload) {
        setGameState(latest.payload as XOGameState)

        // Trigger confetti on player victory
        const state = latest.payload as XOGameState
        if (state.status === 'finished' && state.result?.winner_id === user?.id) {
          confetti({
            particleCount: 80,
            spread: 70,
            origin: { y: 0.6 },
            colors: ['#1A365D', '#15803D', '#B45309', '#FEF08A'],
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

  // Join room on mount
  useEffect(() => {
    if (status === 'OPEN' && currentRoom !== targetRoom) {
      joinRoom(targetRoom)
    }
  }, [status, targetRoom, currentRoom, joinRoom])

  // Handle Multiplayer Move
  const handleCellClick = (row: number, col: number) => {
    if (isBotMode) {
      handleBotMove(row, col)
      return
    }

    if (!gameState || gameState.status !== 'active') {
      if (members.length < 2) {
        toast.info('Waiting for Classmate', 'Invite a classmate to join this room code to start.')
      } else if (!isReady) {
        toast.info('Mark Yourself Ready', 'Click "Ready Up" to start the match.')
      }
      return
    }

    if (gameState.current_turn !== user?.id) {
      toast.error('Not Your Turn!', 'Wait for your classmate to place their mark.')
      return
    }

    if (gameState.board_state?.grid?.[row]?.[col]) {
      toast.error('Cell Occupied', 'That square is already marked.')
      return
    }

    sendEvent('game.move', { row, col })
  }

  // Handle Rematch Request
  const handleRematch = () => {
    if (isBotMode) {
      resetBotGame()
      return
    }
    sendEvent('game.rematch', {})
    toast.success('Rematch Requested', 'Starting a fresh classroom duel!')
  }

  // Handle Desk Note Submission
  const handleSendNote = (e: React.FormEvent) => {
    e.preventDefault()
    if (!chatInput.trim()) return
    sendMessage(chatInput.trim())
    setChatInput('')
  }

  // Bot Practice Engine (Client-side offline mode)
  const handleBotMove = (row: number, col: number) => {
    if (botWinner || botGrid[row][col] !== '' || botTurn !== 'X') return

    const newGrid = botGrid.map((r) => [...r])
    newGrid[row][col] = 'X'
    setBotGrid(newGrid)

    const winCheck = checkLocalWin(newGrid, 'X')
    if (winCheck) {
      setBotWinner('You Win!')
      setBotWinLine(winCheck)
      confetti({ particleCount: 60, spread: 60 })
      return
    }

    if (newGrid.flat().every((c) => c !== '')) {
      setBotWinner('Draw Game!')
      return
    }

    setBotTurn('O')

    // AI bot responds after short natural delay
    setTimeout(() => {
      const emptyCells: [number, number][] = []
      for (let r = 0; r < 3; r++) {
        for (let c = 0; c < 3; c++) {
          if (newGrid[r][c] === '') emptyCells.push([r, c])
        }
      }

      if (emptyCells.length > 0) {
        // Pick center if open, otherwise random available cell
        const chosen =
          emptyCells.find(([r, c]) => r === 1 && c === 1) ||
          emptyCells[Math.floor(Math.random() * emptyCells.length)]

        newGrid[chosen[0]][chosen[1]] = 'O'
        setBotGrid([...newGrid])

        const botWin = checkLocalWin(newGrid, 'O')
        if (botWin) {
          setBotWinner('Class Bot Wins!')
          setBotWinLine(botWin)
        } else if (newGrid.flat().every((c) => c !== '')) {
          setBotWinner('Draw Game!')
        } else {
          setBotTurn('X')
        }
      }
    }, 450)
  }

  const resetBotGame = () => {
    setBotGrid([
      ['', '', ''],
      ['', '', ''],
      ['', '', ''],
    ])
    setBotTurn('X')
    setBotWinner(null)
    setBotWinLine(null)
  }

  const checkLocalWin = (grid: string[][], sym: string): number[][] | null => {
    for (let r = 0; r < 3; r++) {
      if (grid[r][0] === sym && grid[r][1] === sym && grid[r][2] === sym)
        return [[r, 0], [r, 1], [r, 2]]
    }
    for (let c = 0; c < 3; c++) {
      if (grid[0][c] === sym && grid[1][c] === sym && grid[2][c] === sym)
        return [[0, c], [1, c], [2, c]]
    }
    if (grid[0][0] === sym && grid[1][1] === sym && grid[2][2] === sym)
      return [[0, 0], [1, 1], [2, 2]]
    if (grid[0][2] === sym && grid[1][1] === sym && grid[2][0] === sym)
      return [[0, 2], [1, 1], [2, 0]]
    return null
  }

  // Active board extraction
  const grid = isBotMode
    ? botGrid
    : gameState?.board_state?.grid || [
        ['', '', ''],
        ['', '', ''],
        ['', '', ''],
      ]

  const mySymbol = isBotMode
    ? 'X'
    : (user && gameState?.board_state?.player_symbols?.[user.id]) || 'X'

  const isMyTurn = isBotMode
    ? botTurn === 'X'
    : gameState?.status === 'active' && gameState?.current_turn === user?.id

  const winningLine = isBotMode ? botWinLine : gameState?.board_state?.winning_line

  const isCellInWinningLine = (r: number, c: number) => {
    if (!winningLine) return false
    return winningLine.some(([wr, wc]) => wr === r && wc === c)
  }

  const otherMember = members.find((m) => m.user_id !== user?.id)

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16">
      {/* 1. Header Navigation & Room Status */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b-2 border-[#1E242B] pb-4">
        <div className="flex items-center gap-3">
          <Link to="/games">
            <Button variant="secondary" size="sm" leftIcon={<ArrowLeft className="w-4 h-4" />}>
              Syllabus
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-extrabold text-[#1E242B]">XO / Tic-Tac-Toe Arena</h1>
              <Stamp tone="red">MARGIN DUEL</Stamp>
            </div>
            <p className="font-hand text-sm text-[#475569]">
              Room: <span className="font-mono font-bold text-[#1A365D]">{targetRoom}</span> • 3x3 Speed Mark
            </p>
          </div>
        </div>

        {/* Mode Toggle & WSS Status */}
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
                  <WifiOff className="w-3.5 h-3.5" /> RECONNECTING
                </span>
              )}
            </div>
          )}
        </div>
      </div>

      {/* 2. Main Arena & Side Panel */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left Side: XO Gameboard (7 cols) */}
        <div className="lg:col-span-7 space-y-4">
          {/* Turn Banner */}
          <PaperCard variant="sticky" stickyColor="yellow" showTape className="p-3 text-center">
            {isBotMode ? (
              <div className="font-hand text-lg font-bold text-[#1A365D]">
                {botWinner
                  ? botWinner
                  : botTurn === 'X'
                  ? '✎ Your Turn (Mark X with Pencil)'
                  : '⏳ Class Bot is thinking...'}
              </div>
            ) : gameState?.status === 'active' ? (
              <div className="flex items-center justify-center gap-2">
                <span className="w-2.5 h-2.5 rounded-full bg-[#15803D] animate-ping" />
                <span className="font-hand text-xl font-bold text-[#1A365D]">
                  {isMyTurn ? '✎ Your Turn! Place your mark.' : `⏳ Waiting for @${otherMember?.username || 'Opponent'}...`}
                </span>
              </div>
            ) : gameState?.status === 'finished' ? (
              <div className="font-hand text-xl font-bold text-[#991B1B]">
                {gameState.result?.is_draw
                  ? '🤝 Match Ended in a Draw (Cat’s Game)!'
                  : gameState.result?.winner_id === user?.id
                  ? '🏆 Victory! You claimed 3 in a row!'
                  : '✎ Match Concluded. Better luck next period!'}
              </div>
            ) : (
              <div className="font-hand text-base text-[#475569]">
                {members.length < 2
                  ? `Invite a classmate to join code "${targetRoom}"`
                  : 'Both players must toggle "Ready Up" to start.'}
              </div>
            )}
          </PaperCard>

          {/* Graph Paper 3x3 Arena */}
          <PaperCard
            variant="graph"
            className="p-6 sm:p-8 flex flex-col items-center justify-center relative select-none"
          >
            {/* 3x3 Tic Tac Toe Grid */}
            <div className="grid grid-cols-3 gap-3 w-72 sm:w-80 h-72 sm:h-80 relative">
              {grid.map((rowArr, r) =>
                rowArr.map((cell, c) => {
                  const isWinningCell = isCellInWinningLine(r, c)
                  return (
                    <button
                      key={`${r}-${c}`}
                      onClick={() => handleCellClick(r, c)}
                      disabled={
                        (!isBotMode && gameState?.status !== 'active') ||
                        (isBotMode && !!botWinner)
                      }
                      className={`relative flex items-center justify-center rounded-lg border-2 text-5xl sm:text-6xl font-bold font-hand transition-all duration-150 active:scale-95 ${
                        isWinningCell
                          ? 'bg-[#FEF08A] border-[#EAB308] text-[#15803D] shadow-[3px_3px_0px_0px_#B45309] animate-bounce'
                          : cell === 'X'
                          ? 'bg-[#FFFFFF] border-[#1A365D] text-[#1A365D] shadow-[3px_3px_0px_0px_#1A365D]'
                          : cell === 'O'
                          ? 'bg-[#FFFFFF] border-[#991B1B] text-[#991B1B] shadow-[3px_3px_0px_0px_#991B1B]'
                          : 'bg-[#FBF9F3]/90 border-[#94A3B8] hover:border-[#1A365D] hover:bg-white shadow-[2px_2px_0px_0px_#CBD5E1] cursor-pointer'
                      }`}
                    >
                      {cell === 'X' && (
                        <span className="scale-in-center animate-in zoom-in-75 duration-100">
                          ✕
                        </span>
                      )}
                      {cell === 'O' && (
                        <span className="scale-in-center animate-in zoom-in-75 duration-100">
                          ◯
                        </span>
                      )}
                    </button>
                  )
                })
              )}
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
                    Play Rematch Round
                  </Button>
                ) : (
                  <Button
                    onClick={() => toggleReady()}
                    variant={isReady ? 'stamp' : 'primary'}
                    size="md"
                    className="w-full"
                  >
                    {isReady ? '✓ Ready (Waiting for Opponent)' : 'Ready Up to Play'}
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

        {/* Right Side: Players & Desk Notes (5 cols) */}
        <div className="lg:col-span-5 space-y-4">
          {/* Players Roster */}
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
                      Assigned: {mySymbol} (Pencil)
                    </span>
                  </div>
                </div>
                <Badge variant={isReady ? 'green' : 'default'} size="sm">
                  {isReady ? 'READY' : 'NOT READY'}
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
                      {isBotMode
                        ? 'ClassBot (AI)'
                        : otherMember
                        ? `@${otherMember.username}`
                        : 'Empty Desk...'}
                    </p>
                    <span className="font-hand text-sm text-[#991B1B] font-bold">
                      Assigned: {mySymbol === 'X' ? 'O' : 'X'} (Ballpoint)
                    </span>
                  </div>
                </div>
                <Badge
                  variant={isBotMode || otherMember?.is_ready ? 'green' : 'default'}
                  size="sm"
                >
                  {isBotMode || otherMember?.is_ready ? 'READY' : 'WAITING'}
                </Badge>
              </div>
            </div>
          </PaperCard>

          {/* Desk Notes / Quick Chat */}
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
