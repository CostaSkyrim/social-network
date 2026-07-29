'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Avatar } from '@/components/ui/Avatar'
import { Button } from '@/components/ui/Button'
import { format_date } from '@/lib/format'
import { useAuth } from '@/context/AuthProvider'
import { useEditComment, useDeleteComment } from '@/hooks/useComments'
import type { Comment } from '@/types/comment'

interface CommentItemProps {
  comment: Comment
  depth?: number
}

export function CommentItem({ comment, depth = 0 }: CommentItemProps) {
  const { user } = useAuth()
  const edit_mutation = useEditComment()
  const delete_mutation = useDeleteComment()

  const [is_editing, set_editing] = useState(false)
  const [edit_content, set_edit_content] = useState(comment.content ?? '')

  const author_name = comment.author
    ? `${comment.author.first_name} ${comment.author.last_name}`
    : 'Unknown'

  const is_owner = user && String(comment.author_id) === user.id

  const handle_edit = async () => {
    await edit_mutation.mutateAsync({
      id: comment.id,
      content: edit_content,
    })
    set_editing(false)
  }

  const handle_delete = () => {
    if (confirm('Delete this comment?')) {
      delete_mutation.mutate(comment.id)
    }
  }

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
            {is_owner && !is_editing && (
              <div className="ml-auto flex items-center gap-1">
                <button
                  onClick={() => set_editing(true)}
                  className="rounded p-1 text-gray-400 hover:bg-gray-200 hover:text-gray-600"
                  title="Edit comment"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </button>
                <button
                  onClick={handle_delete}
                  className="rounded p-1 text-gray-400 hover:bg-gray-200 hover:text-red-500"
                  title="Delete comment"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            )}
          </div>
          {is_editing ? (
            <div className="mt-2 space-y-2">
              <textarea
                value={edit_content}
                onChange={(e) => set_edit_content(e.target.value)}
                className="w-full rounded-lg border border-gray-300 p-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                rows={2}
                disabled={edit_mutation.isPending}
              />
              <div className="flex items-center justify-end gap-2">
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
          ) : (
            <p className="mt-0.5 text-sm text-gray-700 whitespace-pre-wrap">{comment.content}</p>
          )}
        </div>
      </div>
    </div>
  )
}
