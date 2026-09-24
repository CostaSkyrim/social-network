'use client'

import { useState, useCallback } from 'react'
import Link from 'next/link'
import { Avatar } from '@/components/ui/Avatar'
import { Button } from '@/components/ui/Button'
import { ImageUpload } from '@/components/common/ImageUpload'
import { format_date } from '@/lib/format'
import { get_media_url } from '@/lib/media'
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
  const [edit_image, set_edit_image] = useState<File | null>(null)
  const [edit_preview_url, set_edit_preview_url] = useState<string | null>(null)
  const [edit_remove_image, set_edit_remove_image] = useState(false)

  const author_name = comment.author
    ? `${comment.author.first_name} ${comment.author.last_name}`
    : 'Unknown'

  const is_deleted = comment.is_deleted
  const is_owner = user && String(comment.author_id) === String(user.id)

  const handle_edit = async () => {
    try {
      await edit_mutation.mutateAsync({
        id: comment.id,
        content: edit_content,
        image: edit_image ?? undefined,
        remove_image: edit_remove_image,
      })
      set_editing(false)
      set_edit_image(null)
      set_edit_remove_image(false)
      if (edit_preview_url) URL.revokeObjectURL(edit_preview_url)
      set_edit_preview_url(null)
    } catch {
      // error toast handled by mutation
    }
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
        <div className="rounded-xl bg-purple-400/15 px-4 py-2.5">
          <div className="flex items-center gap-2">
            <Link
              href={`/profile/${comment.author_id}`}
              className="text-sm font-medium text-gray-100 hover:underline"
            >
              {author_name}
            </Link>
            <span className="text-xs text-gray-300">{format_date(comment.created_at)}</span>
            {is_deleted && (
              <span className="text-xs text-red-400 font-medium">deleted</span>
            )}
            {!is_deleted && is_owner && !is_editing && (
              <div className="ml-auto flex items-center gap-1">
                <button
                  onClick={() => set_editing(true)}
                  className="rounded p-1 text-gray-300 hover:bg-purple-400/20 hover:text-gray-300"
                  title="Edit comment"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                  </svg>
                </button>
                <button
                  onClick={handle_delete}
                  className="rounded p-1 text-gray-300 hover:bg-purple-400/20 hover:text-red-500"
                  title="Delete comment"
                >
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            )}
          </div>
          {is_deleted ? (
            <p className="mt-0.5 text-sm text-gray-300 italic">[deleted]</p>
          ) : is_editing ? (
            <div className="mt-2 space-y-2">
              <textarea
                value={edit_content}
                onChange={(e) => set_edit_content(e.target.value)}
                className="w-full rounded-lg border border-purple-400/30 p-2 text-sm focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-500"
                rows={2}
                disabled={edit_mutation.isPending}
              />
              {(edit_preview_url || (comment.image_path && !edit_remove_image && !edit_image)) && (
                <div className="relative">
                  {/* eslint-disable-next-line @next/next/no-img-element */}
                  <img
                    src={edit_preview_url ?? get_media_url(comment.image_path)}
                    alt="Comment image"
                    className="max-h-48 w-full rounded-lg object-contain"
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
              <div className="flex items-center justify-between gap-2">
                <ImageUpload on_select={handle_edit_image}>
                  <button
                    type="button"
                    className="rounded-lg px-2 py-1.5 text-base text-gray-300 hover:bg-purple-400/15 hover:text-gray-100"
                    title={comment.image_path && !edit_remove_image ? 'Change image' : 'Add image'}
                  >
                    📷
                  </button>
                </ImageUpload>
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
                    disabled={edit_mutation.isPending || (edit_content.trim() === '' && !edit_image && !(comment.image_path && !edit_remove_image))}
                  >
                    Save
                  </Button>
                </div>
              </div>
            </div>
          ) : (
            <div>
              <p className="mt-0.5 text-sm text-gray-300 whitespace-pre-wrap">{comment.content}</p>
              {comment.image_path && (
                /* eslint-disable-next-line @next/next/no-img-element */
                <img
                  src={get_media_url(comment.image_path)}
                  alt="Comment image"
                  className="mt-2 max-h-48 w-full rounded-lg object-contain"
                />
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
