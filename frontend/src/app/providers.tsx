'use client'

import { type ReactNode, useState } from 'react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { AuthProvider } from '@/context/AuthProvider'
import { UIProvider } from '@/context/UIProvider'
import { ErrorBoundary } from '@/components/common/ErrorBoundary'
import { ToastContainer } from '@/components/ui/Toast'

export function Providers({ children }: { children: ReactNode }) {
  const [query_client] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            staleTime: 30_000,
            retry: 1,
            refetchOnWindowFocus: true,
          },
        },
      }),
  )

  return (
    <ErrorBoundary>
      <QueryClientProvider client={query_client}>
        <AuthProvider>
          <UIProvider>
            {children}
            <ToastContainer />
          </UIProvider>
        </AuthProvider>
      </QueryClientProvider>
    </ErrorBoundary>
  )
}
