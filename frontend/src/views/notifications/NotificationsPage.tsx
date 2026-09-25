'use client'

import { useState } from 'react'
import Link from 'next/link'
import { useNotifications } from '@/context/NotificationProvider'
import { useAuth } from '@/context/AuthProvider'
import { useAcceptGroupMember, useRejectGroupMember } from '@/hooks/useGroups'
import { EmptyState } from '@/components/common/EmptyState'
import { cn } from '@/lib/cn'

const notif_icons: Record<string, string> = {
  follow_request: '👋',
  follow_accepted: '✅',
  new_follower: '👤',
  new_post: '📝',
  new_comment: '💬',
  group_invitation: '👥',
  group_join_request: '🚪',
  group_accepted: '🎉',
  group_kicked: '🚫',
  new_event: '📅',
  event_reminder: '⏰',
}

const follow_types = new Set(['follow_request', 'follow_accepted', 'new_follower'])
const post_types = new Set(['new_post', 'new_comment'])
const group_types = new Set([
  'group_invitation',
  'group_join_request',
  'group_accepted',
  'group_kicked',
])
const event_types = new Set(['new_event', 'event_reminder'])

function notifHref(notif: {
  type: string
  related_id?: string
  from_user_id?: string
}): string | null {
  if (follow_types.has(notif.type) && notif.from_user_id) {
    return `/profile/${notif.from_user_id}`
  }
  if (post_types.has(notif.type) && notif.related_id) {
    return `/posts/${notif.related_id}`
  }
  if (group_types.has(notif.type) && notif.related_id) {
    return `/groups/${notif.related_id}`
  }
  if (event_types.has(notif.type) && notif.related_id) {
    return `/groups/${notif.related_id}?tab=events`
  }
  return null
}

function formatTimeAgo(iso: string): string {
  const now = Date.now()
  const then = new Date(iso).getTime()
  const diff = Math.floor((now - then) / 1000)

  if (diff < 60) return 'just now'
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  if (diff < 604800) return `${Math.floor(diff / 86400)}d ago`
  return new Date(iso).toLocaleDateString()
}

export default function NotificationsPage() {
  const {
    notifications,
    unread_count,
    is_loading,
    mark_read,
    mark_all_read,
    refresh,
  } = useNotifications()

  if (is_loading) {
    return (
      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-gray-100">Notifications</h2>
        <div className="divide-y divide-purple-400/20 rounded-xl border border-purple-400/25 bg-gradient-to-br from-[#3a1f74] to-[#241146] shadow-lg shadow-black/30">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-center gap-3 px-4 py-3">
              <div className="h-8 w-8 animate-pulse rounded-full bg-purple-400/20" />
              <div className="flex-1 space-y-1">
                <div className="h-4 w-48 animate-pulse rounded bg-purple-400/20" />
                <div className="h-3 w-16 animate-pulse rounded bg-purple-400/20" />
              </div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-gray-100">Notifications</h2>
        {unread_count > 0 && (
          <button
            onClick={mark_all_read}
            className="text-sm font-medium text-violet-600 hover:text-violet-700"
          >
            Mark all as read
          </button>
        )}
      </div>

      {notifications.length === 0 ? (
        <EmptyState
          title="No notifications"
          description="You're all caught up!"
        />
      ) : (
        <div className="divide-y divide-purple-400/20 rounded-xl border border-purple-400/25 bg-gradient-to-br from-[#3a1f74] to-[#241146] shadow-lg shadow-black/30">
          {notifications.map((notif) => (
            <NotificationRow
              key={notif.id}
              notif={notif}
              on_mark_read={() => !notif.is_read && mark_read(notif.id)}
              on_after_action={refresh}
            />
          ))}
        </div>
      )}
    </div>
  )
}

function NotificationRow({
  notif,
  on_mark_read,
  on_after_action,
}: {
  notif: {
    id: number
    type: string
    content: string
    related_id?: string
    from_user_id?: string
    is_read: boolean
    created_at: string
  }
  on_mark_read: () => void
  on_after_action: () => Promise<void>
}) {
  const { user } = useAuth()
  const is_join_request =
    notif.type === 'group_join_request' && !notif.is_read
  const is_invitation =
    notif.type === 'group_invitation' && !notif.is_read
  const is_actionable = is_join_request || is_invitation
  const group_id = notif.related_id
  // A join request targets the requester and is decided by the creator; an
  // invitation targets (and is decided by) the invited user.
  const target_id = is_invitation ? user?.id : notif.from_user_id
  const [action, set_action] = useState<'accept' | 'decline' | null>(null)
  const accept_mutation = useAcceptGroupMember(group_id ?? '')
  const reject_mutation = useRejectGroupMember(group_id ?? '')

  async function handle_accept() {
    if (!group_id || !target_id) return
    set_action('accept')
    try {
      await accept_mutation.mutateAsync(target_id)
      await on_mark_read()
      await on_after_action()
    } catch {
      // error toast handled by mutation
    } finally {
      set_action(null)
    }
  }

  async function handle_decline() {
    if (!group_id || !target_id) return
    set_action('decline')
    try {
      await reject_mutation.mutateAsync(target_id)
      await on_mark_read()
      await on_after_action()
    } catch {
      // error toast handled by mutation
    } finally {
      set_action(null)
    }
  }

  const is_pending =
    action !== null ||
    accept_mutation.isPending ||
    reject_mutation.isPending

  const href = notifHref(notif)
  const content = (
    <>
      <span className="mt-0.5 text-xl">
        {notif_icons[notif.type] || '🔔'}
      </span>
      <div className="min-w-0 flex-1">
        <p
          className={cn(
            'text-sm',
            !notif.is_read ? 'font-semibold text-gray-100' : 'text-gray-300',
          )}
        >
          {notif.content}
        </p>
        <p className="mt-0.5 text-xs text-gray-300">
          {formatTimeAgo(notif.created_at)}
        </p>
      </div>
    </>
  )

  return (
    <div
      className={cn(
        'flex w-full items-start gap-3 px-4 py-3 text-left transition-colors',
        !notif.is_read && 'bg-purple-400/15',
      )}
    >
      {href ? (
        <Link
          href={href}
          onClick={() => !notif.is_read && on_mark_read()}
          className="flex w-full min-w-0 items-start gap-3 hover:opacity-80"
        >
          {content}
        </Link>
      ) : (
        <div className="flex w-full min-w-0 items-start gap-3">{content}</div>
      )}

      {is_actionable && (
        <div className="mt-2 flex flex-shrink-0 items-center gap-2">
          <button
            onClick={handle_accept}
            disabled={is_pending}
            className="rounded-lg bg-green-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-green-700 disabled:opacity-50"
          >
            {action === 'accept' ? 'Accepting…' : 'Accept'}
          </button>
          <button
            onClick={handle_decline}
            disabled={is_pending}
            className="rounded-lg bg-red-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-red-700 disabled:opacity-50"
          >
            {action === 'decline' ? 'Declining…' : 'Decline'}
          </button>
        </div>
      )}
      {!notif.is_read && (
        <span className="mt-2 h-2 w-2 flex-shrink-0 rounded-full bg-violet-500" />
      )}
    </div>
  )
}
