import React from 'react'
import { cn } from '../../lib/utils'

export interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string
  handwrittenLabel?: boolean
  error?: string
  hint?: string
  leftIcon?: React.ReactNode
  rightIcon?: React.ReactNode
  variant?: 'sketch' | 'notebook-line'
}

export const Input = React.forwardRef<HTMLInputElement, InputProps>(
  (
    {
      className,
      label,
      handwrittenLabel = false,
      error,
      hint,
      leftIcon,
      rightIcon,
      variant = 'sketch',
      id,
      ...props
    },
    ref
  ) => {
    const inputId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined)

    return (
      <div className="w-full space-y-1.5 text-left">
        {label && (
          <label
            htmlFor={inputId}
            className={cn(
              'block text-sm font-semibold text-[#1E242B] dark:text-[#F8FAFC]',
              handwrittenLabel && 'font-hand text-base text-[#1A365D] dark:text-[#93C5FD]'
            )}
          >
            {label}
          </label>
        )}

        <div className="relative flex items-center">
          {leftIcon && (
            <div className="absolute left-3 flex items-center pointer-events-none text-[#475569] dark:text-[#94A3B8]">
              {leftIcon}
            </div>
          )}

          <input
            id={inputId}
            ref={ref}
            className={cn(
              'w-full text-sm text-[#1E242B] dark:text-[#F8FAFC] placeholder:text-[#94A3B8] dark:placeholder:text-[#64748B] transition-all outline-none bg-white dark:bg-[#101726]',
              variant === 'sketch' &&
                'px-3.5 py-2.5 rounded-md border-2 border-[#475569] dark:border-[#334155] shadow-[2px_2px_0px_0px_#475569] dark:shadow-[2px_2px_0px_0px_#020617] focus:border-[#1A365D] dark:focus:border-[#60A5FA] focus:shadow-[3px_3px_0px_0px_#1A365D] dark:focus:shadow-[3px_3px_0px_0px_#020617]',
              variant === 'notebook-line' &&
                'bg-transparent border-b-2 border-[#CBD5E1] dark:border-[#334155] rounded-none py-1.5 focus:border-[#1A365D] dark:focus:border-[#60A5FA]',
              leftIcon && 'pl-9',
              rightIcon && 'pr-9',
              error && 'border-[#991B1B] dark:border-[#EF4444] text-[#991B1B] dark:text-[#FCA5A5] focus:border-[#991B1B]',
              className
            )}
            {...props}
          />

          {rightIcon && (
            <div className="absolute right-3 flex items-center text-[#475569]">
              {rightIcon}
            </div>
          )}
        </div>

        {error && (
          <p className="font-hand text-sm font-bold text-[#991B1B] flex items-center gap-1 mt-1">
            <span>✎</span> {error}
          </p>
        )}

        {hint && !error && (
          <p className="text-xs text-[#475569]">{hint}</p>
        )}
      </div>
    )
  }
)

Input.displayName = 'Input'
