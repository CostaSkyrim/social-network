import { Outlet } from 'react-router-dom'

export function AuthLayout() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gray-50 px-4">
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <h1 className="text-3xl font-bold text-blue-600">Social</h1>
          <p className="mt-1 text-sm text-gray-500">Connect with the world</p>
        </div>
        <Outlet />
      </div>
    </div>
  )
}
