import client from './client'
import type { Group, GroupDetail, GroupEvent, CreateEventInput } from '@/types/group'

interface ListResponse<T> {
  message: string
  data?: T[]
}

interface SingleResponse<T> {
  message: string
  data?: T
}

export async function browseGroups(limit = 20, offset = 0): Promise<Group[]> {
  const res = await client.get<ListResponse<Group>>('/api/groups/browse', {
    params: { limit, offset },
  })
  return res.data.data ?? []
}

export async function getGroup(id: string): Promise<GroupDetail> {
  const res = await client.get<SingleResponse<GroupDetail>>(`/api/groups/${id}`)
  if (!res.data.data) throw new Error('Group not found')
  return res.data.data
}

export async function createGroup(input: {
  title: string
  description?: string
  image?: File
}): Promise<string> {
  const form = new FormData()
  form.append('title', input.title)
  if (input.description) form.append('description', input.description)
  if (input.image) form.append('image', input.image)
  const res = await client.post<SingleResponse<{ id: string }>>('/api/groups', form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
  if (!res.data.data?.id) throw new Error('Failed to create group')
  return res.data.data.id
}

export async function getGroupEvents(groupId: string): Promise<GroupEvent[]> {
  const res = await client.get<ListResponse<GroupEvent>>(`/api/groups/${groupId}/events`)
  return res.data.data ?? []
}

export async function createEvent(groupId: string, input: CreateEventInput): Promise<void> {
  const form = new FormData()
  form.append('title', input.title)
  if (input.description) form.append('description', input.description)
  form.append('event_datetime', input.event_datetime)
  if (input.image) form.append('image', input.image)
  await client.post(`/api/groups/${groupId}/events`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export async function rsvpEvent(
  eventId: string,
  response: 'going' | 'not_going',
): Promise<GroupEvent> {
  const res = await client.post<SingleResponse<GroupEvent>>(`/api/events/${eventId}/rsvp`, {
    response,
  })
  if (!res.data.data) throw new Error('Failed to update RSVP')
  return res.data.data
}

export async function acceptGroupMember(groupId: string, userId: string): Promise<void> {
  await client.post(`/api/groups/${groupId}/accept`, { user_id: userId })
}

export async function rejectGroupMember(groupId: string, userId: string): Promise<void> {
  await client.post(`/api/groups/${groupId}/reject`, { user_id: userId })
}

export async function joinGroup(groupId: string): Promise<void> {
  await client.post(`/api/groups/${groupId}/join`)
}

export async function leaveGroup(groupId: string): Promise<void> {
  await client.post(`/api/groups/${groupId}/leave`)
}

export async function inviteGroupMember(groupId: string, nickname: string): Promise<void> {
  await client.post(`/api/groups/${groupId}/invite`, { nickname })
}

export async function uploadGroupAvatar(groupId: string, file: File): Promise<string> {
  const form = new FormData()
  form.append('image', file)
  const res = await client.post<SingleResponse<{ avatar_path: string }>>(
    `/api/groups/${groupId}/avatar`,
    form,
    { headers: { 'Content-Type': 'multipart/form-data' } },
  )
  if (!res.data.data?.avatar_path) throw new Error('Failed to upload avatar')
  return res.data.data.avatar_path
}
