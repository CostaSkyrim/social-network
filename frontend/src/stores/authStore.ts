import { create } from 'zustand'
import type { User } from '@/types/user'

interface AuthState {
  user: User | null
  is_authenticated: boolean
  set_user: (user: User | null) => void
  logout: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  is_authenticated: false,
  set_user: (user) => set({ user, is_authenticated: user !== null }),
  logout: () => set({ user: null, is_authenticated: false }),
}))
