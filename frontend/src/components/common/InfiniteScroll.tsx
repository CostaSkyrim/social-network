import { useRef, useEffect, type ReactNode } from 'react'
import { Spinner } from '@/components/ui/Spinner'

interface InfiniteScrollProps {
  has_next: boolean
  is_loading: boolean
  on_load_more: () => void
  children: ReactNode
}

export function InfiniteScroll({ has_next, is_loading, on_load_more, children }: InfiniteScrollProps) {
  const sentinel = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = sentinel.current
    if (!el) return

    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0].isIntersecting && has_next && !is_loading) {
          on_load_more()
        }
      },
      { threshold: 0.1 },
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [has_next, is_loading, on_load_more])

  return (
    <div>
      {children}
      <div ref={sentinel} className="flex justify-center py-4">
        {is_loading && <Spinner />}
      </div>
    </div>
  )
}
