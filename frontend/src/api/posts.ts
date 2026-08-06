import client from './client'
import type { Post } from '@/types/post'

interface FeedResponse {
  message: string
  data?: Post[]
}

interface PostResponse {
  message: string
  data?: Post
}

interface CreatePostData {
  content: string
  privacy_level: string
  group_id?: string
  image?: File
}

export async function getFeed(page: number, limit: number = 10): Promise<Post[]> {
  const offset = (page - 1) * limit
  const res = await client.get<FeedResponse>(`/api/feed?limit=${limit}&offset=${offset}`)
  return res.data.data ?? []
}

export async function getGroupPosts(groupId: string, page: number, limit: number = 10): Promise<Post[]> {
  const offset = (page - 1) * limit
  const res = await client.get<FeedResponse>(`/api/groups/${groupId}/posts?limit=${limit}&offset=${offset}`)
  return res.data.data ?? []
}

export async function getPost(id: string): Promise<Post> {
  const res = await client.get<PostResponse>(`/api/post/${id}`)
  if (!res.data.data) throw new Error('Post not found')
  return res.data.data
}

export async function createPost(data: CreatePostData): Promise<void> {
  const form = new FormData()
  form.append('content', data.content)
  form.append('privacy_level', data.privacy_level)
  if (data.group_id) form.append('group_id', data.group_id)
  if (data.image) form.append('image', data.image)
  await client.post('/api/posts', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export async function deletePost(id: string): Promise<void> {
  await client.delete(`/api/posts/${id}`)
}

export async function editPost(id: string, data: { content: string; privacy_level: string; image?: File; remove_image?: boolean }): Promise<void> {
  const form = new FormData()
  form.append('content', data.content)
  form.append('privacy_level', data.privacy_level)
  if (data.image) form.append('image', data.image)
  if (data.remove_image) form.append('remove_image', '1')
  await client.put(`/api/posts/${id}/edit`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}
