'use client'

import { useState, useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import { useWebSocket } from '@/hooks/useWebSocket'
import { fetchFollowers, fetchFollowing } from '@/api/chat'
import { Avatar } from '@/components/ui/Avatar'
import { EmptyState } from '@/components/common/EmptyState'
import { cn } from '@/lib/cn'

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

type Tab = 'followers' | 'following'

export default function FollowersPage() {
  const { user } = useAuth()
  const { subscribe } = useWebSocket()
  const router = useRouter()
  const searchParams = useSearchParams()
  const [followers, set_followers] = useState<FollowerItem[]>([])
  const [following, set_following] = useState<FollowerItem[]>([])
  const [is_loading, set_is_loading] = useState(true)
  const [tab, set_tab] = useState<Tab>(() =>
    searchParams.get('tab') === 'following' ? 'following' : 'followers',
  )

  useEffect(() => {
    const userID = user?.id ?? ''
    if (!userID) return

    let cancelled = false
    async function load() {
      try {
        console.log('[FollowersPage] Fetching followers for user_id=', userID)
        const data = await fetchFollowers(userID)
        console.log('[FollowersPage] Got followers:', data)
        if (!cancelled) set_followers(data)
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
    const userID = user?.id ?? ''
    if (!userID) return

    let cancelled = false
    async function load() {
      try {
        const data = await fetchFollowing(userID)
        if (!cancelled) set_following(data)
      } catch {
        // silently fail
      }
    }
    load()
    return () => { cancelled = true }
  }, [user?.id])

  useEffect(() => {
    const unsub = subscribe('presence_update', (msg) => {
      const payload = msg.payload as { user_id: number; user_uuid: string; is_online: boolean } | undefined
      if (!payload) return
      const update = (prev: FollowerItem[]) =>
        prev.map((f) =>
          f.id === payload.user_uuid
            ? { ...f, is_online: payload.is_online }
            : f,
        )
      set_followers(update)
      set_following(update)
    })
    return unsub
  }, [subscribe])

  useEffect(() => {
    const unsub = subscribe('chat_message', () => {
      const userID = user?.id ?? ''
      if (!userID) return
      fetchFollowers(userID).then(set_followers).catch(() => {})
      fetchFollowing(userID).then(set_following).catch(() => {})
    })
    return unsub
  }, [user?.id, subscribe])

  const list = tab === 'followers' ? followers : following
  const title = tab === 'followers' ? 'Followers' : 'Following'

  if (is_loading) {
    return (
      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-gray-100">{title}</h2>
        <div className="divide-y divide-purple-400/15 rounded-lg border border-purple-400/20 bg-[#241748]">
          {Array.from({ length: 5 }).map((_, i) => (
            <div key={i} className="flex items-center gap-3 px-4 py-3">
              <div className="h-10 w-10 animate-pulse rounded-full bg-purple-400/20" />
              <div className="flex-1 space-y-1">
                <div className="h-4 w-32 animate-pulse rounded bg-purple-400/20" />
                <div className="h-3 w-20 animate-pulse rounded bg-purple-400/20" />
              </div>
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-1 rounded-lg bg-purple-400/15 p-1">
        <button
          onClick={() => set_tab('followers')}
          className={cn(
            'flex-1 rounded-md py-1.5 text-sm font-medium transition-colors',
            tab === 'followers'
              ? 'bg-[#241748] text-gray-100 shadow-sm'
              : 'text-gray-300 hover:text-gray-300',
          )}
        >
          Followers ({followers.length})
        </button>
        <button
          onClick={() => set_tab('following')}
          className={cn(
            'flex-1 rounded-md py-1.5 text-sm font-medium transition-colors',
            tab === 'following'
              ? 'bg-[#241748] text-gray-100 shadow-sm'
              : 'text-gray-300 hover:text-gray-300',
          )}
        >
          Following ({following.length})
        </button>
      </div>

      {list.length === 0 ? (
        <EmptyState
          title={`No ${tab} yet`}
          description={
            tab === 'followers'
              ? "When someone follows you, they'll appear here."
              : "People you follow will appear here."
          }
        />
      ) : (
        <div className="divide-y divide-purple-400/15 rounded-lg border border-purple-400/20 bg-[#241748]">
          {list.map((f) => (
            <button
              key={f.id}
              onClick={() => router.push(`/chat/${f.id}`)}
              className="flex w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-purple-400/10"
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
                <p className="truncate text-sm font-medium text-gray-100">
                  {f.first_name} {f.last_name}
                </p>
                <p className="truncate text-xs text-gray-300">
                  {f.nickname ? `@${f.nickname}` : ''}
                  {f.is_public ? ' · public' : ' · private'}
                </p>
              </div>
              <div className="flex items-center gap-2">
                {f.unread_count > 0 && (
                  <span className="flex h-5 min-w-5 items-center justify-center rounded-full bg-red-500 px-1.5 text-[11px] font-bold text-white">
                    {f.unread_count > 99 ? '99+' : f.unread_count}
                  </span>
                )}
                <svg className="h-5 w-5 text-gray-300" fill="none" viewBox="0 0 24 24" stroke="currentColor">
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
