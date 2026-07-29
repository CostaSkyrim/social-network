import client from './client'
import type { Comment } from '@/types/comment'

interface CommentsResponse {
  message: string
  data?: Comment[]
}

export async function getComments(postId: number): Promise<Comment[]> {
  const res = await client.get<CommentsResponse>(`/api/posts/${postId}/comments`)
  return res.data.data ?? []
}

export async function createComment(data: { post_id: number; content: string; image_path?: string }): Promise<void> {
  await client.post('/api/comments', data)
}

export async function deleteComment(id: number): Promise<void> {
  await client.delete(`/api/comments/${id}`)
}

export async function editComment(id: number, data: { content: string; image_path?: string }): Promise<void> {
  await client.put(`/api/comments/${id}/edit`, data)
}
