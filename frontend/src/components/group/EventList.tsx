'use client'

import { EventCard } from './EventCard'
import { Spinner } from '@/components/ui/Spinner'
import { EmptyState } from '@/components/common/EmptyState'
import type { GroupEvent } from '@/types/group'

interface EventListProps {
  events: GroupEvent[]
  groupId: string
  is_loading: boolean
}

export function EventList({ events, groupId, is_loading }: EventListProps) {
  if (is_loading) {
    return (
      <div className="flex justify-center py-8">
        <Spinner />
      </div>
    )
  }

  if (events.length === 0) {
    return (
      <EmptyState
        title="No events yet"
        description="Create the first event for this group!"
      />
    )
  }

  return (
    <div className="space-y-3">
      {events.map((event) => (
        <EventCard key={event.id} event={event} groupId={groupId} />
      ))}
    </div>
  )
}
