'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { Card, CardContent } from '@/components/ui/Card'
import { useCreateEvent } from '@/hooks/useGroups'

interface EventFormProps {
  groupId: string
}

export function EventForm({ groupId }: EventFormProps) {
  const [title, set_title] = useState('')
  const [description, set_description] = useState('')
  const [datetime, set_datetime] = useState('')
  const [errors, set_errors] = useState<Record<string, string>>({})
  const create_event = useCreateEvent(groupId)

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
      })
      set_title('')
      set_description('')
      set_datetime('')
      set_errors({})
    } catch {
      // error toast handled by mutation
    }
  }

  return (
    <Card>
      <CardContent>
        <h3 className="mb-3 text-sm font-semibold text-gray-900">Create an event</h3>
        <form onSubmit={handle_submit} className="space-y-3">
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
          <Button type="submit" loading={create_event.isPending} disabled={!title.trim()}>
            Create event
          </Button>
        </form>
      </CardContent>
    </Card>
  )
}
