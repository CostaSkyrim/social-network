'use client'

import Link from 'next/link'
import { Avatar } from '@/components/ui/Avatar'
import { Badge } from '@/components/ui/Badge'
import { useAcceptGroupMember, useRejectGroupMember } from '@/hooks/useGroups'
import type { GroupMember } from '@/types/group'

interface MemberListProps {
  members: GroupMember[]
  groupId: string
  isCreator: boolean
}

const status_meta: Record<string, { label: string; variant: 'default' | 'success' | 'warning' | 'danger' }> = {
  accepted: { label: 'Member', variant: 'success' },
  pending: { label: 'Pending', variant: 'warning' },
  invited: { label: 'Invited', variant: 'warning' },
  declined: { label: 'Declined', variant: 'danger' },
}

export function MemberList({ members, groupId, isCreator }: MemberListProps) {
  const accepted = members.filter((m) => m.status === 'accepted')
  const pending = members.filter(
    (m) => m.status === 'pending' || m.status === 'invited',
  )

  function render_member(m: GroupMember) {
    const name = `${m.user.first_name} ${m.user.last_name}`
    const meta = status_meta[m.status] ?? status_meta.accepted
    const show_actions =
      isCreator && (m.status === 'pending' || m.status === 'invited')

    return (
      <li key={m.user.id} className="flex items-center gap-3 py-2">
        <Link href={`/profile/${m.user.id}`}>
          <Avatar src={m.user.avatar_path} alt={name} size="sm" />
        </Link>
        <div className="min-w-0 flex-1">
          <Link
            href={`/profile/${m.user.id}`}
            className="text-sm font-medium text-gray-900 hover:underline"
          >
            {name}
          </Link>
          {m.user.nickname && (
            <p className="text-xs text-gray-500">@{m.user.nickname}</p>
          )}
        </div>
        {show_actions ? (
          <div className="flex items-center gap-1.5">
            <AcceptDeclineButtons member={m} groupId={groupId} />
          </div>
        ) : m.status !== 'accepted' ? (
          <Badge variant={meta.variant}>{meta.label}</Badge>
        ) : null}
      </li>
    )
  }

  return (
    <div>
      {accepted.length > 0 && (
        <div>
          <h4 className="mb-1 text-xs font-semibold uppercase tracking-wide text-gray-400">
            Members ({accepted.length})
          </h4>
          <ul>{accepted.map(render_member)}</ul>
        </div>
      )}

      {pending.length > 0 && (
        <div className="mt-3 border-t border-gray-100 pt-3">
          <h4 className="mb-1 text-xs font-semibold uppercase tracking-wide text-gray-400">
            Requests & Invitations ({pending.length})
          </h4>
          <ul>{pending.map(render_member)}</ul>
        </div>
      )}
    </div>
  )
}

function AcceptDeclineButtons({ member, groupId }: { member: GroupMember; groupId: string }) {
  const accept_mutation = useAcceptGroupMember(groupId)
  const reject_mutation = useRejectGroupMember(groupId)
  const is_pending = accept_mutation.isPending || reject_mutation.isPending

  return (
    <>
      <button
        onClick={() => accept_mutation.mutate(member.user.id)}
        disabled={is_pending}
        className="rounded-lg bg-green-600 px-2.5 py-1 text-xs font-medium text-white hover:bg-green-700 disabled:opacity-50"
      >
        Accept
      </button>
      <button
        onClick={() => reject_mutation.mutate(member.user.id)}
        disabled={is_pending}
        className="rounded-lg bg-red-600 px-2.5 py-1 text-xs font-medium text-white hover:bg-red-700 disabled:opacity-50"
      >
        Decline
      </button>
    </>
  )
}
