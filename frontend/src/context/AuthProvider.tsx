import { createContext, useContext, useEffect, useRef, useState, type ReactNode } from 'react'
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
  const check_ref = useRef<Promise<void> | null>(null)

  const check_session = async () => {
    if (check_ref.current) return check_ref.current
    set_loading(true)
    check_ref.current = apiCheckSession().then((u) => {
      set_user(u)
      set_loading(false)
      check_ref.current = null
    }).catch(() => {
      set_user(null)
      set_loading(false)
      check_ref.current = null
    })
    return check_ref.current
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
