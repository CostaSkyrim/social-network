import { cn } from '@/lib/cn'
import { get_media_url } from '@/lib/media'
import { useState } from 'react'

interface AvatarProps {
  src?: string
  alt: string
  size?: 'sm' | 'md' | 'lg' | 'xl'
  className?: string
}

const size_map = {
  sm: 'h-8 w-8 text-xs',
  md: 'h-10 w-10 text-sm',
  lg: 'h-14 w-14 text-lg',
  xl: 'h-20 w-20 text-2xl',
}

export function Avatar({ src, alt, size = 'md', className }: AvatarProps) {
  const [error, setError] = useState(false)

  if (src && !error) {
    return (
      <img
        src={get_media_url(src)}
        alt={alt}
        onError={() => setError(true)}
        className={cn('rounded-full object-cover', size_map[size], className)}
      />
    )
  }

  const initials = alt
    .split(' ')
    .map((n) => n[0])
    .join('')
    .toUpperCase()
    .slice(0, 2)

  return (
    <div
      className={cn(
        'flex items-center justify-center rounded-full bg-purple-300 font-medium text-purple-900',
        size_map[size],
        className,
      )}
    >
      {initials}
    </div>
  )
}
