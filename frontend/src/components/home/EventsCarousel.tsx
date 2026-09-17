'use client'

import { useRef, useEffect, useState } from 'react'
import Link from 'next/link'
import { Card, CardContent } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Spinner } from '@/components/ui/Spinner'
import { EmptyState } from '@/components/common/EmptyState'
import { format_datetime, format_starts_in } from '@/lib/format'
import { get_media_url } from '@/lib/media'
import type { GroupEvent } from '@/types/group'

interface EventsCarouselProps {
  events: GroupEvent[]
  is_loading: boolean
  has_next: boolean | undefined
  on_load_more: () => void
}

export function EventsCarousel({ events, is_loading, has_next, on_load_more }: EventsCarouselProps) {
  const scroll_ref = useRef<HTMLDivElement>(null)
  const [can_scroll_left, set_can_scroll_left] = useState(false)
  const [can_scroll_right, set_can_scroll_right] = useState(false)

  const update_scroll_state = () => {
    const el = scroll_ref.current
    if (!el) return
    set_can_scroll_left(el.scrollLeft > 4)
    set_can_scroll_right(el.scrollLeft + el.clientWidth < el.scrollWidth - 4)
  }

  useEffect(() => {
    update_scroll_state()
    const el = scroll_ref.current
    if (!el) return
    el.addEventListener('scroll', update_scroll_state, { passive: true })
    window.addEventListener('resize', update_scroll_state)
    return () => {
      el.removeEventListener('scroll', update_scroll_state)
      window.removeEventListener('resize', update_scroll_state)
    }
  }, [events.length])

  const scroll_by = (dir: 1 | -1) => {
    const el = scroll_ref.current
    if (!el) return
    el.scrollBy({ left: dir * el.clientWidth * 0.8, behavior: 'smooth' })
  }

  if (is_loading && events.length === 0) {
    return (
      <div className="flex justify-center py-8">
        <Spinner />
      </div>
    )
  }

  if (events.length === 0) {
    return (
      <EmptyState
        title="No upcoming events"
        description="Events from your groups will appear here."
      />
    )
  }

  return (
    <div className="relative">
      {can_scroll_left && (
        <button
          type="button"
          onClick={() => scroll_by(-1)}
          aria-label="Scroll to earlier events"
          className="absolute left-0 top-1/2 z-10 flex h-9 w-9 -translate-y-1/2 items-center justify-center rounded-full border border-purple-400/20 bg-[#241748] text-gray-300 shadow-sm hover:bg-purple-400/10"
        >
          ‹
        </button>
      )}
      {can_scroll_right && (
        <button
          type="button"
          onClick={() => scroll_by(1)}
          aria-label="Scroll to later events"
          className="absolute right-0 top-1/2 z-10 flex h-9 w-9 -translate-y-1/2 items-center justify-center rounded-full border border-purple-400/20 bg-[#241748] text-gray-300 shadow-sm hover:bg-purple-400/10"
        >
          ›
        </button>
      )}

      <div
        ref={scroll_ref}
        className="flex snap-x snap-mandatory gap-4 overflow-x-auto scroll-smooth pb-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden"
      >
        {events.map((event) => (
          <EventSlide key={event.id} event={event} />
        ))}
        {has_next && (
          <button
            type="button"
            onClick={on_load_more}
            disabled={is_loading}
            className="flex w-40 shrink-0 snap-start items-center justify-center rounded-xl border border-dashed border-purple-400/30 bg-[#241748] text-sm text-gray-300 hover:bg-purple-400/10"
          >
            {is_loading ? <Spinner size="sm" /> : 'Load older events'}
          </button>
        )}
      </div>
    </div>
  )
}

function EventSlide({ event }: { event: GroupEvent }) {
  const starts_in = format_starts_in(event.event_datetime)

  return (
    <Link
      href={`/groups/${event.group_uuid ?? event.group_id}`}
      className="w-64 shrink-0 snap-start"
    >
      <Card className="h-full transition-shadow hover:shadow-md">
        {event.image_path && (
          /* eslint-disable-next-line @next/next/no-img-element */
          <img
            src={get_media_url(event.image_path)}
            alt={event.title}
            className="h-28 w-full rounded-t-xl object-cover"
          />
        )}
        <CardContent className="space-y-2">
          <div className="flex items-start justify-between gap-2">
            <h4 className="line-clamp-2 text-sm font-semibold text-gray-100">{event.title}</h4>
          </div>
          {event.group_title && (
            <p className="text-xs font-semibold text-fuchsia-300">{event.group_title}</p>
          )}
          <p className="text-xs text-gray-300">{format_datetime(event.event_datetime)}</p>
          {starts_in && (
            <Badge variant="success" className="text-[10px]">
              {starts_in}
            </Badge>
          )}
          <p className="text-xs text-gray-300">{event.going} going</p>
        </CardContent>
      </Card>
    </Link>
  )
}
