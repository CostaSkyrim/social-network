import type { User } from './user'

export interface Post {
  id: string
  db_id: number
  author_id: string
  author?: User
  group_id?: string
  content?: string
  image_path?: string
  privacy_level: 'public' | 'followers' | 'private'
  created_at: string
  updated_at: string
  comment_count?: number
  comment_list?: Comment[]
}
