'use client'

import Link from 'next/link'
import { Avatar } from '@/components/ui/Avatar'
import { Badge } from '@/components/ui/Badge'
import type { GroupMember } from '@/types/group'

interface MemberListProps {
  members: GroupMember[]
}

const status_meta: Record<string, { label: string; variant: 'default' | 'success' | 'warning' | 'danger' }> = {
  accepted: { label: 'Member', variant: 'success' },
  pending: { label: 'Pending', variant: 'warning' },
  invited: { label: 'Invited', variant: 'warning' },
  declined: { label: 'Declined', variant: 'danger' },
}

export function MemberList({ members }: MemberListProps) {
  const accepted = members.filter((m) => m.status === 'accepted')
  const pending = members.filter((m) => m.status !== 'accepted')

  function render_member(m: GroupMember) {
    const name = `${m.user.first_name} ${m.user.last_name}`
    const meta = status_meta[m.status] ?? status_meta.accepted

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
        {m.status !== 'accepted' && (
          <Badge variant={meta.variant}>{meta.label}</Badge>
        )}
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
