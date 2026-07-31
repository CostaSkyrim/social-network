import client from './client'
import type { User } from '@/types/user'

interface AuthResponse {
  message: string
  data?: User
  error?: string
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
): Promise<User> {
  const body: Record<string, string> = {
    email,
    password,
    first_name,
    last_name,
    date_of_birth,
  }
  if (about_me) body.about_me = about_me

  const res = await client.post<AuthResponse>('/api/signup', body)
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
