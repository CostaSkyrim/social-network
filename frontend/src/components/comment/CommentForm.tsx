'use client'

import { useState, useCallback } from 'react'
import { Button } from '@/components/ui/Button'
import { Avatar } from '@/components/ui/Avatar'
import { ImageUpload } from '@/components/common/ImageUpload'
import { useCreateComment } from '@/hooks/useComments'
import { useAuth } from '@/context/AuthProvider'

interface CommentFormProps {
  postId: string
  parentCommentId?: string
  onSubmitted?: () => void
}

export function CommentForm({ postId, parentCommentId, onSubmitted }: CommentFormProps) {
  const [content, set_content] = useState('')
  const [image, set_image] = useState<File | null>(null)
  const [preview_url, set_preview_url] = useState<string | null>(null)
  const create_comment = useCreateComment()
  const { user } = useAuth()

  const clear_image = useCallback(() => {
    if (preview_url) URL.revokeObjectURL(preview_url)
    set_image(null)
    set_preview_url(null)
  }, [preview_url])

  async function handle_submit(e: React.FormEvent) {
    e.preventDefault()
    if (!content.trim() && !image) return
    await create_comment.mutateAsync({
      post_id: postId,
      parent_comment_id: parentCommentId,
      content: content.trim(),
      image: image ?? undefined,
    })
    set_content('')
    clear_image()
    onSubmitted?.()
  }

  return (
    <form onSubmit={handle_submit} className="flex items-start gap-3">
      <Avatar src={user?.avatar_path} alt={`${user?.first_name} ${user?.last_name}`} size="sm" />
      <div className="flex-1 space-y-2">
        <textarea
          value={content}
          onChange={(e) => set_content(e.target.value)}
          placeholder="Write a comment..."
          rows={2}
          maxLength={2000}
          className="w-full resize-none rounded-lg border border-gray-300 p-2.5 text-sm outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-200"
        />
        {preview_url && (
          <div className="relative">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={preview_url}
              alt="Selected image"
              className="max-h-40 w-full rounded-lg object-cover"
            />
            <button
              type="button"
              onClick={clear_image}
              className="absolute right-2 top-2 rounded-full bg-black/60 px-2 py-1 text-xs text-white hover:bg-black/80"
            >
              Remove
            </button>
          </div>
        )}
        <div className="flex items-center justify-between">
          <ImageUpload on_select={(file) => {
            if (preview_url) URL.revokeObjectURL(preview_url)
            set_image(file)
            set_preview_url(URL.createObjectURL(file))
          }}>
            <button
              type="button"
              className="rounded-lg px-2.5 py-1.5 text-sm text-gray-500 hover:bg-gray-100 hover:text-gray-700"
              title="Add image"
            >
              📷
            </button>
          </ImageUpload>
          <Button
            type="submit"
            size="sm"
            loading={create_comment.isPending}
            disabled={!content.trim() && !image}
          >
            Comment
          </Button>
        </div>
      </div>
    </form>
  )
}
