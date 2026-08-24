'use client'

import { Suspense } from 'react'

export const dynamic = 'force-dynamic'

import LoginPage from '@/views/auth/LoginPage'

export default function LoginRoute() {
  return (
    <Suspense fallback={null}>
      <LoginPage />
    </Suspense>
  )
}
