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
  group_id?: number
}

export async function getFeed(page: number, limit: number = 10): Promise<Post[]> {
  const offset = (page - 1) * limit
  const res = await client.get<FeedResponse>(`/api/feed?limit=${limit}&offset=${offset}`)
  return res.data.data ?? []
}

export async function getPost(id: number): Promise<Post> {
  const res = await client.get<PostResponse>(`/api/post/${id}`)
  if (!res.data.data) throw new Error('Post not found')
  return res.data.data
}

export async function createPost(data: CreatePostData): Promise<void> {
  await client.post('/api/posts', data)
}

export async function deletePost(id: number): Promise<void> {
  await client.delete(`/api/posts/${id}`)
}

export async function editPost(id: number, data: { content: string; image_path?: string; privacy_level: string }): Promise<void> {
  await client.put(`/api/posts/${id}/edit`, data)
}
