'use client'

import { useParams } from 'next/navigation'
import { useGroup, useGroupEvents } from '@/hooks/useGroups'
import { EventList } from '@/components/group/EventList'
import { EventForm } from '@/components/group/EventForm'
import { MemberList } from '@/components/group/MemberList'
import { Card, CardContent, CardHeader } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import { EmptyState } from '@/components/common/EmptyState'

export default function GroupDetailPage() {
  const params = useParams()
  const uuid = (params?.uuid as string) || ''

  const {
    data: group_data,
    isLoading: group_loading,
    isError: group_error,
  } = useGroup(uuid)

  const { data: events, isLoading: events_loading } = useGroupEvents(uuid)

  if (group_loading) {
    return (
      <div className="flex justify-center py-16">
        <Spinner size="lg" />
      </div>
    )
  }

  if (group_error || !group_data) {
    return (
      <EmptyState
        title="Group not found"
        description="This group may have been deleted, or you don't have access to it."
      />
    )
  }

  const { group, members } = group_data
  const accepted_count = members.filter((m) => m.status === 'accepted').length

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between gap-3">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">{group.title}</h2>
              {group.description && (
                <p className="mt-1 text-sm text-gray-500">{group.description}</p>
              )}
            </div>
            <span className="text-xs text-gray-400">{accepted_count} members</span>
          </div>
        </CardHeader>
      </Card>

      <Card>
        <CardHeader>
          <h3 className="text-sm font-semibold text-gray-900">Members</h3>
        </CardHeader>
        <CardContent>
          <MemberList members={members} />
        </CardContent>
      </Card>

      <div className="space-y-4">
        <h3 className="text-sm font-semibold text-gray-900">Events</h3>
        <EventForm groupId={group.id} />
        <EventList events={events ?? []} groupId={group.id} is_loading={events_loading} />
      </div>
    </div>
  )
}
