import React, { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useWebSocket } from '../hooks/useWebSocket'
import { useAuthStore } from '../store/authStore'
import { PaperCard } from '../components/ui/PaperCard'
import { Button } from '../components/ui/Button'
import { Badge, Stamp } from '../components/ui/Badge'
import { Avatar } from '../components/ui/Avatar'
import { Input } from '../components/ui/Input'
import { toast } from '../store/toastStore'
import {
  RotateCcw,
  Sparkles,
  Users,
  Send,
  ArrowLeft,
  Wifi,
  WifiOff,
  CheckCircle,
  Hand,
} from 'lucide-react'
import confetti from 'canvas-confetti'

interface CategoryEntries {
  name: string
  place: string
  animal: string
  thing: string
}

interface RoundScore {
  name_points: number
  place_points: number
  animal_points: number
  thing_points: number
  total_points: number
  entries: CategoryEntries
}

interface NPATBoardState {
  current_round: number
  total_rounds: number
  current_letter: string
  stop_caller_id?: string
  submitted_players: string[]
  cumulative_scores: Record<string, number>
  last_round_scores?: Record<string, RoundScore>
  phase: 'writing' | 'round_summary' | 'finished'
}

interface NPATGameState {
  game_id: string
  game_type: string
  status: 'waiting' | 'active' | 'finished' | 'abandoned'
  current_turn: string
  move_count: number
  board_state: NPATBoardState
  result?: {
    winner_id: string
    is_draw: boolean
    scores: Record<string, number>
    reason: string
  }
  version: number
}

const botSampleDatabase: Record<string, CategoryEntries> = {
  S: { name: 'Sam', place: 'Sydney', animal: 'Snake', thing: 'Spoon' },
  A: { name: 'Alice', place: 'Austin', animal: 'Alligator', thing: 'Apple' },
  M: { name: 'Max', place: 'Madrid', animal: 'Monkey', thing: 'Mirror' },
  R: { name: 'Ryan', place: 'Rome', animal: 'Rabbit', thing: 'Ruler' },
  P: { name: 'Peter', place: 'Paris', animal: 'Penguin', thing: 'Pencil' },
  B: { name: 'Bob', place: 'Boston', animal: 'Bear', thing: 'Book' },
  C: { name: 'Charlie', place: 'Cairo', animal: 'Cat', thing: 'Clock' },
  D: { name: 'David', place: 'Dallas', animal: 'Dog', thing: 'Desk' },
  T: { name: 'Tom', place: 'Tokyo', animal: 'Tiger', thing: 'Table' },
  L: { name: 'Leo', place: 'London', animal: 'Lion', thing: 'Lamp' },
}

export const NPATArenaPage: React.FC = () => {
  const { roomId } = useParams<{ roomId?: string }>()
  const { user } = useAuthStore()

  const targetRoom = roomId || 'RECESS-NPAT-DESK'
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

  const [gameState, setGameState] = useState<NPATGameState | null>(null)
  const [chatInput, setChatInput] = useState('')
  const [deskNotes, setDeskNotes] = useState<Array<{ sender: string; text: string; time: string }>>([])

  // Category Inputs
  const [nameVal, setNameVal] = useState('')
  const [placeVal, setPlaceVal] = useState('')
  const [animalVal, setAnimalVal] = useState('')
  const [thingVal, setThingVal] = useState('')
  const [isSubmitted, setIsSubmitted] = useState(false)

  // Solo Bot State
  const [isBotMode, setIsBotMode] = useState(false)
  const [botRound, setBotRound] = useState(1)
  const [botLetter, setBotLetter] = useState('S')
  const [botPlayerScore, setBotPlayerScore] = useState(0)
  const [botScore, setBotScore] = useState(0)
  const [botLastRound, setBotLastRound] = useState<{ player: RoundScore; bot: RoundScore } | null>(null)
  const [botWinnerMsg, setBotWinnerMsg] = useState<string | null>(null)

  // Listen to WebSocket game.state updates
  useEffect(() => {
    if (events.length > 0) {
      const latest = events[0]
      if (latest.type === 'game.state' && latest.payload) {
        const state = latest.payload as NPATGameState
        setGameState(state)

        // Reset inputs on round change
        if (state.board_state?.current_letter !== gameState?.board_state?.current_letter) {
          setNameVal('')
          setPlaceVal('')
          setAnimalVal('')
          setThingVal('')
          setIsSubmitted(false)
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
  }, [events, user, gameState])

  useEffect(() => {
    if (status === 'OPEN' && currentRoom !== targetRoom) {
      joinRoom(targetRoom)
    }
  }, [status, targetRoom, currentRoom, joinRoom])

  // Move Submissions
  const handleSubmitEntries = (callStop: boolean) => {
    if (isBotMode) {
      handleBotRoundSubmit(callStop)
      return
    }

    if (!gameState || gameState.status !== 'active') {
      if (members.length < 2) {
        toast.info('Waiting for Classmates', 'Invite classmates to join this desk room to start.')
      } else if (!isReady) {
        toast.info('Ready Up to Play', 'Toggle "Ready Up" to begin Name–Place–Animal–Thing.')
      }
      return
    }

    const action = callStop ? 'call_stop' : 'submit_entries'
    sendEvent('game.move', {
      name: nameVal.trim(),
      place: placeVal.trim(),
      animal: animalVal.trim(),
      thing: thingVal.trim(),
      action,
    })

    setIsSubmitted(true)
    if (callStop) {
      toast.success('STOP Called!', 'You called STOP! Scoring round entries...')
    } else {
      toast.info('Entries Submitted', 'Waiting for classmates to finish writing.')
    }
  }

  const handleRematch = () => {
    if (isBotMode) {
      resetBotGame()
      return
    }
    sendEvent('game.rematch', {})
    toast.success('Rematch Requested', 'Starting a fresh 3-round match!')
  }

  const handleSendNote = (e: React.FormEvent) => {
    e.preventDefault()
    if (!chatInput.trim()) return
    sendMessage(chatInput.trim())
    setChatInput('')
  }

  // Offline Bot Logic
  const handleBotRoundSubmit = (callStop: boolean) => {
    if (botWinnerMsg) return

    const targetLetter = botLetter
    const botAns = botSampleDatabase[targetLetter] || {
      name: targetLetter + 'am',
      place: targetLetter + 'pain',
      animal: targetLetter + 'eal',
      thing: targetLetter + 'poon',
    }

    // Score player
    const scoreField = (word: string, botWord: string) => {
      const clean = word.trim().toUpperCase()
      if (clean.length < 2 || !clean.startsWith(targetLetter)) return 0
      if (clean === botWord.toUpperCase()) return 5
      return 10
    }

    const pName = scoreField(nameVal, botAns.name)
    const pPlace = scoreField(placeVal, botAns.place)
    const pAnimal = scoreField(animalVal, botAns.animal)
    const pThing = scoreField(thingVal, botAns.thing)
    const pTotal = pName + pPlace + pAnimal + pThing

    const bName = scoreField(botAns.name, nameVal)
    const bPlace = scoreField(botAns.place, placeVal)
    const bAnimal = scoreField(botAns.animal, animalVal)
    const bThing = scoreField(botAns.thing, thingVal)
    const bTotal = bName + bPlace + bAnimal + bThing

    const nextPScore = botPlayerScore + pTotal
    const nextBScore = botScore + bTotal

    setBotPlayerScore(nextPScore)
    setBotScore(nextBScore)
    setBotLastRound({
      player: {
        name_points: pName,
        place_points: pPlace,
        animal_points: pAnimal,
        thing_points: pThing,
        total_points: pTotal,
        entries: { name: nameVal, place: placeVal, animal: animalVal, thing: thingVal },
      },
      bot: {
        name_points: bName,
        place_points: bPlace,
        animal_points: bAnimal,
        thing_points: bThing,
        total_points: bTotal,
        entries: botAns,
      },
    })

    if (botRound >= 3) {
      if (nextPScore > nextBScore) {
        setBotWinnerMsg(`🏆 Victory! You won ${nextPScore} to ${nextBScore} points!`)
        confetti({ particleCount: 80, spread: 70 })
      } else if (nextBScore > nextPScore) {
        setBotWinnerMsg(`ClassBot won ${nextBScore} to ${nextPScore} points!`)
      } else {
        setBotWinnerMsg('Match ended in a tie!')
      }
    } else {
      const letters = ['S', 'A', 'M', 'R', 'P']
      const nextR = botRound + 1
      setBotRound(nextR)
      setBotLetter(letters[nextR - 1])
      setNameVal('')
      setPlaceVal('')
      setAnimalVal('')
      setThingVal('')
      setIsSubmitted(false)
      toast.success(callStop ? 'STOP! Round Scored' : 'Round Scored', `Scored +${pTotal} points. Next Letter: ${letters[nextR - 1]}`)
    }
  }

  const resetBotGame = () => {
    setBotRound(1)
    setBotLetter('S')
    setBotPlayerScore(0)
    setBotScore(0)
    setBotLastRound(null)
    setBotWinnerMsg(null)
    setNameVal('')
    setPlaceVal('')
    setAnimalVal('')
    setThingVal('')
    setIsSubmitted(false)
  }

  // Active Board Data
  const board = gameState?.board_state
  const currentLetter = isBotMode ? botLetter : board?.current_letter || 'S'
  const currentRound = isBotMode ? botRound : board?.current_round || 1
  const totalRounds = isBotMode ? 3 : board?.total_rounds || 3
  const otherMember = members.find((m) => m.user_id !== user?.id)
  const myScore = isBotMode ? botPlayerScore : (user && board?.cumulative_scores?.[user.id]) || 0
  const oppScore = isBotMode ? botScore : (otherMember && board?.cumulative_scores?.[otherMember.user_id]) || 0

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
              <h1 className="text-2xl font-extrabold text-[#1E242B]">Name–Place–Animal–Thing</h1>
              <Stamp tone="blue">NOTEBOOK LEDGER</Stamp>
            </div>
            <p className="font-hand text-sm text-[#475569]">
              Room: <span className="font-mono font-bold text-[#1A365D]">{targetRoom}</span> • Round {currentRound} of {totalRounds} • Unique Word = 10 Pts, Shared = 5 Pts
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
        {/* Left: Chalkboard Letter Announcement & Notebook Column Table (7 cols) */}
        <div className="lg:col-span-7 space-y-4">
          {/* Chalkboard Letter Card */}
          <PaperCard variant="chalkboard" className="p-6 text-center text-white relative overflow-hidden">
            <div className="flex items-center justify-between text-xs font-mono text-white/70 mb-2">
              <span>ROUND {currentRound} OF {totalRounds}</span>
              <span>4 CATEGORIES • 40 MAX PTS</span>
            </div>

            <div className="py-2">
              <span className="text-xs font-mono uppercase tracking-widest text-[#86EFAC] block">
                TARGET ALPHABET LETTER
              </span>
              <span className="font-serif text-6xl sm:text-7xl font-black text-[#FDE047] drop-shadow-md tracking-wider">
                {currentLetter}
              </span>
            </div>

            <p className="font-hand text-base text-white/90 mt-1">
              Write words starting with letter <span className="font-bold underline text-[#FDE047]">{currentLetter}</span> for all 4 columns!
            </p>
          </PaperCard>

          {/* Notebook Column Sheet */}
          <PaperCard variant="ruled" className="p-6 space-y-5">
            {/* Score HUD */}
            <div className="flex items-center justify-between bg-white/90 p-3 rounded-lg border border-[#CBD5E1] font-mono font-bold text-sm">
              <span className="text-[#1A365D]">You: {myScore} PTS</span>
              <span className="text-[#475569]">vs</span>
              <span className="text-[#991B1B]">Opponent: {oppScore} PTS</span>
            </div>

            {/* Inputs Grid */}
            <div className="space-y-4">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div>
                  <label className="block text-xs font-mono font-bold uppercase text-[#1E242B] mb-1">
                    1. Name (Person):
                  </label>
                  <Input
                    placeholder={`e.g. ${currentLetter}arah`}
                    value={nameVal}
                    onChange={(e) => setNameVal(e.target.value)}
                    disabled={isSubmitted || (isBotMode && !!botWinnerMsg)}
                    className="font-hand text-lg"
                  />
                </div>

                <div>
                  <label className="block text-xs font-mono font-bold uppercase text-[#1E242B] mb-1">
                    2. Place (City/Country):
                  </label>
                  <Input
                    placeholder={`e.g. ${currentLetter}eattle`}
                    value={placeVal}
                    onChange={(e) => setPlaceVal(e.target.value)}
                    disabled={isSubmitted || (isBotMode && !!botWinnerMsg)}
                    className="font-hand text-lg"
                  />
                </div>

                <div>
                  <label className="block text-xs font-mono font-bold uppercase text-[#1E242B] mb-1">
                    3. Animal:
                  </label>
                  <Input
                    placeholder={`e.g. ${currentLetter}nake`}
                    value={animalVal}
                    onChange={(e) => setAnimalVal(e.target.value)}
                    disabled={isSubmitted || (isBotMode && !!botWinnerMsg)}
                    className="font-hand text-lg"
                  />
                </div>

                <div>
                  <label className="block text-xs font-mono font-bold uppercase text-[#1E242B] mb-1">
                    4. Thing (Object):
                  </label>
                  <Input
                    placeholder={`e.g. ${currentLetter}cissors`}
                    value={thingVal}
                    onChange={(e) => setThingVal(e.target.value)}
                    disabled={isSubmitted || (isBotMode && !!botWinnerMsg)}
                    className="font-hand text-lg"
                  />
                </div>
              </div>
            </div>

            {/* Action Buttons: Submit or Call STOP */}
            <div className="flex flex-col sm:flex-row gap-3 pt-3">
              <Button
                onClick={() => handleSubmitEntries(false)}
                disabled={isSubmitted || (isBotMode && !!botWinnerMsg)}
                variant="secondary"
                size="lg"
                className="flex-1"
                leftIcon={<CheckCircle className="w-4 h-4" />}
              >
                {isSubmitted ? 'Submitted (Waiting...)' : 'Submit Words'}
              </Button>

              <Button
                onClick={() => handleSubmitEntries(true)}
                disabled={isSubmitted || (isBotMode && !!botWinnerMsg)}
                variant="stamp"
                size="lg"
                className="flex-1"
                leftIcon={<Hand className="w-5 h-5" />}
              >
                CALL STOP! ✋
              </Button>
            </div>

            {/* Last Round Score Breakdown Box */}
            {isBotMode && botLastRound && (
              <div className="p-3 bg-[#FEF9C3] rounded-lg border border-[#FDE047] text-xs font-mono">
                <div className="flex items-center justify-between font-bold text-[#854D0E] mb-1">
                  <span>PREVIOUS ROUND BREAKDOWN:</span>
                  <span>+{botLastRound.player.total_points} PTS</span>
                </div>
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 text-center text-[11px] mt-1">
                  <div className="bg-white p-1 rounded border border-[#FDE047]">
                    <span className="block text-[#64748B]">Name</span>
                    <span className="font-bold text-[#15803D]">+{botLastRound.player.name_points}</span>
                  </div>
                  <div className="bg-white p-1 rounded border border-[#FDE047]">
                    <span className="block text-[#64748B]">Place</span>
                    <span className="font-bold text-[#15803D]">+{botLastRound.player.place_points}</span>
                  </div>
                  <div className="bg-white p-1 rounded border border-[#FDE047]">
                    <span className="block text-[#64748B]">Animal</span>
                    <span className="font-bold text-[#15803D]">+{botLastRound.player.animal_points}</span>
                  </div>
                  <div className="bg-white p-1 rounded border border-[#FDE047]">
                    <span className="block text-[#64748B]">Thing</span>
                    <span className="font-bold text-[#15803D]">+{botLastRound.player.thing_points}</span>
                  </div>
                </div>
              </div>
            )}

            {/* Rematch Controls */}
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
                    Play Rematch Match
                  </Button>
                ) : (
                  <Button
                    onClick={() => toggleReady()}
                    variant={isReady ? 'stamp' : 'primary'}
                    size="md"
                    className="w-full"
                  >
                    {isReady ? '✓ Ready (Waiting for Opponent)' : 'Ready Up for Round 1'}
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

        {/* Right: Score Breakdown & Classroom Bench (5 cols) */}
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
                      Total: {myScore} PTS
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
                      Total: {oppScore} PTS
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
