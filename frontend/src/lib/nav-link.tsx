'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import type { ReactNode } from 'react'

export function NavLink({
  href,
  className,
  children,
}: {
  href: string
  className: (props: { isActive: boolean }) => string
  children: ReactNode
}) {
  const pathname = usePathname()
  const isActive = pathname ? pathname === href || pathname.startsWith(href + '/') : false

  return (
    <Link href={href} className={className({ isActive })}>
      {children}
    </Link>
  )
}
