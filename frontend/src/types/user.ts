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

export interface SearchUser {
  id: string
  first_name: string
  last_name: string
  nickname?: string
  avatar_path?: string
  is_public: boolean
  is_online?: boolean
  is_following: boolean
  is_follow_pending: boolean
}
