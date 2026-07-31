'use client'

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  browseGroups,
  getGroup,
  getGroupEvents,
  createEvent as apiCreateEvent,
  rsvpEvent as apiRsvpEvent,
} from '@/api/groups'
import { useUI } from '@/context/UIProvider'
import type { CreateEventInput, GroupEvent } from '@/types/group'

export function useGroups() {
  return useQuery({
    queryKey: ['groups'],
    queryFn: () => browseGroups(),
    staleTime: 60_000,
  })
}

export function useGroup(id: string) {
  return useQuery({
    queryKey: ['group', id],
    queryFn: () => getGroup(id),
    enabled: !!id,
  })
}

export function useGroupEvents(groupId: string) {
  return useQuery({
    queryKey: ['group-events', groupId],
    queryFn: () => getGroupEvents(groupId),
    enabled: !!groupId,
  })
}

export function useCreateEvent(groupId: string) {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: (input: CreateEventInput) => apiCreateEvent(groupId, input),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['group-events', groupId] })
      show_toast({ message: 'Event created', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to create event',
        type: 'error',
      })
    },
  })
}

export function useRSVP(eventId: string, groupId: string) {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: (response: 'going' | 'not_going') => apiRsvpEvent(eventId, response),
    onMutate: async (response) => {
      await query_client.cancelQueries({ queryKey: ['group-events', groupId] })

      const previous = query_client.getQueryData<GroupEvent[]>(['group-events', groupId])

      query_client.setQueryData<GroupEvent[]>(['group-events', groupId], (old) => {
        if (!old) return old
        return old.map((ev) => {
          if (ev.id !== eventId) return ev

          const prev = ev.my_response
          const next = { ...ev, my_response: response }

          // adjust counts when changing/toggling response
          if (prev === 'going') next.going = Math.max(0, next.going - 1)
          if (prev === 'not_going') next.not_going = Math.max(0, next.not_going - 1)
          if (response === 'going') next.going = (next.going ?? 0) + 1
          if (response === 'not_going') next.not_going = (next.not_going ?? 0) + 1
          next.total = (next.going ?? 0) + (next.not_going ?? 0)

          return next
        })
      })

      return { previous }
    },
    onError: (err: any, _vars, context: any) => {
      if (context?.previous) {
        query_client.setQueryData(['group-events', groupId], context.previous)
      }
      show_toast({
        message: err?.response?.data?.error || 'Failed to update RSVP',
        type: 'error',
      })
    },
    onSettled: () => {
      query_client.invalidateQueries({ queryKey: ['group-events', groupId] })
    },
  })
}
