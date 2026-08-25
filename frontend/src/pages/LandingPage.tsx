import React from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { Button } from '../components/ui/Button'
import { PaperCard } from '../components/ui/PaperCard'
import { Badge, Stamp } from '../components/ui/Badge'
import {
  Sparkles,
  Users,
  Swords,
  Trophy,
  ArrowRight,
  ShieldCheck,
  Zap,
  RotateCcw,
} from 'lucide-react'

export const LandingPage: React.FC = () => {
  const navigate = useNavigate()
  const { isAuthenticated, guestLogin, isLoading } = useAuthStore()

  const handleQuickGuestPlay = async () => {
    try {
      await guestLogin()
      navigate('/dashboard')
    } catch {
      // Handled in store
    }
  }

  const featuredGames = [
    {
      id: 'hand_cricket',
      title: 'Hand Cricket',
      tagline: 'The ultimate last-bench finger cricket showdown',
      players: '2 Players',
      type: 'Turn-Based',
      duration: '3-5 mins',
      paperType: 'ruled' as const,
      stampText: 'MOST POPULAR',
      stampTone: 'green' as const,
      accentColor: '#1A365D',
      icon: '🏏',
      rules: 'Throw runs 1-6 simultaneously. If batsman and bowler match numbers, batsman is OUT!',
    },
    {
      id: 'dots_boxes',
      title: 'Dots & Boxes',
      tagline: 'Territory conquest on graph paper grids',
      players: '2 Players',
      type: 'Strategy',
      duration: '4-8 mins',
      paperType: 'graph' as const,
      stampText: 'TACTICAL',
      stampTone: 'blue' as const,
      accentColor: '#15803D',
      icon: '⚄',
      rules: 'Connect adjacent dots with pencil lines. Completing the 4th side captures the box!',
    },
    {
      id: 'tic_tac_toe',
      title: 'XO / Tic-Tac-Toe',
      tagline: 'Lightning speed classroom margin duel',
      players: '2 Players',
      type: 'Speed Duel',
      duration: '1-2 mins',
      paperType: 'ruled' as const,
      stampText: 'CLASSIC',
      stampTone: 'amber' as const,
      accentColor: '#991B1B',
      icon: '✕◯',
      rules: 'Get 3 X or O symbols in a row before the 5-second turn timer runs out.',
    },
    {
      id: 'connect_4',
      title: 'Connect 4',
      tagline: 'Vertical wooden slot disc strategy',
      players: '2 Players',
      type: 'Strategy',
      duration: '3-6 mins',
      paperType: 'plain' as const,
      stampText: 'RIVALRY',
      stampTone: 'red' as const,
      accentColor: '#B45309',
      icon: '⚪',
      rules: 'Drop discs into the 7x6 wooden rack. First to link 4 horizontally, vertically, or diagonally wins.',
    },
    {
      id: 'paper_football',
      title: 'Paper Football',
      tagline: 'Tabletop triangle flick & field goal showdown',
      players: '2 Players',
      type: 'Physics Skill',
      duration: '3-5 mins',
      paperType: 'ruled' as const,
      stampText: 'DESK PHYSICS',
      stampTone: 'blue' as const,
      accentColor: '#475569',
      icon: '📐',
      rules: 'Flick the folded paper football across the desk without overshooting the table edge!',
    },
    {
      id: 'name_place_animal_thing',
      title: 'Name–Place–Animal–Thing',
      tagline: 'Multiplayer rapid alphabet buzzer sheets',
      players: '2-6 Players',
      type: 'Word Puzzle',
      duration: '5-10 mins',
      paperType: 'ruled' as const,
      stampText: 'PARTY GAME',
      stampTone: 'green' as const,
      accentColor: '#7C3AED',
      icon: '📝',
      rules: 'A random letter is picked. Fill all 4 columns before someone yells STOP!',
    },
  ]

  return (
    <div className="space-y-16 pb-12">
      {/* 1. Hero Notice Board Section */}
      <section className="relative">
        <PaperCard
          variant="cork"
          className="p-8 sm:p-12 relative overflow-hidden"
        >
          {/* Main Pinned Paper Poster in center of notice board */}
          <div className="relative bg-white dark:bg-[#141C2E] rounded-md p-6 sm:p-10 border-2 border-[#1E242B] dark:border-slate-700 shadow-[4px_4px_0px_0px_#1E242B] dark:shadow-[4px_4px_0px_0px_#020617] max-w-4xl mx-auto text-center">
            {/* Top center push-pin */}
            <div className="absolute -top-3 left-1/2 -translate-x-1/2 z-20">
              <div className="push-pin-red" />
            </div>

            {/* School Seal Header */}
            <div className="flex items-center justify-center gap-2 mb-3">
              <Stamp tone="red">RECESS PERIOD IS ACTIVE</Stamp>
            </div>

            <h1 className="text-3xl sm:text-5xl font-extrabold text-[#1E242B] dark:text-slate-100 tracking-tight leading-tight">
              Classroom Multiplayer Games, <br />
              <span className="font-hand text-4xl sm:text-6xl text-[#1A365D] dark:text-sky-300 underline decoration-[#991B1B]/40 dark:decoration-red-500/40 decoration-wavy">
                Recreated for Back-Benchers.
              </span>
            </h1>

            <p className="font-sans text-base sm:text-lg text-[#475569] dark:text-slate-300 max-w-2xl mx-auto mt-4 leading-relaxed">
              No generic mini-games. No cartoon nonsense. Pure server-authoritative multiplayer
              with real-time WebSockets, Elo rankings, and authentic school stationery aesthetics.
            </p>

            {/* CTA Action Buttons */}
            <div className="flex flex-wrap items-center justify-center gap-4 mt-8">
              {isAuthenticated ? (
                <Link to="/dashboard">
                  <Button size="lg" variant="primary" rightIcon={<ArrowRight className="w-5 h-5" />}>
                    Go to Classroom Hub
                  </Button>
                </Link>
              ) : (
                <>
                  <Button
                    size="lg"
                    variant="primary"
                    onClick={handleQuickGuestPlay}
                    isLoading={isLoading}
                    leftIcon={<Sparkles className="w-5 h-5 text-[#FEF08A]" />}
                  >
                    Play Instantly as Guest
                  </Button>
                  <Link to="/register">
                    <Button size="lg" variant="secondary">
                      Enroll Student Account
                    </Button>
                  </Link>
                </>
              )}
              <Link to="/games">
                <Button size="lg" variant="ghost">
                  Browse All 6 Games →
                </Button>
              </Link>
            </div>

            {/* Schoolyard Trust Badges */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 pt-8 mt-8 border-t border-[#E2E8F0] dark:border-slate-700/60 text-xs font-mono text-[#475569] dark:text-slate-300">
              <div className="flex items-center justify-center gap-1.5">
                <Zap className="w-4 h-4 text-[#15803D] dark:text-emerald-400" />
                <span>&lt;20ms WebSockets</span>
              </div>
              <div className="flex items-center justify-center gap-1.5">
                <ShieldCheck className="w-4 h-4 text-[#1A365D] dark:text-sky-400" />
                <span>Anti-Cheat Authoritative</span>
              </div>
              <div className="flex items-center justify-center gap-1.5">
                <RotateCcw className="w-4 h-4 text-[#B45309] dark:text-amber-400" />
                <span>Instant Reconnects</span>
              </div>
              <div className="flex items-center justify-center gap-1.5">
                <Trophy className="w-4 h-4 text-[#991B1B] dark:text-red-400" />
                <span>Ranked Elo Matchmaking</span>
              </div>
            </div>
          </div>
        </PaperCard>
      </section>

      {/* 2. Featured School Games Catalog */}
      <section className="space-y-6">
        <div className="flex flex-col sm:flex-row sm:items-end justify-between gap-2 border-b-2 border-[#1E242B] dark:border-slate-700 pb-3">
          <div>
            <div className="flex items-center gap-2">
              <h2 className="text-2xl sm:text-3xl font-extrabold text-[#1E242B] dark:text-slate-100">
                The 6 Schoolyard Classics
              </h2>
              <Stamp tone="blue">OFFICIAL SYLLABUS</Stamp>
            </div>
            <p className="font-hand text-lg text-[#475569] dark:text-slate-300">
              Pick your game, challenge a classmate, or hop into ranked matchmaking.
            </p>
          </div>
          <Link to="/games">
            <Button variant="ghost" size="sm">
              View Syllabus & Rules →
            </Button>
          </Link>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {featuredGames.map((game) => (
            <PaperCard
              key={game.id}
              variant={game.paperType}
              className="flex flex-col justify-between hover:shadow-[5px_5px_0px_0px_#1E242B] dark:hover:shadow-[5px_5px_0px_0px_#020617] transition-all group cursor-pointer"
              onClick={() => navigate(`/games`)}
            >
              <div>
                <div className="flex items-start justify-between gap-2 mb-3">
                  <span className="text-3xl">{game.icon}</span>
                  <Stamp tone={game.stampTone}>{game.stampText}</Stamp>
                </div>

                <h3 className="text-xl font-bold text-[#1E242B] dark:text-slate-100 group-hover:text-[#1A365D] group-hover:dark:text-sky-300 transition-colors">
                  {game.title}
                </h3>
                <p className="font-hand text-base text-[#1A365D] dark:text-sky-300/90 mt-0.5">{game.tagline}</p>

                <p className="text-xs text-[#475569] dark:text-slate-300 mt-3 leading-relaxed bg-[#F8FAFC]/90 dark:bg-slate-900/60 p-2.5 rounded border border-[#E2E8F0] dark:border-slate-700/60">
                  <strong className="text-[#1E242B] dark:text-amber-300">How to play:</strong> {game.rules}
                </p>
              </div>

              <div className="pt-4 mt-4 border-t border-[#CBD5E1] dark:border-slate-700/60 flex items-center justify-between text-xs font-mono">
                <div className="flex items-center gap-2">
                  <Badge variant="default" size="sm">
                    {game.players}
                  </Badge>
                  <Badge variant="pencil" size="sm">
                    {game.duration}
                  </Badge>
                </div>
                <span className="font-bold text-[#1A365D] dark:text-sky-400 group-hover:text-blue-700 group-hover:dark:text-sky-300 group-hover:translate-x-1 transition-all">
                  Enter Arena →
                </span>
              </div>
            </PaperCard>
          ))}
        </div>
      </section>

      {/* 3. How Recess Works (3-Step Notebook Section) */}
      <section>
        <PaperCard variant="ruled" className="p-8 sm:p-10">
          <div className="max-w-3xl">
            <Stamp tone="green" className="mb-2">
              DESK PROTOCOL
            </Stamp>
            <h2 className="text-2xl sm:text-3xl font-extrabold text-[#1E242B] dark:text-slate-100">
              How To Start A Rivalry In 3 Steps
            </h2>
            <p className="font-hand text-lg text-[#475569] dark:text-slate-300 mt-1">
              Zero downloads. Zero installs. Works instantly on mobile and desktop browsers.
            </p>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-6 mt-8">
              {/* Step 1 */}
              <div className="space-y-2">
                <div className="w-9 h-9 rounded-full bg-[#1A365D] dark:bg-sky-900 text-white flex items-center justify-center font-bold text-base border-2 border-[#0F2238] dark:border-sky-700">
                  1
                </div>
                <h4 className="font-bold text-base text-[#1E242B] dark:text-slate-100">Pick Your Game</h4>
                <p className="text-xs text-[#475569] dark:text-slate-300 leading-relaxed">
                  Choose from Hand Cricket, Dots & Boxes, XO, Connect 4, Paper Football, or NPAT.
                </p>
              </div>

              {/* Step 2 */}
              <div className="space-y-2">
                <div className="w-9 h-9 rounded-full bg-[#15803D] dark:bg-emerald-900 text-white flex items-center justify-center font-bold text-base border-2 border-[#14532D] dark:border-emerald-700">
                  2
                </div>
                <h4 className="font-bold text-base text-[#1E242B] dark:text-slate-100">Share Room Code</h4>
                <p className="text-xs text-[#475569] dark:text-slate-300 leading-relaxed">
                  Generate a clean custom code like <code className="font-mono text-[11px] bg-[#F2EDE0] dark:bg-slate-800 text-[#1E242B] dark:text-amber-300 px-1 py-0.5 rounded">RECESS-7X9P</code> and text it to your friend.
                </p>
              </div>

              {/* Step 3 */}
              <div className="space-y-2">
                <div className="w-9 h-9 rounded-full bg-[#991B1B] dark:bg-red-900 text-white flex items-center justify-center font-bold text-base border-2 border-[#7F1D1D] dark:border-red-700">
                  3
                </div>
                <h4 className="font-bold text-base text-[#1E242B] dark:text-slate-100">Settle The Score</h4>
                <p className="text-xs text-[#475569] dark:text-slate-300 leading-relaxed">
                  Play in real-time with sub-20ms WebSocket response and climb the classroom honor roll.
                </p>
              </div>
            </div>
          </div>
        </PaperCard>
      </section>

      {/* 4. Principal's Honor Roll Preview */}
      <section className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <PaperCard variant="sticky" stickyColor="yellow" showTape tilt className="p-6">
          <div className="flex items-center gap-2 mb-2">
            <Trophy className="w-5 h-5 text-[#B45309]" />
            <h3 className="font-bold text-base text-[#1E242B]">Principal's Honor Roll</h3>
          </div>
          <p className="font-hand text-lg text-[#1A365D]">
            Top student rankers across all classrooms this week:
          </p>
          <ul className="space-y-2 mt-4 text-xs font-mono">
            <li className="flex items-center justify-between pb-1.5 border-b border-yellow-300">
              <span className="font-bold">1. @ArbiterRecess</span>
              <span className="font-bold text-[#15803D]">1480 ELO (A+)</span>
            </li>
            <li className="flex items-center justify-between pb-1.5 border-b border-yellow-300">
              <span>2. @ClassChampion</span>
              <span className="font-bold text-[#15803D]">1390 ELO (A)</span>
            </li>
            <li className="flex items-center justify-between">
              <span>3. @DeskCaptain</span>
              <span className="font-bold text-[#1A365D]">1310 ELO (B+)</span>
            </li>
          </ul>
        </PaperCard>

        <PaperCard variant="sticky" stickyColor="blue" showTape className="p-6">
          <div className="flex items-center gap-2 mb-2">
            <Swords className="w-5 h-5 text-[#1A365D]" />
            <h3 className="font-bold text-base text-[#1E242B]">Classroom Rivalries</h3>
          </div>
          <p className="font-hand text-lg text-[#1A365D]">
            Instant matchmaking matches you against players of identical skill ratings.
          </p>
          <div className="mt-4 pt-2">
            <Link to="/dashboard">
              <Button size="sm" variant="primary" className="w-full">
                Enter Matchmaking Queue
              </Button>
            </Link>
          </div>
        </PaperCard>

        <PaperCard variant="sticky" stickyColor="pink" showTape tilt className="p-6">
          <div className="flex items-center gap-2 mb-2">
            <Users className="w-5 h-5 text-[#991B1B]" />
            <h3 className="font-bold text-base text-[#1E242B]">Classroom Rooms</h3>
          </div>
          <p className="font-hand text-lg text-[#1A365D]">
            Host a private desk room with a passcode for friendly lunch break tournaments.
          </p>
          <div className="mt-4 pt-2">
            <Link to="/dashboard">
              <Button size="sm" variant="secondary" className="w-full">
                Create Private Room
              </Button>
            </Link>
          </div>
        </PaperCard>
      </section>
    </div>
  )
}
