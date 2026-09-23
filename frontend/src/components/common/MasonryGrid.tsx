'use client'

import { useEffect, useRef, useState, type ReactNode } from 'react'
import { cn } from '@/lib/cn'

/**
 * Row-major masonry grid.
 *
 * CSS columns pack top-to-bottom per column (which scrambles chronological
 * order), and a normal grid aligns every row to its tallest card (which leaves
 * big empty gaps). Instead we use a very small implicit row height and let each
 * item span just enough rows for its content, measured with a ResizeObserver.
 * Order stays left-to-right / top-to-bottom, like a normal grid.
 */

// Must match the container's `gap-4` (1rem).
const ROW_GAP = 16
// Small implicit row height keeps the packing tight (less rounding slack).
const ROW_HEIGHT = 1

interface MasonryGridProps {
  children: ReactNode
  className?: string
}

export function MasonryGrid({ children, className }: MasonryGridProps) {
  return (
    <div
      className={cn('grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3', className)}
      style={{ gridAutoRows: `${ROW_HEIGHT}px` }}
    >
      {children}
    </div>
  )
}

interface MasonryItemProps {
  children: ReactNode
}

export function MasonryItem({ children }: MasonryItemProps) {
  const ref = useRef<HTMLDivElement>(null)
  const [span, set_span] = useState(1)

  useEffect(() => {
    const el = ref.current
    if (!el) return

    const update = () => {
      const height = el.getBoundingClientRect().height
      const next = Math.max(
        1,
        Math.ceil((height + ROW_GAP) / (ROW_HEIGHT + ROW_GAP)),
      )
      set_span((prev) => (prev === next ? prev : next))
    }

    update()
    const observer = new ResizeObserver(update)
    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  return (
    <div style={{ gridRowEnd: `span ${span}` }}>
      <div ref={ref}>{children}</div>
    </div>
  )
}
