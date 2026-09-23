'use client'

import { useState, useCallback } from 'react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ImageUpload } from '@/components/common/ImageUpload'
import { Card, CardContent } from '@/components/ui/Card'
import { useCreateEvent } from '@/hooks/useGroups'

interface EventFormProps {
  groupId: string
}

export function EventForm({ groupId }: EventFormProps) {
  const [open, set_open] = useState(false)
  const [title, set_title] = useState('')
  const [description, set_description] = useState('')
  const [datetime, set_datetime] = useState('')
  const [image, set_image] = useState<File | null>(null)
  const [preview_url, set_preview_url] = useState<string | null>(null)
  const [errors, set_errors] = useState<Record<string, string>>({})
  const create_event = useCreateEvent(groupId)

  const clear_image = useCallback(() => {
    if (preview_url) URL.revokeObjectURL(preview_url)
    set_image(null)
    set_preview_url(null)
  }, [preview_url])

  const handle_select_image = (file: File) => {
    if (preview_url) URL.revokeObjectURL(preview_url)
    set_image(file)
    set_preview_url(URL.createObjectURL(file))
  }

  async function handle_submit(e: React.FormEvent) {
    e.preventDefault()
    const new_errors: Record<string, string> = {}

    if (!title.trim()) new_errors.title = 'Title is required'
    else if (title.trim().length < 5) new_errors.title = 'Title must be at least 5 characters'

    if (!datetime) {
      new_errors.datetime = 'Date and time are required'
    } else {
      const dt = new Date(datetime)
      if (isNaN(dt.getTime())) {
        new_errors.datetime = 'Invalid date and time'
      } else if (dt <= new Date()) {
        new_errors.datetime = 'Date and time must be in the future'
      }
    }

    if (Object.keys(new_errors).length > 0) {
      set_errors(new_errors)
      return
    }

    try {
      await create_event.mutateAsync({
        title: title.trim(),
        description: description.trim() || undefined,
        event_datetime: new Date(datetime).toISOString(),
        image: image ?? undefined,
      })
      set_title('')
      set_description('')
      set_datetime('')
      clear_image()
      set_errors({})
    } catch {
      // error toast handled by mutation
    }
  }

  return (
    <Card>
      <CardContent>
        <button
          type="button"
          onClick={() => set_open((o) => !o)}
          className="mb-3 flex w-full items-center justify-between text-sm font-semibold text-gray-100"
        >
          <span>Create an event</span>
          <span className="text-gray-300">{open ? '−' : '+'}</span>
        </button>
        {open && (
          <form onSubmit={handle_submit} className="space-y-3">
            {preview_url && (
              <div className="relative">
                {/* eslint-disable-next-line @next/next/no-img-element */}
                <img
                  src={preview_url}
                  alt="Event image preview"
                  className="max-h-48 w-full rounded-lg object-contain"
                />
                <button
                  type="button"
                  onClick={clear_image}
                  className="absolute right-2 top-2 rounded-full bg-black/60 px-2 py-1 text-xs text-white hover:bg-black/80"
                >
                  Remove
                </button>
              </div>
            )}
            <Input
              id="event-title"
              label="Title"
              value={title}
              onChange={(e) => set_title(e.target.value)}
              placeholder="e.g. Go Meetup: Concurrency Patterns"
              error={errors.title}
            />
            <Input
              id="event-description"
              label="Description (optional)"
              value={description}
              onChange={(e) => set_description(e.target.value)}
              placeholder="What's this event about?"
            />
            <Input
              id="event-datetime"
              label="Date and time"
              type="datetime-local"
              value={datetime}
              onChange={(e) => set_datetime(e.target.value)}
              error={errors.datetime}
            />
            <div className="flex items-center gap-3">
              <ImageUpload on_select={handle_select_image}>
                <button
                  type="button"
                  className="rounded-lg px-2 py-1.5 text-base text-gray-300 hover:bg-purple-400/15 hover:text-gray-100"
                  title={image ? 'Change image' : 'Add image'}
                >
                  📷
                </button>
              </ImageUpload>
              {image && (
                <button
                  type="button"
                  onClick={clear_image}
                  className="text-sm text-red-600 hover:text-red-200"
                >
                  Remove photo
                </button>
              )}
            </div>
            <Button type="submit" loading={create_event.isPending} disabled={!title.trim()}>
              Create event
            </Button>
          </form>
        )}
      </CardContent>
    </Card>
  )
}
