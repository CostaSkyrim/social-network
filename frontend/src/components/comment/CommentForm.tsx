'use client'

import { useState, useRef, useCallback } from 'react'
import { Button } from '@/components/ui/Button'
import { Avatar } from '@/components/ui/Avatar'
import { ImageUpload } from '@/components/common/ImageUpload'
import { EmojiPicker } from '@/components/ui/EmojiPicker'
import { EmojiSuggestions } from '@/components/ui/EmojiSuggestions'
import { useEmojiAutocomplete } from '@/hooks/useEmojiAutocomplete'
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
  const [cursor_pos, set_cursor_pos] = useState(0)
  const create_comment = useCreateComment()
  const { user } = useAuth()
  const textarea_ref = useRef<HTMLTextAreaElement>(null)

  const { word, start, matches, selected_index, set_selected_index, reset } =
    useEmojiAutocomplete(content, cursor_pos)

  const show_suggestions = matches.length > 0 && word !== null

  const clear_image = useCallback(() => {
    if (preview_url) URL.revokeObjectURL(preview_url)
    set_image(null)
    set_preview_url(null)
  }, [preview_url])

  function insert_at_cursor(text: string) {
    const ta = textarea_ref.current
    if (!ta) {
      set_content((prev) => prev + text)
      return
    }
    const pos = ta.selectionStart
    const new_content = content.slice(0, pos) + text + content.slice(ta.selectionEnd)
    set_content(new_content)
    const new_pos = pos + text.length
    setTimeout(() => {
      ta.setSelectionRange(new_pos, new_pos)
      set_cursor_pos(new_pos)
    }, 0)
  }

  function replace_word(emoji: string) {
    if (start === -1 || !word) return
    const new_content = content.slice(0, start) + emoji + content.slice(start + word.length + 1)
    set_content(new_content)
    const new_pos = start + emoji.length
    setTimeout(() => {
      textarea_ref.current?.setSelectionRange(new_pos, new_pos)
      set_cursor_pos(new_pos)
    }, 0)
  }

  function handle_keydown(e: React.KeyboardEvent<HTMLTextAreaElement>) {
    if (!show_suggestions) return

    if (e.key === 'ArrowDown') {
      e.preventDefault()
      set_selected_index((selected_index + 1) % matches.length)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      set_selected_index((selected_index - 1 + matches.length) % matches.length)
    } else if (e.key === 'Enter' || e.key === 'Tab') {
      e.preventDefault()
      replace_word(matches[selected_index].native)
    } else if (e.key === 'Escape') {
      e.preventDefault()
      reset()
    }
  }

  function track_cursor() {
    set_cursor_pos(textarea_ref.current?.selectionStart ?? 0)
  }

  async function handle_submit(e: React.FormEvent) {
    e.preventDefault()
    if (!content.trim() && !image) return
    try {
      await create_comment.mutateAsync({
        post_id: postId,
        parent_comment_id: parentCommentId,
        content: content.trim(),
        image: image ?? undefined,
      })
      set_content('')
      clear_image()
      onSubmitted?.()
    } catch {
      // error toast handled by mutation
    }
  }

  return (
    <form onSubmit={handle_submit} className="flex items-start gap-3">
      <Avatar src={user?.avatar_path} alt={`${user?.first_name} ${user?.last_name}`} size="sm" />
      <div className="flex-1 space-y-2">
        <div className="relative">
          <textarea
            ref={textarea_ref}
            value={content}
            onChange={(e) => set_content(e.target.value)}
            onKeyDown={handle_keydown}
            onKeyUp={track_cursor}
            onClick={track_cursor}
            onMouseUp={track_cursor}
            placeholder="Write a comment..."
            rows={2}
            maxLength={2000}
            className="w-full resize-none rounded-lg border border-purple-400/30 p-2.5 text-sm outline-none focus:border-violet-500 focus:ring-1 focus:ring-violet-300/40"
          />
          {show_suggestions && (
            <EmojiSuggestions
              matches={matches}
              selected_index={selected_index}
              on_select={replace_word}
            />
          )}
        </div>
        {preview_url && (
          <div className="relative">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={preview_url}
              alt="Selected image"
              className="max-h-40 w-full rounded-lg object-contain"
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
          <div className="flex items-center gap-1">
            <EmojiPicker on_select={insert_at_cursor} />
            <ImageUpload on_select={(file) => {
              if (preview_url) URL.revokeObjectURL(preview_url)
              set_image(file)
              set_preview_url(URL.createObjectURL(file))
            }}>
              <button
                type="button"
                className="rounded-lg px-2.5 py-1.5 text-sm text-gray-300 hover:bg-purple-400/15 hover:text-gray-300"
                title="Add image"
              >
                📷
              </button>
            </ImageUpload>
          </div>
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
