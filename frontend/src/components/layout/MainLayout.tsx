import { Navigate, Outlet } from 'react-router-dom'
import { useAuth } from '@/context/AuthProvider'
import { Sidebar } from './Sidebar'
import { TopBar } from './TopBar'
import { MobileNav } from './MobileNav'
import { Spinner } from '@/components/ui/Spinner'

export function MainLayout() {
  const { is_authenticated, is_loading } = useAuth()

  if (is_loading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner size="lg" />
      </div>
    )
  }

  if (!is_authenticated) {
    return <Navigate to="/login" replace />
  }

  return (
    <div className="flex min-h-screen bg-gray-50">
      <Sidebar />
      <div className="flex flex-1 flex-col">
        <TopBar />
        <main className="flex-1 overflow-y-auto pb-16 md:pb-0">
          <div className="mx-auto max-w-2xl px-4 py-6">
            <Outlet />
          </div>
        </main>
      </div>
      <MobileNav />
    </div>
  )
}
