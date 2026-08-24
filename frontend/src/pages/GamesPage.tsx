import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { PaperCard } from '../components/ui/PaperCard'
import { Button } from '../components/ui/Button'
import { Badge, Stamp } from '../components/ui/Badge'
import { MatchmakingModal } from '../components/ui/MatchmakingModal'
import { Swords, Users, BookOpen, Eye, Radio } from 'lucide-react'

export const GamesPage: React.FC = () => {
  const navigate = useNavigate()
  const [filter, setFilter] = useState<'all' | '2p' | 'turn' | 'strategy'>('all')
  const [matchmakingGame, setMatchmakingGame] = useState<{ id: string; title: string; icon: string } | null>(null)
  const [spectateRoomId, setSpectateRoomId] = useState('')

  const handleSpectate = (e: React.FormEvent) => {
    e.preventDefault()
    const clean = spectateRoomId.trim().toUpperCase()
    if (!clean) return
    navigate(`/spectate/${clean}`)
  }

  const games = [
    {
      id: 'hand_cricket',
      title: 'Hand Cricket',
      category: 'turn',
      is2P: true,
      tagline: 'Last-bench finger cricket with run calls and sudden-death wickets',
      icon: '🏏',
      players: '2 Players',
      duration: '3-5 Mins',
      difficulty: 'Easy',
      surface: 'Ruled Pitch',
      paperVariant: 'ruled' as const,
      rules: [
        'Both players throw a number from 1 to 6 simultaneously.',
        'If the bowler and batsman throw different numbers, the batsman scores runs equal to their number.',
        'If both players throw the exact same number, the batsman is OUT!',
        'Innings switch after wicket. Highest run total wins the match.',
      ],
      proTip: 'Watch your opponent’s past 5 throws — back-benchers often develop subconscious rhythm patterns.',
    },
    {
      id: 'dots_boxes',
      title: 'Dots & Boxes',
      category: 'strategy',
      is2P: true,
      tagline: 'Territorial warfare on graph paper grids',
      icon: '⚄',
      players: '2 Players',
      duration: '5-8 Mins',
      difficulty: 'Medium',
      surface: 'Graph Paper',
      paperVariant: 'graph' as const,
      rules: [
        'Players take turns drawing horizontal or vertical pencil lines connecting adjacent dots.',
        'Completing the 4th side of a 1x1 box claims it and awards an immediate bonus turn.',
        'Chains of boxes can be captured consecutively in a single master turn.',
        'Player with the highest box count when the grid is full wins.',
      ],
      proTip: 'Avoid giving away 3-sided boxes early. Sacrificing 2 boxes to gain a long chain is the classic gambit.',
    },
    {
      id: 'tic_tac_toe',
      title: 'XO / Tic-Tac-Toe',
      category: 'turn',
      is2P: true,
      tagline: 'High-speed classroom margin duel with strict turn timers',
      icon: '✕◯',
      players: '2 Players',
      duration: '1-2 Mins',
      difficulty: 'Easy',
      surface: 'Chalkboard / Margin',
      paperVariant: 'ruled' as const,
      rules: [
        'Player 1 places X, Player 2 places O on a 3x3 grid.',
        'Turn timer is strictly 5 seconds to force rapid instinct play.',
        'First to align 3 consecutive marks horizontally, vertically, or diagonally wins.',
        'Played in best-of-3 classroom series.',
      ],
      proTip: 'Control the center square or create a dual-fork opportunity from opposing corners.',
    },
    {
      id: 'connect_4',
      title: 'Connect 4',
      category: 'strategy',
      is2P: true,
      tagline: 'Gravity drop columns in a wooden classroom rack',
      icon: '⚪',
      players: '2 Players',
      duration: '3-6 Mins',
      difficulty: 'Medium',
      surface: 'Wooden Frame',
      paperVariant: 'plain' as const,
      rules: [
        'Discs drop straight down into the 7-column x 6-row vertical frame.',
        'Players alternate dropping colored discs.',
        'First player to connect 4 discs in a line (horizontal, vertical, or diagonal) wins.',
      ],
      proTip: 'Building vertical threats in column 4 (the center column) gives maximum geometric advantage.',
    },
    {
      id: 'paper_football',
      title: 'Paper Football',
      category: 'strategy',
      is2P: true,
      tagline: 'Tabletop triangle flick physics & field goal showdowns',
      icon: '📐',
      players: '2 Players',
      duration: '3-5 Mins',
      difficulty: 'Hard',
      surface: 'Oak Wood Table',
      paperVariant: 'ruled' as const,
      rules: [
        'Flick the folded paper football across the desk in 4 downs.',
        'If the triangle stops hanging over the edge without falling, you score a Touchdown (6 pts)!',
        'Follow up with a field goal kick through your opponent’s finger goal posts for extra points.',
      ],
      proTip: 'Gentle angled flicks with your thumb and index finger offer better stopping friction.',
    },
    {
      id: 'name_place_animal_thing',
      title: 'Name–Place–Animal–Thing',
      category: 'turn',
      is2P: false,
      tagline: 'Rapid multiplayer alphabet buzzer category sheet',
      icon: '📝',
      players: '2-6 Players',
      duration: '5-10 Mins',
      difficulty: 'Medium',
      surface: 'Exam Sheet',
      paperVariant: 'ruled' as const,
      rules: [
        'A random letter of the alphabet is announced.',
        'All players frantically write a Name, Place, Animal, and Thing starting with that letter.',
        'First player to finish hits STOP, giving others 10 seconds to submit.',
        'Unique answers score 10 pts; shared duplicate answers score 5 pts.',
      ],
      proTip: 'Obscure geography and animal species yield maximum unique 10-point scores.',
    },
  ]

  const filteredGames = games.filter((g) => {
    if (filter === '2p') return g.is2P
    if (filter === 'turn') return g.category === 'turn'
    if (filter === 'strategy') return g.category === 'strategy'
    return true
  })

  return (
    <div className="space-y-8 pb-12">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b-2 border-[#1E242B] pb-4">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-3xl font-extrabold text-[#1E242B]">Classroom Games Syllabus</h1>
            <Stamp tone="blue">OFFICIAL RULES</Stamp>
          </div>
          <p className="font-hand text-lg text-[#475569] mt-0.5">
            Full game rules, mechanics, and strategies for all 6 classic schoolyard competitions.
          </p>
        </div>

        {/* Filter Tabs */}
        <div className="flex items-center gap-1 bg-[#F2EDE0] p-1 rounded-md border border-[#CBD5E1] text-xs font-bold">
          <button
            onClick={() => setFilter('all')}
            className={`px-3 py-1.5 rounded transition-all ${
              filter === 'all'
                ? 'bg-white text-[#1A365D] shadow-xs border border-[#CBD5E1]'
                : 'text-[#475569] hover:text-[#1E242B]'
            }`}
          >
            All 6 Games
          </button>
          <button
            onClick={() => setFilter('2p')}
            className={`px-3 py-1.5 rounded transition-all ${
              filter === '2p'
                ? 'bg-white text-[#1A365D] shadow-xs border border-[#CBD5E1]'
                : 'text-[#475569] hover:text-[#1E242B]'
            }`}
          >
            2-Player
          </button>
          <button
            onClick={() => setFilter('turn')}
            className={`px-3 py-1.5 rounded transition-all ${
              filter === 'turn'
                ? 'bg-white text-[#1A365D] shadow-xs border border-[#CBD5E1]'
                : 'text-[#475569] hover:text-[#1E242B]'
            }`}
          >
            Turn-Based
          </button>
          <button
            onClick={() => setFilter('strategy')}
            className={`px-3 py-1.5 rounded transition-all ${
              filter === 'strategy'
                ? 'bg-white text-[#1A365D] shadow-xs border border-[#CBD5E1]'
                : 'text-[#475569] hover:text-[#1E242B]'
            }`}
          >
            Tactical Strategy
          </button>
        </div>
      </div>

      {/* Sideline Spectator Live Bar */}
      <div className="bg-white border-2 border-[#4A6B82]/30 rounded-2xl p-4 sm:p-5 shadow-md flex flex-col md:flex-row items-center justify-between gap-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-red-100 border border-red-300 flex items-center justify-center text-red-600">
            <Radio className="w-5 h-5 animate-pulse" />
          </div>
          <div>
            <h3 className="font-bold text-sm text-[#2C3E50] flex items-center gap-2">
              Schoolyard Sideline Spectate
              <span className="text-[10px] px-2 py-0.5 bg-red-100 text-red-700 font-extrabold rounded-full">
                LIVE
              </span>
            </h3>
            <p className="font-hand text-xs text-[#475569]">
              Watch any active desk duel in real time as a bystander without making moves.
            </p>
          </div>
        </div>

        <form onSubmit={handleSpectate} className="flex items-center gap-2 w-full md:w-auto">
          <input
            type="text"
            value={spectateRoomId}
            onChange={(e) => setSpectateRoomId(e.target.value)}
            placeholder="Enter Room Code (e.g. RECESS-XO)"
            className="px-3.5 py-2 text-xs font-mono bg-[#FBF9F1] border-2 border-[#4A6B82]/30 rounded-xl focus:outline-none focus:ring-2 focus:ring-[#4A6B82]/40 w-full sm:w-64"
          />
          <Button
            type="submit"
            variant="primary"
            size="sm"
            disabled={!spectateRoomId.trim()}
            leftIcon={<Eye className="w-4 h-4" />}
          >
            Watch Match
          </Button>
        </form>
      </div>

      {/* Games List */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {filteredGames.map((game) => (
          <PaperCard
            key={game.id}
            variant={game.paperVariant}
            className="p-6 sm:p-8 flex flex-col justify-between"
          >
            <div>
              <div className="flex items-start justify-between gap-3 mb-2">
                <div className="flex items-center gap-3">
                  <span className="text-4xl">{game.icon}</span>
                  <div>
                    <h3 className="text-2xl font-bold text-[#1E242B]">{game.title}</h3>
                    <p className="font-hand text-base text-[#1A365D]">{game.tagline}</p>
                  </div>
                </div>
                <Stamp tone="green">VERIFIED</Stamp>
              </div>

              {/* Game Metadata Chips */}
              <div className="flex flex-wrap items-center gap-2 my-4">
                <Badge variant="ink-blue" size="sm">
                  {game.players}
                </Badge>
                <Badge variant="pencil" size="sm">
                  {game.duration}
                </Badge>
                <Badge variant="amber" size="sm">
                  {game.difficulty}
                </Badge>
                <Badge variant="default" size="sm">
                  {game.surface}
                </Badge>
              </div>

              {/* Rules List */}
              <div className="bg-[#F8FAFC]/90 p-4 rounded-md border border-[#CBD5E1] my-4 space-y-2">
                <h4 className="font-bold text-xs uppercase tracking-wider text-[#1E242B] font-mono flex items-center gap-1.5">
                  <BookOpen className="w-3.5 h-3.5 text-[#1A365D]" />
                  Rules & Mechanics:
                </h4>
                <ul className="space-y-1 text-xs text-[#475569] list-disc list-inside leading-relaxed">
                  {game.rules.map((rule, idx) => (
                    <li key={idx}>
                      <span>{rule}</span>
                    </li>
                  ))}
                </ul>
              </div>

              {/* Pro Tip Callout */}
              <div className="p-3 rounded bg-[#FEF9C3]/80 border border-[#FDE047] text-xs">
                <p className="font-hand text-sm text-[#854D0E]">
                  <strong className="font-bold">Back-Bench Secret:</strong> {game.proTip}
                </p>
              </div>
            </div>

            {/* Actions */}
            <div className="pt-6 mt-6 border-t border-[#CBD5E1] flex items-center justify-between gap-3">
              <Button
                onClick={() => {
                  setMatchmakingGame({
                    id: game.id,
                    title: game.title,
                    icon: game.icon,
                  })
                }}
                variant="primary"
                size="md"
                className="flex-1"
                leftIcon={<Swords className="w-4 h-4" />}
              >
                Find Match
              </Button>
              <Button
                onClick={() => {
                  if (game.id === 'tic_tac_toe') {
                    navigate('/games/xo')
                  } else if (game.id === 'hand_cricket') {
                    navigate('/games/hand-cricket')
                  } else if (game.id === 'dots_boxes') {
                    navigate('/games/dots-and-boxes')
                  } else if (game.id === 'connect_4') {
                    navigate('/games/connect-4')
                  } else if (game.id === 'paper_football') {
                    navigate('/games/paper-football')
                  } else if (game.id === 'name_place_animal_thing' || game.id === 'npat') {
                    navigate('/games/npat')
                  } else {
                    navigate('/dashboard')
                  }
                }}
                variant="secondary"
                size="md"
                className="flex-1"
                leftIcon={<Users className="w-4 h-4" />}
              >
                Enter Arena
              </Button>
            </div>
          </PaperCard>
        ))}
      </div>

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
