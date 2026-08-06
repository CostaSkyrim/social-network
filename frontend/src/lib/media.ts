const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

export function get_media_url(path?: string): string | undefined {
  if (!path) return undefined
  if (/^(https?:|blob:|data:)/.test(path)) return path
  return `${API_BASE}/${path}`
}
