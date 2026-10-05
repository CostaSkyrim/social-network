import client from './client'
import type { User } from '@/types/user'

interface AuthResponse {
  message: string
  data?: User
  error?: string
}

export interface OAuthProviders {
  google: boolean
  github: boolean
}

/**
 * Reports which OAuth providers the backend has credentials for. Providers
 * that are not configured are hidden in the UI so the app runs fine without
 * any OAuth secrets.
 */
export async function getOAuthProviders(): Promise<OAuthProviders> {
  const res = await client.get<{ data?: OAuthProviders }>('/api/auth/providers')
  return res.data.data ?? { google: false, github: false }
}

export async function login(email: string, password: string): Promise<User> {
  const res = await client.post<AuthResponse>('/api/login', { email, password })
  if (!res.data.data) throw new Error(res.data.error || 'Login failed')
  return res.data.data
}

export async function signup(
  email: string,
  password: string,
  first_name: string,
  last_name: string,
  date_of_birth: string,
  about_me?: string,
  nickname?: string,
  avatar?: File,
): Promise<User> {
  const body: Record<string, string> = {
    email,
    password,
    first_name,
    last_name,
    date_of_birth,
  }
  if (about_me) body.about_me = about_me
  if (nickname) body.nickname = nickname

  let res
  if (avatar) {
    const form_data = new FormData()
    for (const [key, value] of Object.entries(body)) {
      form_data.append(key, value)
    }
    form_data.append('image', avatar)
    res = await client.post<AuthResponse>('/api/signup', form_data)
  } else {
    res = await client.post<AuthResponse>('/api/signup', body)
  }

  if (!res.data.data) throw new Error(res.data.error || 'Signup failed')
  return res.data.data
}

export async function logout(): Promise<void> {
  await client.post('/api/logout')
}

export async function checkSession(): Promise<User | null> {
  try {
    const res = await client.get<AuthResponse>('/api/auth/check')
    return res.data.data ?? null
  } catch {
    return null
  }
}
