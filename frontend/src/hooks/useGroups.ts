'use client'

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import {
  browseGroups,
  getGroup,
  createGroup as apiCreateGroup,
  getGroupEvents,
  createEvent as apiCreateEvent,
  rsvpEvent as apiRsvpEvent,
  acceptGroupMember as apiAcceptGroupMember,
  rejectGroupMember as apiRejectGroupMember,
  joinGroup as apiJoinGroup,
  leaveGroup as apiLeaveGroup,
  inviteGroupMember as apiInviteGroupMember,
  uploadGroupAvatar as apiUploadGroupAvatar,
} from '@/api/groups'
import { useUI } from '@/context/UIProvider'
import { useNotifications } from '@/context/NotificationProvider'
import type { CreateEventInput, GroupEvent } from '@/types/group'

export function useGroups() {
  return useQuery({
    queryKey: ['groups'],
    queryFn: () => browseGroups(),
    staleTime: 60_000,
  })
}

export function useCreateGroup() {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: (input: { title: string; description?: string; image?: File }) => apiCreateGroup(input),
    onSuccess: (id) => {
      query_client.invalidateQueries({ queryKey: ['groups'] })
      query_client.invalidateQueries({ queryKey: ['group', id] })
      show_toast({ message: 'Group created', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to create group',
        type: 'error',
      })
    },
  })
}

export function useGroup(id: string) {
  return useQuery({
    queryKey: ['group', id],
    queryFn: () => getGroup(id),
    enabled: !!id,
  })
}

export function useGroupEvents(groupId: string, enabled = true) {
  return useQuery({
    queryKey: ['group-events', groupId],
    queryFn: () => getGroupEvents(groupId),
    enabled: !!groupId && enabled,
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

export function useAcceptGroupMember(groupId: string) {
  const query_client = useQueryClient()
  const { show_toast } = useUI()
  const { refresh: refresh_notifications } = useNotifications()

  return useMutation({
    mutationFn: (userId: string) => apiAcceptGroupMember(groupId, userId),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['group', groupId] })
      refresh_notifications()
      show_toast({ message: 'Member accepted', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to accept member',
        type: 'error',
      })
    },
  })
}

export function useRejectGroupMember(groupId: string) {
  const query_client = useQueryClient()
  const { show_toast } = useUI()
  const { refresh: refresh_notifications } = useNotifications()

  return useMutation({
    mutationFn: (userId: string) => apiRejectGroupMember(groupId, userId),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['group', groupId] })
      refresh_notifications()
      show_toast({ message: 'Member declined', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to decline member',
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

export function useJoinGroup(groupId: string) {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: () => apiJoinGroup(groupId),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['group', groupId] })
      show_toast({ message: 'Join request sent', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to join group',
        type: 'error',
      })
    },
  })
}

export function useLeaveGroup(groupId: string) {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: () => apiLeaveGroup(groupId),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['group', groupId] })
      show_toast({ message: 'Left the group', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to leave group',
        type: 'error',
      })
    },
  })
}

export function useInviteGroupMember(groupId: string) {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: (nickname: string) => apiInviteGroupMember(groupId, nickname),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['group', groupId] })
      show_toast({ message: 'Invitation sent', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to send invitation',
        type: 'error',
      })
    },
  })
}

export function useUpdateGroupAvatar(groupId: string) {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: (file: File) => apiUploadGroupAvatar(groupId, file),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['group', groupId] })
      query_client.invalidateQueries({ queryKey: ['groups'] })
      show_toast({ message: 'Avatar updated', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to update avatar',
        type: 'error',
      })
    },
  })
}
