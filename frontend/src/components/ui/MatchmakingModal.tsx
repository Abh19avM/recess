import React, { useState, useEffect, useRef } from 'react'
import { useNavigate } from 'react-router-dom'
import { Modal } from './Modal'
import { Button } from './Button'
import { Badge, Stamp } from './Badge'
import { Avatar } from './Avatar'
import { api } from '../../lib/api'
import { toast } from '../../store/toastStore'
import { Swords, X, Clock } from 'lucide-react'
import confetti from 'canvas-confetti'

interface MatchmakingModalProps {
  isOpen: boolean
  onClose: () => void
  gameId: string
  gameTitle: string
  gameIcon?: string
}

interface MatchResult {
  match_id: string
  room_id: string
  game_type: string
  mode: 'casual' | 'ranked'
  player1: { user_id: string; username: string; rating: number }
  player2: { user_id: string; username: string; rating: number }
}

export const MatchmakingModal: React.FC<MatchmakingModalProps> = ({
  isOpen,
  onClose,
  gameId,
  gameTitle,
  gameIcon = '🎮',
}) => {
  const navigate = useNavigate()
  const [mode, setMode] = useState<'casual' | 'ranked'>('casual')
  const [isSearching, setIsSearching] = useState(false)
  const [searchSeconds, setSearchSeconds] = useState(0)
  const [matchedGame, setMatchedGame] = useState<MatchResult | null>(null)
  const [countdown, setCountdown] = useState<number | null>(null)

  const pollIntervalRef = useRef<number | null>(null)
  const timerIntervalRef = useRef<number | null>(null)

  // Start Searching Queue
  const handleStartSearch = async () => {
    try {
      setIsSearching(true)
      setSearchSeconds(0)
      setMatchedGame(null)

      // Map game IDs to canonical backend game identifiers
      let backendGameType = gameId
      if (gameId === 'tic_tac_toe') backendGameType = 'xo'
      if (gameId === 'hand_cricket') backendGameType = 'hand_cricket'
      if (gameId === 'dots_boxes') backendGameType = 'dots_boxes'
      if (gameId === 'connect_4') backendGameType = 'connect4'
      if (gameId === 'paper_football') backendGameType = 'paper_football'
      if (gameId === 'name_place_animal_thing' || gameId === 'npat') backendGameType = 'npat'

      const res = await api.post<{
        status: 'queued' | 'matched'
        ticket_id: string
        match?: MatchResult
      }>('/matchmaking/join', {
        game_type: backendGameType,
        mode,
      })

      if (res.status === 'matched' && res.match) {
        handleMatchFound(res.match)
      } else {
        // Start polling status
        startPolling(res.ticket_id)
      }

      // Start elapsed timer
      timerIntervalRef.current = window.setInterval(() => {
        setSearchSeconds((s) => s + 1)
      }, 1000)
    } catch (err: any) {
      setIsSearching(false)
      toast.error('Queue Failed', err.message || 'Could not join matchmaking queue')
    }
  }

  // Poll for Match Result
  const startPolling = (tktId: string) => {
    if (pollIntervalRef.current) clearInterval(pollIntervalRef.current)

    pollIntervalRef.current = window.setInterval(async () => {
      try {
        const res = await api.get<{
          status: 'queued' | 'matched'
          ticket_id: string
          match?: MatchResult
        }>(`/matchmaking/status?ticket_id=${tktId}`)

        if (res.status === 'matched' && res.match) {
          handleMatchFound(res.match)
        }
      } catch {
        // Continue polling
      }
    }, 1500)
  }

  // Handle Match Found
  const handleMatchFound = (match: MatchResult) => {
    if (pollIntervalRef.current) clearInterval(pollIntervalRef.current)
    if (timerIntervalRef.current) clearInterval(timerIntervalRef.current)

    setMatchedGame(match)
    setIsSearching(false)
    confetti({ particleCount: 60, spread: 60 })

    // 3-second countdown to join room
    setCountdown(3)
    let cd = 3
    const cdInterval = window.setInterval(() => {
      cd--
      setCountdown(cd)
      if (cd <= 0) {
        clearInterval(cdInterval)
        navigateToArena(match.room_id)
      }
    }, 1000)
  }

  const navigateToArena = (roomId: string) => {
    onClose()
    if (gameId === 'tic_tac_toe' || gameId === 'xo') {
      navigate(`/games/xo/${roomId}`)
    } else if (gameId === 'hand_cricket') {
      navigate(`/games/hand-cricket/${roomId}`)
    } else if (gameId === 'dots_boxes') {
      navigate(`/games/dots-and-boxes/${roomId}`)
    } else if (gameId === 'connect_4' || gameId === 'connect4') {
      navigate(`/games/connect-4/${roomId}`)
    } else if (gameId === 'paper_football') {
      navigate(`/games/paper-football/${roomId}`)
    } else if (gameId === 'name_place_animal_thing' || gameId === 'npat') {
      navigate(`/games/npat/${roomId}`)
    } else {
      navigate(`/dashboard`)
    }
  }

  // Cancel Queue
  const handleCancelSearch = async () => {
    try {
      if (pollIntervalRef.current) clearInterval(pollIntervalRef.current)
      if (timerIntervalRef.current) clearInterval(timerIntervalRef.current)

      let backendGameType = gameId
      if (gameId === 'tic_tac_toe') backendGameType = 'xo'

      await api.post('/matchmaking/leave', { game_type: backendGameType })
    } catch {
      // Handled
    } finally {
      setIsSearching(false)
      setSearchSeconds(0)
    }
  }

  useEffect(() => {
    return () => {
      if (pollIntervalRef.current) clearInterval(pollIntervalRef.current)
      if (timerIntervalRef.current) clearInterval(timerIntervalRef.current)
    }
  }, [])

  return (
    <Modal
      isOpen={isOpen}
      onClose={() => {
        if (isSearching) handleCancelSearch()
        onClose()
      }}
      title={`Classroom Matchmaking: ${gameTitle}`}
    >
      <div className="space-y-6">
        {!isSearching && !matchedGame ? (
          /* Step 1: Mode Selection & Queue Start */
          <div className="space-y-5">
            <div className="p-4 rounded-xl bg-[#FEF9C3] dark:bg-[#2D2106] border border-[#FDE047] dark:border-[#854D0E] flex items-center gap-4">
              <span className="text-4xl">{gameIcon}</span>
              <div>
                <h4 className="font-bold text-base text-[#1E242B] dark:text-[#F8FAFC]">{gameTitle}</h4>
                <p className="font-hand text-sm text-[#854D0E] dark:text-[#FEF08A]">
                  Find an online classmate for a live 1v1 duel across the school desks!
                </p>
              </div>
            </div>

            {/* Mode Selector */}
            <div className="space-y-2">
              <label className="text-xs font-bold font-mono uppercase text-[#475569] dark:text-[#94A3B8]">
                Select Play Mode:
              </label>
              <div className="grid grid-cols-2 gap-3">
                <button
                  type="button"
                  onClick={() => setMode('casual')}
                  className={`p-3 rounded-lg border-2 text-left transition-all ${
                    mode === 'casual'
                      ? 'border-[#1A365D] dark:border-[#3B82F6] bg-[#E0F2FE] dark:bg-[#0B2545] shadow-[2px_2px_0px_0px_#1A365D] dark:shadow-[2px_2px_0px_0px_#020617]'
                      : 'border-[#CBD5E1] dark:border-[#334155] bg-white dark:bg-[#101726] hover:bg-slate-50 dark:hover:bg-[#1E293B]'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span className="font-bold text-sm text-[#1E242B] dark:text-[#F8FAFC]">Casual Match</span>
                    <Stamp tone="green">QUICK</Stamp>
                  </div>
                  <p className="text-xs text-[#475569] dark:text-[#94A3B8] font-hand mt-1">Instant pairing without rating stakes.</p>
                </button>

                <button
                  type="button"
                  onClick={() => setMode('ranked')}
                  className={`p-3 rounded-lg border-2 text-left transition-all ${
                    mode === 'ranked'
                      ? 'border-[#991B1B] dark:border-[#EF4444] bg-[#FFE4E6] dark:bg-[#3B1123] shadow-[2px_2px_0px_0px_#991B1B] dark:shadow-[2px_2px_0px_0px_#020617]'
                      : 'border-[#CBD5E1] dark:border-[#334155] bg-white dark:bg-[#101726] hover:bg-slate-50 dark:hover:bg-[#1E293B]'
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span className="font-bold text-sm text-[#1E242B] dark:text-[#F8FAFC]">Ranked Duel</span>
                    <Stamp tone="red">ELO</Stamp>
                  </div>
                  <p className="text-xs text-[#475569] dark:text-[#94A3B8] font-hand mt-1">Climb the classroom leaderboard report card.</p>
                </button>
              </div>
            </div>

            {/* Start Button */}
            <Button
              onClick={handleStartSearch}
              variant="primary"
              size="lg"
              className="w-full"
              leftIcon={<Swords className="w-5 h-5" />}
            >
              Search for Opponent
            </Button>
          </div>
        ) : isSearching ? (
          /* Step 2: Live Queue Searching Screen */
          <div className="py-8 flex flex-col items-center justify-center text-center space-y-5">
            <div className="relative">
              <div className="w-20 h-20 rounded-full border-4 border-[#1A365D] border-t-transparent animate-spin flex items-center justify-center" />
              <div className="absolute inset-0 flex items-center justify-center text-2xl">
                {gameIcon}
              </div>
            </div>

            <div>
              <h3 className="text-xl font-bold text-[#1E242B]">Searching the Classroom...</h3>
              <p className="font-hand text-base text-[#475569] mt-0.5">
                Looking for a classmate ready to play {gameTitle}.
              </p>
            </div>

            <div className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-white border border-[#CBD5E1] font-mono text-xs font-bold text-[#1A365D]">
              <Clock className="w-3.5 h-3.5" />
              <span>Time in Queue: {Math.floor(searchSeconds / 60)}:{searchSeconds % 60 < 10 ? '0' : ''}{searchSeconds % 60}</span>
            </div>

            <Button
              onClick={handleCancelSearch}
              variant="secondary"
              size="sm"
              leftIcon={<X className="w-4 h-4" />}
            >
              Cancel Search
            </Button>
          </div>
        ) : (
          /* Step 3: Match Found & Countdown Screen */
          <div className="py-6 flex flex-col items-center justify-center text-center space-y-4">
            <Stamp tone="green">OPPONENT FOUND!</Stamp>

            <div className="flex items-center justify-center gap-6 my-2">
              <div className="text-center">
                <Avatar username={matchedGame?.player1.username || 'P1'} size="md" />
                <p className="font-bold text-xs text-[#1E242B] mt-1">@{matchedGame?.player1.username}</p>
                <Badge variant="pencil" size="sm">Rating: {matchedGame?.player1.rating || 1000}</Badge>
              </div>

              <div className="font-hand text-2xl font-bold text-[#991B1B]">VS</div>

              <div className="text-center">
                <Avatar username={matchedGame?.player2.username || 'P2'} size="md" />
                <p className="font-bold text-xs text-[#1E242B] mt-1">@{matchedGame?.player2.username}</p>
                <Badge variant="pencil" size="sm">Rating: {matchedGame?.player2.rating || 1000}</Badge>
              </div>
            </div>

            <div className="p-3 rounded-lg bg-[#DCFCE7] border border-[#86EFAC] w-full">
              <p className="font-hand text-lg font-bold text-[#15803D]">
                Match starting in {countdown} second{countdown === 1 ? '' : 's'}...
              </p>
            </div>
          </div>
        )}
      </div>
    </Modal>
  )
}
