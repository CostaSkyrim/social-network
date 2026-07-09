import { createContext, useContext, useEffect, type ReactNode } from 'react'
import { useAuthStore } from '@/stores/authStore'
import type { User } from '@/types/user'

interface AuthContextValue {
  check_session: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const set_user = useAuthStore((s) => s.set_user)

  const check_session = async () => {
    try {
      const { default: client } = await import('@/api/client')
      const res = await client.get<User>('/profile/me')
      set_user(res.data)
    } catch {
      set_user(null)
    }
  }

  useEffect(() => {
    check_session()
  }, [])

  return (
    <AuthContext.Provider value={{ check_session }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuthContext() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuthContext must be used within AuthProvider')
  return ctx
}
