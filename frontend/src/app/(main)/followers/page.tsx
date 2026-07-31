'use client'

import { Suspense } from 'react'

export const dynamic = 'force-dynamic'

import FollowersPage from '@/views/followers/FollowersPage'

export default function FollowersRoute() {
  return (
    <Suspense fallback={null}>
      <FollowersPage />
    </Suspense>
  )
}
