'use client'

import { useState, useEffect } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import client from '@/api/client'
import { Avatar } from '@/components/ui/Avatar'
import { Button } from '@/components/ui/Button'
import { Card, CardContent, CardHeader } from '@/components/ui/Card'
import { ImageUpload } from '@/components/common/ImageUpload'
import { InfiniteScroll } from '@/components/common/InfiniteScroll'
import { Input } from '@/components/ui/Input'
import { Spinner } from '@/components/ui/Spinner'
import { useUserPosts } from '@/hooks/usePosts'
import { PostCard } from '@/components/post/PostCard'
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
  const { user: currentUser, set_user: set_auth_user } = useAuth()
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
  const [avatar_loading, set_avatar_loading] = useState(false)
  const [avatar_error, set_avatar_error] = useState('')

  const is_own = uuid === 'me' || uuid === currentUser?.id
  const resolved_uuid = is_own ? currentUser?.id : uuid

  const user_posts = useUserPosts(resolved_uuid ?? '')
  const all_posts = user_posts.data?.pages.flatMap((page) => page) ?? []

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

  const handle_avatar_upload = async (file: File) => {
    if (!currentUser || avatar_loading) return
    set_avatar_loading(true)
    set_avatar_error('')
    try {
      const form = new FormData()
      form.append('image', file)
      const res = await client.post(
        `/api/users/${currentUser.id}/avatar`,
        form,
        { headers: { 'Content-Type': 'multipart/form-data' } },
      )
      const avatar_path: string = res.data.data.avatar_path
      set_profile((prev) =>
        prev
          ? { ...prev, user: { ...prev.user, avatar_path } }
          : prev,
      )
      set_auth_user(currentUser ? { ...currentUser, avatar_path } : currentUser)
    } catch (err: any) {
      set_avatar_error(err?.response?.data?.error || 'Failed to update avatar')
    } finally {
      set_avatar_loading(false)
    }
  }

  if (is_loading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-4">
          <div className="h-20 w-20 animate-pulse rounded-full bg-purple-400/20" />
          <div className="space-y-2">
            <div className="h-6 w-48 animate-pulse rounded bg-purple-400/20" />
            <div className="h-4 w-32 animate-pulse rounded bg-purple-400/20" />
          </div>
        </div>
      </div>
    )
  }

  if (!profile) {
    return (
      <div className="rounded-lg border border-purple-400/20 bg-[#241748] p-8 text-center">
        <p className="text-sm text-gray-300">User not found</p>
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
      <div className="rounded-lg border border-purple-400/20 bg-[#241748] p-6">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex items-start gap-4">
            <div className="relative">
              {is_own ? (
                <ImageUpload on_select={handle_avatar_upload}>
                  <button
                    type="button"
                    className="group relative block overflow-hidden rounded-full"
                    title="Change profile photo"
                  >
                    <Avatar
                      src={profile.user.avatar_path}
                      alt={`${profile.user.first_name} ${profile.user.last_name}`}
                      size="xl"
                    />
                    {avatar_loading ? (
                      <span className="absolute inset-0 flex items-center justify-center bg-black/50">
                        <Spinner size="sm" />
                      </span>
                    ) : (
                      <span className="absolute inset-0 flex items-center justify-center bg-black/40 text-xs font-medium text-white opacity-0 transition-opacity group-hover:opacity-100">
                        Change
                      </span>
                    )}
                  </button>
                </ImageUpload>
              ) : (
                <Avatar
                  src={profile.user.avatar_path}
                  alt={`${profile.user.first_name} ${profile.user.last_name}`}
                  size="xl"
                />
              )}
              {profile.user.is_online && (
                <span className="absolute bottom-1 right-1 h-4 w-4 rounded-full border-2 border-white bg-green-500" />
              )}
            </div>
            <div>
              <h1 className="text-xl font-bold text-gray-100">
                {profile.user.first_name} {profile.user.last_name}
              </h1>
              <p className="text-sm text-gray-300">@{profile.user.nickname}</p>
              {(profile.user.is_public ? 'Public' : 'Private')}
              <span className="mx-1.5 text-gray-300">·</span>
              <span className="text-xs text-gray-300">
                Joined{' '}
                {new Date(profile.user.created_at).toLocaleDateString()}
              </span>
              {profile.user.about_me && (
                <p className="mt-2 text-sm text-gray-300">{profile.user.about_me}</p>
              )}
            </div>
          </div>

          <div className="flex items-center gap-2">
            {is_own ? (
              <Button
                size="sm"
                variant="outline"
                className="shrink-0 whitespace-nowrap"
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

        {avatar_error && (
          <div className="mt-3 rounded-lg bg-red-500/15 p-3 text-sm text-red-200">
            {avatar_error}
          </div>
        )}

        <div className="mt-4 flex gap-6">
          {is_own ? (
            <>
              <button
                onClick={() => router.push('/followers')}
                className="text-sm text-gray-300 hover:text-gray-100"
              >
                <span className="font-semibold text-gray-100">
                  {profile.follower_count}
                </span>{' '}
                followers
              </button>
              <button
                onClick={() => router.push('/followers?tab=following')}
                className="text-sm text-gray-300 hover:text-gray-100"
              >
                <span className="font-semibold text-gray-100">
                  {profile.following_count}
                </span>{' '}
                following
              </button>
            </>
          ) : (
            <>
              <span className="text-sm text-gray-300">
                <span className="font-semibold text-gray-100">
                  {profile.follower_count}
                </span>{' '}
                followers
              </span>
              <span className="text-sm text-gray-300">
                <span className="font-semibold text-gray-100">
                  {profile.following_count}
                </span>{' '}
                following
              </span>
            </>
          )}
          <span className="text-sm text-gray-300">
            <span className="font-semibold text-gray-100">
              {profile.post_count}
            </span>{' '}
            posts
          </span>
        </div>
      </div>

      {show_edit && (
        <Card>
          <CardHeader>
            <h2 className="text-base font-semibold text-gray-100">Edit profile</h2>
          </CardHeader>
          <CardContent className="space-y-4">
            {edit_error && (
              <div className="rounded-lg bg-red-500/15 p-3 text-sm text-red-200">
                {edit_error}
              </div>
            )}
            <Input
              id="nickname"
              label="Nickname"
              value={edit_form.nickname}
              onChange={(e) =>
                set_edit_form((f) => ({ ...f, nickname: e.target.value }))
              }
            />
            <label className="block">
              <span className="text-sm font-medium text-gray-300">
                About me
              </span>
              <textarea
                value={edit_form.about_me}
                onChange={(e) =>
                  set_edit_form((f) => ({ ...f, about_me: e.target.value }))
                }
                rows={3}
                className="mt-1 block w-full resize-none rounded-lg border border-purple-400/30 px-3 py-2 text-sm focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-500"
              />
            </label>
            <label className="flex items-center gap-2">
              <input
                type="checkbox"
                checked={edit_form.is_public}
                onChange={(e) =>
                  set_edit_form((f) => ({ ...f, is_public: e.target.checked }))
                }
                className="h-4 w-4 rounded border-purple-400/30 text-violet-600 focus:ring-violet-500"
              />
              <span className="text-sm text-gray-300">Public profile</span>
            </label>
            <div className="flex justify-end gap-2 border-t border-purple-400/20 pt-4">
              <Button
                variant="outline"
                onClick={() => set_show_edit(false)}
                disabled={edit_saving}
              >
                Cancel
              </Button>
              <Button onClick={handleSave} loading={edit_saving}>
                Save changes
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <div>
        <h2 className="mb-3 text-lg font-semibold text-gray-100">Posts</h2>
        {user_posts.isLoading ? (
          <div className="flex justify-center py-8">
            <Spinner />
          </div>
        ) : all_posts.length === 0 ? (
          <div className="rounded-lg border border-purple-400/20 bg-[#241748] p-6 text-center">
            <p className="text-sm text-gray-300">No posts yet</p>
          </div>
        ) : (
          <InfiniteScroll
            has_next={!!user_posts.hasNextPage}
            is_loading={user_posts.isFetchingNextPage}
            on_load_more={() => user_posts.fetchNextPage()}
          >
            <div className="space-y-3">
              {all_posts.map((post) => (
                <PostCard key={post.id} post={post} hide_author hide_actions />
              ))}
            </div>
          </InfiniteScroll>
        )}
      </div>
    </div>
  )
}
