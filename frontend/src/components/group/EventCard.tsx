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
            <h4 className="text-sm font-semibold text-gray-900">{event.title}</h4>
            <p className="mt-0.5 text-xs text-gray-500">
              {format_datetime(event.event_datetime)} · by {creator_name}
            </p>
            {format_starts_in(event.event_datetime) && (
              <span className="mt-1 inline-block rounded bg-blue-50 px-1.5 py-0.5 text-[10px] font-medium text-blue-700">
                {format_starts_in(event.event_datetime)}
              </span>
            )}
          </div>
          <Badge variant="warning">{event.total} going</Badge>
        </div>

        {event.description && (
          <p className="text-sm text-gray-700 whitespace-pre-wrap">{event.description}</p>
        )}

        {event.image_path && (
          /* eslint-disable-next-line @next/next/no-img-element */
          <img
            src={get_media_url(event.image_path)}
            alt="Event image"
            className="max-h-64 w-full rounded-lg object-cover"
          />
        )}

        <div className="flex items-center gap-2">
          <button
            onClick={() => handle_rsvp('going')}
            disabled={is_pending}
            className={cn(
              'rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors disabled:opacity-50',
              event.my_response === 'going'
                ? 'border-green-500 bg-green-50 text-green-700'
                : 'border-gray-300 text-gray-700 hover:bg-green-50',
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
                ? 'border-red-500 bg-red-50 text-red-700'
                : 'border-gray-300 text-gray-700 hover:bg-red-50',
            )}
          >
            ✗ Not going ({event.not_going})
          </button>
        </div>
      </CardContent>
    </Card>
  )
}
