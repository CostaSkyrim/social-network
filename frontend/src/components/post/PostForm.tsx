'use client'

import { useState, useRef, useCallback } from 'react'
import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { PrivacySelector } from './PrivacySelector'
import { EmojiPicker } from '@/components/ui/EmojiPicker'
import { EmojiSuggestions } from '@/components/ui/EmojiSuggestions'
import { useEmojiAutocomplete } from '@/hooks/useEmojiAutocomplete'
import { useCreatePost } from '@/hooks/usePosts'
import { useAuth } from '@/context/AuthProvider'
import { Avatar } from '@/components/ui/Avatar'

export function PostForm() {
  const [content, set_content] = useState('')
  const [privacy, set_privacy] = useState('public')
  const [cursor_pos, set_cursor_pos] = useState(0)
  const create_post = useCreatePost()
  const { user } = useAuth()
  const textarea_ref = useRef<HTMLTextAreaElement>(null)

  const { word, start, matches, selected_index, set_selected_index, reset } =
    useEmojiAutocomplete(content, cursor_pos)

  const show_suggestions = matches.length > 0 && word !== null

  function insert_at_cursor(text: string) {
    const ta = textarea_ref.current
    if (!ta) return
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

  function handle_emoji(emoji: string) {
    insert_at_cursor(emoji)
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
    if (!content.trim()) return
    await create_post.mutateAsync({ content: content.trim(), privacy_level: privacy })
    set_content('')
    set_privacy('public')
  }

  return (
    <Card>
      <CardContent>
        <form onSubmit={handle_submit} className="space-y-3">
          <div className="flex items-start gap-3">
            <Avatar
              src={user?.avatar_path}
              alt={`${user?.first_name} ${user?.last_name}`}
              size="md"
            />
            <div className="relative flex-1">
              <textarea
                ref={textarea_ref}
                value={content}
                onChange={(e) => set_content(e.target.value)}
                onKeyDown={handle_keydown}
                onKeyUp={track_cursor}
                onClick={track_cursor}
                onMouseUp={track_cursor}
                placeholder="What's on your mind?"
                rows={3}
                className="w-full resize-none rounded-lg border border-gray-300 p-3 text-sm outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-200"
                maxLength={7000}
              />
              {show_suggestions && (
                <EmojiSuggestions
                  matches={matches}
                  selected_index={selected_index}
                  on_select={replace_word}
                />
              )}
            </div>
          </div>

          <div className="flex items-center justify-between">
            <EmojiPicker on_select={handle_emoji} />
            <PrivacySelector value={privacy} onChange={set_privacy} />
            <Button
              type="submit"
              size="sm"
              loading={create_post.isPending}
              disabled={!content.trim()}
            >
              Post
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
