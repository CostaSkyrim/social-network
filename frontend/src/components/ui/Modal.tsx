import { useEffect, type ReactNode } from 'react'
import { cn } from '@/lib/cn'

interface ModalProps {
  open: boolean
  on_close: () => void
  title?: string
  children: ReactNode
  className?: string
}

export function Modal({ open, on_close, title, children, className }: ModalProps) {
  useEffect(() => {
    if (open) {
      document.body.style.overflow = 'hidden'
    } else {
      document.body.style.overflow = ''
    }
    return () => {
      document.body.style.overflow = ''
    }
  }, [open])

  useEffect(() => {
    function on_key(e: KeyboardEvent) {
      if (e.key === 'Escape') on_close()
    }
    if (open) window.addEventListener('keydown', on_key)
    return () => window.removeEventListener('keydown', on_key)
  }, [open, on_close])

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div className="fixed inset-0 bg-black/50" onClick={on_close} />
      <div
        className={cn(
          'relative z-10 w-full max-w-lg rounded-xl bg-[#241748] p-6 shadow-xl',
          className,
        )}
      >
        {title && (
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-lg font-semibold">{title}</h2>
            <button onClick={on_close} className="text-gray-300 hover:text-gray-300">
              ✕
            </button>
          </div>
        )}
        {children}
      </div>
    </div>
  )
}
