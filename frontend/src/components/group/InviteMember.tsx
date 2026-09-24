'use client'

import { useState } from 'react'
import { useInviteGroupMember } from '@/hooks/useGroups'

interface InviteMemberProps {
  groupId: string
}

// Nicknames are displayed elsewhere as "@handle", so accept both spellings.
function normalize_nickname(raw: string): string {
  return raw.trim().replace(/^@+/, '')
}

export function InviteMember({ groupId }: InviteMemberProps) {
  const [nickname, set_nickname] = useState('')
  const [error, set_error] = useState<string | null>(null)
  const invite_mutation = useInviteGroupMember(groupId)

  async function handle_submit(e: React.FormEvent) {
    e.preventDefault()
    const trimmed = normalize_nickname(nickname)
    if (!trimmed) return

    set_error(null)
    try {
      await invite_mutation.mutateAsync(trimmed)
      set_nickname('')
    } catch (err: any) {
      // The mutation's onError already raises a toast; keep the input so the
      // handle can be corrected, show the reason inline, and swallow the
      // rejection so it does not surface as an unhandled promise error.
      set_error(err?.response?.data?.error || 'Failed to send invitation')
    }
  }

  return (
    <form onSubmit={handle_submit} className="space-y-1.5">
      <div className="flex items-center gap-2">
        <input
          type="text"
          value={nickname}
          onChange={(e) => {
            set_nickname(e.target.value)
            if (error) set_error(null)
          }}
          placeholder="Invite by nickname (e.g. alicej)…"
          className="flex-1 rounded-lg border border-purple-400/30 px-3 py-1.5 text-sm focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-500"
        />
        <button
          type="submit"
          disabled={invite_mutation.isPending || !normalize_nickname(nickname)}
          className="rounded-lg bg-violet-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-violet-700 disabled:opacity-50"
        >
          {invite_mutation.isPending ? 'Inviting…' : 'Invite'}
        </button>
      </div>
      {error && <p className="text-xs text-red-400">{error}</p>}
    </form>
  )
}
