import type { User } from './user'

export interface Notification {
  id: number
  from_user_id?: string
  from_user?: User
  type: string
  content: string
  related_id?: number
  is_read: boolean
  read_at?: string
  created_at: string
}
