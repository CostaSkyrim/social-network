import client from './client'
import type { Notification } from '@/types/notification'

export async function fetchNotifications(
  limit = 20,
  offset = 0,
): Promise<Notification[]> {
  const res = await client.get('/api/notifications', {
    params: { limit, offset },
  })
  return res.data.data ?? []
}

export async function fetchUnreadCount(): Promise<number> {
  const res = await client.get('/api/notifications/unread-count')
  return res.data.data?.count ?? 0
}

export async function markNotificationRead(id: number): Promise<void> {
  await client.put(`/api/notifications/${id}/read`)
}

export async function markAllNotificationsRead(): Promise<void> {
  await client.put('/api/notifications/read-all')
}
