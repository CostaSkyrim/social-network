'use client'

import Link from 'next/link'
import { Avatar } from '@/components/ui/Avatar'
import { format_date } from '@/lib/format'
import type { Comment } from '@/types/comment'

interface CommentItemProps {
  comment: Comment
  depth?: number
}

export function CommentItem({ comment, depth = 0 }: CommentItemProps) {
  const author_name = comment.author
    ? `${comment.author.first_name} ${comment.author.last_name}`
    : 'Unknown'

  const is_deleted = comment.is_deleted

  return (
    <div className={`flex items-start gap-3 ${depth > 0 ? 'ml-8' : ''}`}>
      <Link href={`/profile/${comment.author_id}`}>
        <Avatar
          src={comment.author?.avatar_path}
          alt={author_name}
          size="sm"
        />
      </Link>
      <div className="min-w-0 flex-1">
        <div className="rounded-xl bg-gray-100 px-4 py-2.5">
          <div className="flex items-center gap-2">
            <Link
              href={`/profile/${comment.author_id}`}
              className="text-sm font-medium text-gray-900 hover:underline"
            >
              {author_name}
            </Link>
            <span className="text-xs text-gray-500">{format_date(comment.created_at)}</span>
            {is_deleted && (
              <span className="text-xs text-red-400 font-medium">deleted</span>
            )}
          </div>
          {is_deleted ? (
            <p className="mt-0.5 text-sm text-gray-400 italic">[deleted]</p>
          ) : (
            <p className="mt-0.5 text-sm text-gray-700 whitespace-pre-wrap">{comment.content}</p>
          )}
        </div>
      </div>
    </div>
  )
}
