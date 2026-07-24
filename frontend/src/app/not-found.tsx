import Link from 'next/link'

export default function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-gray-50 text-center">
      <div className="text-6xl font-bold text-gray-300">404</div>
      <h2 className="mt-4 text-xl font-semibold text-gray-900">Page not found</h2>
      <p className="mt-1 text-sm text-gray-500">
        The page you're looking for doesn't exist or has been moved.
      </p>
      <Link
        href="/home"
        className="mt-6 rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
      >
        Go home
      </Link>
    </div>
  )
}
