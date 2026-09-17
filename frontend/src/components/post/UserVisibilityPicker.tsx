'use client'

import { useEffect, useMemo, useState } from 'react'
import { fetchFollowers } from '@/api/chat'
import { Avatar } from '@/components/ui/Avatar'
import { useAuth } from '@/context/AuthProvider'

interface Follower {
  id: string
  first_name: string
  last_name: string
  nickname?: string
  avatar_path?: string
}

interface UserVisibilityPickerProps {
  selected: string[]
  onChange: (ids: string[]) => void
}

export function UserVisibilityPicker({ selected, onChange }: UserVisibilityPickerProps) {
  const { user } = useAuth()
  const [followers, set_followers] = useState<Follower[]>([])
  const [is_loading, set_is_loading] = useState(true)
  const [query, set_query] = useState('')

  useEffect(() => {
    const userID = user?.id
    if (!userID) return

    let cancelled = false
    fetchFollowers(userID)
      .then((data) => {
        if (!cancelled) set_followers(data as Follower[])
      })
      .catch(() => {
        // silently fail
      })
      .finally(() => {
        if (!cancelled) set_is_loading(false)
      })

    return () => {
      cancelled = true
    }
  }, [user?.id])

  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase()
    if (!q) return followers
    return followers.filter((f) => {
      const name = `${f.first_name} ${f.last_name}`.toLowerCase()
      const nickname = f.nickname?.toLowerCase() ?? ''
      return name.includes(q) || nickname.includes(q)
    })
  }, [followers, query])

  function toggle(id: string) {
    if (selected.includes(id)) onChange(selected.filter((x) => x !== id))
    else onChange([...selected, id])
  }

  return (
    <div className="rounded-lg border border-purple-400/20 p-3">
      <p className="mb-2 text-xs font-medium text-gray-300">
        Who can see this post?
      </p>

      {is_loading ? (
        <p className="text-xs text-gray-300">Loading followers…</p>
      ) : followers.length === 0 ? (
        <p className="text-xs text-gray-300">
          You have no followers to choose from yet.
        </p>
      ) : (
        <>
          <div className="relative mb-2">
            <input
              type="text"
              value={query}
              onChange={(e) => set_query(e.target.value)}
              placeholder="Search by name or @nickname..."
              className="w-full rounded-lg border border-purple-400/30 py-1.5 pl-8 pr-2 text-sm focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-300/40"
            />
            <svg
              className="pointer-events-none absolute left-2.5 top-1/2 h-4 w-4 -translate-y-1/2 text-gray-300"
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
          </div>

          {filtered.length === 0 ? (
            <p className="py-2 text-xs text-gray-300">
              No matches for &ldquo;{query.trim()}&rdquo;.
            </p>
          ) : (
            <ul className="max-h-40 space-y-1 overflow-y-auto">
              {filtered.map((f) => {
                const name = `${f.first_name} ${f.last_name}`
                const checked = selected.includes(f.id)
                return (
                  <li key={f.id}>
                    <label className="flex cursor-pointer items-center gap-2 rounded p-1 hover:bg-purple-400/10">
                      <input
                        type="checkbox"
                        checked={checked}
                        onChange={() => toggle(f.id)}
                        className="h-4 w-4 rounded border-purple-400/30 text-violet-600 focus:ring-violet-500"
                      />
                      <Avatar src={f.avatar_path} alt={name} size="sm" />
                      <span className="min-w-0 flex-1 truncate text-sm text-gray-100">
                        {name}
                        {f.nickname ? (
                          <span className="ml-1 text-xs text-gray-300">@{f.nickname}</span>
                        ) : null}
                      </span>
                    </label>
                  </li>
                )
              })}
            </ul>
          )}
        </>
      )}

      {!is_loading && followers.length > 0 && (
        <p className="mt-2 text-[11px] text-gray-300">
          {selected.length} selected
        </p>
      )}
    </div>
  )
}
