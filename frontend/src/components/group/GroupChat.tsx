'use client'

import { useState, useEffect, useRef } from 'react'
import { useAuth } from '@/context/AuthProvider'
import { useWebSocket } from '@/hooks/useWebSocket'
import { fetchGroupMessages, sendGroupMessage as apiSendGroupMessage } from '@/api/groups'
import { Avatar } from '@/components/ui/Avatar'
import { ImageUpload } from '@/components/common/ImageUpload'
import { EmojiPicker } from '@/components/ui/EmojiPicker'
import { EmojiSuggestions } from '@/components/ui/EmojiSuggestions'
import { useEmojiAutocomplete } from '@/hooks/useEmojiAutocomplete'
import { cn } from '@/lib/cn'
import { get_media_url } from '@/lib/media'

interface GroupMessageData {
  id: number
  uuid: string
  sender_id: number
  sender?: {
    id?: string
    first_name: string
    last_name: string
    nickname?: string
    avatar_path?: string
  }
  content: string
  image_path?: string
  is_read: boolean
  created_at: string
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

interface GroupChatProps {
  groupId: string
  groupTitle: string
  groupAvatar?: string
}

export function GroupChat({ groupId, groupTitle, groupAvatar }: GroupChatProps) {
  const { user } = useAuth()
  const { subscribe } = useWebSocket()

  const [messages, set_messages] = useState<GroupMessageData[]>([])
  const [input, set_input] = useState('')
  const [pending_image, set_pending_image] = useState<File | null>(null)
  const [pending_preview, set_pending_preview] = useState<string | null>(null)
  const [is_loading, set_is_loading] = useState(true)
  const [cursor_pos, set_cursor_pos] = useState(0)
  const listRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)

  const { word, start, matches, selected_index, set_selected_index, reset } =
    useEmojiAutocomplete(input, cursor_pos)

  const show_suggestions = matches.length > 0 && word !== null

  const scrollToBottom = () => {
    const list = listRef.current
    if (list) {
      list.scrollTop = list.scrollHeight
    }
  }

  useEffect(() => {
    if (!groupId) return
    let cancelled = false

    async function load() {
      try {
        const msgs = await fetchGroupMessages(groupId)
        if (!cancelled) {
          set_messages(msgs)
          scrollToBottom()
          window.dispatchEvent(new Event('messages-read'))
        }
      } catch {
        // silently fail
      } finally {
        if (!cancelled) set_is_loading(false)
      }
    }
    load()
    return () => {
      cancelled = true
    }
  }, [groupId])

  useEffect(() => {
    if (!groupId) return

    const unsub = subscribe('group_message', (msg) => {
      const payload = msg.payload as {
        message_id: number
        group_id: number
        group_uuid?: string
        content: string
        image_path?: string
        sender: {
          id?: string
          first_name: string
          last_name: string
          nickname?: string
          avatar_path?: string
        }
        created_at: string
      }
      const matchesGroup =
        payload?.group_uuid === groupId || String(payload?.group_id) === String(groupId)
      if (!payload || !matchesGroup) return

      const new_msg: GroupMessageData = {
        id: payload.message_id,
        uuid: crypto.randomUUID(),
        sender_id: msg.sender_id ?? 0,
        sender: payload.sender,
        content: payload.content,
        image_path: payload.image_path,
        is_read: false,
        created_at: payload.created_at,
      }

      set_messages((prev) =>
        prev.some((m) => m.id === payload.message_id) ? prev : [...prev, new_msg],
      )
      setTimeout(scrollToBottom, 100)
    })

    return unsub
  }, [groupId, subscribe])

  const clear_pending_image = () => {
    if (pending_preview) URL.revokeObjectURL(pending_preview)
    set_pending_image(null)
    set_pending_preview(null)
  }

  const handle_select_image = (file: File) => {
    if (pending_preview) URL.revokeObjectURL(pending_preview)
    set_pending_image(file)
    set_pending_preview(URL.createObjectURL(file))
  }

  const handleSend = async () => {
    const content = input.trim()
    if ((!content && !pending_image) || !groupId) return

    set_input('')
    const image = pending_image
    const image_preview = pending_preview
    clear_pending_image()

    const tempID = Date.now()
    set_messages((prev) => [
      ...prev,
      {
        id: tempID,
        uuid: crypto.randomUUID(),
        sender_id: user?.id ? Number(user.id) : 0,
        sender: {
          id: user?.id,
          first_name: user?.first_name ?? '',
          last_name: user?.last_name ?? '',
          nickname: user?.nickname,
          avatar_path: user?.avatar_path,
        },
        content,
        image_path: image ? (image_preview ?? undefined) : undefined,
        is_read: true,
        created_at: new Date().toISOString(),
      },
    ])
    setTimeout(scrollToBottom, 100)

    try {
      const result = await apiSendGroupMessage(groupId, content, image ?? undefined)
      set_messages((prev) =>
        prev.map((m) => (m.id === tempID ? { ...m, id: result.id } : m)),
      )
    } catch {
      set_messages((prev) => prev.filter((m) => m.id !== tempID))
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (show_suggestions) {
      if (e.key === 'ArrowDown') {
        e.preventDefault()
        set_selected_index((selected_index + 1) % matches.length)
        return
      } else if (e.key === 'ArrowUp') {
        e.preventDefault()
        set_selected_index((selected_index - 1 + matches.length) % matches.length)
        return
      } else if (e.key === 'Enter' || e.key === 'Tab') {
        e.preventDefault()
        replace_word(matches[selected_index].native)
        return
      } else if (e.key === 'Escape') {
        e.preventDefault()
        reset()
        return
      }
    }

    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  function track_cursor() {
    set_cursor_pos(inputRef.current?.selectionStart ?? 0)
  }

  function replace_word(emoji: string) {
    if (start === -1 || !word) return
    const new_content = input.slice(0, start) + emoji + input.slice(start + word.length + 1)
    set_input(new_content)
    const new_pos = start + emoji.length
    setTimeout(() => {
      inputRef.current?.setSelectionRange(new_pos, new_pos)
      set_cursor_pos(new_pos)
    }, 0)
  }

  function insert_at_cursor(text: string) {
    const ta = inputRef.current
    if (!ta) {
      set_input((prev) => prev + text)
      return
    }
    const pos = ta.selectionStart
    const new_content = input.slice(0, pos) + text + input.slice(ta.selectionEnd)
    set_input(new_content)
    const new_pos = pos + text.length
    setTimeout(() => {
      ta.setSelectionRange(new_pos, new_pos)
      set_cursor_pos(new_pos)
    }, 0)
  }

  return (
    <div className="flex h-[calc(100vh-16rem)] min-h-[400px] flex-col">
      <div className="flex items-center gap-2 border-b border-purple-400/20 pb-3">
        <Avatar src={groupAvatar} alt={groupTitle} size="sm" />
        <div>
          <p className="text-sm font-medium text-gray-100">{groupTitle}</p>
          <p className="text-xs text-gray-300">Group chat</p>
        </div>
      </div>

      <div ref={listRef} className="flex-1 space-y-2 overflow-y-auto py-3">
        {is_loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-violet-500 border-t-transparent" />
          </div>
        ) : messages.length === 0 ? (
          <p className="py-8 text-center text-sm text-gray-300">
            No messages yet. Start the conversation!
          </p>
        ) : (
          messages.map((msg) => {
            const isMine =
              (msg.sender?.id ?? String(msg.sender_id)) === String(user?.id ?? '')
            return (
              <div
                key={msg.uuid}
                className={cn('flex', isMine ? 'justify-end' : 'justify-start')}
              >
                {!isMine && (
                  <Avatar
                    src={msg.sender?.avatar_path}
                    alt={msg.sender?.first_name ?? ''}
                    size="sm"
                    className="mr-2 mt-1 flex-shrink-0"
                  />
                )}
                <div className={cn('max-w-[75%]', isMine ? 'items-end' : 'items-start')}>
                  {!isMine && (
                    <p className="mb-0.5 px-1 text-[11px] text-gray-300">
                      {msg.sender?.nickname ||
                        `${msg.sender?.first_name ?? ''} ${msg.sender?.last_name ?? ''}`.trim()}
                    </p>
                  )}
                  <div
                    className={cn(
                      'max-w-[75%] rounded-lg px-3 py-2',
                      isMine ? 'ml-auto bg-violet-500 text-white' : 'bg-purple-400/15 text-gray-100',
                    )}
                  >
                    {msg.image_path && (
                      /* eslint-disable-next-line @next/next/no-img-element */
                      <img
                        src={get_media_url(msg.image_path)}
                        alt="Message image"
                        className="mb-1 max-h-60 w-full rounded-md object-cover"
                      />
                    )}
                    {msg.content && (
                      <p className="whitespace-pre-wrap break-words text-sm">{msg.content}</p>
                    )}
                    <p
                      className={cn(
                        'mt-0.5 text-right text-[10px]',
                        isMine ? 'text-violet-200' : 'text-gray-300',
                      )}
                    >
                      {formatTime(msg.created_at)}
                    </p>
                  </div>
                </div>
                {isMine && (
                  <Avatar
                    src={user?.avatar_path}
                    alt={user?.first_name ?? ''}
                    size="sm"
                    className="ml-2 mt-1 flex-shrink-0"
                  />
                )}
              </div>
            )
          })
        )}
      </div>

      <div className="border-t border-purple-400/20 pt-3">
        {pending_preview && (
          <div className="relative mb-2">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={pending_preview}
              alt="Selected image"
              className="max-h-32 w-48 rounded-lg object-cover"
            />
            <button
              type="button"
              onClick={clear_pending_image}
              className="absolute right-2 top-2 rounded-full bg-black/60 px-2 py-1 text-xs text-white hover:bg-black/80"
            >
              Remove
            </button>
          </div>
        )}
        <div className="flex items-end gap-2">
          <ImageUpload on_select={handle_select_image}>
            <button
              type="button"
              className="rounded-lg p-2 text-gray-300 hover:bg-purple-400/15 hover:text-gray-300"
              title="Attach image"
            >
              <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z"
                />
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
              </svg>
            </button>
          </ImageUpload>
          <EmojiPicker on_select={insert_at_cursor} />
          <div className="relative flex-1">
            <textarea
              ref={inputRef}
              value={input}
              onChange={(e) => set_input(e.target.value)}
              onKeyDown={handleKeyDown}
              onKeyUp={track_cursor}
              onClick={track_cursor}
              placeholder="Type a message..."
              rows={1}
              className="max-h-32 min-h-[2.5rem] w-full resize-none rounded-lg border border-purple-400/30 px-3 py-2 text-sm focus:border-violet-500 focus:outline-none focus:ring-1 focus:ring-violet-500"
            />
            {show_suggestions && (
              <EmojiSuggestions
                matches={matches}
                selected_index={selected_index}
                on_select={replace_word}
              />
            )}
          </div>
          <button
            onClick={handleSend}
            disabled={!input.trim() && !pending_image}
            className="rounded-lg bg-violet-500 p-2 text-white hover:bg-violet-600 disabled:opacity-50"
          >
            <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"
              />
            </svg>
          </button>
        </div>
      </div>
    </div>
  )
}
