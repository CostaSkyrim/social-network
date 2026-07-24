import { useParams } from 'next/navigation'

export default function GroupDetailPage() {
  const params = useParams()
  const uuid = params?.uuid
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900">Group</h2>
      <p className="text-sm text-gray-500">Group {uuid} — implement in Phase 6.</p>
    </div>
  )
}
