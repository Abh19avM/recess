import React, { useState, useEffect, useCallback } from 'react'
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
  Grid,
} from 'lucide-react'
import confetti from 'canvas-confetti'

interface EdgeKey {
  type: 'h' | 'v'
  row: number
  col: number
}

interface DotsBoxesBoardState {
  rows: number
  cols: number
  horizontal_edges: string[][]
  vertical_edges: string[][]
  boxes: string[][]
  player_initials: Record<string, string>
  player_colors: Record<string, string>
  scores: Record<string, number>
  last_edge?: EdgeKey
  completed_boxes: number
  total_boxes: number
}

interface DotsBoxesGameState {
  game_id: string
  game_type: string
  status: 'waiting' | 'active' | 'finished' | 'abandoned'
  current_turn: string
  move_count: number
  board_state: DotsBoxesBoardState
  result?: {
    winner_id: string
    is_draw: boolean
    scores: Record<string, number>
    reason: string
  }
  version: number
}

interface GridPreset {
  size: number
  name: string
  tag: string
  boxes: number
  dots: string
  desc: string
}

const GRID_PRESETS: GridPreset[] = [
  { size: 3, name: '3×3 Pocket', tag: 'QUICK', boxes: 9, dots: '4×4 Dots', desc: 'Fast 2-min classroom skirmish' },
  { size: 4, name: '4×4 Classic', tag: 'STANDARD', boxes: 16, dots: '5×5 Dots', desc: 'Balanced notebook margin battle' },
  { size: 5, name: '5×5 Large', tag: 'TACTICAL', boxes: 25, dots: '6×6 Dots', desc: 'Deep strategic graph paper match' },
  { size: 6, name: '6×6 Jumbo', tag: 'MASTER', boxes: 36, dots: '7×7 Dots', desc: 'Full classroom tournament grid' },
]

export const DotsBoxesArenaPage: React.FC = () => {
  const { roomId } = useParams<{ roomId?: string }>()
  const { user } = useAuthStore()

  const targetRoom = roomId || 'RECESS-DOTS-DESK'
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

  const [gameState, setGameState] = useState<DotsBoxesGameState | null>(null)
  const [chatInput, setChatInput] = useState('')
  const [deskNotes, setDeskNotes] = useState<Array<{ sender: string; text: string; time: string }>>([])
  const [selectedGridSize, setSelectedGridSize] = useState<number>(4)

  // Solo Bot Practice State
  const [isBotMode, setIsBotMode] = useState(false)
  const [botRows, setBotRows] = useState(4)
  const [botCols, setBotCols] = useState(4)
  const [botHEdges, setBotHEdges] = useState<string[][]>(() =>
    Array.from({ length: 5 }, () => Array(4).fill(''))
  )
  const [botVEdges, setBotVEdges] = useState<string[][]>(() =>
    Array.from({ length: 4 }, () => Array(5).fill(''))
  )
  const [botBoxes, setBotBoxes] = useState<string[][]>(() =>
    Array.from({ length: 4 }, () => Array(4).fill(''))
  )
  const [botTurn, setBotTurn] = useState<'player' | 'bot'>('player')
  const [botPlayerScore, setBotPlayerScore] = useState(0)
  const [botScore, setBotScore] = useState(0)
  const [botWinnerMsg, setBotWinnerMsg] = useState<string | null>(null)

  const resetBotGameForSize = useCallback((size: number) => {
    setBotRows(size)
    setBotCols(size)
    setBotHEdges(Array.from({ length: size + 1 }, () => Array(size).fill('')))
    setBotVEdges(Array.from({ length: size }, () => Array(size + 1).fill('')))
    setBotBoxes(Array.from({ length: size }, () => Array(size).fill('')))
    setBotPlayerScore(0)
    setBotScore(0)
    setBotTurn('player')
    setBotWinnerMsg(null)
  }, [])

  const handleSelectGridSize = (size: number) => {
    setSelectedGridSize(size)
    if (isBotMode) {
      resetBotGameForSize(size)
      toast.success(`Grid Updated to ${size}×${size}`, `${size * size} boxes ready to conquer!`)
    } else {
      sendEvent('game.config', { rows: size, cols: size })
      toast.info(`Grid Dimension Requested: ${size}×${size}`, 'Will be used for the next match round.')
    }
  }

  // Listen to WebSocket game.state updates
  useEffect(() => {
    if (events.length > 0) {
      const latest = events[0]
      if (latest.type === 'game.state' && latest.payload) {
        const state = latest.payload as DotsBoxesGameState
        setGameState(state)

        if (state.board_state?.rows) {
          setSelectedGridSize(state.board_state.rows)
        }

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
  const handleDrawEdge = (type: 'h' | 'v', row: number, col: number) => {
    if (isBotMode) {
      handleBotEdgeClick(type, row, col)
      return
    }

    if (!gameState || gameState.status !== 'active') {
      if (members.length < 2) {
        toast.info('Waiting for Classmate', 'Invite a classmate to join this desk room to start.')
      } else if (!isReady) {
        toast.info('Ready Up to Play', 'Toggle "Ready Up" to begin the Dots & Boxes duel.')
      }
      return
    }

    if (gameState.current_turn !== user?.id) {
      toast.error('Not Your Turn!', 'Wait for your classmate to draw their line.')
      return
    }

    sendEvent('game.move', { type, row, col })
  }

  const handleRematch = () => {
    if (isBotMode) {
      resetBotGameForSize(selectedGridSize)
      return
    }
    sendEvent('game.rematch', { config: { rows: selectedGridSize, cols: selectedGridSize } })
    toast.success('Rematch Requested', `Starting a fresh ${selectedGridSize}×${selectedGridSize} grid match!`)
  }

  const handleSendNote = (e: React.FormEvent) => {
    e.preventDefault()
    if (!chatInput.trim()) return
    sendMessage(chatInput.trim())
    setChatInput('')
  }

  // Offline Bot Logic
  const handleBotEdgeClick = (type: 'h' | 'v', row: number, col: number) => {
    if (botWinnerMsg || botTurn !== 'player') return

    const newH = botHEdges.map((r) => [...r])
    const newV = botVEdges.map((r) => [...r])
    const newB = botBoxes.map((r) => [...r])

    if (type === 'h') {
      if (newH[row][col]) return
      newH[row][col] = 'player'
    } else {
      if (newV[row][col]) return
      newV[row][col] = 'player'
    }

    // Check completed boxes
    let newlyClosed = 0
    for (let r = 0; r < botRows; r++) {
      for (let c = 0; c < botCols; c++) {
        if (!newB[r][c] && newH[r][c] && newH[r + 1][c] && newV[r][c] && newV[r][c + 1]) {
          newB[r][c] = 'player'
          newlyClosed++
        }
      }
    }

    const nextPlayerScore = botPlayerScore + newlyClosed
    setBotHEdges(newH)
    setBotVEdges(newV)
    setBotBoxes(newB)
    setBotPlayerScore(nextPlayerScore)

    const totalClaimed = nextPlayerScore + botScore
    if (totalClaimed >= botRows * botCols) {
      finishBotGame(nextPlayerScore, botScore)
      return
    }

    if (newlyClosed > 0) {
      toast.success('Bonus Turn!', 'You closed a box! Take another line.')
    } else {
      setBotTurn('bot')
      triggerBotMove(newH, newV, newB, nextPlayerScore, botScore)
    }
  }

  const triggerBotMove = (
    curH: string[][],
    curV: string[][],
    curB: string[][],
    pScore: number,
    bScore: number
  ) => {
    setTimeout(() => {
      const openEdges: Array<{ type: 'h' | 'v'; row: number; col: number }> = []

      for (let r = 0; r <= botRows; r++) {
        for (let c = 0; c < botCols; c++) {
          if (!curH[r][c]) openEdges.push({ type: 'h', row: r, col: c })
        }
      }
      for (let r = 0; r < botRows; r++) {
        for (let c = 0; c <= botCols; c++) {
          if (!curV[r][c]) openEdges.push({ type: 'v', row: r, col: c })
        }
      }

      if (openEdges.length === 0) return

      // Pick random open edge
      const chosen = openEdges[Math.floor(Math.random() * openEdges.length)]
      const nextH = curH.map((r) => [...r])
      const nextV = curV.map((r) => [...r])
      const nextB = curB.map((r) => [...r])

      if (chosen.type === 'h') {
        nextH[chosen.row][chosen.col] = 'bot'
      } else {
        nextV[chosen.row][chosen.col] = 'bot'
      }

      let botClosed = 0
      for (let r = 0; r < botRows; r++) {
        for (let c = 0; c < botCols; c++) {
          if (!nextB[r][c] && nextH[r][c] && nextH[r + 1][c] && nextV[r][c] && nextV[r][c + 1]) {
            nextB[r][c] = 'bot'
            botClosed++
          }
        }
      }

      const nextBotScore = bScore + botClosed
      setBotHEdges(nextH)
      setBotVEdges(nextV)
      setBotBoxes(nextB)
      setBotScore(nextBotScore)

      const totalClaimed = pScore + nextBotScore
      if (totalClaimed >= botRows * botCols) {
        finishBotGame(pScore, nextBotScore)
        return
      }

      if (botClosed > 0) {
        triggerBotMove(nextH, nextV, nextB, pScore, nextBotScore)
      } else {
        setBotTurn('player')
      }
    }, 450)
  }

  const finishBotGame = (pScore: number, bScore: number) => {
    if (pScore > bScore) {
      setBotWinnerMsg(`You Won! Claimed ${pScore} vs ${bScore} boxes.`)
      confetti({ particleCount: 80, spread: 70 })
    } else if (bScore > pScore) {
      setBotWinnerMsg(`ClassBot Won! Claimed ${bScore} vs ${pScore} boxes.`)
    } else {
      setBotWinnerMsg('Match Ended in a Draw!')
    }
  }

  // Active Board Data
  const board = gameState?.board_state
  const rows = isBotMode ? botRows : board?.rows || selectedGridSize
  const cols = isBotMode ? botCols : board?.cols || selectedGridSize
  const hEdges = isBotMode ? botHEdges : board?.horizontal_edges || []
  const vEdges = isBotMode ? botVEdges : board?.vertical_edges || []
  const boxes = isBotMode ? botBoxes : board?.boxes || []
  const isMyTurn = isBotMode
    ? botTurn === 'player'
    : gameState?.status === 'active' && gameState?.current_turn === user?.id

  const otherMember = members.find((m) => m.user_id !== user?.id)
  const myScore = isBotMode ? botPlayerScore : (user && board?.scores?.[user.id]) || 0
  const oppScore = isBotMode ? botScore : (otherMember && board?.scores?.[otherMember.user_id]) || 0

  // Adaptive Sizing depending on grid density
  const getDimensionStyles = (dimension: number) => {
    if (dimension >= 6) {
      return {
        box: 'w-10 sm:w-12 h-10 sm:h-12 text-xl sm:text-2xl',
        hEdge: 'h-2 w-10 sm:w-12',
        vEdge: 'w-2 h-10 sm:h-12',
        dot: 'w-2.5 h-2.5',
      }
    }
    if (dimension === 5) {
      return {
        box: 'w-12 sm:w-14 h-12 sm:h-14 text-2xl sm:text-3xl',
        hEdge: 'h-2.5 w-12 sm:w-14',
        vEdge: 'w-2.5 h-12 sm:h-14',
        dot: 'w-3 h-3',
      }
    }
    if (dimension === 4) {
      return {
        box: 'w-14 sm:w-16 h-14 sm:h-16 text-3xl sm:text-4xl',
        hEdge: 'h-2.5 w-14 sm:w-16',
        vEdge: 'w-2.5 h-14 sm:h-16',
        dot: 'w-3.5 h-3.5',
      }
    }
    // 3x3
    return {
      box: 'w-16 sm:w-20 h-16 sm:h-20 text-3xl sm:text-4xl',
      hEdge: 'h-2.5 w-16 sm:w-20',
      vEdge: 'w-2.5 h-16 sm:h-20',
      dot: 'w-3.5 h-3.5',
    }
  }

  const dim = getDimensionStyles(cols)

  return (
    <div className="space-y-6 max-w-5xl mx-auto pb-16">
      {/* 1. Header Navigation */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b-2 border-[#1E242B] dark:border-[#334155] pb-4">
        <div className="flex items-center gap-3">
          <Link to="/games">
            <Button variant="secondary" size="sm" leftIcon={<ArrowLeft className="w-4 h-4" />}>
              Syllabus
            </Button>
          </Link>
          <div>
            <div className="flex items-center gap-2">
              <h1 className="text-2xl font-extrabold text-[#1E242B] dark:text-[#F8FAFC]">Dots & Boxes Arena</h1>
              <Stamp tone="blue">NOTEBOOK GRAPH</Stamp>
            </div>
            <p className="font-hand text-sm text-[#475569] dark:text-[#94A3B8]">
              Room: <span className="font-mono font-bold text-[#1A365D] dark:text-[#60A5FA]">{targetRoom}</span> • {rows}×{cols} Grid ({rows * cols} Boxes) • Close Boxes for Bonus Turns!
            </p>
          </div>
        </div>

        {/* Solo Practice Mode Toggle */}
        <div className="flex items-center gap-2">
          <Button
            onClick={() => {
              setIsBotMode(!isBotMode)
              if (!isBotMode) resetBotGameForSize(selectedGridSize)
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
                <span className="flex items-center gap-1 text-[#15803D] dark:text-[#86EFAC] bg-[#DCFCE7] dark:bg-[#063323] px-2 py-1 rounded border border-[#86EFAC] dark:border-[#065F46]">
                  <Wifi className="w-3.5 h-3.5" /> LIVE
                </span>
              ) : (
                <span className="flex items-center gap-1 text-[#991B1B] dark:text-[#FCA5A5] bg-[#FFE4E6] dark:bg-[#3B1123] px-2 py-1 rounded border border-[#FECDD3] dark:border-[#9D174D]">
                  <WifiOff className="w-3.5 h-3.5" /> OFFLINE
                </span>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Grid Size Selection Toolbar */}
      <PaperCard variant="plain" className="p-3.5 flex flex-col md:flex-row md:items-center justify-between gap-3">
        <div className="flex items-center gap-2">
          <Grid className="w-4 h-4 text-[#1A365D] dark:text-[#60A5FA]" />
          <span className="text-xs font-bold font-mono uppercase tracking-wider text-[#1E242B] dark:text-[#F8FAFC]">
            Grid Dimension:
          </span>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 flex-1 max-w-2xl">
          {GRID_PRESETS.map((preset) => {
            const isSelected = selectedGridSize === preset.size
            return (
              <button
                key={preset.size}
                type="button"
                onClick={() => handleSelectGridSize(preset.size)}
                className={`px-3 py-2 rounded-lg border-2 text-left transition-all cursor-pointer ${
                  isSelected
                    ? 'border-[#1A365D] dark:border-[#38BDF8] bg-[#E0F2FE] dark:bg-[#0B2545] shadow-[2px_2px_0px_0px_#1A365D] dark:shadow-[2px_2px_0px_0px_#020617]'
                    : 'border-[#CBD5E1] dark:border-[#334155] bg-white dark:bg-[#141C2E] hover:bg-slate-50 dark:hover:bg-[#1E293B]'
                }`}
              >
                <div className="flex items-center justify-between">
                  <span className="font-bold text-xs text-[#1E242B] dark:text-[#F8FAFC]">{preset.name}</span>
                  <span className="text-[9px] font-mono font-semibold text-[#15803D] dark:text-[#86EFAC]">{preset.boxes} B</span>
                </div>
                <p className="text-[10px] text-[#475569] dark:text-[#94A3B8] font-hand leading-tight mt-0.5 truncate">{preset.desc}</p>
              </button>
            )
          })}
        </div>
      </PaperCard>

      {/* 2. Main Arena Layout */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left: Interactive Graph Board (7 cols) */}
        <div className="lg:col-span-7 space-y-4">
          {/* Turn Banner */}
          <PaperCard variant="sticky" stickyColor="blue" showTape className="p-3 text-center">
            {isBotMode ? (
              <div className="font-hand text-lg font-bold text-[#1A365D] dark:text-[#93C5FD]">
                {botWinnerMsg
                  ? botWinnerMsg
                  : botTurn === 'player'
                  ? '✎ Your Turn! Draw a line between two dots.'
                  : '⏳ ClassBot is drawing a line...'}
              </div>
            ) : gameState?.status === 'active' ? (
              <div className="flex items-center justify-center gap-2">
                <span className="w-2.5 h-2.5 rounded-full bg-[#15803D] dark:bg-[#4ADE80] animate-ping" />
                <span className="font-hand text-xl font-bold text-[#1A365D] dark:text-[#93C5FD]">
                  {isMyTurn ? '✎ Your Turn! Draw a pencil line.' : `⏳ Waiting for @${otherMember?.username || 'Opponent'}...`}
                </span>
              </div>
            ) : gameState?.status === 'finished' ? (
              <div className="font-hand text-xl font-bold text-[#991B1B] dark:text-[#F87171]">
                {gameState.result?.is_draw
                  ? '🤝 Match Ended in a Draw (Equal Boxes Claimed)!'
                  : gameState.result?.winner_id === user?.id
                  ? '🏆 Victory! You claimed the most boxes!'
                  : '✎ Match Concluded. Better luck next period!'}
              </div>
            ) : (
              <div className="font-hand text-base text-[#475569] dark:text-[#94A3B8]">
                {members.length < 2
                  ? `Invite a classmate to join code "${targetRoom}"`
                  : 'Both players must toggle "Ready Up" to start.'}
              </div>
            )}
          </PaperCard>

          {/* Graph Paper Dots Arena */}
          <PaperCard variant="graph" className="p-4 sm:p-6 flex flex-col items-center justify-center relative select-none overflow-x-auto">
            {/* Score Tracker Bar */}
            <div className="flex items-center justify-between w-full max-w-md mb-5 px-4 py-2 rounded-lg bg-white/90 dark:bg-[#101726]/90 border border-[#CBD5E1] dark:border-[#334155] shadow-2xs font-mono font-bold text-sm">
              <span className="text-[#1A365D] dark:text-[#60A5FA]">You: {myScore} Boxes</span>
              <span className="text-xs text-[#475569] dark:text-[#94A3B8]">Total {rows * cols}</span>
              <span className="text-[#991B1B] dark:text-[#F87171]">Opponent: {oppScore} Boxes</span>
            </div>

            {/* Dots Grid Canvas */}
            <div className="relative inline-block p-2 sm:p-4 bg-white/50 dark:bg-[#0B0F19]/60 rounded-xl border border-[#E2E8F0] dark:border-[#1E293B]">
              {Array.from({ length: rows }).map((_, r) => (
                <div key={r} className="flex flex-col">
                  {/* Row of Dots and Horizontal Edges */}
                  <div className="flex items-center">
                    {Array.from({ length: cols }).map((_, c) => {
                      const isClaimed = hEdges[r]?.[c]
                      const isPlayer = isClaimed === (isBotMode ? 'player' : user?.id)
                      return (
                        <React.Fragment key={`h-${r}-${c}`}>
                          {/* Dot */}
                          <div className={`${dim.dot} rounded-full bg-[#1E242B] dark:bg-[#F8FAFC] border-2 border-white dark:border-[#334155] shadow-xs z-10`} />

                          {/* Horizontal Edge Button */}
                          <button
                            onClick={() => handleDrawEdge('h', r, c)}
                            disabled={!!isClaimed || (!isBotMode && gameState?.status !== 'active') || (isBotMode && !!botWinnerMsg)}
                            className={`${dim.hEdge} transition-all rounded-xs cursor-pointer ${
                              isClaimed
                                ? isPlayer
                                  ? 'bg-[#1A365D] dark:bg-[#38BDF8] shadow-[0_1px_3px_rgba(26,54,93,0.4)] dark:shadow-[0_0_8px_rgba(56,189,248,0.5)]'
                                  : 'bg-[#991B1B] dark:bg-[#F87171] shadow-[0_1px_3px_rgba(153,27,27,0.4)] dark:shadow-[0_0_8px_rgba(248,113,113,0.5)]'
                                : 'bg-[#E2E8F0] dark:bg-[#334155] hover:bg-[#94A3B8]/80 dark:hover:bg-[#60A5FA]/60 active:scale-95'
                            }`}
                          />
                        </React.Fragment>
                      )
                    })}
                    {/* Last dot in row */}
                    <div className={`${dim.dot} rounded-full bg-[#1E242B] dark:bg-[#F8FAFC] border-2 border-white dark:border-[#334155] shadow-xs z-10`} />
                  </div>

                  {/* Vertical Edges and Box Centers */}
                  <div className="flex items-center">
                    {Array.from({ length: cols }).map((_, c) => {
                      const vLeft = vEdges[r]?.[c]
                      const isVLeftPlayer = vLeft === (isBotMode ? 'player' : user?.id)
                      const boxOwner = boxes[r]?.[c]
                      const isBoxPlayer = boxOwner === (isBotMode ? 'player' : user?.id)

                      return (
                        <React.Fragment key={`v-b-${r}-${c}`}>
                          {/* Vertical Edge Left */}
                          <button
                            onClick={() => handleDrawEdge('v', r, c)}
                            disabled={!!vLeft || (!isBotMode && gameState?.status !== 'active') || (isBotMode && !!botWinnerMsg)}
                            className={`${dim.vEdge} transition-all rounded-xs cursor-pointer ${
                              vLeft
                                ? isVLeftPlayer
                                  ? 'bg-[#1A365D] dark:bg-[#38BDF8] shadow-[1px_0_3px_rgba(26,54,93,0.4)] dark:shadow-[0_0_8px_rgba(56,189,248,0.5)]'
                                  : 'bg-[#991B1B] dark:bg-[#F87171] shadow-[1px_0_3px_rgba(153,27,27,0.4)] dark:shadow-[0_0_8px_rgba(248,113,113,0.5)]'
                                : 'bg-[#E2E8F0] dark:bg-[#334155] hover:bg-[#94A3B8]/80 dark:hover:bg-[#60A5FA]/60 active:scale-95'
                            }`}
                          />

                          {/* Box Inner Area */}
                          <div
                            className={`${dim.box} flex items-center justify-center font-hand font-bold transition-all duration-200 ${
                              boxOwner
                                ? isBoxPlayer
                                  ? 'bg-[#E0F2FE]/80 dark:bg-[#0B2545]/80 text-[#1A365D] dark:text-[#38BDF8] animate-in zoom-in-75'
                                  : 'bg-[#FFE4E6]/80 dark:bg-[#3B1123]/80 text-[#991B1B] dark:text-[#F87171] animate-in zoom-in-75'
                                : 'bg-transparent'
                            }`}
                          >
                            {boxOwner && (
                              <span className="drop-shadow-xs select-none">
                                {isBoxPlayer ? (user?.username?.[0]?.toUpperCase() || 'A') : (otherMember?.username?.[0]?.toUpperCase() || 'B')}
                              </span>
                            )}
                          </div>
                        </React.Fragment>
                      )
                    })}

                    {/* Rightmost Vertical Edge */}
                    {(() => {
                      const vRight = vEdges[r]?.[cols]
                      const isVRightPlayer = vRight === (isBotMode ? 'player' : user?.id)
                      return (
                        <button
                          onClick={() => handleDrawEdge('v', r, cols)}
                          disabled={!!vRight || (!isBotMode && gameState?.status !== 'active') || (isBotMode && !!botWinnerMsg)}
                          className={`${dim.vEdge} transition-all rounded-xs cursor-pointer ${
                            vRight
                              ? isVRightPlayer
                                ? 'bg-[#1A365D] dark:bg-[#38BDF8] shadow-[1px_0_3px_rgba(26,54,93,0.4)] dark:shadow-[0_0_8px_rgba(56,189,248,0.5)]'
                                : 'bg-[#991B1B] dark:bg-[#F87171] shadow-[1px_0_3px_rgba(153,27,27,0.4)] dark:shadow-[0_0_8px_rgba(248,113,113,0.5)]'
                              : 'bg-[#E2E8F0] dark:bg-[#334155] hover:bg-[#94A3B8]/80 dark:hover:bg-[#60A5FA]/60 active:scale-95'
                          }`}
                        />
                      )
                    })()}
                  </div>
                </div>
              ))}

              {/* Bottommost Row of Dots and Horizontal Edges */}
              <div className="flex items-center">
                {Array.from({ length: cols }).map((_, c) => {
                  const isClaimed = hEdges[rows]?.[c]
                  const isPlayer = isClaimed === (isBotMode ? 'player' : user?.id)
                  return (
                    <React.Fragment key={`bottom-h-${c}`}>
                      <div className={`${dim.dot} rounded-full bg-[#1E242B] dark:bg-[#F8FAFC] border-2 border-white dark:border-[#334155] shadow-xs z-10`} />
                      <button
                        onClick={() => handleDrawEdge('h', rows, c)}
                        disabled={!!isClaimed || (!isBotMode && gameState?.status !== 'active') || (isBotMode && !!botWinnerMsg)}
                        className={`${dim.hEdge} transition-all rounded-xs cursor-pointer ${
                          isClaimed
                            ? isPlayer
                              ? 'bg-[#1A365D] dark:bg-[#38BDF8] shadow-[0_1px_3px_rgba(26,54,93,0.4)] dark:shadow-[0_0_8px_rgba(56,189,248,0.5)]'
                              : 'bg-[#991B1B] dark:bg-[#F87171] shadow-[0_1px_3px_rgba(153,27,27,0.4)] dark:shadow-[0_0_8px_rgba(248,113,113,0.5)]'
                            : 'bg-[#E2E8F0] dark:bg-[#334155] hover:bg-[#94A3B8]/80 dark:hover:bg-[#60A5FA]/60 active:scale-95'
                        }`}
                      />
                    </React.Fragment>
                  )
                })}
                <div className={`${dim.dot} rounded-full bg-[#1E242B] dark:bg-[#F8FAFC] border-2 border-white dark:border-[#334155] shadow-xs z-10`} />
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
                    Play Rematch ({selectedGridSize}×{selectedGridSize})
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
                  onClick={() => resetBotGameForSize(selectedGridSize)}
                  variant="secondary"
                  size="md"
                  className="w-full"
                  leftIcon={<RotateCcw className="w-4 h-4" />}
                >
                  Restart Practice ({selectedGridSize}×{selectedGridSize})
                </Button>
              )}
            </div>
          </PaperCard>
        </div>

        {/* Right: Roster & Desk Notes (5 cols) */}
        <div className="lg:col-span-5 space-y-4">
          {/* Players Roster */}
          <PaperCard variant="ruled" className="p-5">
            <div className="flex items-center justify-between pb-3 border-b border-[#CBD5E1] dark:border-[#334155] mb-3">
              <div className="flex items-center gap-2">
                <Users className="w-4 h-4 text-[#1A365D] dark:text-[#60A5FA]" />
                <h3 className="font-bold text-sm text-[#1E242B] dark:text-[#F8FAFC]">Classroom Bench</h3>
              </div>
              <Stamp tone="blue">{members.length}/2 SEATED</Stamp>
            </div>

            <div className="space-y-3">
              {/* Player 1 (You) */}
              <div className="flex items-center justify-between p-2.5 rounded bg-white dark:bg-[#101726] border border-[#CBD5E1] dark:border-[#334155] shadow-2xs">
                <div className="flex items-center gap-2.5">
                  <Avatar username={user?.username || 'You'} size="sm" isOnline />
                  <div>
                    <p className="text-xs font-bold text-[#1E242B] dark:text-[#F8FAFC]">
                      @{user?.username || 'You'} <span className="text-[#1A365D] dark:text-[#60A5FA]">(You)</span>
                    </p>
                    <span className="font-hand text-sm text-[#1A365D] dark:text-[#60A5FA] font-bold">
                      Lines: Navy Pencil • Score: {myScore}
                    </span>
                  </div>
                </div>
                <Badge variant={isReady || isBotMode ? 'green' : 'default'} size="sm">
                  {isReady || isBotMode ? 'READY' : 'NOT READY'}
                </Badge>
              </div>

              {/* Player 2 (Opponent) */}
              <div className="flex items-center justify-between p-2.5 rounded bg-white dark:bg-[#101726] border border-[#CBD5E1] dark:border-[#334155] shadow-2xs">
                <div className="flex items-center gap-2.5">
                  <Avatar
                    username={otherMember?.username || (isBotMode ? 'ClassBot' : 'Waiting...')}
                    size="sm"
                    isOnline={!!otherMember || isBotMode}
                  />
                  <div>
                    <p className="text-xs font-bold text-[#1E242B] dark:text-[#F8FAFC]">
                      {isBotMode ? 'ClassBot (AI)' : otherMember ? `@${otherMember.username}` : 'Empty Desk...'}
                    </p>
                    <span className="font-hand text-sm text-[#991B1B] dark:text-[#F87171] font-bold">
                      Lines: Red Pen • Score: {oppScore}
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
            <h4 className="font-bold text-xs uppercase tracking-wider text-[#1E242B] dark:text-[#F8FAFC] font-mono mb-2 pb-1 border-b border-[#CBD5E1] dark:border-[#334155]">
              Pass Desk Notes:
            </h4>

            {/* Note log */}
            <div className="flex-1 overflow-y-auto space-y-2 pr-1 text-xs">
              {deskNotes.length === 0 ? (
                <p className="text-[#94A3B8] dark:text-[#64748B] font-hand text-sm text-center py-6">
                  No notes passed yet. Whisper something to your benchmate!
                </p>
              ) : (
                deskNotes.map((note, idx) => (
                  <div key={idx} className="bg-[#FEF9C3] dark:bg-[#2D2106] p-2 rounded border border-[#FDE047] dark:border-[#854D0E] text-left">
                    <div className="flex items-center justify-between text-[10px] text-[#854D0E] dark:text-[#FEF08A] font-mono font-bold">
                      <span>@{note.sender}</span>
                      <span>{note.time}</span>
                    </div>
                    <p className="font-hand text-sm text-[#1E242B] dark:text-[#F8FAFC] mt-0.5">{note.text}</p>
                  </div>
                ))
              )}
            </div>

            {/* Note Input */}
            <form onSubmit={handleSendNote} className="flex gap-2 pt-2 mt-2 border-t border-[#CBD5E1] dark:border-[#334155]">
              <input
                type="text"
                placeholder="Whisper a quick note..."
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                className="flex-1 px-2.5 py-1.5 text-xs bg-white dark:bg-[#101726] rounded border border-[#94A3B8] dark:border-[#475569] text-[#1E242B] dark:text-[#F8FAFC] outline-none"
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
