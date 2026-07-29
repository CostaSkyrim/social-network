'use client'

import Link from 'next/link'
import { Avatar } from '@/components/ui/Avatar'
import { Card, CardContent } from '@/components/ui/Card'
import { format_date } from '@/lib/format'
import type { Post } from '@/types/post'

interface PostCardProps {
  post: Post
}

export function PostCard({ post }: PostCardProps) {
  const author_name = post.author
    ? `${post.author.first_name} ${post.author.last_name}`
    : 'Unknown'

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
        </div>

        <Link href={`/posts/${post.db_id}`}>
          <p className="text-sm text-gray-800 whitespace-pre-wrap">{post.content}</p>
        </Link>

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
