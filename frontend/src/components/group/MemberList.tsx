'use client'

import { useEffect, useState } from 'react'
import Link from 'next/link'
import { Avatar } from '@/components/ui/Avatar'
import { Badge } from '@/components/ui/Badge'
import { useAcceptGroupMember, useRejectGroupMember } from '@/hooks/useGroups'
import { useWebSocket } from '@/hooks/useWebSocket'
import type { GroupMember } from '@/types/group'

interface MemberListProps {
  members: GroupMember[]
  groupId: string
  isCreator: boolean
  creatorId: string
}

const status_meta: Record<string, { label: string; variant: 'default' | 'success' | 'warning' | 'danger' }> = {
  accepted: { label: 'Member', variant: 'success' },
  pending: { label: 'Pending', variant: 'warning' },
  invited: { label: 'Invited', variant: 'warning' },
  declined: { label: 'Declined', variant: 'danger' },
}

function sort_members(members: GroupMember[], creatorId: string): GroupMember[] {
  const by_name = (a: GroupMember, b: GroupMember) => {
    const name_a = `${a.user.first_name} ${a.user.last_name}`.toLowerCase()
    const name_b = `${b.user.first_name} ${b.user.last_name}`.toLowerCase()
    return name_a.localeCompare(name_b)
  }
  const creator = members.filter((m) => m.user.id === creatorId)
  const rest = members
    .filter((m) => m.user.id !== creatorId)
    .sort(by_name)
  return [...creator.sort(by_name), ...rest]
}

export function MemberList({ members, groupId, isCreator, creatorId }: MemberListProps) {
  const { subscribe } = useWebSocket()
  const [online_overrides, set_online_overrides] = useState<Record<string, boolean>>({})

  useEffect(() => {
    const unsub = subscribe('presence_update', (msg) => {
      const payload = msg.payload as
        | { user_uuid?: string; is_online: boolean }
        | undefined
      if (!payload?.user_uuid) return
      set_online_overrides((prev) => ({
        ...prev,
        [payload.user_uuid as string]: payload.is_online,
      }))
    })
    return unsub
  }, [subscribe])

  const accepted = sort_members(
    members.filter((m) => m.status === 'accepted'),
    creatorId,
  )
  const pending = members.filter(
    (m) => m.status === 'pending' || m.status === 'invited',
  )

  function render_member(m: GroupMember) {
    const name = `${m.user.first_name} ${m.user.last_name}`
    const meta = status_meta[m.status] ?? status_meta.accepted
    const show_actions =
      isCreator && (m.status === 'pending' || m.status === 'invited')
    const is_creator = m.user.id === creatorId
    const is_online = online_overrides[m.user.id] ?? m.user.is_online ?? false

    return (
      <li key={m.user.id} className="flex items-center gap-3 py-2">
        <div className="relative">
          <Link href={`/profile/${m.user.id}`}>
            <Avatar src={m.user.avatar_path} alt={name} size="sm" />
          </Link>
          {is_online && (
            <span className="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-white bg-green-500" />
          )}
        </div>
        <div className="min-w-0 flex-1">
          <Link
            href={`/profile/${m.user.id}`}
            className="text-sm font-medium text-gray-100 hover:underline"
          >
            {name}
            {is_creator && (
              <span className="ml-1.5" title="Group creator">
                👑
              </span>
            )}
          </Link>
          {m.user.nickname && (
            <p className="text-xs text-gray-300">@{m.user.nickname}</p>
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
          <h4 className="mb-1 text-xs font-semibold uppercase tracking-wide text-gray-300">
            Members ({accepted.length})
          </h4>
          <ul>{accepted.map(render_member)}</ul>
        </div>
      )}

      {pending.length > 0 && (
        <div className="mt-3 border-t border-purple-400/15 pt-3">
          <h4 className="mb-1 text-xs font-semibold uppercase tracking-wide text-gray-300">
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
