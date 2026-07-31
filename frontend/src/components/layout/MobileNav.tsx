'use client'

import { NavLink } from '@/lib/nav-link'
import { cn } from '@/lib/cn'

const items = [
  { to: '/home', label: 'Home', icon: '🏠' },
  { to: '/profile/me', label: 'Profile', icon: '👤' },
  { to: '/groups', label: 'Groups', icon: '👥' },
  { to: '/followers', label: 'Followers', icon: '💬' },
]

export function MobileNav() {
  return (
    <nav className="fixed bottom-0 left-0 right-0 flex border-t border-gray-200 bg-white md:hidden">
      {items.map((item) => (
        <NavLink
          key={item.to}
          href={item.to}
          className={({ isActive }) =>
            cn(
              'flex flex-1 flex-col items-center gap-0.5 py-2 text-xs',
              isActive ? 'text-blue-600' : 'text-gray-500',
            )
          }
        >
          <span className="text-lg">{item.icon}</span>
          {item.label}
        </NavLink>
      ))}
    </nav>
  )
}
