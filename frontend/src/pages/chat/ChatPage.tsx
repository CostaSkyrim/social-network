import { useParams } from 'react-router-dom'

export default function ChatPage() {
  const { id } = useParams()
  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900">Messages</h2>
      <p className="text-sm text-gray-500">
        {id ? `Conversation ${id}` : 'Select a conversation'} — implement in Phase 8.
      </p>
    </div>
  )
}
