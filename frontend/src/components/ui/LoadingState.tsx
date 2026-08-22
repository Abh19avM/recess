import React from 'react'
import { cn } from '../../lib/utils'

export interface LoadingStateProps {
  message?: string
  fullPage?: boolean
  className?: string
}

export const LoadingState: React.FC<LoadingStateProps> = ({
  message = 'Sharpening pencils & loading...',
  fullPage = false,
  className,
}) => {
  const content = (
    <div className={cn('flex flex-col items-center justify-center p-8 text-center', className)}>
      {/* Animated Desk Bell / Pencil Icon */}
      <div className="relative w-12 h-12 mb-4">
        <div className="w-12 h-12 rounded-full border-3 border-[#1A365D] border-t-transparent animate-spin" />
        <div className="absolute inset-0 flex items-center justify-center font-hand text-lg text-[#1A365D] font-bold">
          ✎
        </div>
      </div>

      <p className="font-hand text-xl font-bold text-[#1A365D] tracking-wide animate-pulse">
        {message}
      </p>
    </div>
  )

  if (fullPage) {
    return (
      <div className="min-h-[60vh] flex items-center justify-center">
        {content}
      </div>
    )
  }

  return content
}

export const SkeletonLines: React.FC<{ count?: number }> = ({ count = 3 }) => {
  return (
    <div className="space-y-3 w-full animate-pulse py-2">
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="h-4 bg-[#E2E8F0] rounded-sm"
          style={{ width: `${85 - (i % 3) * 15}%` }}
        />
      ))}
    </div>
  )
}
