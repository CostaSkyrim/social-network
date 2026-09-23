'use client'

import { useState } from 'react'
import { useAuth } from '@/context/AuthProvider'
import { useRSVP, useUpdateEvent, useDeleteEvent } from '@/hooks/useGroups'
import { Card, CardContent } from '@/components/ui/Card'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ImageUpload } from '@/components/common/ImageUpload'
import { format_datetime, format_starts_in } from '@/lib/format'
import { get_media_url } from '@/lib/media'
import { cn } from '@/lib/cn'
import type { GroupEvent } from '@/types/group'

interface EventCardProps {
  event: GroupEvent
  groupId: string
}

// Converts an ISO timestamp into the value format a datetime-local input expects.
function to_local_input(iso: string): string {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(
    d.getHours(),
  )}:${pad(d.getMinutes())}`
}

export function EventCard({ event, groupId }: EventCardProps) {
  const { user } = useAuth()
  const rsvp = useRSVP(event.id, groupId)
  const update_event = useUpdateEvent(groupId)
  const delete_event = useDeleteEvent(groupId)
  const is_pending = rsvp.isPending

  const [is_editing, set_editing] = useState(false)
  const [title, set_title] = useState(event.title)
  const [description, set_description] = useState(event.description ?? '')
  const [datetime, set_datetime] = useState(to_local_input(event.event_datetime))
  const [image, set_image] = useState<File | null>(null)
  const [preview_url, set_preview_url] = useState<string | null>(null)
  const [remove_image, set_remove_image] = useState(false)

  const is_owner = !!user && event.creator_id === user.id
  const is_cancelled = !!event.is_cancelled

  function handle_rsvp(response: 'going' | 'not_going') {
    if (event.my_response === response) return
    rsvp.mutate(response)
  }

  function handle_image(file: File) {
    if (preview_url) URL.revokeObjectURL(preview_url)
    set_image(file)
    set_preview_url(URL.createObjectURL(file))
    set_remove_image(false)
  }

  function handle_remove_image() {
    if (preview_url) URL.revokeObjectURL(preview_url)
    set_image(null)
    set_preview_url(null)
    set_remove_image(true)
  }

  async function handle_save() {
    await update_event.mutateAsync({
      eventId: event.id,
      input: {
        title: title.trim(),
        description: description.trim() || undefined,
        event_datetime: new Date(datetime).toISOString(),
        image: image ?? undefined,
        remove_image,
      },
    })
    set_editing(false)
    set_image(null)
    set_remove_image(false)
    if (preview_url) URL.revokeObjectURL(preview_url)
    set_preview_url(null)
  }

  function handle_cancel_event() {
    if (
      confirm(
        'Cancel this event? It stays visible (marked as cancelled) and RSVPs close.',
      )
    ) {
      delete_event.mutate(event.id)
    }
  }

  const creator_name = event.creator
    ? `${event.creator.first_name} ${event.creator.last_name}`
    : 'Unknown'

  const show_existing_image = !!event.image_path && !remove_image && !image
  const show_preview = !!preview_url

  return (
    <Card>
      <CardContent className="space-y-3">
        {is_editing ? (
          <div className="space-y-3">
            <Input
              id={`event-title-${event.id}`}
              label="Title"
              value={title}
              onChange={(e) => set_title(e.target.value)}
              maxLength={300}
            />
            <Input
              id={`event-desc-${event.id}`}
              label="Description (optional)"
              value={description}
              onChange={(e) => set_description(e.target.value)}
            />
            <Input
              id={`event-datetime-${event.id}`}
              label="Date and time"
              type="datetime-local"
              value={datetime}
              onChange={(e) => set_datetime(e.target.value)}
            />
            {(show_preview || show_existing_image) && (
              <div className="relative">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={preview_url ?? get_media_url(event.image_path)}
                  alt="Event image"
                  className="max-h-48 w-full rounded-lg object-contain"
                />
                <button
                  type="button"
                  onClick={handle_remove_image}
                  className="absolute right-2 top-2 rounded-full bg-black/60 px-2 py-1 text-xs text-white hover:bg-black/80"
                >
                  Remove
                </button>
              </div>
            )}
            <div className="flex items-center justify-between">
              <ImageUpload on_select={handle_image}>
                <button
                  type="button"
                  className="rounded-lg px-2 py-1.5 text-base text-gray-300 hover:bg-purple-400/15 hover:text-gray-100"
                  title={event.image_path && !remove_image ? 'Change image' : 'Add image'}
                >
                  📷
                </button>
              </ImageUpload>
              <div className="flex items-center gap-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => set_editing(false)}
                  disabled={update_event.isPending}
                >
                  Cancel
                </Button>
                <Button
                  size="sm"
                  onClick={handle_save}
                  loading={update_event.isPending}
                  disabled={!title.trim() || !datetime}
                >
                  Save
                </Button>
              </div>
            </div>
          </div>
        ) : (
          <>
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0">
                <h4
                  className={cn(
                    'text-sm font-semibold',
                    is_cancelled ? 'text-gray-400 line-through' : 'text-gray-100',
                  )}
                >
                  {event.title}
                </h4>
                <p className="mt-0.5 text-xs text-gray-300">
                  {format_datetime(event.event_datetime)} · by {creator_name}
                </p>
                {!is_cancelled && format_starts_in(event.event_datetime) && (
                  <span className="mt-1 inline-block rounded bg-violet-400/20 px-1.5 py-0.5 text-[10px] font-medium text-violet-200">
                    {format_starts_in(event.event_datetime)}
                  </span>
                )}
              </div>
              <div className="flex shrink-0 items-center gap-2">
                {is_cancelled ? (
                  <Badge variant="danger">Cancelled</Badge>
                ) : (
                  <Badge variant="warning">{event.total} going</Badge>
                )}
                {is_owner && !is_cancelled && (
                  <div className="flex items-center gap-1">
                    <button
                      onClick={() => set_editing(true)}
                      className="rounded p-1 text-gray-300 hover:bg-purple-400/15 hover:text-gray-100"
                      title="Edit event"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" />
                      </svg>
                    </button>
                    <button
                      onClick={handle_cancel_event}
                      className="rounded p-1 text-gray-300 hover:bg-purple-400/15 hover:text-red-400"
                      title="Cancel event"
                    >
                      <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                      </svg>
                    </button>
                  </div>
                )}
              </div>
            </div>

            {event.description && (
              <p className="text-sm text-gray-300 whitespace-pre-wrap">{event.description}</p>
            )}

            {event.image_path && (
              /* eslint-disable-next-line @next/next/no-img-element */
              <img
                src={get_media_url(event.image_path)}
                alt="Event image"
                className="max-h-64 w-full rounded-lg object-contain"
              />
            )}

            {is_cancelled ? (
              <p className="text-sm italic text-gray-400">
                This event was cancelled. RSVPs are closed.
              </p>
            ) : (
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handle_rsvp('going')}
                  disabled={is_pending}
                  className={cn(
                    'rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors disabled:opacity-50',
                    event.my_response === 'going'
                      ? 'border-green-500 bg-green-500/15 text-green-200'
                      : 'border-purple-400/30 text-gray-300 hover:bg-green-500/15',
                  )}
                >
                  ✓ Going ({event.going})
                </button>
                <button
                  onClick={() => handle_rsvp('not_going')}
                  disabled={is_pending}
                  className={cn(
                    'rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors disabled:opacity-50',
                    event.my_response === 'not_going'
                      ? 'border-red-500 bg-red-500/15 text-red-200'
                      : 'border-purple-400/30 text-gray-300 hover:bg-red-500/15',
                  )}
                >
                  ✗ Not going ({event.not_going})
                </button>
              </div>
            )}
          </>
        )}
      </CardContent>
    </Card>
  )
}
