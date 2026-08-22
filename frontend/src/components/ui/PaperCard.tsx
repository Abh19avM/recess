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
      yellow: 'bg-[#FEF9C3] border-[#FDE047] text-[#1E242B]',
      blue: 'bg-[#E0F2FE] border-[#BAE6FD] text-[#1E242B]',
      pink: 'bg-[#FFE4E6] border-[#FECDD3] text-[#1E242B]',
      green: 'bg-[#DCFCE7] border-[#BBF7D0] text-[#1E242B]',
    }

    return (
      <div
        ref={ref}
        className={cn(
          'relative transition-all duration-200',
          variant === 'plain' &&
            'bg-white rounded-lg p-5 border-2 border-[#475569] shadow-[3px_3px_0px_0px_#475569]',
          variant === 'ruled' &&
            'paper-ruled rounded-lg p-6 border-2 border-[#CBD5E1] shadow-[3px_3px_0px_0px_rgba(71,85,105,0.2)] pl-16',
          variant === 'graph' &&
            'paper-graph rounded-lg p-5 border-2 border-[#CBD5E1] shadow-[3px_3px_0px_0px_#CBD5E1]',
          variant === 'sticky' &&
            cn(
              'p-5 rounded-sm border shadow-[3px_3px_6px_rgba(0,0,0,0.08)]',
              stickyColors[stickyColor],
              tilt && 'rotate-[-1deg] hover:rotate-0'
            ),
          variant === 'chalkboard' &&
            'paper-graph-dark rounded-lg p-6 border-4 border-[#5c4033] shadow-[inset_0_2px_8px_rgba(0,0,0,0.6)] text-[#F8FAFC]',
          variant === 'cork' &&
            'paper-cork rounded-lg p-6 border-4 border-[#8B5A2B] shadow-[inset_0_2px_6px_rgba(0,0,0,0.4)]',
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
