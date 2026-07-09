import { NavLink } from 'react-router-dom'
import { cn } from '@/lib/cn'
import { useAuthStore } from '@/stores/authStore'
import { Avatar } from '@/components/ui/Avatar'

const nav_items = [
  { to: '/home', label: 'Home', icon: '🏠' },
  { to: '/profile/me', label: 'My Profile', icon: '👤' },
  { to: '/groups', label: 'Groups', icon: '👥' },
  { to: '/chat', label: 'Messages', icon: '💬' },
  { to: '/notifications', label: 'Notifications', icon: '🔔' },
  { to: '/search', label: 'Search', icon: '🔍' },
]

export function Sidebar() {
  const user = useAuthStore((s) => s.user)

  return (
    <aside className="hidden w-64 border-r border-gray-200 bg-white md:flex md:flex-col">
      <div className="flex items-center gap-3 border-b border-gray-200 p-4">
        <Avatar src={user?.avatar_path} alt={`${user?.first_name} ${user?.last_name}`} />
        <div className="min-w-0">
          <p className="truncate text-sm font-medium text-gray-900">
            {user?.first_name} {user?.last_name}
          </p>
          <p className="truncate text-xs text-gray-500">@{user?.nickname}</p>
        </div>
      </div>

      <nav className="flex-1 space-y-1 p-3">
        {nav_items.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) =>
              cn(
                'flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors',
                isActive
                  ? 'bg-blue-50 text-blue-700'
                  : 'text-gray-700 hover:bg-gray-100',
              )
            }
          >
            <span>{item.icon}</span>
            {item.label}
          </NavLink>
        ))}
      </nav>

      <div className="border-t border-gray-200 p-3">
        <NavLink
          to="/login"
          onClick={() => useAuthStore.getState().logout()}
          className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
        >
          <span>🚪</span>
          Logout
        </NavLink>
      </div>
    </aside>
  )
}
