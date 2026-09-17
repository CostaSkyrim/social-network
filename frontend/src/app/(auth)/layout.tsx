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
    <div className="flex min-h-screen flex-col items-center justify-center bg-transparent px-4">
      <div className="mb-8 text-center">
        <h1 className="bg-gradient-to-r from-fuchsia-400 to-violet-400 bg-clip-text text-6xl font-black tracking-wide text-transparent font-display sm:text-8xl md:text-9xl">
          Andromeda
        </h1>
        <p className="mt-1 text-sm text-purple-200/80">Connect with the world</p>
      </div>

      <div className="w-full max-w-sm">{children}</div>
    </div>
  )
}
