export interface Group {
  id: string
  creator_id: string
  title: string
  description?: string
  avatar_path?: string
  last_message_at?: string
  created_at: string
  updated_at: string
}

export interface GroupMember {
  group_id: string
  user_id: string
  status: 'pending' | 'accepted' | 'declined' | 'invited'
  invited_by?: string
  joined_at?: string
}

export interface GroupEvent {
  id: string
  group_id: string
  creator_id: string
  title: string
  description?: string
  event_datetime: string
  created_at: string
  updated_at: string
}

export interface GroupEventResponse {
  event_id: string
  user_id: string
  response: 'going' | 'not_going' | 'maybe'
}
