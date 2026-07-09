export interface User {
  id: string
  email: string
  first_name: string
  last_name: string
  nickname?: string
  date_of_birth: string
  about_me?: string
  avatar_path?: string
  is_public: boolean
  last_seen?: string
  is_online?: boolean
}

export interface Session {
  session_id: string
  user_id: number
  expires_at: string
}

export interface Follow {
  follower_id: string
  following_id: string
  status: 'pending' | 'accepted' | 'declined'
}
