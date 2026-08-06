'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useGroups } from '@/hooks/useGroups'
import { Card, CardContent } from '@/components/ui/Card'
import { Avatar } from '@/components/ui/Avatar'
import { Spinner } from '@/components/ui/Spinner'
import { EmptyState } from '@/components/common/EmptyState'

export default function GroupsPage() {
  const { data: groups, isLoading, isError } = useGroups()
  const router = useRouter()

  if (isLoading) {
    return (
      <div className="flex justify-center py-16">
        <Spinner size="lg" />
      </div>
    )
  }

  if (isError) {
    return <EmptyState title="Failed to load groups" description="Please try again later." />
  }

  if (!groups || groups.length === 0) {
    return (
      <div className="space-y-6">
        <GroupsHeader />
        <EmptyState
          title="No groups yet"
          description="Create the first group to get started!"
          action={{ label: '+ Create group', onClick: () => router.push('/groups/new') }}
        />
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <GroupsHeader />
      <div className="space-y-10">
        {groups.map((group) => (
          <Link key={group.id} href={`/groups/${group.id}`}>
            <Card className="transition-shadow hover:shadow-md">
              <CardContent className="flex items-center gap-4 px-6 py-6">
                <Avatar src={group.avatar_path} alt={group.title} size="md" />
                <div className="min-w-0">
                  <h3 className="text-base font-semibold text-gray-900">{group.title}</h3>
                  {group.description && (
                    <p className="mt-1 text-sm text-gray-500 line-clamp-2">{group.description}</p>
                  )}
                </div>
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  )
}

function GroupsHeader() {
  return (
    <div className="flex items-center justify-between">
      <h2 className="text-xl font-semibold text-gray-900">Groups</h2>
      <Link
        href="/groups/new"
        className="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
      >
        + Create group
      </Link>
    </div>
  )
}
