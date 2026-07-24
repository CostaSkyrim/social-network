'use client'

export const dynamic = 'force-dynamic'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import { Spinner } from '@/components/ui/Spinner'

export default function RootPage() {
  const { is_authenticated, is_loading } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (is_loading) return
    router.replace(is_authenticated ? '/home' : '/login')
  }, [is_authenticated, is_loading, router])

  return (
    <div className="flex min-h-screen items-center justify-center">
      <Spinner size="lg" />
    </div>
  )
}
