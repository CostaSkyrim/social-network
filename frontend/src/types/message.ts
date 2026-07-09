import type { User } from './user'

export interface Message {
  uuid: string
  sender_id: string
  sender?: User
  direct_message_id?: number
  group_id?: string
  content: string
  is_read: boolean
  created_at: string
}

export interface DirectMessage {
  id: number
  other_user: User
  last_message?: Message
  last_message_at?: string
  created_at: string
}

export interface MessageRead {
  message_id: string
  user_id: string
  read_at: string
}
