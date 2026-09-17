import type { ReactNode } from 'react'
import { cn } from '@/lib/cn'

interface BadgeProps {
  children: ReactNode
  variant?: 'default' | 'success' | 'warning' | 'danger'
  className?: string
}

const variant_styles = {
  default: 'bg-purple-400/15 text-gray-300',
  success: 'bg-green-500/20 text-green-200',
  warning: 'bg-pink-500/20 text-pink-200',
  danger: 'bg-red-500/20 text-red-200',
}

export function Badge({ children, variant = 'default', className }: BadgeProps) {
  return (
    <span
      className={cn(
        'inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium',
        variant_styles[variant],
        className,
      )}
    >
      {children}
    </span>
  )
}
