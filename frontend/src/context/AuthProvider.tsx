import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import type { User } from '@/types/user'
import client from '@/api/client'

interface AuthContextValue {
  user: User | null
  is_authenticated: boolean
  set_user: (user: User | null) => void
  logout: () => void
  check_session: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, set_user] = useState<User | null>(null)

  const check_session = async () => {
    try {
      const res = await client.get<User>('/profile/me')
      set_user(res.data)
    } catch {
      set_user(null)
    }
  }

  useEffect(() => {
    check_session()
  }, [])

  const logout = () => set_user(null)

  return (
    <AuthContext.Provider
      value={{
        user,
        is_authenticated: user !== null,
        set_user,
        logout,
        check_session,
      }}
    >
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
