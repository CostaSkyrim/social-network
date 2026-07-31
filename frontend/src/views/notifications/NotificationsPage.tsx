'use client'

import { useNotifications } from '@/context/NotificationProvider'
import { EmptyState } from '@/components/common/EmptyState'
import { cn } from '@/lib/cn'

const notif_icons: Record<string, string> = {
  follow_request: '👋',
  follow_accepted: '✅',
  new_follower: '👤',
  new_post: '📝',
  group_invitation: '👥',
  group_join_request: '🚪',
  group_accepted: '🎉',
  new_event: '📅',
  event_reminder: '⏰',
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
  } = useNotifications()

  if (is_loading) {
    return (
      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-gray-900">Notifications</h2>
        <div className="divide-y divide-gray-100 rounded-lg border border-gray-200 bg-white">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-center gap-3 px-4 py-3">
              <div className="h-8 w-8 animate-pulse rounded-full bg-gray-200" />
              <div className="flex-1 space-y-1">
                <div className="h-4 w-48 animate-pulse rounded bg-gray-200" />
                <div className="h-3 w-16 animate-pulse rounded bg-gray-200" />
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
        <h2 className="text-xl font-semibold text-gray-900">Notifications</h2>
        {unread_count > 0 && (
          <button
            onClick={mark_all_read}
            className="text-sm font-medium text-blue-600 hover:text-blue-700"
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
        <div className="divide-y divide-gray-100 rounded-lg border border-gray-200 bg-white">
          {notifications.map((notif) => (
            <button
              key={notif.id}
              onClick={() => !notif.is_read && mark_read(notif.id)}
              className={cn(
                'flex w-full items-start gap-3 px-4 py-3 text-left transition-colors hover:bg-gray-50',
                !notif.is_read && 'bg-blue-50/50',
              )}
            >
              <span className="mt-0.5 text-xl">
                {notif_icons[notif.type] || '🔔'}
              </span>
              <div className="min-w-0 flex-1">
                <p
                  className={cn(
                    'text-sm',
                    !notif.is_read
                      ? 'font-semibold text-gray-900'
                      : 'text-gray-700',
                  )}
                >
                  {notif.content}
                </p>
                <p className="mt-0.5 text-xs text-gray-500">
                  {formatTimeAgo(notif.created_at)}
                </p>
              </div>
              {!notif.is_read && (
                <span className="mt-2 h-2 w-2 flex-shrink-0 rounded-full bg-blue-500" />
              )}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
