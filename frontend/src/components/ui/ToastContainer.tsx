import React from 'react'
import { X, CheckCircle, AlertCircle, Info, AlertTriangle } from 'lucide-react'
import { useToastStore, type ToastItem } from '../../store/toastStore'
import { cn } from '../../lib/utils'

export const ToastContainer: React.FC = () => {
  const { toasts, removeToast } = useToastStore()

  if (toasts.length === 0) return null

  return (
    <div className="fixed bottom-5 right-5 z-50 flex flex-col gap-3 max-w-sm w-full pointer-events-none">
      {toasts.map((toast) => (
        <ToastCard key={toast.id} toast={toast} onDismiss={() => removeToast(toast.id)} />
      ))}
    </div>
  )
}

interface ToastCardProps {
  toast: ToastItem
  onDismiss: () => void
}

const ToastCard: React.FC<ToastCardProps> = ({ toast, onDismiss }) => {
  const icons = {
    success: <CheckCircle className="w-5 h-5 text-[#15803D] shrink-0" />,
    error: <AlertCircle className="w-5 h-5 text-[#991B1B] shrink-0" />,
    warning: <AlertTriangle className="w-5 h-5 text-[#B45309] shrink-0" />,
    info: <Info className="w-5 h-5 text-[#1A365D] shrink-0" />,
  }

  const bgStyles = {
    success: 'bg-[#DCFCE7] border-[#15803D] text-[#14532D]',
    error: 'bg-[#FFE4E6] border-[#991B1B] text-[#881337]',
    warning: 'bg-[#FEF9C3] border-[#B45309] text-[#78350F]',
    info: 'bg-[#E0F2FE] border-[#1A365D] text-[#0C4A6E]',
  }

  return (
    <div
      className={cn(
        'pointer-events-auto relative p-4 rounded-md border-2 shadow-[4px_4px_0px_0px_#1E242B] flex items-start gap-3 transition-all animate-in slide-in-from-bottom-5 duration-200',
        bgStyles[toast.type]
      )}
    >
      {/* Push-pin at top right */}
      <div className="absolute -top-1.5 right-6">
        <div className="push-pin-red" />
      </div>

      <div>{icons[toast.type]}</div>

      <div className="flex-1 pr-2">
        <h4 className="font-bold text-sm leading-tight">{toast.title}</h4>
        {toast.message && (
          <p className="text-xs mt-1 font-medium opacity-90">{toast.message}</p>
        )}
      </div>

      <button
        onClick={onDismiss}
        className="text-current opacity-70 hover:opacity-100 p-0.5"
        aria-label="Dismiss toast"
      >
        <X className="w-4 h-4" />
      </button>
    </div>
  )
}
