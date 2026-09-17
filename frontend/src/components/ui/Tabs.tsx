import { type ReactNode } from 'react'
import { cn } from '@/lib/cn'

interface Tab {
  id: string
  label: string
}

interface TabsProps {
  tabs: Tab[]
  active: string
  on_change: (id: string) => void
  className?: string
}

export function Tabs({ tabs, active, on_change, className }: TabsProps) {
  return (
    <div className={cn('flex border-b border-purple-400/20', className)}>
      {tabs.map((tab) => (
        <button
          key={tab.id}
          onClick={() => on_change(tab.id)}
          className={cn(
            'px-4 py-3 text-sm font-medium transition-colors',
            active === tab.id
              ? 'border-b-2 border-violet-600 text-violet-600'
              : 'text-gray-300 hover:text-gray-300',
          )}
        >
          {tab.label}
        </button>
      ))}
    </div>
  )
}

export function TabPanel({ id, active, children }: { id: string; active: string; children: ReactNode }) {
  if (id !== active) return null
  return <div className="py-4">{children}</div>
}
