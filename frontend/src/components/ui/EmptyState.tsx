import React from 'react'
import { Button } from './Button'
import { PaperCard } from './PaperCard'
import { cn } from '../../lib/utils'

export interface EmptyStateProps {
  title: string
  description: string
  actionLabel?: string
  onAction?: () => void
  icon?: React.ReactNode
  className?: string
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  title,
  description,
  actionLabel,
  onAction,
  icon,
  className,
}) => {
  return (
    <PaperCard
      variant="ruled"
      className={cn('text-center py-10 px-6 max-w-md mx-auto my-6', className)}
    >
      <div className="flex flex-col items-center">
        <div className="w-16 h-16 rounded-full bg-[#FEF9C3] border-2 border-[#EAB308] flex items-center justify-center text-2xl mb-4 shadow-[2px_2px_0px_0px_#475569]">
          {icon || <span className="font-hand font-bold text-[#A16207]">✎</span>}
        </div>

        <h3 className="font-bold text-lg text-[#1E242B] font-sans">{title}</h3>
        <p className="font-hand text-base text-[#475569] mt-1 max-w-xs">{description}</p>

        {actionLabel && onAction && (
          <div className="mt-5">
            <Button onClick={onAction} size="sm" variant="primary">
              {actionLabel}
            </Button>
          </div>
        )}
      </div>
    </PaperCard>
  )
}
