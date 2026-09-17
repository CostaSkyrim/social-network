'use client'

import { useRouter } from 'next/navigation'
import { NAV_ITEMS } from '@/lib/nav'
import { Dropdown, DropdownItem } from '@/components/ui/Dropdown'
import { useAuth } from '@/context/AuthProvider'
import { useLogout } from '@/hooks/useAuth'
import { disconnectWebSocket } from '@/hooks/useWebSocket'
import { Avatar } from '@/components/ui/Avatar'
import type { ReactNode } from 'react'

export function NavMenu({
  trigger,
  align = 'left',
}: {
  trigger: ReactNode
  align?: 'left' | 'right'
}) {
  const router = useRouter()
  const { user } = useAuth()
  const logout_mutation = useLogout()

  async function handle_logout() {
    await logout_mutation.mutateAsync()
    disconnectWebSocket()
    router.push('/login')
  }

  return (
    <Dropdown align={align} trigger={trigger}>
      <div className="flex items-center gap-2 border-b border-purple-400/20 px-4 py-3">
        <Avatar
          src={user?.avatar_path}
          alt={`${user?.first_name} ${user?.last_name}`}
          size="sm"
        />
        <div className="min-w-0">
          <p className="truncate text-sm font-medium text-gray-100">
            {user?.first_name} {user?.last_name}
          </p>
          <p className="truncate text-xs text-gray-300">@{user?.nickname}</p>
        </div>
      </div>

      {NAV_ITEMS.map((item) => (
        <DropdownItem key={item.to} onClick={() => router.push(item.to)}>
          <span className="mr-2">{item.icon}</span>
          {item.label}
        </DropdownItem>
      ))}

      <div className="my-1 border-t border-purple-400/20" />
      <DropdownItem danger onClick={handle_logout}>
        <span className="mr-2">🚪</span>
        {logout_mutation.isPending ? 'Logging out...' : 'Logout'}
      </DropdownItem>
    </Dropdown>
  )
}
