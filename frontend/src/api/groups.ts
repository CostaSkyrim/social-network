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

export async function getUserGroupEvents(
  page: number,
  limit: number = 10,
): Promise<GroupEvent[]> {
  const offset = (page - 1) * limit
  const res = await client.get<ListResponse<GroupEvent>>('/api/events', {
    params: { limit, offset },
  })
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

export async function kickGroupMember(groupId: string, userId: string): Promise<void> {
  await client.post(`/api/groups/${groupId}/kick`, { user_id: userId })
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

export interface GroupMessage {
  id: number
  uuid: string
  sender_id: number
  sender?: {
    id?: string
    first_name: string
    last_name: string
    nickname?: string
    avatar_path?: string
  }
  content: string
  image_path?: string
  is_read: boolean
  created_at: string
}

export async function fetchGroupMessages(
  groupId: string,
  limit = 50,
): Promise<GroupMessage[]> {
  const res = await client.get<SingleResponse<GroupMessage[]>>(
    `/api/groups/${groupId}/messages`,
    { params: { limit } },
  )
  return res.data.data ?? []
}

export async function sendGroupMessage(
  groupId: string,
  content: string,
  image?: File,
): Promise<{ id: number; group_id: number; content: string; sender_id: number }> {
  const form = new FormData()
  form.append('content', content)
  if (image) form.append('image', image)
  const res = await client.post<SingleResponse<{ id: number; group_id: number; content: string; sender_id: number }>>(
    `/api/groups/${groupId}/messages/send`,
    form,
    { headers: { 'Content-Type': 'multipart/form-data' } },
  )
  if (!res.data.data) throw new Error('Failed to send message')
  return res.data.data
}

export async function updateEvent(
  eventId: string,
  input: {
    title: string
    description?: string
    event_datetime: string
    image?: File
    remove_image?: boolean
  },
): Promise<void> {
  const form = new FormData()
  form.append('title', input.title)
  if (input.description) form.append('description', input.description)
  form.append('event_datetime', input.event_datetime)
  if (input.image) form.append('image', input.image)
  if (input.remove_image) form.append('remove_image', '1')
  await client.put(`/api/events/${eventId}/edit`, form, {
    headers: { 'Content-Type': 'multipart/form-data' },
  })
}

export async function deleteEvent(eventId: string): Promise<void> {
  await client.delete(`/api/events/${eventId}/delete`)
}
