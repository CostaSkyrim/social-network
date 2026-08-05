'use client'

import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import { useNotifications } from '@/context/NotificationProvider'
import { Avatar } from '@/components/ui/Avatar'
import { useMessageBadge } from '@/hooks/useMessageBadge'
import { NavMenu } from '@/components/layout/NavMenu'

export function TopBar() {
  const { user } = useAuth()

  return (
    <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-gray-200 bg-white px-4">
      <div className="flex items-center gap-3">
        <NavMenu
          trigger={
            <button className="rounded-lg p-1 text-gray-500 hover:bg-gray-100">
              <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
              </svg>
            </button>
          }
        />
        <h1 className="text-lg font-bold text-blue-600">Social</h1>
      </div>

      <div className="flex items-center gap-3">
        <NavFollowersBadge />
        <NavNotificationBell />
        <Avatar
          src={user?.avatar_path}
          alt={`${user?.first_name} ${user?.last_name}`}
          size="sm"
        />
      </div>
    </header>
  )
}

function NavFollowersBadge() {
  const { message_count } = useMessageBadge()
  const router = useRouter()

  return (
    <button
      onClick={() => router.push('/followers')}
      className="relative rounded-lg p-1 text-gray-500 hover:bg-gray-100"
      title="Followers & Messages"
    >
      <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"
        />
      </svg>
      {message_count > 0 && (
        <span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-blue-500 px-1 text-[10px] font-bold text-white">
          {message_count > 99 ? '99+' : message_count}
        </span>
      )}
    </button>
  )
}

function NavNotificationBell() {
  const { unread_count } = useNotifications()
  const router = useRouter()

  return (
    <button
      onClick={() => router.push('/notifications')}
      className="relative rounded-lg p-1 text-gray-500 hover:bg-gray-100"
    >
      <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
        />
      </svg>
      {unread_count > 0 && (
        <span className="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-red-500 px-1 text-[10px] font-bold text-white">
          {unread_count > 99 ? '99+' : unread_count}
        </span>
      )}
    </button>
  )
}
