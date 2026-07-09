import { useParams } from 'react-router-dom'

export default function GroupDetailPage() {
  const { uuid } = useParams()
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900">Group</h2>
      <p className="text-sm text-gray-500">Group {uuid} — implement in Phase 6.</p>
    </div>
  )
}
