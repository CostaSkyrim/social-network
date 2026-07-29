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
