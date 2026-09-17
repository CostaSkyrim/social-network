'use client'

import { useState, useCallback } from 'react'
import Link from 'next/link'
import { Avatar } from '@/components/ui/Avatar'
import { Card, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { ImageUpload } from '@/components/common/ImageUpload'
import { UserVisibilityPicker } from './UserVisibilityPicker'
import { format_date } from '@/lib/format'
import { get_media_url } from '@/lib/media'
import { getPost } from '@/api/posts'
import { useAuth } from '@/context/AuthProvider'
import { useEditPost, useDeletePost } from '@/hooks/usePosts'
import { PrivacySelector } from './PrivacySelector'
import type { Post } from '@/types/post'

interface PostCardProps {
  post: Post
}

export function PostCard({ post }: PostCardProps) {
  const { user } = useAuth()
  const edit_mutation = useEditPost()
  const delete_mutation = useDeletePost()

  const [is_editing, set_editing] = useState(false)
  const [edit_content, set_edit_content] = useState(post.content ?? '')
  const [edit_privacy, set_edit_privacy] = useState(post.privacy_level)
  const [edit_image, set_edit_image] = useState<File | null>(null)
  const [edit_preview_url, set_edit_preview_url] = useState<string | null>(null)
  const [edit_remove_image, set_edit_remove_image] = useState(false)
  const [edit_visible_user_ids, set_edit_visible_user_ids] = useState<string[]>(
    post.visible_user_ids ?? [],
  )

  const author_name = post.author
    ? `${post.author.first_name} ${post.author.last_name}`
    : 'Unknown'

  const is_deleted = post.is_deleted
  const is_owner = user && Number(post.author_id) === Number(user.id)

  const start_edit = async () => {
    set_editing(true)
    // The feed doesn't include the visibility list; fetch it for private posts.
    if (post.privacy_level === 'private' && post.visible_user_ids === undefined) {
      try {
        const full = await getPost(post.id)
        set_edit_visible_user_ids(full.visible_user_ids ?? [])
      } catch {
        // silently fail
      }
    }
  }

  const handle_edit = async () => {
    await edit_mutation.mutateAsync({
      id: post.id,
      content: edit_content,
      privacy_level: edit_privacy,
      image: edit_image ?? undefined,
      remove_image: edit_remove_image,
      visible_user_ids: edit_privacy === 'private' ? edit_visible_user_ids : [],
    })
    set_editing(false)
    set_edit_image(null)
    set_edit_remove_image(false)
    if (edit_preview_url) URL.revokeObjectURL(edit_preview_url)
    set_edit_preview_url(null)
  }

  const handle_edit_image = (file: File) => {
    if (edit_preview_url) URL.revokeObjectURL(edit_preview_url)
    set_edit_image(file)
    set_edit_preview_url(URL.createObjectURL(file))
    set_edit_remove_image(false)
  }

  const handle_edit_remove_image = useCallback(() => {
    if (edit_preview_url) URL.revokeObjectURL(edit_preview_url)
    set_edit_image(null)
    set_edit_preview_url(null)
    set_edit_remove_image(true)
  }, [edit_preview_url])

  const handle_delete = () => {
    if (confirm('Delete this post?')) {
      delete_mutation.mutate(post.id)
    }
  }

  return (
    <Card>
      <CardContent className="space-y-3">
        <div className="flex items-center gap-3">
          <Link href={`/profile/${post.author_id}`}>
            <Avatar
              src={post.author?.avatar_path}
              alt={author_name}
              size="md"
            />
          </Link>
          <div className="min-w-0 flex-1">
            <Link
              href={`/profile/${post.author_id}`}
              className="text-sm font-medium text-gray-100 hover:underline"
            >
              {author_name}
            </Link>
            <p className="text-xs text-gray-300">{format_date(post.created_at)}</p>
          </div>
          {is_deleted ? (
            <span className="text-xs text-red-400 font-medium">deleted</span>
          ) : post.privacy_level !== 'public' ? (
            <span className="text-xs text-gray-300">
              {post.privacy_level === 'followers' ? '🫂' : '🔒'}
            </span>
          ) : null}
          {!is_deleted && is_owner && !is_editing && (
            <div className="flex items-center gap-1">
              <button
                onClick={start_edit}
                className="rounded p-1 text-gray-300 hover:bg-purple-400/15 hover:text-gray-300"
                title="Edit post"
              >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                </svg>
              </button>
              <button
                onClick={handle_delete}
                className="rounded p-1 text-gray-300 hover:bg-purple-400/15 hover:text-red-500"
                title="Delete post"
              >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          )}
        </div>

        {is_deleted ? (
          <Link href={`/posts/${post.id}`}>
            <p className="text-sm text-gray-300 italic">[deleted]</p>
          </Link>
        ) : is_editing ? (
          <div className="space-y-3">
            <textarea
              value={edit_content}
              onChange={(e) => set_edit_content(e.target.value)}
              className="w-full rounded-lg border border-purple-400/30 p-3 text-sm focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-500"
              rows={3}
              disabled={edit_mutation.isPending}
            />
            {(edit_preview_url || (post.image_path && !edit_remove_image && !edit_image)) && (
              <div className="relative">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={edit_preview_url ?? get_media_url(post.image_path)}
                  alt="Post image"
                  className="max-h-64 w-full rounded-lg object-contain"
                />
                <button
                  type="button"
                  onClick={handle_edit_remove_image}
                  className="absolute right-2 top-2 rounded-full bg-black/60 px-2 py-1 text-xs text-white hover:bg-black/80"
                >
                  Remove
                </button>
              </div>
            )}
            {edit_privacy === 'private' && (
              <UserVisibilityPicker
                selected={edit_visible_user_ids}
                onChange={set_edit_visible_user_ids}
              />
            )}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <ImageUpload on_select={handle_edit_image}>
                  <Button type="button" variant="ghost" size="sm">
                    {post.image_path && !edit_remove_image ? 'Change image' : 'Add image'}
                  </Button>
                </ImageUpload>
                <PrivacySelector value={edit_privacy} onChange={(v) => set_edit_privacy(v as typeof edit_privacy)} />
              </div>
              <div className="flex items-center gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => set_editing(false)}
                  disabled={edit_mutation.isPending}
                >
                  Cancel
                </Button>
                <Button
                  size="sm"
                  onClick={handle_edit}
                  disabled={edit_mutation.isPending || (edit_content.trim() === '' && !edit_image && !(post.image_path && !edit_remove_image))}
                >
                  Save
                </Button>
              </div>
            </div>
          </div>
        ) : (
          <Link href={`/posts/${post.id}`}>
            <p className="text-sm text-gray-100 whitespace-pre-wrap">{post.content}</p>
          </Link>
        )}

        {!is_deleted && !is_editing && post.image_path && (
          /* eslint-disable-next-line @next/next/no-img-element */
          <img
            src={get_media_url(post.image_path)}
            alt="Post image"
            className="max-h-96 w-full rounded-lg object-contain"
          />
        )}

        <Link href={`/posts/${post.id}`} className="flex items-center gap-4 text-sm text-gray-300 hover:text-gray-300">
          <span>{post.comment_count ?? 0} comments</span>
        </Link>
      </CardContent>
    </Card>
  )
}
