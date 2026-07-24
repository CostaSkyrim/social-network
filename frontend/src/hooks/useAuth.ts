import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useAuth } from '@/context/AuthProvider'
import { login as apiLogin, signup as apiSignup, logout as apiLogout } from '@/api/auth'
import type { User } from '@/types/user'

export function useLogin() {
  const { set_user } = useAuth()
  const query_client = useQueryClient()

  return useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) =>
      apiLogin(email, password),
    onSuccess: (user: User) => {
      set_user(user)
      query_client.invalidateQueries()
    },
  })
}

export function useSignup() {
  const { set_user } = useAuth()
  const query_client = useQueryClient()

  return useMutation({
    mutationFn: (data: {
      email: string
      password: string
      first_name: string
      last_name: string
      date_of_birth: string
      nickname?: string
      about_me?: string
    }) => apiSignup(data.email, data.password, data.first_name, data.last_name,
      data.date_of_birth, data.nickname, data.about_me),
    onSuccess: (user: User) => {
      set_user(user)
      query_client.invalidateQueries()
    },
  })
}

export function useLogout() {
  const { logout: clearUser } = useAuth()
  const query_client = useQueryClient()

  return useMutation({
    mutationFn: () => apiLogout(),
    onSuccess: () => {
      clearUser()
      query_client.clear()
    },
  })
}
