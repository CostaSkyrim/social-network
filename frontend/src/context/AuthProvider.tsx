import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import type { User } from '@/types/user'
import { checkSession as apiCheckSession } from '@/api/auth'

interface AuthContextValue {
  user: User | null
  is_authenticated: boolean
  is_loading: boolean
  set_user: (user: User | null) => void
  logout: () => void
  check_session: () => Promise<void>
}

const AuthContext = createContext<AuthContextValue | null>(null)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, set_user] = useState<User | null>(null)
  const [is_loading, set_loading] = useState(true)

  const check_session = async () => {
    set_loading(true)
    const u = await apiCheckSession()
    set_user(u)
    set_loading(false)
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
        is_loading,
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
