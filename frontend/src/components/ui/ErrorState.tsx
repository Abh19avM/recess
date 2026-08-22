import React from 'react'
import { AlertCircle, RefreshCw } from 'lucide-react'
import { Button } from './Button'
import { PaperCard } from './PaperCard'
import { Stamp } from './Badge'
import { cn } from '../../lib/utils'

export interface ErrorStateProps {
  title?: string
  message: string
  onRetry?: () => void
  className?: string
}

export const ErrorState: React.FC<ErrorStateProps> = ({
  title = 'Classroom Disruption!',
  message,
  onRetry,
  className,
}) => {
  return (
    <PaperCard
      variant="ruled"
      className={cn('text-left py-8 px-6 max-w-md mx-auto my-6 border-[#991B1B]/40', className)}
    >
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-2">
          <AlertCircle className="w-6 h-6 text-[#991B1B]" />
          <h3 className="font-bold text-lg text-[#991B1B]">{title}</h3>
        </div>
        <Stamp tone="red">DETENTION</Stamp>
      </div>

      <div className="bg-[#FFE4E6]/50 p-3 rounded border border-[#FECDD3] text-sm text-[#881337] font-medium my-3">
        <p className="font-hand text-base">
          <span className="font-bold text-[#991B1B]">Teacher's Note:</span> {message}
        </p>
      </div>

      {onRetry && (
        <div className="mt-4 flex justify-end">
          <Button
            onClick={onRetry}
            size="sm"
            variant="secondary"
            leftIcon={<RefreshCw className="w-4 h-4" />}
          >
            Try Again
          </Button>
        </div>
      )}
    </PaperCard>
  )
}
