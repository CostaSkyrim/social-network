'use client'

import { useState, useEffect, useRef } from 'react'
import { useAuth } from '@/context/AuthProvider'
import { useWebSocket } from '@/hooks/useWebSocket'
import { fetchGroupMessages, sendGroupMessage as apiSendGroupMessage } from '@/api/groups'
import { Avatar } from '@/components/ui/Avatar'
import { ImageUpload } from '@/components/common/ImageUpload'
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
  const messagesEndRef = useRef<HTMLDivElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
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
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex h-[calc(100vh-16rem)] min-h-[400px] flex-col">
      <div className="flex items-center gap-2 border-b border-gray-200 pb-3">
        <Avatar src={groupAvatar} alt={groupTitle} size="sm" />
        <div>
          <p className="text-sm font-medium text-gray-900">{groupTitle}</p>
          <p className="text-xs text-gray-500">Group chat</p>
        </div>
      </div>

      <div className="flex-1 space-y-2 overflow-y-auto py-3">
        {is_loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
          </div>
        ) : messages.length === 0 ? (
          <p className="py-8 text-center text-sm text-gray-400">
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
                    <p className="mb-0.5 px-1 text-[11px] text-gray-500">
                      {msg.sender?.nickname ||
                        `${msg.sender?.first_name ?? ''} ${msg.sender?.last_name ?? ''}`.trim()}
                    </p>
                  )}
                  <div
                    className={cn(
                      'max-w-[75%] rounded-lg px-3 py-2',
                      isMine ? 'ml-auto bg-blue-500 text-white' : 'bg-gray-100 text-gray-900',
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
                        isMine ? 'text-blue-200' : 'text-gray-400',
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
        <div ref={messagesEndRef} />
      </div>

      <div className="border-t border-gray-200 pt-3">
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
        <div className="flex items-center gap-2">
          <ImageUpload on_select={handle_select_image}>
            <button
              type="button"
              className="rounded-lg p-2 text-gray-500 hover:bg-gray-100 hover:text-gray-700"
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
          <input
            type="text"
            value={input}
            onChange={(e) => set_input(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Type a message..."
            className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          />
          <button
            onClick={handleSend}
            disabled={!input.trim() && !pending_image}
            className="rounded-lg bg-blue-500 p-2 text-white hover:bg-blue-600 disabled:opacity-50"
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
