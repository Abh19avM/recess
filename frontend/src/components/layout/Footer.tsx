import React from 'react'
import { Link } from 'react-router-dom'
import { Stamp } from '../ui/Badge'
import { Heart } from 'lucide-react'

export const Footer: React.FC = () => {
  const gamesList = [
    { name: 'Hand Cricket', id: 'hand_cricket' },
    { name: 'Dots & Boxes', id: 'dots_boxes' },
    { name: 'XO / Tic-Tac-Toe', id: 'tic_tac_toe' },
    { name: 'Connect 4', id: 'connect_4' },
    { name: 'Paper Football', id: 'paper_football' },
    { name: 'Name-Place-Animal-Thing', id: 'name_place_animal_thing' },
  ]

  return (
    <footer className="mt-auto bg-[#F2EDE0] dark:bg-[#131C16] border-t-2 border-[#1E242B] dark:border-[#334155] text-[#475569] dark:text-[#94A3B8] text-sm py-12 transition-colors">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8 mb-8">
          {/* Col 1: Brand & Nostalgia */}
          <div className="space-y-3 md:col-span-2">
            <div className="flex items-center gap-2">
              <span className="font-bold text-xl text-[#1E242B]">RECESS</span>
              <Stamp tone="green">A+ VERIFIED</Stamp>
            </div>
            <p className="font-hand text-lg text-[#1A365D] max-w-md">
              "The best rivalries, strategies, and memories were born between classes, on ruled notebook margins, with graphite pencils."
            </p>
            <div className="flex items-center gap-2 text-xs font-mono text-[#64748B]">
              <span>SERVER AUTHORITATIVE</span>
              <span>•</span>
              <span>REAL-TIME WEBSOCKETS</span>
              <span>•</span>
              <span>ZERO ADS</span>
            </div>
          </div>

          {/* Col 2: Classic Games Directory */}
          <div>
            <h4 className="font-bold text-xs uppercase tracking-wider text-[#1E242B] mb-3 font-mono">
              Schoolyard Games
            </h4>
            <ul className="space-y-1.5 text-xs font-medium">
              {gamesList.map((g) => (
                <li key={g.id}>
                  <Link
                    to={`/games`}
                    className="hover:text-[#1A365D] hover:underline transition-colors"
                  >
                    {g.name}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          {/* Col 3: Classroom Links */}
          <div>
            <h4 className="font-bold text-xs uppercase tracking-wider text-[#1E242B] mb-3 font-mono">
              Classroom Directory
            </h4>
            <ul className="space-y-1.5 text-xs font-medium">
              <li>
                <Link to="/dashboard" className="hover:text-[#1A365D] hover:underline">
                  Classroom Hub
                </Link>
              </li>
              <li>
                <Link to="/games" className="hover:text-[#1A365D] hover:underline">
                  Game Lobby Directory
                </Link>
              </li>
              <li>
                <Link to="/profile" className="hover:text-[#1A365D] hover:underline">
                  Student Report Card
                </Link>
              </li>
              <li>
                <Link to="/register" className="hover:text-[#1A365D] hover:underline">
                  Enroll New Student
                </Link>
              </li>
            </ul>
          </div>
        </div>

        {/* Bottom divider */}
        <div className="pt-6 border-t border-[#CBD5E1] flex flex-col sm:flex-row items-center justify-between gap-4 text-xs">
          <p>© {new Date().getFullYear()} Recess Multiplayer. Built for pure competitive nostalgia.</p>
          <div className="flex items-center gap-1 font-hand text-sm text-[#1E242B]">
            <span>Crafted with</span>
            <Heart className="w-3.5 h-3.5 text-[#991B1B] fill-[#991B1B]" />
            <span>for back-benchers worldwide</span>
          </div>
        </div>
      </div>
    </footer>
  )
}
