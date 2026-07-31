'use client'

import Link from 'next/link'
import { useGroups } from '@/hooks/useGroups'
import { Card, CardContent } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'
import { EmptyState } from '@/components/common/EmptyState'

export default function GroupsPage() {
  const { data: groups, isLoading, isError } = useGroups()

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
    return <EmptyState title="No groups yet" description="No groups are available right now." />
  }

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900">Groups</h2>
      <div className="space-y-4">
        {groups.map((group) => (
          <Link key={group.id} href={`/groups/${group.id}`}>
            <Card className="transition-shadow hover:shadow-md">
              <CardContent>
                <h3 className="text-sm font-semibold text-gray-900">{group.title}</h3>
                {group.description && (
                  <p className="mt-1 text-sm text-gray-500 line-clamp-2">{group.description}</p>
                )}
              </CardContent>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  )
}
