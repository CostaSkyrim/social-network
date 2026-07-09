import { Button } from '@/components/ui/Button'

export default function SignupPage() {
  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold text-gray-900">Create account</h2>
      <p className="text-sm text-gray-500">Signup page — wire up with useAuth hook in Phase 2.</p>
      <Button className="w-full">Create account</Button>
    </div>
  )
}
