import client from './client'
import type { Comment } from '@/types/comment'

interface CommentsResponse {
  message: string
  data?: Comment[]
}

export async function getComments(postId: string): Promise<Comment[]> {
  const res = await client.get<CommentsResponse>(`/api/posts/${postId}/comments`)
  return res.data.data ?? []
}

export async function createComment(data: { post_id: string; parent_comment_id?: string; content: string; image?: File }): Promise<void> {
  const form = new FormData()
  form.append('post_id', data.post_id)
  if (data.parent_comment_id) form.append('parent_comment_id', data.parent_comment_id)
  form.append('content', data.content)
  if (data.image) form.append('image', data.image)
  await client.post('/api/comments', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export async function deleteComment(id: string): Promise<void> {
  await client.delete(`/api/comments/${id}`)
}

export async function editComment(id: string, data: { content: string; image?: File; remove_image?: boolean }): Promise<void> {
  const form = new FormData()
  form.append('content', data.content)
  if (data.image) form.append('image', data.image)
  if (data.remove_image) form.append('remove_image', '1')
  await client.put(`/api/comments/${id}/edit`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
