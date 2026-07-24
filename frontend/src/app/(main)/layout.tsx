'use client'

import { useEffect, type ReactNode } from 'react'

export const dynamic = 'force-dynamic'
import { useRouter } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import { Sidebar } from '@/components/layout/Sidebar'
import { TopBar } from '@/components/layout/TopBar'
import { MobileNav } from '@/components/layout/MobileNav'
import { Spinner } from '@/components/ui/Spinner'

export default function MainLayout({ children }: { children: ReactNode }) {
  const { is_authenticated, is_loading } = useAuth()
  const router = useRouter()

  useEffect(() => {
    if (!is_loading && !is_authenticated) {
      router.replace('/login')
    }
  }, [is_authenticated, is_loading, router])

  if (is_loading || !is_authenticated) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner size="lg" />
      </div>
    )
  }

  return (
    <div className="flex min-h-screen bg-gray-50">
      <Sidebar />
      <div className="flex flex-1 flex-col">
        <TopBar />
        <main className="flex-1 overflow-y-auto pb-16 md:pb-0">
          <div className="mx-auto max-w-2xl px-4 py-6">
            {children}
          </div>
        </main>
      </div>
      <MobileNav />
    </div>
  )
}
