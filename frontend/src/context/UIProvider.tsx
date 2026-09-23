import { createContext, useContext, useState, type ReactNode } from 'react'

interface Toast {
  id: string
  message: string
  type: 'success' | 'error' | 'info'
}

interface UIContextValue {
  toasts: Toast[]
  show_toast: (toast: Omit<Toast, 'id'>) => void
  dismiss_toast: (id: string) => void
}

const UIContext = createContext<UIContextValue | null>(null)

export function UIProvider({ children }: { children: ReactNode }) {
  const [toasts, set_toasts] = useState<Toast[]>([])

  const show_toast = (toast: Omit<Toast, 'id'>) =>
    set_toasts((s) => [...s, { ...toast, id: crypto.randomUUID() }])
  const dismiss_toast = (id: string) =>
    set_toasts((s) => s.filter((t) => t.id !== id))

  return (
    <UIContext.Provider
      value={{
        toasts,
        show_toast,
        dismiss_toast,
      }}
    >
      {children}
    </UIContext.Provider>
  )
}

export function useUI() {
  const ctx = useContext(UIContext)
  if (!ctx) throw new Error('useUI must be used within UIProvider')
  return ctx
}
