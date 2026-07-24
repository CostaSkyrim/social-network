import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useSignup } from '@/hooks/useAuth'
import { VALIDATION, validate_length } from '@/lib/validators'

export default function SignupPage() {
  const [form, set_form] = useState({
    email: '',
    password: '',
    first_name: '',
    last_name: '',
    date_of_birth: '',
    nickname: '',
    about_me: '',
  })
  const [errors, set_errors] = useState<Record<string, string>>({})
  const signup = useSignup()

  function set(field: string, value: string) {
    set_form((prev) => ({ ...prev, [field]: value }))
    set_errors((prev) => ({ ...prev, [field]: '' }))
  }

  async function handle_submit(e: React.FormEvent) {
    e.preventDefault()
    const new_errors: Record<string, string> = {}

    if (!form.email.trim()) new_errors.email = 'Email is required'
    if (!form.password.trim()) new_errors.password = 'Password is required'
    else {
      const pw_err = validate_length(form.password, 'Password', VALIDATION.password.min, VALIDATION.password.max)
      if (pw_err) new_errors.password = pw_err
    }
    if (!form.first_name.trim()) new_errors.first_name = 'First name is required'
    if (!form.last_name.trim()) new_errors.last_name = 'Last name is required'
    if (!form.date_of_birth) new_errors.date_of_birth = 'Date of birth is required'

    if (Object.keys(new_errors).length > 0) {
      set_errors(new_errors)
      return
    }

    try {
      await signup.mutateAsync({
        email: form.email.trim(),
        password: form.password,
        first_name: form.first_name.trim(),
        last_name: form.last_name.trim(),
        date_of_birth: new Date(form.date_of_birth).toISOString(),
        nickname: form.nickname.trim() || undefined,
        about_me: form.about_me.trim() || undefined,
      })
    } catch (err: any) {
      set_errors({ form: err?.response?.data?.error || err?.message || 'Signup failed' })
    }
  }

  return (
    <form onSubmit={handle_submit} className="space-y-4">
      <h2 className="text-xl font-semibold text-gray-900">Create account</h2>

      {errors.form && (
        <div className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{errors.form}</div>
      )}

      <Input
        id="email"
        label="Email"
        type="email"
        value={form.email}
        onChange={(e) => set('email', e.target.value)}
        error={errors.email}
        required
      />

      <Input
        id="password"
        label="Password"
        type="password"
        value={form.password}
        onChange={(e) => set('password', e.target.value)}
        placeholder={`${VALIDATION.password.min}–${VALIDATION.password.max} characters`}
        error={errors.password}
        required
      />

      <div className="flex gap-3">
        <Input
          id="first_name"
          label="First name"
          value={form.first_name}
          onChange={(e) => set('first_name', e.target.value)}
          error={errors.first_name}
          required
        />
        <Input
          id="last_name"
          label="Last name"
          value={form.last_name}
          onChange={(e) => set('last_name', e.target.value)}
          error={errors.last_name}
          required
        />
      </div>

      <Input
        id="date_of_birth"
        label="Date of birth"
        type="date"
        value={form.date_of_birth}
        onChange={(e) => set('date_of_birth', e.target.value)}
        error={errors.date_of_birth}
        required
      />

      <Input
        id="nickname"
        label="Nickname (optional)"
        value={form.nickname}
        onChange={(e) => set('nickname', e.target.value)}
      />

      <Input
        id="about_me"
        label="About me (optional)"
        value={form.about_me}
        onChange={(e) => set('about_me', e.target.value)}
      />

      <Button type="submit" className="w-full" loading={signup.isPending}>
        Create account
      </Button>

      <p className="text-center text-sm text-gray-500">
        Already have an account?{' '}
        <Link to="/login" className="font-medium text-blue-600 hover:text-blue-700">
          Sign in
        </Link>
      </p>
    </form>
  )
}
