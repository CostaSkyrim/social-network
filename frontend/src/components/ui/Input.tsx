import { forwardRef, type InputHTMLAttributes } from 'react'
import { cn } from '@/lib/cn'

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ className, label, error, id, ...props }, ref) => {
    return (
      <div className="space-y-1">
        {label && (
          <label htmlFor={id} className="block text-sm font-medium text-gray-300">
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={id}
          className={cn(
            'block w-full rounded-lg border bg-transparent px-3 py-2 text-sm text-gray-100 shadow-sm transition-colors placeholder:text-gray-300 [color-scheme:dark] focus:outline-none focus:ring-2',
            error
              ? 'border-red-500 focus:border-red-500 focus:ring-red-200'
              : 'border-purple-400/30 focus:border-violet-500 focus:ring-violet-300/40',
            className,
          )}
          {...props}
        />
        {error && <p className="text-sm text-red-600">{error}</p>}
      </div>
    )
  },
)
Input.displayName = 'Input'
