'use client'

import { useState } from 'react'
import { useInviteGroupMember } from '@/hooks/useGroups'

interface InviteMemberProps {
  groupId: string
}

export function InviteMember({ groupId }: InviteMemberProps) {
  const [nickname, set_nickname] = useState('')
  const invite_mutation = useInviteGroupMember(groupId)

  async function handle_submit(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = nickname.trim()
    if (!trimmed) return
    await invite_mutation.mutateAsync(trimmed)
    set_nickname('')
  }

  return (
    <form onSubmit={handle_submit} className="flex items-center gap-2">
      <input
        type="text"
        value={nickname}
        onChange={(e) => set_nickname(e.target.value)}
        placeholder="Invite by nickname…"
        className="flex-1 rounded-lg border border-purple-400/30 px-3 py-1.5 text-sm focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-500"
      />
      <button
        type="submit"
        disabled={invite_mutation.isPending || !nickname.trim()}
        className="rounded-lg bg-violet-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-violet-700 disabled:opacity-50"
      >
        {invite_mutation.isPending ? 'Inviting…' : 'Invite'}
      </button>
    </form>
  )
}
