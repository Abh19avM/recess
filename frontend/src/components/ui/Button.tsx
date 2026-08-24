import React from 'react'
import { cn } from '../../lib/utils'

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'chalk' | 'danger' | 'ghost' | 'stamp'
  size?: 'sm' | 'md' | 'lg'
  isLoading?: boolean
  leftIcon?: React.ReactNode
  rightIcon?: React.ReactNode
}

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      className,
      variant = 'primary',
      size = 'md',
      isLoading = false,
      leftIcon,
      rightIcon,
      children,
      disabled,
      ...props
    },
    ref
  ) => {
    const baseStyles =
      'inline-flex items-center justify-center font-medium transition-all duration-150 active:translate-y-[1px] disabled:opacity-50 disabled:pointer-events-none disabled:cursor-not-allowed select-none'

    const sizeStyles = {
      sm: 'text-xs px-3 py-1.5 rounded gap-1.5 font-semibold',
      md: 'text-sm px-4 py-2 rounded-md gap-2 font-semibold',
      lg: 'text-base px-6 py-3 rounded-md gap-2.5 font-bold',
    }

    const variantStyles = {
      primary:
        'bg-[#1A365D] dark:bg-[#2563EB] text-white border-2 border-[#0F2238] dark:border-[#1D4ED8] shadow-[2px_2px_0px_0px_#0F2238] dark:shadow-[2px_2px_0px_0px_#020617] hover:shadow-[3px_3px_0px_0px_#0F2238] dark:hover:shadow-[3px_3px_0px_0px_#020617] hover:-translate-x-[1px] hover:-translate-y-[1px] active:shadow-none active:translate-x-[1px] active:translate-y-[1px]',
      secondary:
        'bg-[#FFFFFF] dark:bg-[#1E293B] text-[#1E242B] dark:text-[#F8FAFC] border-2 border-[#475569] dark:border-[#334155] shadow-[2px_2px_0px_0px_#475569] dark:shadow-[2px_2px_0px_0px_#020617] hover:bg-[#FBF9F3] dark:hover:bg-[#27354A] hover:shadow-[3px_3px_0px_0px_#475569] hover:-translate-x-[1px] hover:-translate-y-[1px] active:shadow-none active:translate-x-[1px] active:translate-y-[1px]',
      chalk:
        'bg-[#18231C] dark:bg-[#0F172A] text-[#F8FAFC] border-2 border-[#5c4033] dark:border-[#475569] shadow-[2px_2px_0px_0px_#5c4033] dark:shadow-[2px_2px_0px_0px_#020617] hover:text-[#FEF08A] hover:border-[#FEF08A]/50 active:translate-x-[1px] active:translate-y-[1px]',
      danger:
        'bg-[#991B1B] dark:bg-[#DC2626] text-white border-2 border-[#7F1D1D] dark:border-[#B91C1C] shadow-[2px_2px_0px_0px_#7F1D1D] dark:shadow-[2px_2px_0px_0px_#020617] hover:bg-[#B91C1C] hover:shadow-[3px_3px_0px_0px_#7F1D1D] active:shadow-none',
      ghost:
        'bg-transparent text-[#475569] dark:text-[#94A3B8] hover:bg-[#F2EDE0] dark:hover:bg-[#1E293B] hover:text-[#1E242B] dark:hover:text-[#F8FAFC] border border-transparent',
      stamp:
        'font-hand uppercase tracking-wider text-[#15803D] dark:text-[#86EFAC] bg-[#15803D]/10 dark:bg-[#22C55E]/20 border-2 border-[#15803D] dark:border-[#22C55E] rounded hover:bg-[#15803D]/20 rotate-[-1deg] hover:rotate-[0deg]',
    }

    return (
      <button
        ref={ref}
        className={cn(baseStyles, sizeStyles[size], variantStyles[variant], className)}
        disabled={disabled || isLoading}
        {...props}
      >
        {isLoading ? (
          <svg
            className="animate-spin h-4 w-4 text-current"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            ></circle>
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
        ) : (
          leftIcon
        )}
        <span>{children}</span>
        {!isLoading && rightIcon}
      </button>
    )
  }
)

Button.displayName = 'Button'
