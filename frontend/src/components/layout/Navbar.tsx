import React from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../../store/authStore'
import { Button } from '../ui/Button'
import { Avatar } from '../ui/Avatar'
import { Stamp } from '../ui/Badge'
import { LogOut, LayoutDashboard, Dices, Sparkles, Wifi } from 'lucide-react'

export const Navbar: React.FC = () => {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, isAuthenticated, logout, guestLogin, isLoading } = useAuthStore()

  const navLinks = [
    { name: 'Play Games', path: '/games', icon: <Dices className="w-4 h-4" /> },
    { name: 'Classroom Hub', path: '/dashboard', icon: <LayoutDashboard className="w-4 h-4" /> },
    { name: 'WSS Bench', path: '/ws-test', icon: <Wifi className="w-4 h-4 text-[#15803D]" /> },
  ]

  const handleGuestPlay = async () => {
    try {
      await guestLogin()
      navigate('/dashboard')
    } catch {
      // Handled in store
    }
  }

  return (
    <header className="sticky top-0 z-40 bg-[#FFFFFF] border-b-2 border-[#1E242B] shadow-[0_2px_4px_rgba(0,0,0,0.04)]">
      {/* Top red margin strip */}
      <div className="h-1 bg-[#991B1B] w-full" />

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <div className="flex items-center gap-6">
            <Link to="/" className="flex items-center gap-2 group">
              <div className="w-9 h-9 rounded bg-[#1A365D] text-white flex items-center justify-center font-bold text-lg border-2 border-[#0F2238] shadow-[2px_2px_0px_0px_#0F2238] group-hover:-translate-y-0.5 transition-transform">
                <span className="font-hand text-xl">R</span>
              </div>
              <div className="flex flex-col">
                <div className="flex items-center gap-1.5">
                  <span className="font-bold text-xl tracking-tight text-[#1E242B]">
                    RECESS
                  </span>
                  <Stamp tone="amber" className="hidden sm:inline-flex text-[10px] py-0 px-1.5">
                    BELL ON
                  </Stamp>
                </div>
                <span className="text-[10px] font-mono text-[#475569] tracking-wider uppercase -mt-1 hidden sm:block">
                  Classic Schoolyard Games
                </span>
              </div>
            </Link>

            {/* Navigation Tabs */}
            <nav className="hidden md:flex items-center gap-1 ml-4">
              {navLinks.map((link) => {
                const isActive = location.pathname === link.path
                return (
                  <Link
                    key={link.path}
                    to={link.path}
                    className={`flex items-center gap-2 px-3 py-1.5 rounded-md text-sm font-semibold transition-all ${
                      isActive
                        ? 'bg-[#1A365D] text-white border border-[#0F2238] shadow-[2px_2px_0px_0px_#0F2238]'
                        : 'text-[#475569] hover:bg-[#F2EDE0] hover:text-[#1E242B]'
                    }`}
                  >
                    {link.icon}
                    <span>{link.name}</span>
                  </Link>
                )
              })}
            </nav>
          </div>

          {/* Right Action Items / Student Profile */}
          <div className="flex items-center gap-3">
            {isAuthenticated && user ? (
              <div className="flex items-center gap-3">
                <Link
                  to="/profile"
                  className="flex items-center gap-2 px-3 py-1.5 rounded-md bg-[#FBF9F3] border-2 border-[#CBD5E1] hover:border-[#1A365D] transition-colors shadow-xs"
                >
                  <Avatar username={user.username} preset={user.avatar_preset} size="sm" isOnline />
                  <div className="text-left hidden sm:block">
                    <p className="text-xs font-bold text-[#1E242B] leading-none">
                      {user.username}
                    </p>
                    <p className="text-[10px] font-mono text-[#15803D] font-semibold">
                      {user.rating} ELO {user.is_guest && '(Guest)'}
                    </p>
                  </div>
                </Link>

                <Button
                  onClick={logout}
                  variant="ghost"
                  size="sm"
                  aria-label="Logout"
                  className="text-[#475569] hover:text-[#991B1B]"
                >
                  <LogOut className="w-4 h-4" />
                </Button>
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <Button
                  onClick={handleGuestPlay}
                  isLoading={isLoading}
                  variant="secondary"
                  size="sm"
                  leftIcon={<Sparkles className="w-3.5 h-3.5 text-[#B45309]" />}
                  className="hidden sm:inline-flex"
                >
                  Quick Guest Play
                </Button>
                <Link to="/login">
                  <Button variant="ghost" size="sm">
                    Sign In
                  </Button>
                </Link>
                <Link to="/register">
                  <Button variant="primary" size="sm">
                    Join School
                  </Button>
                </Link>
              </div>
            )}
          </div>
        </div>
      </div>
    </header>
  )
}
