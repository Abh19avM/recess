import React from 'react'
import { cn } from '../../lib/utils'

export interface AvatarProps {
  username?: string
  preset?: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  isOnline?: boolean
  className?: string
}

export const Avatar: React.FC<AvatarProps> = ({
  username = 'Student',
  preset = 'pencil_sketch_1',
  size = 'md',
  isOnline,
  className,
}) => {
  const sizeClasses = {
    sm: 'w-8 h-8 text-xs',
    md: 'w-10 h-10 text-sm',
    lg: 'w-14 h-14 text-lg',
    xl: 'w-20 h-20 text-2xl',
  }

  const getInitials = (name: string) => {
    return name.slice(0, 2).toUpperCase()
  }

  // Pre-designed nostalgic pencil avatars
  const avatarColors: Record<string, { bg: string; text: string; border: string }> = {
    pencil_sketch_1: { bg: 'bg-[#FEF08A]/40', text: 'text-[#1E242B]', border: 'border-[#475569]' },
    pencil_sketch_guest: { bg: 'bg-[#E2E8F0]', text: 'text-[#475569]', border: 'border-[#94A3B8]' },
    chalkboard_star: { bg: 'bg-[#18231C]', text: 'text-[#FEF08A]', border: 'border-[#5c4033]' },
    prefect_badge: { bg: 'bg-[#1A365D]', text: 'text-white', border: 'border-[#0F2238]' },
  }

  const activeTheme = avatarColors[preset] || avatarColors.pencil_sketch_1

  return (
    <div className="relative inline-block select-none">
      <div
        className={cn(
          'rounded-full flex items-center justify-center font-bold font-hand border-2 shadow-[2px_2px_0px_0px_rgba(30,36,43,0.3)]',
          sizeClasses[size],
          activeTheme.bg,
          activeTheme.text,
          activeTheme.border,
          className
        )}
      >
        <span>{getInitials(username)}</span>
      </div>

      {isOnline !== undefined && (
        <span
          className={cn(
            'absolute bottom-0 right-0 rounded-full border-2 border-white',
            size === 'sm' ? 'w-2.5 h-2.5' : 'w-3.5 h-3.5',
            isOnline ? 'bg-[#15803D]' : 'bg-[#94A3B8]'
          )}
        />
      )}
    </div>
  )
}
