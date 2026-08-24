import React from 'react'
import { cn } from '../../lib/utils'

export interface PaperCardProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: 'ruled' | 'graph' | 'sticky' | 'chalkboard' | 'plain' | 'cork'
  stickyColor?: 'yellow' | 'blue' | 'pink' | 'green'
  showPushPin?: boolean
  showTape?: boolean
  tilt?: boolean
}

export const PaperCard = React.forwardRef<HTMLDivElement, PaperCardProps>(
  (
    {
      className,
      variant = 'plain',
      stickyColor = 'yellow',
      showPushPin = false,
      showTape = false,
      tilt = false,
      children,
      ...props
    },
    ref
  ) => {
    const stickyColors = {
      yellow: 'bg-[#FEF9C3] dark:bg-[#2D2106] border-[#FDE047] dark:border-[#854D0E] text-[#1E242B] dark:text-[#FEF08A]',
      blue: 'bg-[#E0F2FE] dark:bg-[#0B2545] border-[#BAE6FD] dark:border-[#1E40AF] text-[#1E242B] dark:text-[#BAE6FD]',
      pink: 'bg-[#FFE4E6] dark:bg-[#3B1123] border-[#FECDD3] dark:border-[#9D174D] text-[#1E242B] dark:text-[#FBCFE8]',
      green: 'bg-[#DCFCE7] dark:bg-[#063323] border-[#BBF7D0] dark:border-[#065F46] text-[#1E242B] dark:text-[#A7F3D0]',
    }

    return (
      <div
        ref={ref}
        className={cn(
          'relative transition-all duration-200',
          variant === 'plain' &&
            'bg-white dark:bg-[#151D2E] rounded-lg p-5 border-2 border-[#475569] dark:border-[#334155] shadow-[3px_3px_0px_0px_#475569] dark:shadow-[3px_3px_0px_0px_#020617] text-[#1E242B] dark:text-[#F8FAFC]',
          variant === 'ruled' &&
            'paper-ruled rounded-lg p-6 border-2 border-[#CBD5E1] dark:border-[#334155] shadow-[3px_3px_0px_0px_rgba(71,85,105,0.2)] dark:shadow-[3px_3px_0px_0px_#020617] pl-16 text-[#1E242B] dark:text-[#F8FAFC]',
          variant === 'graph' &&
            'paper-graph rounded-lg p-5 border-2 border-[#CBD5E1] dark:border-[#334155] shadow-[3px_3px_0px_0px_#CBD5E1] dark:shadow-[3px_3px_0px_0px_#020617] text-[#1E242B] dark:text-[#F8FAFC]',
          variant === 'sticky' &&
            cn(
              'p-5 rounded-sm border shadow-[3px_3px_6px_rgba(0,0,0,0.08)] dark:shadow-[3px_3px_6px_rgba(0,0,0,0.4)]',
              stickyColors[stickyColor],
              tilt && 'rotate-[-1deg] hover:rotate-0'
            ),
          variant === 'chalkboard' &&
            'paper-graph-dark rounded-lg p-6 border-4 border-[#5c4033] dark:border-[#3E2723] shadow-[inset_0_2px_8px_rgba(0,0,0,0.6)] text-[#F8FAFC]',
          variant === 'cork' &&
            'paper-cork rounded-lg p-6 border-4 border-[#8B5A2B] dark:border-[#4A2E18] shadow-[inset_0_2px_6px_rgba(0,0,0,0.4)] text-[#1E242B] dark:text-[#F8FAFC]',
          className
        )}
        {...props}
      >
        {showPushPin && (
          <div className="absolute -top-2 left-1/2 -translate-x-1/2 z-10">
            <div className="push-pin-red" />
          </div>
        )}

        {showTape && (
          <div className="absolute -top-3 left-1/2 -translate-x-1/2 w-16 h-5 masking-tape rotate-[-2deg] rounded-xs border border-yellow-400/40 z-10" />
        )}

        {children}
      </div>
    )
  }
)

PaperCard.displayName = 'PaperCard'
