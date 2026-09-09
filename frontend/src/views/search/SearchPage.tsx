'use client'

import { useState, useEffect, useMemo } from 'react'
import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { searchUsers, followUser, unfollowUser } from '@/api/users'
import { useAuth } from '@/context/AuthProvider'
import { Avatar } from '@/components/ui/Avatar'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { EmptyState } from '@/components/common/EmptyState'
import { cn } from '@/lib/cn'
import { useUI } from '@/context/UIProvider'

export default function SearchPage() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { user: currentUser } = useAuth()
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  const [q, set_q] = useState(() => searchParams.get('q') ?? '')
  const [debounced_q, set_debounced_q] = useState(q)
  const [busy_id, set_busy_id] = useState<string | null>(null)

  useEffect(() => {
    const t = setTimeout(() => {
      set_debounced_q(q.trim())
      if (q.trim()) router.replace(`/search?q=${encodeURIComponent(q.trim())}`, { scroll: false })
      else router.replace('/search', { scroll: false })
    }, 300)
    return () => clearTimeout(t)
  }, [q, router])

  const { data: results, isFetching } = useQuery({
    queryKey: ['user-search', debounced_q],
    queryFn: () => searchUsers(debounced_q),
    enabled: debounced_q.length > 0,
  })

  const visible = useMemo(
    () => (results ?? []).filter((u) => u.id !== currentUser?.id),
    [results, currentUser],
  )

  async function toggle_follow(u: { id: string; is_following: boolean }) {
    set_busy_id(u.id)
    try {
      if (u.is_following) await unfollowUser(u.id)
      else await followUser(u.id)
      await query_client.invalidateQueries({ queryKey: ['user-search', debounced_q] })
    } catch (err: any) {
      show_toast({
        message: err?.response?.data?.error || 'Failed to update follow',
        type: 'error',
      })
    } finally {
      set_busy_id(null)
    }
  }

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900">Search</h2>

      <div className="relative">
        <input
          type="text"
          value={q}
          onChange={(e) => set_q(e.target.value)}
          placeholder="Search people by name or @nickname..."
          autoFocus
          className="w-full rounded-lg border border-gray-300 px-4 py-2.5 pl-10 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-200"
        />
        <svg
          className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M21 21l-4.35-4.35M17 10a7 7 0 11-14 0 7 7 0 0114 0z"
          />
        </svg>
        {isFetching && debounced_q && (
          <span className="absolute right-3 top-1/2 -translate-y-1/2">
            <Spinner size="sm" />
          </span>
        )}
      </div>

      {!debounced_q ? (
        <EmptyState
          title="Find people to follow"
          description="Search by first name, last name, or nickname."
        />
      ) : visible.length === 0 && !isFetching ? (
        <EmptyState
          title="No users found"
          description={`No results for "${debounced_q}".`}
        />
      ) : (
        <div className="divide-y divide-gray-100 rounded-lg border border-gray-200 bg-white">
          {visible.map((u) => {
            const name = `${u.first_name} ${u.last_name}`
            return (
              <div key={u.id} className="flex items-center gap-3 px-4 py-3">
                <div className="relative">
                  <Link href={`/profile/${u.id}`}>
                    <Avatar src={u.avatar_path} alt={name} size="md" />
                  </Link>
                  {u.is_online && (
                    <span className="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-white bg-green-500" />
                  )}
                </div>
                <div className="min-w-0 flex-1">
                  <Link
                    href={`/profile/${u.id}`}
                    className="text-sm font-medium text-gray-900 hover:underline"
                  >
                    {name}
                  </Link>
                  <p className="truncate text-xs text-gray-500">
                    {u.nickname ? `@${u.nickname}` : ''}
                    <span className="mx-1 text-gray-300">·</span>
                    {u.is_public ? 'Public' : 'Private'}
                  </p>
                </div>
                <Button
                  size="sm"
                  loading={busy_id === u.id}
                  disabled={busy_id !== null || u.is_follow_pending}
                  onClick={() => toggle_follow(u)}
                  variant={u.is_following ? 'outline' : 'primary'}
                  className={cn(u.is_follow_pending && 'cursor-not-allowed opacity-60')}
                >
                  {u.is_following ? 'Following' : u.is_follow_pending ? 'Requested' : 'Follow'}
                </Button>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
