import { createContext, useContext, useState, type ReactNode } from 'react'

interface Toast {
  id: string
  message: string
  type: 'success' | 'error' | 'info'
}

interface UIContextValue {
  sidebar_open: boolean
  toggle_sidebar: () => void
  toasts: Toast[]
  show_toast: (toast: Omit<Toast, 'id'>) => void
  dismiss_toast: (id: string) => void
  active_modal: string | null
  open_modal: (id: string) => void
  close_modal: () => void
}

const UIContext = createContext<UIContextValue | null>(null)

export function UIProvider({ children }: { children: ReactNode }) {
  const [sidebar_open, set_sidebar_open] = useState(true)
  const [toasts, set_toasts] = useState<Toast[]>([])
  const [active_modal, set_active_modal] = useState<string | null>(null)

  const toggle_sidebar = () => set_sidebar_open((s) => !s)
  const show_toast = (toast: Omit<Toast, 'id'>) =>
    set_toasts((s) => [...s, { ...toast, id: crypto.randomUUID() }])
  const dismiss_toast = (id: string) =>
    set_toasts((s) => s.filter((t) => t.id !== id))
  const open_modal = (id: string) => set_active_modal(id)
  const close_modal = () => set_active_modal(null)

  return (
    <UIContext.Provider
      value={{
        sidebar_open,
        toggle_sidebar,
        toasts,
        show_toast,
        dismiss_toast,
        active_modal,
        open_modal,
        close_modal,
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
