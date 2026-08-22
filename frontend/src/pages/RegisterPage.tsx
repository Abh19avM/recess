import React, { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'
import { Button } from '../components/ui/Button'
import { Input } from '../components/ui/Input'
import { PaperCard } from '../components/ui/PaperCard'
import { Stamp } from '../components/ui/Badge'
import { UserPlus, User, Mail, Lock } from 'lucide-react'
import { toast } from '../store/toastStore'

export const RegisterPage: React.FC = () => {
  const navigate = useNavigate()
  const { register, isLoading, error, clearError } = useAuthStore()

  const [username, setUsername] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [clientError, setClientError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    clearError()
    setClientError(null)

    if (username.trim().length < 3) {
      setClientError('Username must be at least 3 characters.')
      return
    }

    if (password.length < 6) {
      setClientError('Passcode must be at least 6 characters.')
      return
    }

    if (password !== confirmPassword) {
      setClientError('Passcodes do not match!')
      return
    }

    try {
      await register({
        username: username.trim(),
        email: email.trim() || undefined,
        password,
      })
      toast.success('Enrollment Approved!', `Welcome to Recess, @${username}!`)
      navigate('/dashboard')
    } catch {
      // Handled in store
    }
  }

  const displayError = clientError || error

  return (
    <div className="max-w-md mx-auto py-8">
      <PaperCard variant="ruled" className="p-8 sm:p-10 relative">
        {/* Top Stamp */}
        <div className="flex items-center justify-between mb-4">
          <span className="font-hand text-xl text-[#1A365D] font-bold">New Admission</span>
          <Stamp tone="green">ENROLLMENT</Stamp>
        </div>

        <h2 className="text-2xl sm:text-3xl font-extrabold text-[#1E242B] tracking-tight">
          Student Registration
        </h2>
        <p className="font-hand text-base text-[#475569] mt-0.5">
          Claim your unique classroom handle to track rankings, streaks, and tournament trophies.
        </p>

        <form onSubmit={handleSubmit} className="space-y-4 mt-6">
          <Input
            label="Classroom Handle"
            placeholder="e.g. BackbenchCaptain"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            leftIcon={<User className="w-4 h-4" />}
            hint="3-30 characters, letters, numbers, underscores."
            required
          />

          <Input
            label="School Email (Optional)"
            type="email"
            placeholder="student@school.edu"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            leftIcon={<Mail className="w-4 h-4" />}
            hint="Used for password recovery and tournament notifications."
          />

          <Input
            label="Passcode"
            type="password"
            placeholder="Minimum 6 characters"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            leftIcon={<Lock className="w-4 h-4" />}
            required
          />

          <Input
            label="Confirm Passcode"
            type="password"
            placeholder="Re-enter passcode"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            leftIcon={<Lock className="w-4 h-4" />}
            required
          />

          {displayError && (
            <div className="bg-[#FFE4E6] p-3 rounded border border-[#FECDD3] text-xs font-bold text-[#991B1B] font-hand text-sm flex items-center gap-1.5">
              <span>✎</span> {displayError}
            </div>
          )}

          <div className="pt-2">
            <Button
              type="submit"
              variant="primary"
              size="md"
              className="w-full"
              isLoading={isLoading}
              leftIcon={<UserPlus className="w-4 h-4" />}
            >
              Submit Admission Form
            </Button>
          </div>
        </form>

        <div className="mt-8 pt-4 border-t border-[#CBD5E1] text-center text-xs text-[#475569]">
          <span>Already registered? </span>
          <Link to="/login" className="font-bold text-[#1A365D] hover:underline">
            Sign In Here →
          </Link>
        </div>
      </PaperCard>
    </div>
  )
}
