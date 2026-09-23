export function format_date(iso: string): string {
  const date = new Date(iso)
  const now = new Date()
  const diff = now.getTime() - date.getTime()
  const mins = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)

  if (mins < 1) return 'just now'
  if (mins < 60) return `${mins}m ago`
  if (hours < 24) return `${hours}h ago`
  if (days < 7) return `${days}d ago`
  return date.toLocaleDateString()
}

export function truncate(text: string, max: number): string {
  if (text.length <= max) return text
  return text.slice(0, max) + '...'
}

// format_post_time gives a short "when" stamp. Today → how long ago (e.g.
// "5m ago", "3h ago"); on any earlier day → the day + month, adding the year
// only when it isn't the current year.
export function format_post_time(iso: string): string {
  const date = new Date(iso)
  const now = new Date()
  const start_of_day = (d: Date) =>
    new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime()

  const days_ago = Math.round(
    (start_of_day(now) - start_of_day(date)) / 86_400_000,
  )

  if (days_ago <= 0) {
    const mins = Math.floor((now.getTime() - date.getTime()) / 60_000)
    if (mins < 1) return 'just now'
    if (mins < 60) return `${mins}m ago`
    return `${Math.floor(mins / 60)}h ago`
  }

  if (date.getFullYear() === now.getFullYear()) {
    return date.toLocaleDateString(undefined, { day: 'numeric', month: 'short' })
  }
  return date.toLocaleDateString(undefined, {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}

export function format_datetime(iso: string): string {
  const date = new Date(iso)
  return date.toLocaleString(undefined, {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

// format_starts_in returns a short "starts in …" hint for upcoming events
// (within 24h), or null when the event is in the past or further out.
export function format_starts_in(iso: string): string | null {
  const diff = new Date(iso).getTime() - Date.now()
  if (diff <= 0) return null

  const mins = Math.floor(diff / 60000)
  if (mins < 60) return `starts in ${Math.max(1, mins)}m`

  const hours = Math.floor(mins / 60)
  if (hours < 24) return `starts in ${hours}h`
  return null
}
