'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import client from '@/api/client'
import { Avatar } from '@/components/ui/Avatar'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { cn } from '@/lib/cn'

interface ProfileData {
  user: {
    id: string
    first_name: string
    last_name: string
    nickname?: string
    email: string
    about_me?: string
    avatar_path?: string
    is_public: boolean
    is_online?: boolean
    created_at: string
  }
  follower_count: number
  following_count: number
  post_count: number
  is_following: boolean
  is_follow_pending: boolean
  recent_posts: Array<{
    id: string
    content: string
    privacy_level: string
    created_at: string
    comment_count?: number
  }>
}

export default function ProfilePage() {
  const params = useParams()
  const router = useRouter()
  const { user: currentUser } = useAuth()
  const uuid = params?.uuid as string

  const [profile, set_profile] = useState<ProfileData | null>(null)
  const [is_loading, set_is_loading] = useState(true)
  const [follow_loading, set_follow_loading] = useState(false)
  const [follow_state, set_follow_state] = useState({ following: false, pending: false })
  const [show_edit, set_show_edit] = useState(false)
  const [edit_form, set_edit_form] = useState({
    nickname: '',
    about_me: '',
    is_public: true,
  })
  const [edit_error, set_edit_error] = useState('')
  const [edit_saving, set_edit_saving] = useState(false)

  const is_own = uuid === 'me' || uuid === currentUser?.id
  const resolved_uuid = is_own ? currentUser?.id : uuid

  useEffect(() => {
    if (!resolved_uuid) return

    let cancelled = false
    async function load() {
      try {
        const res = await client.get(`/api/users/${resolved_uuid}`)
        if (!cancelled) {
          const data = res.data.data
          set_profile(data)
          set_follow_state({
            following: data.is_following,
            pending: data.is_follow_pending,
          })
        }
      } catch {
        // silently fail
      } finally {
        if (!cancelled) set_is_loading(false)
      }
    }
    load()
    return () => { cancelled = true }
  }, [resolved_uuid])

  useEffect(() => {
    if (profile && is_own) {
      set_edit_form({
        nickname: profile.user.nickname ?? '',
        about_me: profile.user.about_me ?? '',
        is_public: profile.user.is_public,
      })
    }
  }, [profile, is_own])

  const handleFollow = async () => {
    if (!profile || !currentUser || follow_loading || follow_state.pending) return
    set_follow_loading(true)
    try {
      if (follow_state.following) {
        await client.post('/api/follow/remove', { user_id: profile.user.id })
        set_follow_state({ following: false, pending: false })
      } else {
        await client.post('/api/follow/request', { user_id: profile.user.id })
        if (profile.user.is_public) {
          set_follow_state({ following: true, pending: false })
        } else {
          set_follow_state({ following: false, pending: true })
        }
      }
    } catch {
      // silently fail
    } finally {
      set_follow_loading(false)
    }
  }

  const handleSave = async () => {
    if (!currentUser || edit_saving) return
    set_edit_saving(true)
    set_edit_error('')
    try {
      const res = await client.put(`/api/users/${currentUser.id}/edit`, {
        nickname: edit_form.nickname || null,
        about_me: edit_form.about_me || null,
        is_public: edit_form.is_public,
      })
      const updated = res.data.data
      set_profile((prev) =>
        prev
          ? {
              ...prev,
              user: {
                ...prev.user,
                nickname: updated.nickname,
                about_me: updated.about_me,
                is_public: updated.is_public,
              },
            }
          : prev,
      )
      set_show_edit(false)
    } catch (err: any) {
      set_edit_error(err?.response?.data?.error || 'Failed to save')
    } finally {
      set_edit_saving(false)
    }
  }

  if (is_loading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-4">
          <div className="h-20 w-20 animate-pulse rounded-full bg-gray-200" />
          <div className="space-y-2">
            <div className="h-6 w-48 animate-pulse rounded bg-gray-200" />
            <div className="h-4 w-32 animate-pulse rounded bg-gray-200" />
          </div>
        </div>
      </div>
    )
  }

  if (!profile) {
    return (
      <div className="rounded-lg border border-gray-200 bg-white p-8 text-center">
        <p className="text-sm text-gray-500">User not found</p>
      </div>
    )
  }

  const followLabel = follow_state.following
    ? 'Unfollow'
    : follow_state.pending
      ? 'Requested'
      : 'Follow'

  return (
    <div className="space-y-6">
      <div className="rounded-lg border border-gray-200 bg-white p-6">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex items-start gap-4">
            <div className="relative">
              <Avatar
                src={profile.user.avatar_path}
                alt={`${profile.user.first_name} ${profile.user.last_name}`}
                size="xl"
              />
              {profile.user.is_online && (
                <span className="absolute bottom-1 right-1 h-4 w-4 rounded-full border-2 border-white bg-green-500" />
              )}
            </div>
            <div>
              <h1 className="text-xl font-bold text-gray-900">
                {profile.user.first_name} {profile.user.last_name}
              </h1>
              <p className="text-sm text-gray-500">@{profile.user.nickname}</p>
              {(profile.user.is_public ? 'Public' : 'Private')}
              <span className="mx-1.5 text-gray-300">·</span>
              <span className="text-xs text-gray-500">
                Joined{' '}
                {new Date(profile.user.created_at).toLocaleDateString()}
              </span>
              {profile.user.about_me && (
                <p className="mt-2 text-sm text-gray-700">{profile.user.about_me}</p>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            {is_own ? (
              <Button
                size="sm"
                variant="outline"
                onClick={() => set_show_edit(true)}
              >
                Edit Profile
              </Button>
            ) : currentUser ? (
              <Button
                size="sm"
                onClick={handleFollow}
                loading={follow_loading}
                variant={follow_state.following ? 'outline' : 'primary'}
                disabled={follow_state.pending}
              >
                {followLabel}
              </Button>
            ) : null}
          </div>
        </div>

        <div className="mt-4 flex gap-6">
          {is_own ? (
            <>
              <button
                onClick={() => router.push('/followers')}
                className="text-sm text-gray-600 hover:text-gray-900"
              >
                <span className="font-semibold text-gray-900">
                  {profile.follower_count}
                </span>{' '}
                followers
              </button>
              <button
                onClick={() => router.push('/followers?tab=following')}
                className="text-sm text-gray-600 hover:text-gray-900"
              >
                <span className="font-semibold text-gray-900">
                  {profile.following_count}
                </span>{' '}
                following
              </button>
            </>
          ) : (
            <>
              <span className="text-sm text-gray-600">
                <span className="font-semibold text-gray-900">
                  {profile.follower_count}
                </span>{' '}
                followers
              </span>
              <span className="text-sm text-gray-600">
                <span className="font-semibold text-gray-900">
                  {profile.following_count}
                </span>{' '}
                following
              </span>
            </>
          )}
          <span className="text-sm text-gray-600">
            <span className="font-semibold text-gray-900">
              {profile.post_count}
            </span>{' '}
            posts
          </span>
        </div>
      </div>

      {show_edit && (
        <div className="rounded-lg border border-gray-200 bg-white p-6">
          <h2 className="mb-4 text-lg font-semibold text-gray-900">
            Edit Profile
          </h2>
          {edit_error && (
            <div className="mb-3 rounded-lg bg-red-50 p-3 text-sm text-red-700">
              {edit_error}
            </div>
          )}
          <div className="space-y-4">
            <Input
              id="nickname"
              label="Nickname"
              value={edit_form.nickname}
              onChange={(e) =>
                set_edit_form((f) => ({ ...f, nickname: e.target.value }))
              }
            />
            <label className="block">
              <span className="text-sm font-medium text-gray-700">
                About me
              </span>
              <textarea
                value={edit_form.about_me}
                onChange={(e) =>
                  set_edit_form((f) => ({ ...f, about_me: e.target.value }))
                }
                rows={3}
                className="mt-1 block w-full rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </label>
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={edit_form.is_public}
                onChange={(e) =>
                  set_edit_form((f) => ({ ...f, is_public: e.target.checked }))
                }
                className="h-4 w-4 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
              />
              <span className="text-sm text-gray-700">Public profile</span>
            </label>
            <div className="flex gap-2">
              <Button onClick={handleSave} loading={edit_saving} size="sm">
                Save
              </Button>
              <Button
                variant="outline"
                size="sm"
                onClick={() => set_show_edit(false)}
              >
                Cancel
              </Button>
            </div>
          </div>
        </div>
      )}

      <div>
        <h2 className="mb-3 text-lg font-semibold text-gray-900">Posts</h2>
        {profile.recent_posts.length === 0 ? (
          <div className="rounded-lg border border-gray-200 bg-white p-6 text-center">
            <p className="text-sm text-gray-500">No posts yet</p>
          </div>
        ) : (
          <div className="space-y-3">
            {profile.recent_posts.map((post) => (
              <div
                key={post.id}
                className="rounded-lg border border-gray-200 bg-white p-4"
              >
                <p className="text-sm text-gray-900">{post.content}</p>
                <div className="mt-2 flex items-center gap-3 text-xs text-gray-500">
                  <span>
                    {new Date(post.created_at).toLocaleDateString()}
                  </span>
                  {post.privacy_level !== 'public' && (
                    <span className="rounded bg-gray-100 px-1.5 py-0.5 text-[10px] font-medium uppercase">
                      {post.privacy_level}
                    </span>
                  )}
                  {(post.comment_count ?? 0) > 0 && (
                    <span>{post.comment_count} comments</span>
                  )}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
