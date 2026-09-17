'use client'

import { useRSVP } from '@/hooks/useGroups'
import { Card, CardContent } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { format_datetime, format_starts_in } from '@/lib/format'
import { get_media_url } from '@/lib/media'
import { cn } from '@/lib/cn'
import type { GroupEvent } from '@/types/group'

interface EventCardProps {
  event: GroupEvent
  groupId: string
}

export function EventCard({ event, groupId }: EventCardProps) {
  const rsvp = useRSVP(event.id, groupId)
  const is_pending = rsvp.isPending

  function handle_rsvp(response: 'going' | 'not_going') {
    if (event.my_response === response) return
    rsvp.mutate(response)
  }

  const creator_name = event.creator
    ? `${event.creator.first_name} ${event.creator.last_name}`
    : 'Unknown'

  return (
    <Card>
      <CardContent className="space-y-3">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <h4 className="text-sm font-semibold text-gray-100">{event.title}</h4>
            <p className="mt-0.5 text-xs text-gray-300">
              {format_datetime(event.event_datetime)} · by {creator_name}
            </p>
            {format_starts_in(event.event_datetime) && (
              <span className="mt-1 inline-block rounded bg-violet-400/20 px-1.5 py-0.5 text-[10px] font-medium text-violet-200">
                {format_starts_in(event.event_datetime)}
              </span>
            )}
          </div>
          <Badge variant="warning">{event.total} going</Badge>
        </div>

        {event.description && (
          <p className="text-sm text-gray-300 whitespace-pre-wrap">{event.description}</p>
        )}

        {event.image_path && (
          /* eslint-disable-next-line @next/next/no-img-element */
          <img
            src={get_media_url(event.image_path)}
            alt="Event image"
            className="max-h-64 w-full rounded-lg object-contain"
          />
        )}

        <div className="flex items-center gap-2">
          <button
            onClick={() => handle_rsvp('going')}
            disabled={is_pending}
            className={cn(
              'rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors disabled:opacity-50',
              event.my_response === 'going'
                ? 'border-green-500 bg-green-500/15 text-green-200'
                : 'border-purple-400/30 text-gray-300 hover:bg-green-500/15',
            )}
          >
            ✓ Going ({event.going})
          </button>
          <button
            onClick={() => handle_rsvp('not_going')}
            disabled={is_pending}
            className={cn(
              'rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors disabled:opacity-50',
              event.my_response === 'not_going'
                ? 'border-red-500 bg-red-500/15 text-red-200'
                : 'border-purple-400/30 text-gray-300 hover:bg-red-500/15',
            )}
          >
            ✗ Not going ({event.not_going})
          </button>
        </div>
      </CardContent>
    </Card>
  )
}
