import { Button } from '@/components/ui/Button'

export default function LoginPage() {
  return (
    <div className="space-y-4">
      <h2 className="text-xl font-semibold text-gray-900">Sign in</h2>
      <p className="text-sm text-gray-500">Sign in page — wire up with useAuth hook in Phase 2.</p>
      <div className="space-y-3">
        <Button className="w-full">Sign in</Button>
        <Button variant="secondary" className="w-full">
          Continue with Google
        </Button>
        <Button variant="secondary" className="w-full">
          Continue with GitHub
        </Button>
      </div>
    </div>
  )
}
