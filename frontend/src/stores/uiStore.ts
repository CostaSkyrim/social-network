import { create } from 'zustand'

interface Toast {
  id: string
  message: string
  type: 'success' | 'error' | 'info'
}

interface UIState {
  sidebar_open: boolean
  active_modal: string | null
  toasts: Toast[]
  toggle_sidebar: () => void
  open_modal: (id: string) => void
  close_modal: () => void
  show_toast: (toast: Omit<Toast, 'id'>) => void
  dismiss_toast: (id: string) => void
}

export const useUIStore = create<UIState>((set) => ({
  sidebar_open: true,
  active_modal: null,
  toasts: [],
  toggle_sidebar: () => set((s) => ({ sidebar_open: !s.sidebar_open })),
  open_modal: (id) => set({ active_modal: id }),
  close_modal: () => set({ active_modal: null }),
  show_toast: (toast) =>
    set((s) => ({
      toasts: [...s.toasts, { ...toast, id: crypto.randomUUID() }],
    })),
  dismiss_toast: (id) =>
    set((s) => ({ toasts: s.toasts.filter((t) => t.id !== id) })),
}))
