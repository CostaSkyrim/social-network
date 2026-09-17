'use client'

import { useState } from 'react'
import { useParams } from 'next/navigation'
import { useAuth } from '@/context/AuthProvider'
import { useGroup, useGroupEvents, useUpdateGroupAvatar } from '@/hooks/useGroups'
import { useGroupPosts } from '@/hooks/usePosts'
import { EventList } from '@/components/group/EventList'
import { EventForm } from '@/components/group/EventForm'
import { MemberList } from '@/components/group/MemberList'
import { GroupActions } from '@/components/group/GroupActions'
import { GroupChat } from '@/components/group/GroupChat'
import { InviteMember } from '@/components/group/InviteMember'
import { PostForm } from '@/components/post/PostForm'
import { PostList } from '@/components/post/PostList'
import { Avatar } from '@/components/ui/Avatar'
import { Card, CardContent, CardHeader } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import { EmptyState } from '@/components/common/EmptyState'
import { ImageUpload } from '@/components/common/ImageUpload'
import { cn } from '@/lib/cn'

export default function GroupDetailPage() {
  const params = useParams()
  const uuid = (params?.uuid as string) || ''
  const { user } = useAuth()

  const {
    data: group_data,
    isLoading: group_loading,
    isError: group_error,
  } = useGroup(uuid)

  const { data: events, isLoading: events_loading } = useGroupEvents(uuid)
  const upload_avatar = useUpdateGroupAvatar(uuid)
  const group_posts = useGroupPosts(uuid)
  const [active_tab, set_active_tab] = useState<'posts' | 'events' | 'chat' | 'members'>('posts')

  if (group_loading) {
    return (
      <div className="flex justify-center py-16">
        <Spinner size="lg" />
      </div>
    )
  }

  if (group_error || !group_data) {
    return (
      <EmptyState
        title="Group not found"
        description="This group may have been deleted, or you don't have access to it."
      />
    )
  }

  const { group, members } = group_data
  const accepted_count = members.filter((m) => m.status === 'accepted').length
  const currentUserId = user?.id ?? ''
  const is_creator = group.creator_id === currentUserId
  const is_member = members.some(
    (m) => m.user.id === currentUserId && m.status === 'accepted',
  )

  return (
    <div className="space-y-6">
      <Card>
        <CardHeader>
          <div className="flex items-start justify-between gap-3">
            <div className="flex items-start gap-4">
              {is_creator ? (
                <ImageUpload on_select={(file) => upload_avatar.mutate(file)}>
                  <button
                    type="button"
                    className="group relative block overflow-hidden rounded-full"
                    disabled={upload_avatar.isPending}
                    title="Change group photo"
                  >
                    <Avatar src={group.avatar_path} alt={group.title} size="lg" />
                    {upload_avatar.isPending ? (
                      <span className="absolute inset-0 flex items-center justify-center bg-black/50">
                        <Spinner size="sm" />
                      </span>
                    ) : (
                      <span className="absolute inset-0 flex items-center justify-center bg-black/40 text-xs font-medium text-white opacity-0 transition-opacity group-hover:opacity-100">
                        Change
                      </span>
                    )}
                  </button>
                </ImageUpload>
              ) : (
                <Avatar src={group.avatar_path} alt={group.title} size="lg" />
              )}
              <div>
                <h2 className="text-xl font-semibold text-gray-100">{group.title}</h2>
                {group.description && (
                  <p className="mt-1 text-sm text-gray-300">{group.description}</p>
                )}
              </div>
            </div>
            <span className="text-xs text-gray-300">{accepted_count} members</span>
          </div>
          <div className="mt-3">
            <GroupActions
              groupId={group.id}
              creatorId={group.creator_id}
              members={members}
              currentUserId={currentUserId}
            />
          </div>
        </CardHeader>
      </Card>

      <Card>
        <CardHeader>
          <div className="flex items-center gap-1 border-b border-purple-400/20">
            {([
              ['posts', 'Posts'],
              ['events', 'Events'],
              ['chat', 'Chat'],
              ['members', 'Members'],
            ] as const).map(([key, label]) => (
              <button
                key={key}
                type="button"
                onClick={() => set_active_tab(key)}
                className={cn(
                  'border-b-2 px-3 py-2 text-sm font-medium transition-colors',
                  active_tab === key
                    ? 'border-violet-300 text-violet-300'
                    : 'border-transparent text-gray-300 hover:text-gray-300',
                )}
              >
                {label}
              </button>
            ))}
          </div>
        </CardHeader>
        <CardContent className={active_tab === 'members' ? 'space-y-3' : 'space-y-4'}>
          {active_tab === 'posts' &&
            (is_member ? (
              <>
                <PostForm groupId={group.id} />
                <PostList
                  posts={group_posts.data?.pages.flatMap((page) => page) ?? []}
                  is_loading={group_posts.isLoading}
                  has_next={group_posts.hasNextPage}
                  on_load_more={() => group_posts.fetchNextPage()}
                />
              </>
            ) : (
              <p className="rounded-lg border border-purple-400/20 bg-[#241748] p-6 text-center text-sm text-gray-300">
                Join this group to see its posts.
              </p>
            ))}

          {active_tab === 'events' &&
            (is_member ? (
              <>
                <EventForm groupId={group.id} />
                <EventList
                  events={events ?? []}
                  groupId={group.id}
                  is_loading={events_loading}
                />
              </>
            ) : (
              <p className="rounded-lg border border-purple-400/20 bg-[#241748] p-6 text-center text-sm text-gray-300">
                Join this group to see its events.
              </p>
            ))}

          {active_tab === 'chat' &&
            (is_member ? (
              <GroupChat
                groupId={group.id}
                groupTitle={group.title}
                groupAvatar={group.avatar_path}
              />
            ) : (
              <p className="rounded-lg border border-purple-400/20 bg-[#241748] p-6 text-center text-sm text-gray-300">
                Join this group to participate in the chat.
              </p>
            ))}

          {active_tab === 'members' && (
            <>
              {is_member && <InviteMember groupId={group.id} />}
              <MemberList members={members} groupId={group.id} isCreator={is_creator} creatorId={group.creator_id} />
            </>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
