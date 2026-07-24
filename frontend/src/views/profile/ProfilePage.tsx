import { useParams } from 'next/navigation'

export default function ProfilePage() {
  const params = useParams()
  const uuid = params?.uuid
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900">Profile</h2>
      <p className="text-sm text-gray-500">Profile for user: {uuid} — implement in Phase 5.</p>
    </div>
  )
}
