'use client'

import { useState } from 'react'
import Link from 'next/link'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ImageUpload } from '@/components/common/ImageUpload'
import OAuthButtons from '@/components/common/OAuthButtons'
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
  const [avatar, set_avatar] = useState<File | null>(null)
  const [avatar_preview, set_avatar_preview] = useState<string | null>(null)
  const [errors, set_errors] = useState<Record<string, string>>({})
  const signup = useSignup()

  function set(field: string, value: string) {
    set_form((prev) => ({ ...prev, [field]: value }))
    set_errors((prev) => ({ ...prev, [field]: '' }))
  }

  function handle_avatar(file: File) {
    if (avatar_preview) URL.revokeObjectURL(avatar_preview)
    set_avatar(file)
    set_avatar_preview(URL.createObjectURL(file))
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
        about_me: form.about_me.trim() || undefined,
        nickname: form.nickname.trim() || undefined,
        avatar: avatar ?? undefined,
      })
    } catch (err: any) {
      set_errors({ form: err?.response?.data?.error || err?.message || 'Signup failed' })
    }
  }

  return (
    <form onSubmit={handle_submit} className="space-y-4">
      <h2 className="text-xl font-semibold text-gray-100">Create account</h2>

      {errors.form && (
        <div className="rounded-lg bg-red-500/15 p-3 text-sm text-red-200">{errors.form}</div>
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
        placeholder="your_unique_handle"
      />

      <div>
        <p className="mb-1 text-sm font-medium text-gray-300">Avatar (optional)</p>
        <div className="flex items-center gap-3">
          {avatar_preview ? (
            /* eslint-disable-next-line @next/next/no-img-element */
            <img
              src={avatar_preview}
              alt="Avatar preview"
              className="h-12 w-12 rounded-full object-cover"
            />
          ) : (
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-purple-400/20 text-xs text-gray-300">
              No image
            </div>
          )}
          <ImageUpload on_select={handle_avatar} className="[&>div]:inline-flex">
            <button
              type="button"
              className="rounded-lg px-3 py-1.5 text-sm text-gray-300 hover:bg-purple-400/15 hover:text-gray-100"
            >
              {avatar_preview ? 'Change image' : 'Upload image'}
            </button>
          </ImageUpload>
        </div>
      </div>

      <Input
        id="about_me"
        label="About me (optional)"
        value={form.about_me}
        onChange={(e) => set('about_me', e.target.value)}
      />

      <Button type="submit" className="w-full" loading={signup.isPending}>
        Create account
      </Button>

      <OAuthButtons />

      <p className="text-center text-sm text-gray-300">
        Already have an account?{' '}
        <Link href="/login" className="font-medium text-fuchsia-300 hover:text-fuchsia-200">
          Sign in
        </Link>
      </p>
    </form>
  )
}
