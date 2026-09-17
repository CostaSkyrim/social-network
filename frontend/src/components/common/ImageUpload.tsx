import { useRef, useState, type ChangeEvent } from 'react'
import { cn } from '@/lib/cn'

interface ImageUploadProps {
  on_select: (file: File) => void
  accept?: string
  max_size?: number
  className?: string
  children?: React.ReactNode
}

export function ImageUpload({
  on_select,
  accept = 'image/jpeg,image/png,image/gif,image/webp',
  max_size = 20 * 1024 * 1024,
  className,
  children,
}: ImageUploadProps) {
  const input_ref = useRef<HTMLInputElement>(null)
  const [error, setError] = useState<string | null>(null)

  function handle_change(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0]
    setError(null)

    if (!file) return

    if (file.size > max_size) {
      setError('File is too large. Max 20MB.')
      return
    }

    on_select(file)
    e.target.value = ''
  }

  return (
    <div className={className}>
      <input
        ref={input_ref}
        type="file"
        accept={accept}
        onChange={handle_change}
        className="hidden"
      />
      <div onClick={() => input_ref.current?.click()}>
        {children || (
          <button
            type="button"
            className={cn(
              'flex cursor-pointer items-center justify-center rounded-lg border-2 border-dashed border-purple-400/30 p-6 text-sm text-gray-300 transition-colors hover:border-purple-400/60 hover:text-gray-100',
            )}
          >
            Upload image
          </button>
        )}
      </div>
      {error && <p className="mt-1 text-sm text-red-600">{error}</p>}
    </div>
  )
}
