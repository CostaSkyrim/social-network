'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import { Spinner } from '@/components/ui/Spinner'

export default function MyProfilePage() {
  const { user, is_loading } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (!is_loading && user?.id) {
      router.replace(`/profile/${user.id}`)
    }
  }, [is_loading, user, router])

  return (
    <div className="flex items-center justify-center py-16">
      <Spinner size="lg" />
    </div>
  )
}
