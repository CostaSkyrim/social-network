import type { User } from './user'

export interface Comment {
  id: string
  post_id: string
  author_id: string
  author?: User
  parent_comment_id?: string
  content: string
  image_path?: string
  is_deleted?: boolean
  created_at: string
  updated_at: string
}
