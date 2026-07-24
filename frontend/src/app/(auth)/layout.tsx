'use client'

import { useEffect, type ReactNode } from 'react'

export const dynamic = 'force-dynamic'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import { Spinner } from '@/components/ui/Spinner'

export default function AuthLayout({ children }: { children: ReactNode }) {
  const { is_authenticated, is_loading } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (!is_loading && is_authenticated) {
      router.replace('/home')
    }
  }, [is_authenticated, is_loading, router])

  if (is_loading || is_authenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner size="lg" />
      </div>
    )
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold text-blue-600">Social</h1>
          <p className="mt-1 text-sm text-gray-500">Connect with the world</p>
        </div>
        {children}
      </div>
    </div>
  )
}
