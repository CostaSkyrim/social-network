'use client'

import { PostCard } from '@/components/post/PostCard'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { EmptyState } from '@/components/common/EmptyState'
import type { Post } from '@/types/post'

interface PostGridProps {
  posts: Post[]
  is_loading: boolean
  has_next: boolean | undefined
  on_load_more: () => void
  empty_title?: string
  empty_description?: string
}

export function PostGrid({
  posts,
  is_loading,
  has_next,
  on_load_more,
  empty_title = 'No posts yet',
  empty_description = 'Nothing to show here right now.',
}: PostGridProps) {
  if (!is_loading && posts.length === 0) {
    return <EmptyState title={empty_title} description={empty_description} />
  }

  return (
    <div className="space-y-4">
      {/* Grid flows left-to-right, top-to-bottom (row-major): 1 col on phones,
          2 on md, 3 on lg. */}
      <div className="grid grid-cols-1 items-start gap-4 md:grid-cols-2 lg:grid-cols-3">
        {posts.map((post) => (
          <PostCard key={post.id} post={post} />
        ))}
      </div>

      {has_next && (
        <div className="flex justify-center py-2">
          <Button variant="outline" onClick={on_load_more} disabled={is_loading}>
            {is_loading ? <Spinner size="sm" /> : 'Load more'}
          </Button>
        </div>
      )}

      {is_loading && posts.length > 0 && (
        <div className="flex justify-center py-2">
          <Spinner />
        </div>
      )}
    </div>
  )
}
