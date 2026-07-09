import { useParams } from 'react-router-dom'

export default function ProfilePage() {
  const { uuid } = useParams()
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900">Profile</h2>
      <p className="text-sm text-gray-500">Profile for user: {uuid} — implement in Phase 5.</p>
    </div>
  )
}
