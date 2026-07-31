'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import { useWebSocket } from '@/hooks/useWebSocket'
import { fetchFollowers } from '@/api/chat'
import { Avatar } from '@/components/ui/Avatar'
import { EmptyState } from '@/components/common/EmptyState'

interface FollowerItem {
  id: string
  first_name: string
  last_name: string
  nickname?: string
  avatar_path?: string
  is_public: boolean
  is_online?: boolean
  unread_count: number
  last_dm_at?: string
}

export default function FollowersPage() {
  const { user } = useAuth()
  const { subscribe } = useWebSocket()
  const router = useRouter()
  const [followers, set_followers] = useState<FollowerItem[]>([])
  const [is_loading, set_is_loading] = useState(true)

  useEffect(() => {
    if (!user?.id) return

    let cancelled = false
    async function load() {
      try {
        const data = await fetchFollowers(user.id)
        if (!cancelled) {
          set_followers(data)
        }
      } catch {
        // silently fail
      } finally {
        if (!cancelled) set_is_loading(false)
      }
    }
    load()
    return () => { cancelled = true }
  }, [user?.id])

  useEffect(() => {
    const unsub = subscribe('presence_update', (msg) => {
      const payload = msg.payload as { user_id: number; is_online: boolean } | undefined
      if (!payload) return
      set_followers((prev) =>
        prev.map((f) =>
          f.id === String(payload.user_id)
            ? { ...f, is_online: payload.is_online }
            : f,
        ),
      )
    })
    return unsub
  }, [subscribe])

  if (is_loading) {
    return (
      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-gray-900">Followers</h2>
        <div className="divide-y divide-gray-100 rounded-lg border border-gray-200 bg-white">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-center gap-3 px-4 py-3">
              <div className="h-10 w-10 animate-pulse rounded-full bg-gray-200" />
              <div className="flex-1 space-y-1">
                <div className="h-4 w-32 animate-pulse rounded bg-gray-200" />
                <div className="h-3 w-20 animate-pulse rounded bg-gray-200" />
              </div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold text-gray-900">Followers</h2>
      {followers.length === 0 ? (
        <EmptyState title="No followers yet" description="When someone follows you, they'll appear here." />
      ) : (
        <div className="divide-y divide-gray-100 rounded-lg border border-gray-200 bg-white">
          {followers.map((f) => (
            <button
              key={f.id}
              onClick={() => router.push(`/chat/${f.id}`)}
              className="flex w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-gray-50"
            >
              <div className="relative">
                <Avatar
                  src={f.avatar_path}
                  alt={`${f.first_name} ${f.last_name}`}
                  size="sm"
                />
                {f.is_online && (
                  <span className="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-white bg-green-500" />
                )}
              </div>
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium text-gray-900">
                  {f.first_name} {f.last_name}
                </p>
                <p className="truncate text-xs text-gray-500">
                  {f.nickname ? `@${f.nickname}` : ''}
                  {f.is_public ? ' · public' : ' · private'}
                </p>
              </div>
              <div className="flex items-center gap-2">
                {f.unread_count > 0 && (
                  <span className="flex h-5 min-w-5 items-center justify-center rounded-full bg-blue-500 px-1.5 text-[11px] font-bold text-white">
                    {f.unread_count > 99 ? '99+' : f.unread_count}
                  </span>
                )}
                <svg className="h-5 w-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
