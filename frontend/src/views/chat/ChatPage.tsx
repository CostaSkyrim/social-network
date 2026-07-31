'use client'

import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import { useWebSocket } from '@/hooks/useWebSocket'
import { fetchMessages, fetchDMs, sendMessage as apiSendMessage } from '@/api/chat'
import { Avatar } from '@/components/ui/Avatar'
import { cn } from '@/lib/cn'
import type { User } from '@/types/user'

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

interface DMItem {
  id: number
  other_user: User
  last_message?: string
  last_message_at?: string
  unread_count: number
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
  const targetUserID = params?.id as string | undefined

  const [dms, set_dms] = useState<DMItem[]>([])
  const [messages, set_messages] = useState<MessageData[]>([])
  const [input, set_input] = useState('')
  const [is_loading_dms, set_is_loading_dms] = useState(true)
  const [is_loading_msgs, set_is_loading_msgs] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const data = await fetchDMs()
        if (!cancelled) set_dms(data)
      } catch {
        // silently fail
      } finally {
        if (!cancelled) set_is_loading_dms(false)
      }
    }
    load()
    return () => { cancelled = true }
  }, [])

  useEffect(() => {
    if (!targetUserID || !dms.length) return

    const targetID = Number(targetUserID)
    const dm = dms.find((d) => d.other_user.id === targetID)
    if (!dm) return

    let cancelled = false
    set_is_loading_msgs(true)
    async function load() {
      try {
        const msgs = await fetchMessages(dm.id)
        if (!cancelled) {
          set_messages(msgs)
          scrollToBottom()
        }
      } catch {
        // silently fail
      } finally {
        if (!cancelled) set_is_loading_msgs(false)
      }
    }
    load()
    return () => { cancelled = true }
  }, [targetUserID, dms])

  useEffect(() => {
    if (!targetUserID) return

    const unsub = subscribe('chat_message', (msg) => {
      const payload = msg.payload as {
        message_id: number
        dm_id: number
        content: string
        sender: {
          first_name: string
          last_name: string
          nickname?: string
          avatar_path?: string
        }
        created_at: string
      }
      if (!payload) return

      const currentDM = dms.find(
        (d) => d.other_user.id === targetUserID,
      )

      if (currentDM && payload.dm_id === currentDM.id) {
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
      }

      set_dms((prev) => {
        const idx = prev.findIndex((d) => d.id === payload.dm_id)
        if (idx === -1) return prev
        const updated = [...prev]
        updated[idx] = {
          ...updated[idx],
          last_message: payload.content,
          last_message_at: payload.created_at,
          unread_count:
            currentDM?.id === payload.dm_id
              ? 0
              : updated[idx].unread_count + 1,
        }
        return updated
      })
    })

    return unsub
  }, [targetUserID, dms, subscribe])

  useEffect(() => {
    const unsub = subscribe('presence_update', (msg) => {
      const payload = msg.payload as { user_id: number; is_online: boolean } | undefined
      if (!payload) return
      set_dms((prev) =>
        prev.map((d) =>
          d.other_user.id === String(payload.user_id)
            ? { ...d, other_user: { ...d.other_user, is_online: payload.is_online } }
            : d,
        ),
      )
    })
    return unsub
  }, [subscribe])

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
      const result = await apiSendMessage(Number(targetUserID), content)
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

  const selectedDM = dms.find((d) => d.other_user.id === targetUserID)
  const otherUser = selectedDM?.other_user

  if (!targetUserID) {
    return (
      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-gray-900">Messages</h2>
        {is_loading_dms ? (
          <div className="divide-y divide-gray-100 rounded-lg border border-gray-200 bg-white">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 px-4 py-3">
                <div className="h-10 w-10 animate-pulse rounded-full bg-gray-200" />
                <div className="flex-1 space-y-1">
                  <div className="h-4 w-32 animate-pulse rounded bg-gray-200" />
                  <div className="h-3 w-48 animate-pulse rounded bg-gray-200" />
                </div>
              </div>
            ))}
          </div>
        ) : dms.length === 0 ? (
          <div className="rounded-lg border border-gray-200 bg-white p-8 text-center">
            <p className="text-4xl">💬</p>
            <p className="mt-2 text-sm text-gray-500">
              No conversations yet. Go to your followers page to start chatting.
            </p>
          </div>
        ) : (
          <div className="divide-y divide-gray-100 rounded-lg border border-gray-200 bg-white">
            {dms
              .filter((d) => d.other_user)
              .map((dm) => (
                <button
                  key={dm.id}
                  onClick={() => router.push(`/chat/${dm.other_user.id}`)}
                  className="flex w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-gray-50"
                >
                  <div className="relative">
                    <Avatar
                      src={dm.other_user.avatar_path}
                      alt={`${dm.other_user.first_name} ${dm.other_user.last_name}`}
                      size="sm"
                    />
                    {dm.other_user.is_online && (
                      <span className="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-white bg-green-500" />
                    )}
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium text-gray-900">
                      {dm.other_user.first_name} {dm.other_user.last_name}
                    </p>
                    <p className="truncate text-xs text-gray-500">
                      {dm.last_message ?? 'No messages yet'}
                    </p>
                  </div>
                  {dm.unread_count > 0 && (
                    <span className="flex h-5 min-w-5 items-center justify-center rounded-full bg-blue-500 px-1.5 text-[11px] font-bold text-white">
                      {dm.unread_count}
                    </span>
                  )}
                </button>
              ))}
          </div>
        )}
      </div>
    )
  }

  return (
    <div className="flex h-[calc(100vh-7rem)] flex-col">
      <div className="flex items-center gap-3 border-b border-gray-200 pb-3">
        <button
          onClick={() => router.push('/chat')}
          className="rounded-lg p-1 text-gray-500 hover:bg-gray-100"
        >
          <svg className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
          </svg>
        </button>
        {otherUser && (
          <div className="flex items-center gap-2">
            <div className="relative">
              <Avatar
                src={otherUser.avatar_path}
                alt={`${otherUser.first_name} ${otherUser.last_name}`}
                size="sm"
              />
              {otherUser.is_online && (
                <span className="absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-white bg-green-500" />
              )}
            </div>
            <div>
              <p className="text-sm font-medium text-gray-900">
                {otherUser.first_name} {otherUser.last_name}
              </p>
              <p className="text-xs text-gray-500">
                {otherUser.is_online ? 'Online' : 'Offline'}
              </p>
            </div>
          </div>
        )}
      </div>

      <div className="flex-1 overflow-y-auto py-3 space-y-2">
        {is_loading_msgs ? (
          <div className="flex items-center justify-center py-8">
            <div className="h-6 w-6 animate-spin rounded-full border-2 border-blue-500 border-t-transparent" />
          </div>
        ) : (
          messages.map((msg) => {
            const isMine = msg.sender_id === Number(user?.id ?? 0)
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
