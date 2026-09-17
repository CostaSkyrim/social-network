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
    <div className="flex min-h-screen flex-col items-center justify-center bg-purple-400/10 text-center">
      <div className="text-6xl font-bold text-gray-300">!</div>
      <h2 className="mt-4 text-xl font-semibold text-gray-100">Something went wrong</h2>
      <p className="mt-1 text-sm text-gray-300">{error.message || 'An unexpected error occurred.'}</p>
      <div className="mt-6 flex gap-3">
        <button
          onClick={reset}
          className="rounded-lg bg-violet-600 px-4 py-2 text-sm font-medium text-white hover:bg-violet-700"
        >
          Try again
        </button>
        <Link
          href="/home"
          className="rounded-lg bg-purple-400/20 px-4 py-2 text-sm font-medium text-gray-100 hover:bg-purple-400/30"
        >
          Go home
        </Link>
      </div>
    </div>
  )
}
