import React, { useEffect, useState } from 'react'
import { useAuthStore } from '../store/authStore'
import { PaperCard } from '../components/ui/PaperCard'
import { Avatar } from '../components/ui/Avatar'
import { Badge, Stamp } from '../components/ui/Badge'
import { api } from '../lib/api'
import { toast } from '../store/toastStore'
import {
  BookOpen,
  Check,
  Sparkles,
  Trophy,
  History,
  GraduationCap,
} from 'lucide-react'

interface MatchHistoryItem {
  match_id: string
  game_type: string
  room_code: string
  opponent_id: string
  opponent_username: string
  opponent_avatar: string
  is_winner: boolean
  is_draw: boolean
  score: number
  opponent_score: number
  rating_before: number
  rating_after: number
  rating_delta: number
  ended_at: string
}

interface Achievement {
  id: string
  name: string
  description: string
  category: string
  icon: string
  badge_tone: 'blue' | 'green' | 'amber' | 'red' | 'purple'
  points: number
  is_unlocked: boolean
  unlocked_at?: string
}

export const ProfilePage: React.FC = () => {
  const { user, profileStats, fetchMe } = useAuthStore()
  const [selectedAvatar, setSelectedAvatar] = useState(user?.avatar_preset || 'pencil_sketch_1')
  const [matchHistory, setMatchHistory] = useState<MatchHistoryItem[]>([])
  const [achievements, setAchievements] = useState<Achievement[]>([])
  const [activeTab, setActiveTab] = useState<'report_card' | 'history' | 'achievements'>('report_card')

  useEffect(() => {
    fetchMe()
    loadHistoryAndAchievements()
  }, [])

  const loadHistoryAndAchievements = async () => {
    try {
      if (user?.id) {
        const [histRes, achRes] = await Promise.all([
          api.get<{ matches: MatchHistoryItem[] }>(`/progression/${user.id}/history`),
          api.get<{ achievements: Achievement[] }>(`/progression/${user.id}/achievements`),
        ])
        if (histRes?.matches) setMatchHistory(histRes.matches)
        if (achRes?.achievements) setAchievements(achRes.achievements)
      }
    } catch {
      // Fallback demo data
      setMatchHistory([
        {
          match_id: 'm1',
          game_type: 'hand_cricket',
          room_code: 'RECESS-HC-4821',
          opponent_id: 'u2',
          opponent_username: 'SchoolCaptain',
          opponent_avatar: 'pencil_sketch_1',
          is_winner: true,
          is_draw: false,
          score: 42,
          opponent_score: 28,
          rating_before: 1242,
          rating_after: 1260,
          rating_delta: 18,
          ended_at: new Date(Date.now() - 3600000).toISOString(),
        },
        {
          match_id: 'm2',
          game_type: 'dots_boxes',
          room_code: 'RECESS-DOTS-9012',
          opponent_id: 'u3',
          opponent_username: 'ArbiterRecess',
          opponent_avatar: 'pencil_sketch_2',
          is_winner: false,
          is_draw: false,
          score: 4,
          opponent_score: 5,
          rating_before: 1256,
          rating_after: 1242,
          rating_delta: -14,
          ended_at: new Date(Date.now() - 14400000).toISOString(),
        },
        {
          match_id: 'm3',
          game_type: 'xo',
          room_code: 'RECESS-XO-1134',
          opponent_id: 'u4',
          opponent_username: 'PencilProdigy',
          opponent_avatar: 'pencil_sketch_3',
          is_winner: true,
          is_draw: false,
          score: 1,
          opponent_score: 0,
          rating_before: 1240,
          rating_after: 1256,
          rating_delta: 16,
          ended_at: new Date(Date.now() - 86400000).toISOString(),
        },
      ])

      setAchievements([
        { id: 'first_win', name: 'First Bell Victory', description: 'Win your very first classroom duel in Recess', category: 'general', icon: '🔔', badge_tone: 'green', points: 10, is_unlocked: true },
        { id: 'streak_3', name: 'Hat-Trick Student', description: 'Win 3 matches in a row across any game', category: 'general', icon: '🔥', badge_tone: 'red', points: 25, is_unlocked: true },
        { id: 'streak_5', name: 'Classroom Dominator', description: 'Achieve an unbroken 5-match winning streak', category: 'general', icon: '⚡', badge_tone: 'purple', points: 50, is_unlocked: false },
        { id: 'cricket_centurion', name: 'Finger Cricket Master', description: 'Score 50+ runs in a single Hand Cricket match', category: 'hand_cricket', icon: '🏏', badge_tone: 'amber', points: 30, is_unlocked: true },
        { id: 'clean_sweep_xo', name: 'Flawless Grid', description: 'Win an XO duel without allowing opponent a corner', category: 'xo', icon: '✕', badge_tone: 'blue', points: 20, is_unlocked: true },
        { id: 'box_conqueror', name: 'Territory Mogul', description: 'Claim 6 or more boxes in a single Dots & Boxes duel', category: 'dots_boxes', icon: '⚄', badge_tone: 'green', points: 25, is_unlocked: false },
        { id: 'scholar_1300', name: 'Honor Roll ELO', description: 'Reach an overall rating of 1300 ELO', category: 'rating', icon: '📜', badge_tone: 'amber', points: 50, is_unlocked: false },
        { id: 'veteran_25', name: 'Recess Veteran', description: 'Complete 25 total multiplayer classroom matches', category: 'milestone', icon: '🎓', badge_tone: 'purple', points: 40, is_unlocked: false },
      ])
    }
  }

  const avatarOptions = [
    { id: 'pencil_sketch_1', name: 'Classic Student' },
    { id: 'pencil_sketch_guest', name: 'Hall Monitor' },
    { id: 'chalkboard_star', name: 'Chalkboard Ace' },
    { id: 'prefect_badge', name: 'Class Prefect' },
  ]

  const handleSaveAvatar = (presetId: string) => {
    setSelectedAvatar(presetId)
    toast.success('ID Photo Stamped!', 'Your student avatar has been updated.')
  }

  const gameRatings = profileStats?.game_stats || [
    { game_type: 'hand_cricket', rating: 1260, played: 6, won: 5, lost: 1, drawn: 0 },
    { game_type: 'dots_boxes', rating: 1220, played: 4, won: 3, lost: 1, drawn: 0 },
    { game_type: 'xo', rating: 1190, played: 2, won: 1, lost: 1, drawn: 0 },
    { game_type: 'connect_4', rating: 1200, played: 0, won: 0, lost: 0, drawn: 0 },
    { game_type: 'paper_football', rating: 1200, played: 0, won: 0, lost: 0, drawn: 0 },
    { game_type: 'name_place_animal_thing', rating: 1200, played: 0, won: 0, lost: 0, drawn: 0 },
  ]

  const formatGameName = (type: string) => {
    switch (type) {
      case 'hand_cricket':
        return 'Hand Cricket'
      case 'dots_boxes':
        return 'Dots & Boxes'
      case 'xo':
      case 'tic_tac_toe':
        return 'XO / Tic-Tac-Toe'
      case 'connect_4':
        return 'Connect 4'
      case 'paper_football':
        return 'Paper Football'
      case 'name_place_animal_thing':
        return 'Name–Place–Animal–Thing'
      default:
        return type
    }
  }

  const getSubjectGrade = (rating: number) => {
    if (rating >= 1350) return { grade: 'A+', remark: 'Distinction' }
    if (rating >= 1250) return { grade: 'A', remark: 'Excellent' }
    if (rating >= 1180) return { grade: 'B+', remark: 'Commendable' }
    if (rating >= 1100) return { grade: 'B', remark: 'Satisfactory' }
    return { grade: 'C', remark: 'Needs Practice' }
  }

  const totalPlayed = profileStats?.games_played || 12
  const totalWon = profileStats?.games_won || 9
  const totalLost = totalPlayed - totalWon
  const winRate = profileStats?.win_rate || 75.0
  const overallRating = user?.rating || 1260
  const unlockedAchievementsCount = achievements.filter((a) => a.is_unlocked).length

  return (
    <div className="space-y-8 max-w-5xl mx-auto pb-16">
      {/* 1. Header Navigation Tabs */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b-2 border-[#1E242B] dark:border-slate-700 pb-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl font-extrabold text-[#1E242B] dark:text-slate-100">Student Academic Dossier</h1>
            <Stamp tone="amber">TERM 2026</Stamp>
          </div>
          <p className="font-hand text-sm text-[#475569] dark:text-slate-300">
            Official classroom performance record, duel ledger, and stamped achievements.
          </p>
        </div>

        {/* Tab Switcher */}
        <div className="flex items-center gap-2 bg-[#F1F5F9] dark:bg-slate-900/80 p-1 rounded-lg border border-[#CBD5E1] dark:border-slate-700">
          <button
            onClick={() => setActiveTab('report_card')}
            className={`px-3 py-1.5 rounded-md text-xs font-bold font-mono transition-all ${
              activeTab === 'report_card'
                ? 'bg-[#1A365D] dark:bg-sky-700 text-white shadow-xs'
                : 'text-[#475569] dark:text-slate-400 hover:text-[#1E242B] dark:hover:text-slate-200'
            }`}
          >
            Report Card
          </button>
          <button
            onClick={() => setActiveTab('history')}
            className={`px-3 py-1.5 rounded-md text-xs font-bold font-mono transition-all ${
              activeTab === 'history'
                ? 'bg-[#1A365D] dark:bg-sky-700 text-white shadow-xs'
                : 'text-[#475569] dark:text-slate-400 hover:text-[#1E242B] dark:hover:text-slate-200'
            }`}
          >
            Match Ledger ({matchHistory.length})
          </button>
          <button
            onClick={() => setActiveTab('achievements')}
            className={`px-3 py-1.5 rounded-md text-xs font-bold font-mono transition-all ${
              activeTab === 'achievements'
                ? 'bg-[#1A365D] dark:bg-sky-700 text-white shadow-xs'
                : 'text-[#475569] dark:text-slate-400 hover:text-[#1E242B] dark:hover:text-slate-200'
            }`}
          >
            Stamps ({unlockedAchievementsCount}/{achievements.length})
          </button>
        </div>
      </div>

      {/* 2. Main Tab Content */}
      {activeTab === 'report_card' && (
        <div className="space-y-8">
          {/* Main Official School Report Card Sheet */}
          <PaperCard variant="ruled" className="p-8 sm:p-10 relative overflow-hidden">
            {/* Watermark Crest Stamp */}
            <div className="absolute right-6 top-6 opacity-15 pointer-events-none select-none text-right">
              <span className="font-serif text-6xl font-black block tracking-widest text-[#1A365D] dark:text-sky-400">RECESS</span>
              <span className="font-mono text-xs font-bold uppercase tracking-widest text-[#1A365D] dark:text-sky-400">Academic Board</span>
            </div>

            {/* School Heading */}
            <div className="text-center pb-6 border-b-2 border-[#1E242B] dark:border-slate-700 space-y-1">
              <div className="flex items-center justify-center gap-2">
                <GraduationCap className="w-6 h-6 text-[#1A365D] dark:text-sky-400" />
                <h2 className="font-serif text-2xl sm:text-3xl font-extrabold uppercase tracking-wide text-[#1E242B] dark:text-slate-100">
                  Recess Central Board of Play
                </h2>
              </div>
              <p className="font-mono text-xs text-[#475569] dark:text-slate-400 uppercase tracking-widest">
                Official Student Athletic & Tactical Proficiency Report Card
              </p>
              <div className="flex items-center justify-center gap-6 pt-2 font-mono text-xs font-bold text-[#1A365D] dark:text-sky-300">
                <span>ROLL NO: #REC-2026-{user?.id?.substring(0, 4).toUpperCase() || '4890'}</span>
                <span>•</span>
                <span>HOUSE: EMERALD TACTICIANS</span>
                <span>•</span>
                <span>STATUS: {user?.is_guest ? 'GUEST PASS' : 'ENROLLED REGULAR'}</span>
              </div>
            </div>

            {/* Student ID & Overview Banner */}
            <div className="grid grid-cols-1 md:grid-cols-12 gap-6 pt-6 pb-6 border-b border-[#CBD5E1] dark:border-slate-700/60 items-center">
              {/* Photo & Name (7 cols) */}
              <div className="md:col-span-7 flex items-center gap-5">
                <div className="relative">
                  <Avatar username={user?.username || 'Student'} size="lg" isOnline />
                  <div className="absolute -top-2 -right-2 rotate-12">
                    <Stamp tone="green">VERIFIED</Stamp>
                  </div>
                </div>
                <div>
                  <h3 className="text-2xl font-black text-[#1E242B] dark:text-slate-100">@{user?.username}</h3>
                  <p className="font-hand text-lg text-[#1A365D] dark:text-sky-300 font-bold">
                    Title: {user?.title || 'Classroom Valedictorian'}
                  </p>
                  <p className="font-mono text-xs text-[#64748B] dark:text-slate-400">
                    Student Since: {new Date(user?.created_at || Date.now()).toLocaleDateString([], { month: 'short', year: 'numeric' })}
                  </p>
                </div>
              </div>

              {/* Cumulative ELO Grade Box (5 cols) */}
              <div className="md:col-span-5 bg-[#FEF9C3] dark:bg-amber-950/40 p-4 rounded-xl border-2 border-[#FDE047] dark:border-amber-700/60 shadow-xs text-center flex items-center justify-around">
                <div>
                  <span className="text-[10px] font-mono text-[#854D0E] dark:text-amber-300 uppercase block font-bold">
                    Cumulative Rating
                  </span>
                  <span className="font-mono text-3xl font-black text-[#1E242B] dark:text-slate-100">
                    {overallRating} <span className="text-xs font-normal text-[#854D0E] dark:text-amber-300">ELO</span>
                  </span>
                </div>
                <div className="w-px h-10 bg-[#FDE047] dark:bg-amber-700/60" />
                <div>
                  <span className="text-[10px] font-mono text-[#854D0E] dark:text-amber-300 uppercase block font-bold">
                    Academic Rank
                  </span>
                  <span className="font-serif text-3xl font-black text-[#15803D] dark:text-emerald-400">
                    A+
                  </span>
                </div>
              </div>
            </div>

            {/* Academic Standing Metrics */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 py-6 border-b border-[#CBD5E1] dark:border-slate-700/60 text-center">
              <div className="p-3 bg-white dark:bg-slate-900/70 rounded-lg border border-[#CBD5E1] dark:border-slate-700/60 shadow-2xs">
                <span className="text-[11px] font-mono text-[#64748B] dark:text-slate-400 uppercase block">Total Duels</span>
                <span className="text-2xl font-mono font-black text-[#1E242B] dark:text-slate-100">{totalPlayed}</span>
              </div>
              <div className="p-3 bg-white dark:bg-slate-900/70 rounded-lg border border-[#CBD5E1] dark:border-slate-700/60 shadow-2xs">
                <span className="text-[11px] font-mono text-[#64748B] dark:text-slate-400 uppercase block">Victories</span>
                <span className="text-2xl font-mono font-black text-[#15803D] dark:text-emerald-400">{totalWon}</span>
              </div>
              <div className="p-3 bg-white dark:bg-slate-900/70 rounded-lg border border-[#CBD5E1] dark:border-slate-700/60 shadow-2xs">
                <span className="text-[11px] font-mono text-[#64748B] dark:text-slate-400 uppercase block">Defeats</span>
                <span className="text-2xl font-mono font-black text-[#991B1B] dark:text-red-400">{totalLost}</span>
              </div>
              <div className="p-3 bg-white dark:bg-slate-900/70 rounded-lg border border-[#CBD5E1] dark:border-slate-700/60 shadow-2xs">
                <span className="text-[11px] font-mono text-[#64748B] dark:text-slate-400 uppercase block">Win Ratio</span>
                <span className="text-2xl font-mono font-black text-[#1A365D] dark:text-sky-400">{winRate}%</span>
              </div>
            </div>

            {/* Subject-Wise Performance Breakdown Table */}
            <div className="pt-6 space-y-4">
              <div className="flex items-center justify-between">
                <h4 className="font-bold text-base text-[#1E242B] dark:text-slate-100 flex items-center gap-2">
                  <BookOpen className="w-4 h-4 text-[#1A365D] dark:text-sky-400" />
                  Subject-Wise Athletic Breakdown
                </h4>
                <Stamp tone="blue">6 COURSES</Stamp>
              </div>

              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs font-mono border-collapse">
                  <thead>
                    <tr className="border-b-2 border-[#1E242B] dark:border-slate-700 bg-[#F8FAFC] dark:bg-slate-900 text-[#475569] dark:text-slate-300">
                      <th className="py-2.5 px-3 uppercase font-bold">Course / Game</th>
                      <th className="py-2.5 px-3 uppercase font-bold text-center">Rating</th>
                      <th className="py-2.5 px-3 uppercase font-bold text-center">Played</th>
                      <th className="py-2.5 px-3 uppercase font-bold text-center">Record (W-L)</th>
                      <th className="py-2.5 px-3 uppercase font-bold text-center">Grade</th>
                      <th className="py-2.5 px-3 uppercase font-bold text-right">Remarks</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[#E2E8F0] dark:divide-slate-800">
                    {gameRatings.map((stat) => {
                      const evalGrade = getSubjectGrade(stat.rating)
                      return (
                        <tr key={stat.game_type} className="hover:bg-slate-50/80 dark:hover:bg-slate-800/50 transition-colors">
                          <td className="py-3 px-3 font-sans font-bold text-sm text-[#1E242B] dark:text-slate-100">
                            {formatGameName(stat.game_type)}
                          </td>
                          <td className="py-3 px-3 text-center font-bold text-[#1A365D] dark:text-sky-300">
                            {stat.rating} ELO
                          </td>
                          <td className="py-3 px-3 text-center text-[#475569] dark:text-slate-300">
                            {stat.played}
                          </td>
                          <td className="py-3 px-3 text-center">
                            <span className="text-[#15803D] dark:text-emerald-400 font-bold">{stat.won}W</span> - <span className="text-[#991B1B] dark:text-red-400 font-bold">{stat.lost}L</span>
                          </td>
                          <td className="py-3 px-3 text-center">
                            <Badge variant={stat.rating >= 1250 ? 'green' : stat.rating >= 1200 ? 'ink-blue' : 'default'} size="sm">
                              {evalGrade.grade}
                            </Badge>
                          </td>
                          <td className="py-3 px-3 text-right font-hand text-sm text-[#1A365D] dark:text-sky-300">
                            {evalGrade.remark}
                          </td>
                        </tr>
                      )
                    })}
                  </tbody>
                </table>
              </div>
            </div>

            {/* Arbiter / Teacher's Handwritten Evaluation Remarks */}
            <div className="mt-8 p-5 bg-[#FFFBEB] dark:bg-amber-950/40 rounded-xl border border-[#FDE68A] dark:border-amber-700/60 relative">
              <div className="flex items-center gap-2 mb-1">
                <Sparkles className="w-4 h-4 text-[#B45309] dark:text-amber-400" />
                <h5 className="font-mono text-xs uppercase font-bold text-[#92400E] dark:text-amber-300">
                  Class Teacher & Arbiter Remarks:
                </h5>
              </div>
              <p className="font-hand text-base text-[#1E242B] dark:text-amber-100 leading-relaxed">
                "@{user?.username} demonstrates remarkable hand-eye composure and tactical intuition across all classroom periods. Particularly dominant in Hand Cricket batting innings. Promoted to Senior Recess League with Honors."
              </p>
              <div className="text-right mt-2 font-hand text-sm font-bold text-[#B45309] dark:text-amber-400">
                — Head Arbiter, Recess Examination Board
              </div>
            </div>
          </PaperCard>

          {/* Avatar Stamp Selector Box */}
          <PaperCard variant="plain" className="p-6">
            <h4 className="font-bold text-sm text-[#1E242B] dark:text-slate-100 uppercase font-mono mb-4">
              Select ID Photograph Badge:
            </h4>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              {avatarOptions.map((opt) => (
                <div
                  key={opt.id}
                  onClick={() => handleSaveAvatar(opt.id)}
                  className={`p-4 rounded-xl border-2 text-center cursor-pointer transition-all flex flex-col items-center justify-between ${
                    selectedAvatar === opt.id
                      ? 'border-[#1A365D] dark:border-sky-500 bg-[#E0F2FE] dark:bg-sky-950/50 shadow-[3px_3px_0px_0px_#1A365D] dark:shadow-[3px_3px_0px_0px_#020617]'
                      : 'border-[#CBD5E1] dark:border-slate-700 bg-white dark:bg-slate-900 hover:bg-slate-50 dark:hover:bg-slate-800'
                  }`}
                >
                  <Avatar username={user?.username || 'You'} size="md" />
                  <span className="font-bold text-xs text-[#1E242B] dark:text-slate-100 mt-2 block">{opt.name}</span>
                  {selectedAvatar === opt.id && (
                    <Badge variant="green" size="sm" className="mt-2">
                      <Check className="w-3 h-3 mr-1 inline" /> ACTIVE
                    </Badge>
                  )}
                </div>
              ))}
            </div>
          </PaperCard>
        </div>
      )}

      {/* 3. Match History Ledger Tab */}
      {activeTab === 'history' && (
        <div className="space-y-4">
          <PaperCard variant="ruled" className="p-6 sm:p-8">
            <div className="flex items-center justify-between pb-4 border-b-2 border-[#1E242B] mb-6">
              <div className="flex items-center gap-2">
                <History className="w-5 h-5 text-[#1A365D]" />
                <h3 className="font-bold text-lg text-[#1E242B]">Official Match Duel Ledger</h3>
              </div>
              <Stamp tone="blue">RECORDED MATCHES</Stamp>
            </div>

            <div className="space-y-3">
              {matchHistory.map((item) => (
                <div
                  key={item.match_id}
                  className="p-4 rounded-lg bg-white dark:bg-slate-900/80 border border-[#CBD5E1] dark:border-slate-700 shadow-2xs flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4"
                >
                  {/* Left: Outcome & Game */}
                  <div className="flex items-center gap-3">
                    <Stamp tone={item.is_winner ? 'green' : 'red'}>
                      {item.is_winner ? 'WON' : 'LOST'}
                    </Stamp>
                    <div>
                      <h4 className="font-bold text-sm text-[#1E242B] dark:text-slate-100">
                        {formatGameName(item.game_type)}
                      </h4>
                      <p className="font-mono text-xs text-[#64748B] dark:text-slate-400">
                        Room: <span className="font-bold text-[#1A365D] dark:text-sky-300">{item.room_code}</span> • vs @{item.opponent_username}
                      </p>
                    </div>
                  </div>

                  {/* Right: Scores & ELO Delta */}
                  <div className="flex items-center gap-6 text-right w-full sm:w-auto justify-between sm:justify-end">
                    <div>
                      <span className="text-[10px] font-mono text-[#64748B] dark:text-slate-400 uppercase block">Final Score</span>
                      <span className="font-mono font-bold text-sm text-[#1E242B] dark:text-slate-100">
                        {item.score} - {item.opponent_score}
                      </span>
                    </div>

                    <div className="bg-[#F8FAFC] dark:bg-slate-950/70 px-3 py-1.5 rounded border border-[#CBD5E1] dark:border-slate-700">
                      <span className="text-[10px] font-mono text-[#64748B] dark:text-slate-400 uppercase block">Rating Shift</span>
                      <span
                        className={`font-mono font-black text-sm ${
                          item.rating_delta >= 0 ? 'text-[#15803D] dark:text-emerald-400' : 'text-[#991B1B] dark:text-red-400'
                        }`}
                      >
                        {item.rating_delta >= 0 ? `+${item.rating_delta}` : item.rating_delta} ELO
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </PaperCard>
        </div>
      )}

      {/* 4. Achievements Stamps Tab */}
      {activeTab === 'achievements' && (
        <div className="space-y-4">
          <PaperCard variant="cork" className="p-6 sm:p-8">
            <div className="flex items-center justify-between pb-4 border-b border-white/20 mb-6 text-white">
              <div className="flex items-center gap-2">
                <Trophy className="w-5 h-5 text-[#FDE047]" />
                <h3 className="font-bold text-lg text-white font-sans">Classroom Honor Stamps & Awards</h3>
              </div>
              <Stamp tone="amber">{unlockedAchievementsCount}/{achievements.length} UNLOCKED</Stamp>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              {achievements.map((ach) => (
                <div
                  key={ach.id}
                  className={`p-4 rounded-xl border-2 transition-all flex flex-col justify-between ${
                    ach.is_unlocked
                      ? 'bg-[#FEF9C3] dark:bg-amber-950/70 border-[#FDE047] dark:border-amber-600 shadow-[3px_3px_0px_0px_#854D0E] dark:shadow-[3px_3px_0px_0px_#020617]'
                      : 'bg-white/80 dark:bg-slate-900/60 border-[#CBD5E1] dark:border-slate-700 opacity-60'
                  }`}
                >
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <span className="text-3xl">{ach.icon}</span>
                      <Stamp tone={ach.is_unlocked ? 'green' : 'amber'}>
                        {ach.is_unlocked ? 'UNLOCKED' : `${ach.points} PTS`}
                      </Stamp>
                    </div>
                    <h4 className="font-bold text-sm text-[#1E242B] dark:text-slate-100">{ach.name}</h4>
                    <p className="font-hand text-xs text-[#475569] dark:text-slate-300 mt-1">{ach.description}</p>
                  </div>

                  {ach.is_unlocked && (
                    <div className="mt-3 pt-2 border-t border-[#FDE047]/60 dark:border-amber-700/60 flex items-center justify-between text-[10px] font-mono text-[#854D0E] dark:text-amber-300">
                      <span>Awarded by Arbiter</span>
                      <Check className="w-3.5 h-3.5 text-[#15803D] dark:text-emerald-400" />
                    </div>
                  )}
                </div>
              ))}
            </div>
          </PaperCard>
        </div>
      )}
    </div>
  )
}
