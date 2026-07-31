'use client'

import { useState, useEffect, useCallback } from 'react'
import { useWebSocket } from '@/hooks/useWebSocket'
import { fetchUnreadMessageCount } from '@/api/chat'
import { useAuth } from '@/context/AuthProvider'

export function useMessageBadge() {
  const { is_authenticated } = useAuth()
  const { subscribe } = useWebSocket()
  const [message_count, set_message_count] = useState(0)

  const refresh = useCallback(async () => {
    if (!is_authenticated) return
    try {
      const count = await fetchUnreadMessageCount()
      set_message_count(count)
    } catch {
      // silently fail
    }
  }, [is_authenticated])

  useEffect(() => {
    if (is_authenticated) {
      refresh()
    }
  }, [is_authenticated, refresh])

  useEffect(() => {
    if (!is_authenticated) return

    const unsub = subscribe('chat_message', () => {
      set_message_count((prev) => prev + 1)
    })

    return unsub
  }, [is_authenticated, subscribe])

  return { message_count, refresh }
}
