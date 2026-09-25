'use client'

import { useJoinGroup, useLeaveGroup, useAcceptGroupMember, useRejectGroupMember } from '@/hooks/useGroups'
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
  const accept_mutation = useAcceptGroupMember(groupId)
  const reject_mutation = useRejectGroupMember(groupId)
  const is_responding = accept_mutation.isPending || reject_mutation.isPending

  const is_creator = creatorId === currentUserId
  const my_membership = members.find((m) => m.user.id === currentUserId)
  const status = my_membership?.status

  // Creator: no join/leave buttons
  if (is_creator) {
    return (
      <div className="rounded-lg bg-purple-400/10 px-3 py-2 text-xs text-gray-300">
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
        className="rounded-lg bg-violet-600 px-4 py-2 text-sm font-medium text-white hover:bg-violet-700 disabled:opacity-50"
      >
        {join_mutation.isPending ? 'Joining…' : 'Join group'}
      </button>
    )
  }

  // Invitation: the invited user decides, not the creator.
  if (status === 'invited') {
    return (
      <div className="flex flex-wrap items-center gap-2 rounded-lg bg-purple-400/10 px-3 py-2 text-xs text-gray-300">
        <span>You've been invited to this group.</span>
        <button
          onClick={() => accept_mutation.mutate(currentUserId)}
          disabled={is_responding}
          className="rounded-lg bg-green-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-green-700 disabled:opacity-50"
        >
          {accept_mutation.isPending ? 'Accepting…' : 'Accept'}
        </button>
        <button
          onClick={() => reject_mutation.mutate(currentUserId)}
          disabled={is_responding}
          className="rounded-lg bg-red-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-red-700 disabled:opacity-50"
        >
          {reject_mutation.isPending ? 'Declining…' : 'Decline'}
        </button>
      </div>
    )
  }

  // Join request: awaiting the creator's approval.
  if (status === 'pending') {
    return (
      <div className="rounded-lg bg-purple-400/10 px-3 py-2 text-xs text-gray-300">
        Request sent — awaiting approval
      </div>
    )
  }

  // Accepted member
  return (
    <button
      onClick={() => leave_mutation.mutate()}
      disabled={leave_mutation.isPending}
      className="rounded-lg bg-red-500/15 px-4 py-2 text-sm font-medium text-red-600 hover:bg-red-500/20 disabled:opacity-50"
    >
      {leave_mutation.isPending ? 'Leaving…' : 'Leave group'}
    </button>
  )
}
