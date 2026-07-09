import { useParams } from 'react-router-dom'

export default function PostDetailPage() {
  const { uuid } = useParams()
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900">Post</h2>
      <p className="text-sm text-gray-500">Post {uuid} — implement in Phase 3.</p>
    </div>
  )
}
