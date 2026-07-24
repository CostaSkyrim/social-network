'use client'

import Link from 'next/link'

export default function ErrorPage({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gray-50 text-center">
      <div className="text-6xl font-bold text-gray-300">!</div>
      <h2 className="mt-4 text-xl font-semibold text-gray-900">Something went wrong</h2>
      <p className="mt-1 text-sm text-gray-500">{error.message || 'An unexpected error occurred.'}</p>
      <div className="mt-6 flex gap-3">
        <button
          onClick={reset}
          className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          Try again
        </button>
        <Link
          href="/home"
          className="rounded-lg bg-gray-200 px-4 py-2 text-sm font-medium text-gray-900 hover:bg-gray-300"
        >
          Go home
        </Link>
      </div>
    </div>
  )
}
