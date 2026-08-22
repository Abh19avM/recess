import { useEffect, useState } from 'react'
import { useAuthStore } from '../store/authStore'
import { PaperCard } from '../components/ui/PaperCard'
import { Avatar } from '../components/ui/Avatar'
import { Stamp } from '../components/ui/Badge'
import { Award, Check } from 'lucide-react'
import { toast } from '../store/toastStore'

export const ProfilePage: React.FC = () => {
  const { user, profileStats, fetchMe } = useAuthStore()
  const [selectedAvatar, setSelectedAvatar] = useState(user?.avatar_preset || 'pencil_sketch_1')

  useEffect(() => {
    fetchMe()
  }, [])

  const avatarOptions = [
    { id: 'pencil_sketch_1', name: 'Classic Student' },
    { id: 'pencil_sketch_guest', name: 'Hall Monitor' },
    { id: 'chalkboard_star', name: 'Chalkboard Ace' },
    { id: 'prefect_badge', name: 'Class Prefect' },
  ]

  const handleSaveAvatar = (presetId: string) => {
    setSelectedAvatar(presetId)
    toast.success('Avatar Updated!', 'Your student ID photo has been saved.')
  }

  const gameRatings = profileStats?.game_stats || [
    { game_type: 'hand_cricket', rating: 1260, played: 6, won: 5, lost: 1 },
    { game_type: 'dots_boxes', rating: 1220, played: 4, won: 3, lost: 1 },
    { game_type: 'tic_tac_toe', rating: 1190, played: 2, won: 1, lost: 1 },
    { game_type: 'connect_4', rating: 1200, played: 0, won: 0, lost: 0 },
    { game_type: 'paper_football', rating: 1200, played: 0, won: 0, lost: 0 },
    { game_type: 'name_place_animal_thing', rating: 1200, played: 0, won: 0, lost: 0 },
  ]

  const formatGameName = (type: string) => {
    switch (type) {
      case 'hand_cricket':
        return 'Hand Cricket'
      case 'dots_boxes':
        return 'Dots & Boxes'
      case 'tic_tac_toe':
        return 'XO / Tic-Tac-Toe'
      case 'connect_4':
        return 'Connect 4'
      case 'paper_football':
        return 'Paper Football'
      case 'name_place_animal_thing':
        return 'Name-Place-Animal-Thing'
      default:
        return type
    }
  }

  return (
    <div className="space-y-8 max-w-4xl mx-auto pb-12">
      {/* Student ID Card Sheet */}
      <PaperCard variant="ruled" className="p-8 relative">
        <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-6 pb-6 border-b border-[#CBD5E1]">
          <div className="flex items-center gap-5">
            <Avatar username={user?.username || 'Student'} preset={selectedAvatar} size="xl" />
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-2xl sm:text-3xl font-extrabold text-[#1E242B]">
                  {user?.username || 'Classmate'}
                </h1>
                <Stamp tone="green">A+ MERIT</Stamp>
              </div>
              <p className="font-hand text-lg text-[#1A365D] mt-0.5">
                {user?.title || 'Classroom Rookie'} • Student ID: <span className="font-mono text-sm">{user?.id || 'usr_demo'}</span>
              </p>
              {user?.email && (
                <p className="text-xs text-[#475569] font-mono mt-1">
                  Email: {user.email}
                </p>
              )}
            </div>
          </div>

          <div className="bg-[#FEF9C3] px-4 py-2.5 rounded border border-[#FDE047] text-right">
            <span className="text-xs font-mono text-[#854D0E] uppercase block">
              Global Honor Rating
            </span>
            <span className="font-mono text-2xl font-black text-[#1E242B]">
              {user?.rating || 1200} ELO
            </span>
          </div>
        </div>

        {/* Change Student ID Avatar Preset */}
        <div className="pt-6">
          <h3 className="font-bold text-sm text-[#1E242B] font-mono uppercase tracking-wider mb-3">
            Pencil Sketch Photo Presets:
          </h3>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            {avatarOptions.map((opt) => (
              <button
                key={opt.id}
                type="button"
                onClick={() => handleSaveAvatar(opt.id)}
                className={`p-3 rounded-md border-2 flex items-center gap-3 transition-all text-left ${
                  selectedAvatar === opt.id
                    ? 'border-[#1A365D] bg-[#1A365D]/5 shadow-[2px_2px_0px_0px_#1A365D]'
                    : 'border-[#CBD5E1] bg-white hover:border-[#475569]'
                }`}
              >
                <Avatar username={user?.username || 'S'} preset={opt.id} size="sm" />
                <div className="flex-1 min-w-0">
                  <p className="text-xs font-bold text-[#1E242B] truncate">{opt.name}</p>
                  {selectedAvatar === opt.id && (
                    <span className="text-[10px] text-[#15803D] font-bold flex items-center gap-0.5">
                      <Check className="w-3 h-3" /> Selected
                    </span>
                  )}
                </div>
              </button>
            ))}
          </div>
        </div>
      </PaperCard>

      {/* Per-Game Performance Report Card */}
      <section className="space-y-4">
        <div className="flex items-center justify-between border-b-2 border-[#1E242B] pb-2">
          <div className="flex items-center gap-2">
            <h2 className="text-xl font-bold text-[#1E242B]">
              Subject Performance Ledger
            </h2>
            <Stamp tone="blue">ELO BREAKDOWN</Stamp>
          </div>
          <span className="text-xs font-mono text-[#475569]">
            Updated after each match
          </span>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          {gameRatings.map((stat, idx) => (
            <PaperCard key={idx} variant="plain" className="p-4">
              <div className="flex items-center justify-between mb-2">
                <h4 className="font-bold text-base text-[#1E242B]">
                  {formatGameName(stat.game_type)}
                </h4>
                <span className="font-mono text-sm font-black text-[#15803D] bg-[#DCFCE7] px-2 py-0.5 rounded border border-[#BBF7D0]">
                  {stat.rating} ELO
                </span>
              </div>

              <div className="grid grid-cols-3 gap-2 pt-2 text-center text-xs font-mono">
                <div className="bg-[#F8FAFC] p-1.5 rounded border border-[#E2E8F0]">
                  <span className="text-[#64748B] block text-[10px]">Played</span>
                  <span className="font-bold text-[#1E242B]">{stat.played}</span>
                </div>
                <div className="bg-[#F8FAFC] p-1.5 rounded border border-[#E2E8F0]">
                  <span className="text-[#15803D] block text-[10px]">Won</span>
                  <span className="font-bold text-[#15803D]">{stat.won}</span>
                </div>
                <div className="bg-[#F8FAFC] p-1.5 rounded border border-[#E2E8F0]">
                  <span className="text-[#991B1B] block text-[10px]">Lost</span>
                  <span className="font-bold text-[#991B1B]">{stat.lost}</span>
                </div>
              </div>
            </PaperCard>
          ))}
        </div>
      </section>

      {/* Classroom Honors & Badges */}
      <PaperCard variant="sticky" stickyColor="yellow" showPushPin className="p-6">
        <div className="flex items-center gap-2 mb-3">
          <Award className="w-5 h-5 text-[#B45309]" />
          <h3 className="font-bold text-base text-[#1E242B]">Earned School Badges</h3>
        </div>

        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs">
          <div className="bg-white p-3 rounded border border-yellow-300">
            <span className="font-hand text-lg text-[#15803D] font-bold block">
              🏏 6-Ball Master
            </span>
            <p className="text-[#475569] mt-0.5">Won a Hand Cricket match without losing a single wicket.</p>
          </div>
          <div className="bg-white p-3 rounded border border-yellow-300">
            <span className="font-hand text-lg text-[#1A365D] font-bold block">
              ⚄ Grid Conqueror
            </span>
            <p className="text-[#475569] mt-0.5">Captured a 6-box continuous chain in Dots & Boxes.</p>
          </div>
          <div className="bg-white p-3 rounded border border-yellow-300">
            <span className="font-hand text-lg text-[#991B1B] font-bold block">
              ⚡ Speed Champion
            </span>
            <p className="text-[#475569] mt-0.5">Placed 3 winning XO marks in under 4 seconds.</p>
          </div>
        </div>
      </PaperCard>
    </div>
  )
}
