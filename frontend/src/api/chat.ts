import client from './client'
import type { DirectMessage } from '@/types/message'
import type { User } from '@/types/user'

interface FollowerWithDM extends User {
  unread_count: number
  last_dm_at?: string
}

interface Message {
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

export async function fetchDMs(): Promise<DMItem[]> {
  const res = await client.get('/api/chat/dms')
  return res.data.data ?? []
}

export async function fetchMessages(
  dmID: number,
  limit = 50,
): Promise<Message[]> {
  const res = await client.get(`/api/chat/dms/${dmID}/messages`, {
    params: { limit },
  })
  return res.data.data ?? []
}

export async function sendMessage(
  targetUserID: string,
  content: string,
): Promise<{ id: number; dm_id: number; content: string; sender_id: number }> {
  const res = await client.post(`/api/chat/send/${targetUserID}`, { content })
  return res.data.data
}

export async function fetchUnreadMessageCount(): Promise<number> {
  const res = await client.get('/api/chat/unread-count')
  return res.data.data?.count ?? 0
}

export async function fetchFollowers(
  userID: string,
): Promise<FollowerWithDM[]> {
  const res = await client.get(`/api/followers?user_id=${userID}`)
  return res.data.data ?? []
}

export async function fetchFollowing(
  userID: string,
): Promise<FollowerWithDM[]> {
  const res = await client.get(`/api/following?user_id=${userID}`)
  return res.data.data ?? []
}
