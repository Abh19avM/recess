import React, { useEffect } from 'react'
import { X } from 'lucide-react'
import { cn } from '../../lib/utils'

export interface ModalProps {
  isOpen: boolean
  onClose: () => void
  title?: string
  handwrittenTitle?: boolean
  children: React.ReactNode
  maxWidth?: 'sm' | 'md' | 'lg' | 'xl'
}

export const Modal: React.FC<ModalProps> = ({
  isOpen,
  onClose,
  title,
  handwrittenTitle = false,
  children,
  maxWidth = 'md',
}) => {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    if (isOpen) {
      document.body.style.overflow = 'hidden'
      window.addEventListener('keydown', handleKeyDown)
    }
    return () => {
      document.body.style.overflow = 'unset'
      window.removeEventListener('keydown', handleKeyDown)
    }
  }, [isOpen, onClose])

  if (!isOpen) return null

  const maxWClasses = {
    sm: 'max-w-sm',
    md: 'max-w-md',
    lg: 'max-w-lg',
    xl: 'max-w-2xl',
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      {/* Backdrop */}
      <div
        className="fixed inset-0 bg-[#1E242B]/60 backdrop-blur-xs transition-opacity"
        onClick={onClose}
      />

      {/* Modal Container Sheet */}
      <div
        className={cn(
          'relative w-full bg-white dark:bg-[#141C2E] rounded-lg border-2 border-[#1E242B] dark:border-[#334155] shadow-[6px_6px_0px_0px_#1E242B] dark:shadow-[6px_6px_0px_0px_#020617] p-6 z-10 animate-in fade-in zoom-in-95 duration-150 text-[#1E242B] dark:text-[#F8FAFC]',
          maxWClasses[maxWidth]
        )}
      >
        {/* Paper clip decoration at top left */}
        <div className="absolute -top-3 left-6 w-4 h-8 border-2 border-[#475569] rounded-full z-20 pointer-events-none" />

        {/* Header */}
        <div className="flex items-center justify-between pb-3 border-b border-[#CBD5E1] mb-4">
          {title && (
            <h3
              className={cn(
                'text-lg font-bold text-[#1E242B]',
                handwrittenTitle && 'font-hand text-2xl text-[#1A365D]'
              )}
            >
              {title}
            </h3>
          )}
          <button
            onClick={onClose}
            className="p-1 rounded text-[#475569] hover:bg-[#F2EDE0] hover:text-[#1E242B] transition-colors ml-auto"
            aria-label="Close modal"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Content */}
        <div>{children}</div>
      </div>
    </div>
  )
}
