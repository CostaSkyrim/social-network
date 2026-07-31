'use client'

import {
  createContext,
  useContext,
  useState,
  useEffect,
  useCallback,
  type ReactNode,
} from 'react'
import { useWebSocket } from '@/hooks/useWebSocket'
import {
  fetchUnreadCount,
  fetchNotifications,
  markNotificationRead as apiMarkRead,
  markAllNotificationsRead as apiMarkAllRead,
} from '@/api/notifications'
import { useAuth } from '@/context/AuthProvider'
import type { Notification } from '@/types/notification'

interface NotificationContextValue {
  notifications: Notification[]
  unread_count: number
  is_loading: boolean
  mark_read: (id: number) => Promise<void>
  mark_all_read: () => Promise<void>
  refresh: () => Promise<void>
}

const NotificationContext = createContext<NotificationContextValue | null>(null)

const MAX_NOTIFICATIONS = 50

export function NotificationProvider({ children }: { children: ReactNode }) {
  const { is_authenticated } = useAuth()
  const { subscribe } = useWebSocket()
  const [notifications, set_notifications] = useState<Notification[]>([])
  const [unread_count, set_unread_count] = useState(0)
  const [is_loading, set_is_loading] = useState(true)

  const refresh = useCallback(async () => {
    if (!is_authenticated) return
    try {
      const [notifs, count] = await Promise.all([
        fetchNotifications(MAX_NOTIFICATIONS, 0),
        fetchUnreadCount(),
      ])
      set_notifications(notifs)
      set_unread_count(count)
    } catch {
      // silently fail
    } finally {
      set_is_loading(false)
    }
  }, [is_authenticated])

  useEffect(() => {
    if (is_authenticated) {
      refresh()
    }
  }, [is_authenticated, refresh])

  useEffect(() => {
    if (!is_authenticated) return

    const unsub = subscribe('notification', (msg) => {
      const payload = msg.payload as
        | {
            id: number
            type: string
            content: string
            related_id?: number
            from_user_id?: number
            is_read: boolean
            created_at: string
          }
        | undefined
      if (!payload) return

      const newNotif: Notification = {
        id: payload.id,
        type: payload.type,
        content: payload.content,
        related_id: payload.related_id,
        from_user_id: payload.from_user_id
          ? String(payload.from_user_id)
          : undefined,
        is_read: payload.is_read,
        created_at: payload.created_at,
      }

      set_notifications((prev) => [newNotif, ...prev].slice(0, MAX_NOTIFICATIONS))
      set_unread_count((prev) => prev + 1)
    })

    return unsub
  }, [is_authenticated, subscribe])

  const mark_read = useCallback(async (id: number) => {
    try {
      await apiMarkRead(id)
      set_notifications((prev) =>
        prev.map((n) => (n.id === id ? { ...n, is_read: true } : n)),
      )
      set_unread_count((prev) => Math.max(0, prev - 1))
    } catch {
      // silently fail
    }
  }, [])

  const mark_all_read = useCallback(async () => {
    try {
      await apiMarkAllRead()
      set_notifications((prev) => prev.map((n) => ({ ...n, is_read: true })))
      set_unread_count(0)
    } catch {
      // silently fail
    }
  }, [])

  return (
    <NotificationContext.Provider
      value={{
        notifications,
        unread_count,
        is_loading,
        mark_read,
        mark_all_read,
        refresh,
      }}
    >
      {children}
    </NotificationContext.Provider>
  )
}

export function useNotifications() {
  const ctx = useContext(NotificationContext)
  if (!ctx) {
    throw new Error(
      'useNotifications must be used within NotificationProvider',
    )
  }
  return ctx
}
