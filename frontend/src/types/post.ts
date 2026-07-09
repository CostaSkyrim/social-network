import type { User } from './user'

export interface Post {
  id: string
  author_id: string
  author?: User
  group_id?: string
  content?: string
  image_path?: string
  privacy_level: 'public' | 'private' | 'almost_private'
  created_at: string
  updated_at: string
  comment_count?: number
  comment_list?: Comment[]
}
