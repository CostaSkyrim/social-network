import { useAuth } from '@/context/AuthProvider'
import { useUI } from '@/context/UIProvider'
import { Avatar } from '@/components/ui/Avatar'

export function TopBar() {
  const { toggle_sidebar } = useUI()
  const { user } = useAuth()

  return (
    <header className="flex h-14 items-center justify-between border-b border-gray-200 bg-white px-4">
      <div className="flex items-center gap-3">
        <button
          onClick={toggle_sidebar}
          className="rounded-lg p-1 text-gray-500 hover:bg-gray-100 md:hidden"
        >
          <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6h16M4 12h16M4 18h16" />
          </svg>
        </button>
        <h1 className="text-lg font-bold text-blue-600">Social</h1>
      </div>

      <div className="flex items-center gap-3">
        <NavNotificationBell />
        <Avatar
          src={user?.avatar_path}
          alt={`${user?.first_name} ${user?.last_name}`}
          size="sm"
        />
      </div>
    </header>
  )
}

function NavNotificationBell() {
  return (
    <button className="relative rounded-lg p-1 text-gray-500 hover:bg-gray-100">
      <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path
          strokeLinecap="round"
          strokeLinejoin="round"
          strokeWidth={2}
          d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"
        />
      </svg>
    </button>
  )
}
