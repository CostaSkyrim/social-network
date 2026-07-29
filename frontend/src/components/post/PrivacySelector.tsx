'use client'

interface PrivacySelectorProps {
  value: string
  onChange: (value: string) => void
}

const options = [
  { value: 'public', label: 'Public', desc: 'Everyone can see' },
  { value: 'followers', label: 'Followers', desc: 'Only your followers' },
  { value: 'private', label: 'Private', desc: 'Only specific people' },
]

export function PrivacySelector({ value, onChange }: PrivacySelectorProps) {
  return (
    <div className="flex gap-2">
      {options.map((opt) => (
        <button
          key={opt.value}
          type="button"
          onClick={() => onChange(opt.value)}
          className={`rounded-lg border px-3 py-1.5 text-xs font-medium transition-colors ${
            value === opt.value
              ? 'border-blue-500 bg-blue-50 text-blue-700'
              : 'border-gray-300 text-gray-600 hover:bg-gray-50'
          }`}
          title={opt.desc}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}
