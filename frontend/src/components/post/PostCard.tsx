'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Avatar } from '@/components/ui/Avatar'
import { Card, CardContent } from '@/components/ui/Card'
import { Button } from '@/components/ui/Button'
import { format_date } from '@/lib/format'
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

  const author_name = post.author
    ? `${post.author.first_name} ${post.author.last_name}`
    : 'Unknown'

  const is_owner = user && post.author_id === user.id

  const handle_edit = async () => {
    await edit_mutation.mutateAsync({
      id: post.db_id,
      content: edit_content,
      privacy_level: edit_privacy,
    })
    set_editing(false)
  }

  const handle_delete = () => {
    if (confirm('Delete this post?')) {
      delete_mutation.mutate(post.db_id)
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
              className="text-sm font-medium text-gray-900 hover:underline"
            >
              {author_name}
            </Link>
            <p className="text-xs text-gray-500">{format_date(post.created_at)}</p>
          </div>
          {post.privacy_level !== 'public' && (
            <span className="text-xs text-gray-400">
              {post.privacy_level === 'followers' ? '🫂' : '🔒'}
            </span>
          )}
          {is_owner && !is_editing && (
            <div className="flex items-center gap-1">
              <button
                onClick={() => set_editing(true)}
                className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
                title="Edit post"
              >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                </svg>
              </button>
              <button
                onClick={handle_delete}
                className="rounded p-1 text-gray-400 hover:bg-gray-100 hover:text-red-500"
                title="Delete post"
              >
                <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                </svg>
              </button>
            </div>
          )}
        </div>

        {is_editing ? (
          <div className="space-y-3">
            <textarea
              value={edit_content}
              onChange={(e) => set_edit_content(e.target.value)}
              className="w-full rounded-lg border border-gray-300 p-3 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              rows={3}
              disabled={edit_mutation.isPending}
            />
            <div className="flex items-center justify-between">
              <PrivacySelector value={edit_privacy} onChange={set_edit_privacy} />
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
                  disabled={edit_mutation.isPending || edit_content.trim() === ''}
                >
                  Save
                </Button>
              </div>
            </div>
          </div>
        ) : (
          <Link href={`/posts/${post.db_id}`}>
            <p className="text-sm text-gray-800 whitespace-pre-wrap">{post.content}</p>
          </Link>
        )}

        {post.image_path && (
          <img
            src={post.image_path}
            alt="Post image"
            className="w-full rounded-lg object-cover max-h-96"
          />
        )}

        <div className="flex items-center gap-4 text-sm text-gray-500">
          <span>{post.comment_count ?? 0} comments</span>
        </div>
      </CardContent>
    </Card>
  )
}
