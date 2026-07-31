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

export interface GroupDetail {
  group: Group
  members: GroupMember[]
}

export interface CreateEventInput {
  title: string
  description?: string
  event_datetime: string
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
  db_id: number
  group_id: string
  creator_id: string
  creator?: {
    first_name: string
    last_name: string
    nickname?: string
  }
  title: string
  description?: string
  event_datetime: string
  created_at: string
  updated_at: string
  going: number
  not_going: number
  total: number
  my_response?: 'going' | 'not_going'
}

export interface GroupEventResponse {
  event_id: string
  user_id: string
  response: 'going' | 'not_going'
}
