import client from './client'
import type { SearchUser } from '@/types/user'

interface ListResponse<T> {
  message: string
  data?: T[]
}

export async function searchUsers(q: string): Promise<SearchUser[]> {
  const res = await client.get<ListResponse<SearchUser>>('/api/users/search', {
    params: { q },
  })
  return res.data.data ?? []
}

export async function followUser(userId: string): Promise<void> {
  await client.post('/api/follow/request', { user_id: userId })
}

export async function unfollowUser(userId: string): Promise<void> {
  await client.post('/api/follow/remove', { user_id: userId })
}
