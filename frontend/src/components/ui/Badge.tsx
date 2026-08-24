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
    default: 'bg-[#F2EDE0] dark:bg-[#1E293B] text-[#1E242B] dark:text-[#F8FAFC] border border-[#CBD5E1] dark:border-[#334155]',
    'ink-blue': 'bg-[#1A365D]/10 dark:bg-[#3B82F6]/20 text-[#1A365D] dark:text-[#93C5FD] border border-[#1A365D]/30 dark:border-[#3B82F6]/40',
    pencil: 'bg-white dark:bg-[#1E293B] text-[#475569] dark:text-[#CBD5E1] border border-[#94A3B8] dark:border-[#475569]',
    chalk: 'bg-[#18231C] dark:bg-[#0F172A] text-[#FEF08A] border border-[#FEF08A]/40 font-mono',
    green: 'bg-[#15803D]/10 dark:bg-[#22C55E]/20 text-[#15803D] dark:text-[#86EFAC] border border-[#15803D]/30 dark:border-[#22C55E]/40',
    amber: 'bg-[#B45309]/10 dark:bg-[#F59E0B]/20 text-[#B45309] dark:text-[#FDE047] border border-[#B45309]/30 dark:border-[#F59E0B]/40',
    red: 'bg-[#991B1B]/10 dark:bg-[#EF4444]/20 text-[#991B1B] dark:text-[#FCA5A5] border border-[#991B1B]/30 dark:border-[#EF4444]/40',
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
