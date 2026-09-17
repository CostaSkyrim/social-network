'use client'

import { useEffect, type ReactNode } from 'react'

export const dynamic = 'force-dynamic'
import { useRouter, usePathname } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import { TopBar } from '@/components/layout/TopBar'
import { MobileNav } from '@/components/layout/MobileNav'
import { Spinner } from '@/components/ui/Spinner'

export default function MainLayout({ children }: { children: ReactNode }) {
  const { is_authenticated, is_loading } = useAuth()
  const router = useRouter()
  const pathname = usePathname()

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

  // The home page uses a wider, multi-column layout; every other page stays in
  // the narrow reading column.
  const container_width = pathname === '/home' ? 'max-w-6xl' : 'max-w-2xl'

  return (
    <div className="flex min-h-screen bg-transparent">
      <div className="flex flex-1 flex-col">
        <TopBar />
        <main className="flex-1 overflow-y-auto pb-16 md:pb-0">
          <div className={`mx-auto px-4 py-6 ${container_width}`}>
            {children}
          </div>
        </main>
      </div>
      <MobileNav />
    </div>
  )
}
