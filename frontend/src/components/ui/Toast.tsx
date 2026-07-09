import { useEffect } from 'react'
import { useUIStore } from '@/stores/uiStore'
import { cn } from '@/lib/cn'

const type_styles = {
  success: 'bg-green-600 text-white',
  error: 'bg-red-600 text-white',
  info: 'bg-blue-600 text-white',
}

export function ToastContainer() {
  const toasts = useUIStore((s) => s.toasts)
  const dismiss_toast = useUIStore((s) => s.dismiss_toast)

  return (
    <div className="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
      {toasts.map((t) => (
        <ToastItem key={t.id} id={t.id} message={t.message} type={t.type} on_dismiss={dismiss_toast} />
      ))}
    </div>
  )
}

function ToastItem({
  id,
  message,
  type,
  on_dismiss,
}: {
  id: string
  message: string
  type: 'success' | 'error' | 'info'
  on_dismiss: (id: string) => void
}) {
  useEffect(() => {
    const timer = setTimeout(() => on_dismiss(id), 4000)
    return () => clearTimeout(timer)
  }, [id, on_dismiss])

  return (
    <div
      className={cn('rounded-lg px-4 py-3 text-sm shadow-lg transition-all', type_styles[type])}
      onClick={() => on_dismiss(id)}
    >
      {message}
    </div>
  )
}
