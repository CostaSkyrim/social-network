'use client'

import { useJoinGroup, useLeaveGroup } from '@/hooks/useGroups'
import type { GroupMember } from '@/types/group'

interface GroupActionsProps {
  groupId: string
  creatorId: string
  members: GroupMember[]
  currentUserId: string
}

export function GroupActions({
  groupId,
  creatorId,
  members,
  currentUserId,
}: GroupActionsProps) {
  const join_mutation = useJoinGroup(groupId)
  const leave_mutation = useLeaveGroup(groupId)

  const is_creator = creatorId === currentUserId
  const my_membership = members.find((m) => m.user.id === currentUserId)
  const status = my_membership?.status

  // Creator: no join/leave buttons
  if (is_creator) {
    return (
      <div className="rounded-lg bg-gray-50 px-3 py-2 text-xs text-gray-500">
        You're the creator of this group.
      </div>
    )
  }

  // Not a member (or left/declined previously)
  if (!status || status === 'declined') {
    return (
      <button
        onClick={() => join_mutation.mutate()}
        disabled={join_mutation.isPending}
        className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
      >
        {join_mutation.isPending ? 'Joining…' : 'Join group'}
      </button>
    )
  }

  // Pending / invited
  if (status !== 'accepted') {
    return (
      <div className="rounded-lg bg-gray-50 px-3 py-2 text-xs text-gray-500">
        {status === 'pending' ? 'Request sent — awaiting approval' : 'Invited — awaiting response'}
      </div>
    )
  }

  // Accepted member
  return (
    <button
      onClick={() => leave_mutation.mutate()}
      disabled={leave_mutation.isPending}
      className="rounded-lg bg-red-50 px-4 py-2 text-sm font-medium text-red-600 hover:bg-red-100 disabled:opacity-50"
    >
      {leave_mutation.isPending ? 'Leaving…' : 'Leave group'}
    </button>
  )
}
