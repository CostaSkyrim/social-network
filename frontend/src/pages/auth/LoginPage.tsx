import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useLogin } from '@/hooks/useAuth'

export default function LoginPage() {
  const [email, set_email] = useState('')
  const [password, set_password] = useState('')
  const [error, set_error] = useState<string | null>(null)
  const login = useLogin()

  async function handle_submit(e: React.FormEvent) {
    e.preventDefault()
    set_error(null)

    if (!email.trim() || !password.trim()) {
      set_error('Email and password are required')
      return
    }

    try {
      await login.mutateAsync({ email: email.trim(), password })
    } catch (err: any) {
      set_error(err?.response?.data?.error || err?.message || 'Login failed')
    }
  }

  return (
    <form onSubmit={handle_submit} className="space-y-4">
      <h2 className="text-xl font-semibold text-gray-900">Sign in</h2>

      {error && (
        <div className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</div>
      )}

      <Input
        id="email"
        label="Email"
        type="email"
        value={email}
        onChange={(e) => set_email(e.target.value)}
        placeholder="alice@example.com"
        required
      />

      <Input
        id="password"
        label="Password"
        type="password"
        value={password}
        onChange={(e) => set_password(e.target.value)}
        placeholder="password123"
        required
      />

      <Button type="submit" className="w-full" loading={login.isPending}>
        Sign in
      </Button>

      <p className="text-center text-sm text-gray-500">
        Don't have an account?{' '}
        <Link to="/signup" className="font-medium text-blue-600 hover:text-blue-700">
          Sign up
        </Link>
      </p>
    </form>
  )
}
