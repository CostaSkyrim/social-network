import { useState, useRef, useEffect, type ReactNode } from 'react'
import { cn } from '@/lib/cn'

interface DropdownProps {
  trigger: ReactNode
  children: ReactNode
  align?: 'left' | 'right'
  className?: string
}

export function Dropdown({ trigger, children, align = 'left', className }: DropdownProps) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function on_click(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false)
      }
    }
    document.addEventListener('mousedown', on_click)
    return () => document.removeEventListener('mousedown', on_click)
  }, [])

  return (
    <div ref={ref} className="relative inline-block">
      <div onClick={() => setOpen(!open)}>{trigger}</div>
      {open && (
        <div
          className={cn(
            'absolute z-40 mt-1 min-w-[180px] rounded-lg border border-purple-400/20 bg-[#241748] py-1 shadow-lg',
            align === 'right' ? 'right-0' : 'left-0',
            className,
          )}
          onClick={() => setOpen(false)}
        >
          {children}
        </div>
      )}
    </div>
  )
}

export function DropdownItem({
  onClick,
  children,
  danger,
}: {
  onClick: () => void
  children: ReactNode
  danger?: boolean
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        'flex w-full items-center px-4 py-2 text-left text-sm transition-colors',
        danger ? 'text-red-600 hover:bg-red-500/15' : 'text-gray-300 hover:bg-purple-400/15',
      )}
    >
      {children}
    </button>
  )
}
