import React, { useState } from 'react'
import { Link, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { PaperCard } from '../components/ui/PaperCard'
import { Stamp } from '../components/ui/Badge'
import { LogIn, Sparkles, User, Lock } from 'lucide-react'
import { toast } from '../store/toastStore'

export const LoginPage: React.FC = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { login, guestLogin, isLoading, error, clearError } = useAuthStore()

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [guestName, setGuestName] = useState('')
  const [isGuestMode, setIsGuestMode] = useState(false)

  const from = (location.state as any)?.from?.pathname || '/dashboard'

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    clearError()

    if (!username.trim() || !password) {
      return
    }

    try {
      await login({ username: username.trim(), password })
      toast.success('Welcome back to class!', `Signed in as @${username}`)
      navigate(from, { replace: true })
    } catch {
      // Error is stored in authStore
    }
  }

  const handleGuestSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    clearError()

    try {
      await guestLogin(guestName.trim() || undefined)
      toast.success('Guest Session Active', 'You are now ready to play!')
      navigate(from, { replace: true })
    } catch {
      // Handled in store
    }
  }

  return (
    <div className="max-w-md mx-auto py-8">
      <PaperCard variant="ruled" className="p-8 sm:p-10 relative">
        {/* Top Stamp */}
        <div className="flex items-center justify-between mb-4">
          <span className="font-hand text-xl text-[#1A365D] font-bold">Class Attendance</span>
          <Stamp tone="blue">ROLL CALL</Stamp>
        </div>

        <h2 className="text-2xl sm:text-3xl font-extrabold text-[#1E242B] tracking-tight">
          {isGuestMode ? 'Instant Guest Pass' : 'Student Sign In'}
        </h2>
        <p className="font-hand text-base text-[#475569] mt-0.5">
          {isGuestMode
            ? 'Jump straight into games without registration.'
            : 'Enter your classroom handle and passcode.'}
        </p>

        {/* Tab switcher */}
        <div className="flex rounded-md border border-[#CBD5E1] p-1 bg-[#F2EDE0] my-6">
          <button
            type="button"
            onClick={() => {
              setIsGuestMode(false)
              clearError()
            }}
            className={`flex-1 py-1.5 text-xs font-bold rounded transition-all ${
              !isGuestMode
                ? 'bg-white text-[#1A365D] shadow-xs border border-[#CBD5E1]'
                : 'text-[#475569] hover:text-[#1E242B]'
            }`}
          >
            Registered Student
          </button>
          <button
            type="button"
            onClick={() => {
              setIsGuestMode(true)
              clearError()
            }}
            className={`flex-1 py-1.5 text-xs font-bold rounded transition-all flex items-center justify-center gap-1 ${
              isGuestMode
                ? 'bg-white text-[#B45309] shadow-xs border border-[#CBD5E1]'
                : 'text-[#475569] hover:text-[#1E242B]'
            }`}
          >
            <Sparkles className="w-3 h-3 text-[#B45309]" />
            Guest Pass
          </button>
        </div>

        {/* Regular Login Form */}
        {!isGuestMode ? (
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              label="Student Handle or Email"
              placeholder="e.g. ArbiterRecess"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              leftIcon={<User className="w-4 h-4" />}
              required
            />

            <Input
              label="Passcode"
              type="password"
              placeholder="••••••••"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              leftIcon={<Lock className="w-4 h-4" />}
              required
            />

            {error && (
              <div className="bg-[#FFE4E6] p-3 rounded border border-[#FECDD3] text-xs font-bold text-[#991B1B] font-hand text-sm flex items-center gap-1.5">
                <span>✎</span> {error}
              </div>
            )}

            <div className="pt-2">
              <Button
                type="submit"
                variant="primary"
                size="md"
                className="w-full"
                isLoading={isLoading}
                leftIcon={<LogIn className="w-4 h-4" />}
              >
                Sign In to Classroom
              </Button>
            </div>
          </form>
        ) : (
          /* Guest Form */
          <form onSubmit={handleGuestSubmit} className="space-y-4">
            <Input
              label="Temporary Nickname (Optional)"
              placeholder="e.g. QuickRunner"
              value={guestName}
              onChange={(e) => setGuestName(e.target.value)}
              leftIcon={<User className="w-4 h-4" />}
              hint="You can link an email later to save your rankings permanently."
            />

            {error && (
              <div className="bg-[#FFE4E6] p-3 rounded border border-[#FECDD3] text-xs font-bold text-[#991B1B] font-hand text-sm flex items-center gap-1.5">
                <span>✎</span> {error}
              </div>
            )}

            <div className="pt-2">
              <Button
                type="submit"
                variant="secondary"
                size="md"
                className="w-full"
                isLoading={isLoading}
                leftIcon={<Sparkles className="w-4 h-4 text-[#B45309]" />}
              >
                Launch Guest Session
              </Button>
            </div>
          </form>
        )}

        {/* Footer info */}
        <div className="mt-8 pt-4 border-t border-[#CBD5E1] text-center text-xs text-[#475569]">
          <span>New to Recess? </span>
          <Link to="/register" className="font-bold text-[#1A365D] hover:underline">
            Enroll New Student Account →
          </Link>
        </div>
      </PaperCard>
    </div>
  )
}
