import React from 'react'
import { cn } from '../../lib/utils'

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: 'default' | 'ink-blue' | 'pencil' | 'chalk' | 'green' | 'amber' | 'red'
  size?: 'sm' | 'md'
}

export const Badge: React.FC<BadgeProps> = ({
  className,
  variant = 'default',
  size = 'md',
  children,
  ...props
}) => {
  const sizeStyles = {
    sm: 'text-[11px] px-2 py-0.5 font-medium tracking-wide',
    md: 'text-xs px-2.5 py-1 font-semibold tracking-wider',
  }

  const variantStyles = {
    default: 'bg-[#F2EDE0] text-[#1E242B] border border-[#CBD5E1]',
    'ink-blue': 'bg-[#1A365D]/10 text-[#1A365D] border border-[#1A365D]/30',
    pencil: 'bg-white text-[#475569] border border-[#94A3B8]',
    chalk: 'bg-[#18231C] text-[#FEF08A] border border-[#FEF08A]/40 font-mono',
    green: 'bg-[#15803D]/10 text-[#15803D] border border-[#15803D]/30',
    amber: 'bg-[#B45309]/10 text-[#B45309] border border-[#B45309]/30',
    red: 'bg-[#991B1B]/10 text-[#991B1B] border border-[#991B1B]/30',
  }

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded uppercase select-none font-mono',
        sizeStyles[size],
        variantStyles[variant],
        className
      )}
      {...props}
    >
      {children}
    </span>
  )
}

export interface StampProps extends React.HTMLAttributes<HTMLSpanElement> {
  tone?: 'green' | 'red' | 'blue' | 'amber'
  children: React.ReactNode
}

export const Stamp: React.FC<StampProps> = ({
  tone = 'green',
  className,
  children,
  ...props
}) => {
  const toneClasses = {
    green: 'rubber-stamp-green',
    red: 'rubber-stamp-red',
    blue: 'rubber-stamp-blue',
    amber: 'rubber-stamp-amber',
  }

  return (
    <span
      className={cn(
        'rubber-stamp px-3 py-0.5 text-xs font-bold font-hand',
        toneClasses[tone],
        className
      )}
      {...props}
    >
      {children}
    </span>
  )
}
