'use client'

import { useState, useEffect, useRef } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import { useWebSocket } from '@/hooks/useWebSocket'
import { fetchMessages, sendMessage as apiSendMessage } from '@/api/chat'
import client from '@/api/client'
import { Avatar } from '@/components/ui/Avatar'
import { cn } from '@/lib/cn'

interface MessageData {
  id: number
  uuid: string
  sender_id: number
  sender?: {
    first_name: string
    last_name: string
    nickname?: string
    avatar_path?: string
  }
  content: string
  is_read: boolean
  created_at: string
}

function formatTime(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

export default function ChatPage() {
  const params = useParams()
  const router = useRouter()
  const { user } = useAuth()
  const { subscribe } = useWebSocket()
  const targetUserID = params?.id as string

  const [messages, set_messages] = useState<MessageData[]>([])
  const [input, set_input] = useState('')
  const [is_loading, set_is_loading] = useState(true)
  const [partner, set_partner] = useState<{
    id: string
    first_name: string
    last_name: string
    nickname?: string
    avatar_path?: string
    is_online: boolean
  } | null>(null)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const dmIDRef = useRef<number | null>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    if (!targetUserID) return

    let cancelled = false
    async function load() {
      try {
        const dmsRes = await client.get('/api/chat/dms')
        const dms = dmsRes.data.data ?? []
        const dm = dms.find((d: any) => d.other_user?.id === targetUserID)
        if (dm) {
          dmIDRef.current = dm.id
          set_partner({
            id: dm.other_user.id,
            first_name: dm.other_user.first_name,
            last_name: dm.other_user.last_name,
            nickname: dm.other_user.nickname,
            avatar_path: dm.other_user.avatar_path,
            is_online: dm.other_user.is_online ?? false,
          })
          const msgs = await fetchMessages(dm.id)
          if (!cancelled) {
            set_messages(msgs)
            scrollToBottom()
          }
        } else {
          const userRes = await client.get(`/api/users/${targetUserID}`)
          const other = userRes.data.data
          if (other) {
            set_partner({
              id: other.id,
              first_name: other.first_name,
              last_name: other.last_name,
              nickname: other.nickname,
              avatar_path: other.avatar_path,
              is_online: other.is_online ?? false,
            })
          }
        }
      } catch {
        // silently fail
      } finally {
        if (!cancelled) set_is_loading(false)
      }
    }
    load()
    return () => { cancelled = true }
  }, [targetUserID])

  useEffect(() => {
    if (!targetUserID) return

    const unsub = subscribe('chat_message', (msg) => {
      const payload = msg.payload as {
        message_id: number
        dm_id: number
        content: string
        sender: {
          id?: number
          first_name: string
          last_name: string
          nickname?: string
          avatar_path?: string
        }
        created_at: string
      }
      if (!payload || payload.dm_id !== dmIDRef.current) return

      set_messages((prev) => [
        ...prev,
        {
          id: payload.message_id,
          uuid: crypto.randomUUID(),
          sender_id: msg.sender_id ?? 0,
          sender: payload.sender,
          content: payload.content,
          is_read: false,
          created_at: payload.created_at,
        },
      ])
      setTimeout(scrollToBottom, 100)
    })

    return unsub
  }, [targetUserID, subscribe])

  useEffect(() => {
    if (!partner) return
    const unsub = subscribe('presence_update', (msg) => {
      const payload = msg.payload as { user_id: number; is_online: boolean } | undefined
      if (!payload) return
      if (String(payload.user_id) === partner.id) {
        set_partner((prev) => (prev ? { ...prev, is_online: payload.is_online } : null))
      }
    })
    return unsub
  }, [partner?.id, subscribe])

  const handleSend = async () => {
    if (!input.trim() || !targetUserID) return

    const content = input.trim()
    set_input('')

    const tempID = Date.now()
    set_messages((prev) => [
      ...prev,
      {
        id: tempID,
        uuid: crypto.randomUUID(),
        sender_id: user?.id ? Number(user.id) : 0,
        sender: {
          first_name: user?.first_name ?? '',
          last_name: user?.last_name ?? '',
          nickname: user?.nickname,
          avatar_path: user?.avatar_path,
        },
        content,
        is_read: true,
        created_at: new Date().toISOString(),
      },
    ])
    setTimeout(scrollToBottom, 100)

    try {
      const result = await apiSendMessage(targetUserID, content)
      dmIDRef.current = result.dm_id
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

  if (!targetUserID) {
    return null
  }

  return (
    <div className="flex h-[calc(100vh-7rem)] flex-col">
      <div className="flex items-center gap-3 border-b border-gray-200 pb-3">
        <button
          onClick={() => router.back()}
          className="rounded-lg p-1 text-gray-500 hover:bg-gray-100"
        >
          <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        {partner && (
          <div className="flex items-center gap-2">
            <div className="relative">
              <Avatar
                src={partner.avatar_path}
                alt={`${partner.first_name} ${partner.last_name}`}
                size="sm"
              />
              {partner.is_online && (
                <span className="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-white bg-green-500" />
              )}
            </div>
            <div>
              <p className="text-sm font-medium text-gray-900">
                {partner.first_name} {partner.last_name}
              </p>
              <p className="text-xs text-gray-500">
                {partner.is_online ? 'Online' : 'Offline'}
              </p>
            </div>
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto py-3 space-y-2">
        {is_loading ? (
          <div className="flex items-center justify-center py-8">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
          </div>
        ) : (
          messages.map((msg) => {
            const isMine = String(msg.sender_id) === String(user?.id ?? '')
            return (
              <div
                key={msg.id}
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
                <div
                  className={cn(
                    'max-w-[75%] rounded-lg px-3 py-2',
                    isMine
                      ? 'bg-blue-500 text-white'
                      : 'bg-gray-100 text-gray-900',
                  )}
                >
                  <p className="text-sm whitespace-pre-wrap break-words">
                    {msg.content}
                  </p>
                  <p
                    className={cn(
                      'mt-0.5 text-right text-[10px]',
                      isMine ? 'text-blue-200' : 'text-gray-400',
                    )}
                  >
                    {formatTime(msg.created_at)}
                  </p>
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

      <div className="flex items-center gap-2 border-t border-gray-200 pt-3">
        <input
          ref={inputRef}
          type="text"
          value={input}
          onChange={(e) => set_input(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Type a message..."
          className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
        <button
          onClick={handleSend}
          disabled={!input.trim()}
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
  )
}
