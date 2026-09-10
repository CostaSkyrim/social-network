'use client'

import { useEffect, useState } from 'react'
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

  function toggle(id: string) {
    if (selected.includes(id)) onChange(selected.filter((x) => x !== id))
    else onChange([...selected, id])
  }

  return (
    <div className="rounded-lg border border-gray-200 p-3">
      <p className="mb-2 text-xs font-medium text-gray-600">
        Who can see this post?
      </p>

      {is_loading ? (
        <p className="text-xs text-gray-400">Loading followers…</p>
      ) : followers.length === 0 ? (
        <p className="text-xs text-gray-400">
          You have no followers to choose from yet.
        </p>
      ) : (
        <ul className="max-h-40 space-y-1 overflow-y-auto">
          {followers.map((f) => {
            const name = `${f.first_name} ${f.last_name}`
            const checked = selected.includes(f.id)
            return (
              <li key={f.id}>
                <label className="flex cursor-pointer items-center gap-2 rounded p-1 hover:bg-gray-50">
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={() => toggle(f.id)}
                    className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <Avatar src={f.avatar_path} alt={name} size="sm" />
                  <span className="min-w-0 flex-1 truncate text-sm text-gray-800">
                    {name}
                    {f.nickname ? (
                      <span className="ml-1 text-xs text-gray-400">@{f.nickname}</span>
                    ) : null}
                  </span>
                </label>
              </li>
            )
          })}
        </ul>
      )}

      {!is_loading && followers.length > 0 && (
        <p className="mt-2 text-[11px] text-gray-400">
          {selected.length} selected
        </p>
      )}
    </div>
  )
}
