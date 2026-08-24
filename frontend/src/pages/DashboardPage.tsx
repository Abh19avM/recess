import { useState, useEffect } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { PaperCard } from '../components/ui/PaperCard'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { Badge, Stamp } from '../components/ui/Badge'
import { Avatar } from '../components/ui/Avatar'
import { Modal } from '../components/ui/Modal'
import { MatchmakingModal } from '../components/ui/MatchmakingModal'
import { Swords, Users, PlusCircle, Play } from 'lucide-react'
import { toast } from '../store/toastStore'
import { api } from '../lib/api'

export const DashboardPage: React.FC = () => {
  const navigate = useNavigate()
  const { user, profileStats, fetchMe } = useAuthStore()

  const [roomCode, setRoomCode] = useState('')
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false)
  const [matchmakingGame, setMatchmakingGame] = useState<{ id: string; title: string; icon: string } | null>(null)
  const [selectedGame, setSelectedGame] = useState('hand_cricket')
  const [roomTitle, setRoomTitle] = useState('')
  const [isPrivate, setIsPrivate] = useState(false)
  const [passcode, setPasscode] = useState('')
  const [activeRooms, setActiveRooms] = useState<any[]>([])
  const [isLoadingRooms, setIsLoadingRooms] = useState(false)

  useEffect(() => {
    fetchMe()
    loadRooms()
  }, [])

  const loadRooms = async () => {
    setIsLoadingRooms(true)
    try {
      const resp = await api.get<any[]>('/rooms')
      setActiveRooms(resp || [])
    } catch {
      // Fallback demo rooms if API not seeded
      setActiveRooms([
        {
          id: '1',
          code: 'RECESS-HC-7X9P',
          title: 'Hand Cricket Championship',
          game_type: 'hand_cricket',
          host: 'ArbiterRecess',
          current_players: 1,
          max_players: 2,
        },
        {
          id: '2',
          code: 'RECESS-DOTS-4B1L',
          title: 'Graph Paper Showdown',
          game_type: 'dots_boxes',
          host: 'SchoolCaptain',
          current_players: 1,
          max_players: 2,
        },
      ])
    } finally {
      setIsLoadingRooms(false)
    }
  }

  const handleJoinByCode = (e: React.FormEvent) => {
    e.preventDefault()
    const cleanCode = roomCode.trim().toUpperCase()
    if (!cleanCode) return

    toast.info('Entering Classroom Desk...', cleanCode)
    if (cleanCode.includes('CRICKET') || cleanCode.includes('HC')) {
      navigate(`/games/hand-cricket/${cleanCode}`)
    } else if (cleanCode.includes('DOT') || cleanCode.includes('BOX') || cleanCode.includes('DB')) {
      navigate(`/games/dots-and-boxes/${cleanCode}`)
    } else {
      navigate(`/games/xo/${cleanCode}`)
    }
  }

  const handleCreateRoom = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      const prefix = selectedGame === 'hand_cricket' ? 'HC' : selectedGame === 'dots_boxes' ? 'DOTS' : 'XO'
      const generatedCode = `RECESS-${prefix}-${Math.random().toString(36).substring(2, 6).toUpperCase()}`
      setIsCreateModalOpen(false)
      toast.success('Lobby Created!', `Room code: ${generatedCode}`)
      
      if (selectedGame === 'hand_cricket') {
        navigate(`/games/hand-cricket/${generatedCode}`)
      } else if (selectedGame === 'dots_boxes') {
        navigate(`/games/dots-and-boxes/${generatedCode}`)
      } else {
        navigate(`/games/xo/${generatedCode}`)
      }
    } catch (err: any) {
      toast.error('Failed to create room', err.message)
    }
  }

  const gamesCatalog = [
    { id: 'hand_cricket', name: 'Hand Cricket', icon: '🏏', players: '2P' },
    { id: 'dots_boxes', name: 'Dots & Boxes', icon: '⚄', players: '2P' },
    { id: 'tic_tac_toe', name: 'XO / Tic-Tac-Toe', icon: '✕◯', players: '2P' },
    { id: 'connect_4', name: 'Connect 4', icon: '⚪', players: '2P' },
    { id: 'paper_football', name: 'Paper Football', icon: '📐', players: '2P' },
    { id: 'name_place_animal_thing', name: 'Name–Place–Animal–Thing', icon: '📝', players: '2-6P' },
  ]

  const gamesPlayed = profileStats?.games_played || 12
  const gamesWon = profileStats?.games_won || 9
  const winRate = profileStats?.win_rate || 75.0

  return (
    <div className="space-y-8">
      {/* 1. Student Report Card & Quick Actions */}
      <section className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Student ID / Report Card */}
        <PaperCard variant="ruled" className="p-6 relative lg:col-span-2">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4 pb-4 border-b border-[#CBD5E1]">
            <div className="flex items-center gap-3">
              <Avatar
                username={user?.username || 'Student'}
                preset={user?.avatar_preset}
                size="lg"
                isOnline
              />
              <div>
                <div className="flex items-center gap-2">
                  <h2 className="text-xl font-bold text-[#1E242B]">
                    {user?.username || 'Student'}
                  </h2>
                  <Stamp tone="green">ACTIVE</Stamp>
                </div>
                <p className="font-hand text-base text-[#1A365D]">
                  {user?.title || 'Classroom Rookie'} • {user?.is_guest ? 'Guest Pass' : 'Enrolled Student'}
                </p>
              </div>
            </div>

            <div className="text-right flex items-center gap-2">
              <div className="bg-[#FEF9C3] px-3 py-1.5 rounded border border-[#FDE047] text-left">
                <span className="text-[10px] font-mono text-[#854D0E] uppercase block">
                  Overall Rating
                </span>
                <span className="font-mono text-lg font-black text-[#1E242B]">
                  {user?.rating || 1200} ELO
                </span>
              </div>
            </div>
          </div>

          {/* Academic Report Metrics */}
          <div className="grid grid-cols-3 gap-4 pt-4 text-center">
            <div className="bg-[#F8FAFC] p-3 rounded border border-[#E2E8F0]">
              <span className="text-[11px] font-mono text-[#64748B] uppercase block">
                Games Played
              </span>
              <span className="text-2xl font-black text-[#1E242B] font-mono">
                {gamesPlayed}
              </span>
            </div>
            <div className="bg-[#F8FAFC] p-3 rounded border border-[#E2E8F0]">
              <span className="text-[11px] font-mono text-[#64748B] uppercase block">
                Victories
              </span>
              <span className="text-2xl font-black text-[#15803D] font-mono">
                {gamesWon}
              </span>
            </div>
            <div className="bg-[#F8FAFC] p-3 rounded border border-[#E2E8F0]">
              <span className="text-[11px] font-mono text-[#64748B] uppercase block">
                Win Rate
              </span>
              <span className="text-2xl font-black text-[#1A365D] font-mono">
                {winRate}%
              </span>
            </div>
          </div>
        </PaperCard>

        {/* Custom Room Code Entry Box */}
        <PaperCard variant="sticky" stickyColor="yellow" showPushPin className="p-6 flex flex-col justify-between">
          <div>
            <div className="flex items-center gap-2 mb-2">
              <Swords className="w-5 h-5 text-[#B45309]" />
              <h3 className="font-bold text-base text-[#1E242B]">Enter Classroom Code</h3>
            </div>
            <p className="font-hand text-base text-[#1A365D]">
              Got an invite code from a benchmate? Enter it below to join their desk immediately.
            </p>

            <form onSubmit={handleJoinByCode} className="space-y-3 mt-4">
              <Input
                placeholder="RECESS-XXXX"
                value={roomCode}
                onChange={(e) => setRoomCode(e.target.value.toUpperCase())}
                className="font-mono text-center tracking-widest font-bold uppercase"
              />
              <Button type="submit" variant="primary" size="md" className="w-full">
                Join Classroom Desk →
              </Button>
            </form>
          </div>

          <div className="pt-4 mt-4 border-t border-yellow-300">
            <Button
              onClick={() => setIsCreateModalOpen(true)}
              variant="secondary"
              size="sm"
              className="w-full"
              leftIcon={<PlusCircle className="w-4 h-4" />}
            >
              Host Custom Desk Room
            </Button>
          </div>
        </PaperCard>
      </section>

      {/* 2. Quick Game Select Grid */}
      <section className="space-y-4">
        <div className="flex items-center justify-between border-b-2 border-[#1E242B] pb-2">
          <div className="flex items-center gap-2">
            <h2 className="text-xl font-bold text-[#1E242B]">Choose Game Arena</h2>
            <Stamp tone="amber">6 ARENAS</Stamp>
          </div>
          <Link to="/games" className="text-xs font-bold text-[#1A365D] hover:underline">
            View Rules & Syllabus →
          </Link>
        </div>

        <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
          {gamesCatalog.map((game) => (
            <PaperCard
              key={game.id}
              variant="plain"
              className="p-4 text-center hover:shadow-[4px_4px_0px_0px_#1E242B] hover:-translate-y-1 transition-all cursor-pointer flex flex-col items-center justify-between group"
              onClick={() => {
                setMatchmakingGame({
                  id: game.id,
                  title: game.name,
                  icon: game.icon,
                })
              }}
            >
              <span className="text-3xl mb-2">{game.icon}</span>
              <h4 className="font-bold text-sm text-[#1E242B] group-hover:text-[#1A365D] transition-colors leading-tight">
                {game.name}
              </h4>
              <Badge variant="pencil" size="sm" className="mt-2">
                {game.players}
              </Badge>
            </PaperCard>
          ))}
        </div>
      </section>

      {/* 3. Active Lobbies Notice Board */}
      <section className="space-y-4">
        <PaperCard variant="cork" className="p-6">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-2">
              <Users className="w-5 h-5 text-[#FFFFFF]" />
              <h3 className="font-bold text-lg text-white font-sans">
                Active Classroom Desks
              </h3>
              <Stamp tone="green">LIVE LOBBIES</Stamp>
            </div>
            <Button
              onClick={loadRooms}
              variant="chalk"
              size="sm"
              isLoading={isLoadingRooms}
            >
              Refresh Notice Board
            </Button>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {activeRooms.map((room) => (
              <div
                key={room.id}
                className="bg-white rounded p-4 border-2 border-[#1E242B] shadow-[3px_3px_0px_0px_#1E242B] flex items-center justify-between"
              >
                <div>
                  <div className="flex items-center gap-2">
                    <span className="font-mono font-bold text-xs bg-[#F2EDE0] px-2 py-0.5 rounded border border-[#CBD5E1]">
                      {room.code}
                    </span>
                    <Badge variant="ink-blue" size="sm">
                      {room.game_type}
                    </Badge>
                  </div>
                  <h4 className="font-bold text-sm text-[#1E242B] mt-1">{room.title}</h4>
                  <p className="text-xs text-[#475569] font-hand">
                    Hosted by @{room.host || 'Classmate'}
                  </p>
                </div>

                <Button
                  onClick={() => {
                    if (room.game_type === 'hand_cricket') {
                      navigate(`/games/hand-cricket/${room.code}`)
                    } else if (room.game_type === 'dots_boxes') {
                      navigate(`/games/dots-and-boxes/${room.code}`)
                    } else {
                      navigate(`/games/xo/${room.code}`)
                    }
                  }}
                  variant="primary"
                  size="sm"
                  leftIcon={<Play className="w-3.5 h-3.5" />}
                >
                  Seat In
                </Button>
              </div>
            ))}
          </div>
        </PaperCard>
      </section>

      {/* Host Room Modal */}
      <Modal
        isOpen={isCreateModalOpen}
        onClose={() => setIsCreateModalOpen(false)}
        title="Host A Classroom Desk"
        handwrittenTitle
      >
        <form onSubmit={handleCreateRoom} className="space-y-4">
          <div>
            <label className="block text-sm font-semibold text-[#1E242B] mb-1.5">
              Select School Game
            </label>
            <select
              value={selectedGame}
              onChange={(e) => setSelectedGame(e.target.value)}
              className="w-full text-sm text-[#1E242B] bg-white px-3 py-2.5 rounded-md border-2 border-[#475569] shadow-[2px_2px_0px_0px_#475569] outline-none"
            >
              {gamesCatalog.map((g) => (
                <option key={g.id} value={g.id}>
                  {g.icon} {g.name} ({g.players})
                </option>
              ))}
            </select>
          </div>

          <Input
            label="Lobby Title"
            placeholder="e.g. 5-Over Hand Cricket Showdown"
            value={roomTitle}
            onChange={(e) => setRoomTitle(e.target.value)}
            required
          />

          <div className="flex items-center gap-2 pt-2">
            <input
              type="checkbox"
              id="is_private"
              checked={isPrivate}
              onChange={(e) => setIsPrivate(e.target.checked)}
              className="w-4 h-4 rounded border-2 border-[#475569]"
            />
            <label htmlFor="is_private" className="text-xs font-semibold text-[#1E242B] cursor-pointer">
              Require Desk Passcode (Private Room)
            </label>
          </div>

          {isPrivate && (
            <Input
              label="Room Passcode"
              type="password"
              placeholder="4-digit passcode"
              value={passcode}
              onChange={(e) => setPasscode(e.target.value)}
            />
          )}

          <div className="pt-4 flex justify-end gap-3">
            <Button
              type="button"
              variant="ghost"
              size="md"
              onClick={() => setIsCreateModalOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" variant="primary" size="md">
              Create & Open Lobby
            </Button>
          </div>
        </form>
      </Modal>

      {/* Matchmaking Queue Modal */}
      {matchmakingGame && (
        <MatchmakingModal
          isOpen={!!matchmakingGame}
          onClose={() => setMatchmakingGame(null)}
          gameId={matchmakingGame.id}
          gameTitle={matchmakingGame.title}
          gameIcon={matchmakingGame.icon}
        />
      )}
    </div>
  )
}
