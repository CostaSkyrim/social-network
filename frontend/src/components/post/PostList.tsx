'use client'

import { PostCard } from './PostCard'
import { Spinner } from '@/components/ui/Spinner'
import { EmptyState } from '@/components/common/EmptyState'
import type { Post } from '@/types/post'

interface PostListProps {
  posts: Post[]
  is_loading: boolean
  has_next: boolean | undefined
  on_load_more: () => void
}

export function PostList({ posts, is_loading, has_next, on_load_more }: PostListProps) {
  if (!is_loading && posts.length === 0) {
    return <EmptyState title="No posts yet" description="Be the first to post something!" />
  }

  return (
    <div className="space-y-4">
      {posts.map((post) => (
        <PostCard key={post.id} post={post} />
      ))}

      {has_next && (
        <div className="flex justify-center py-4">
          <button
            onClick={on_load_more}
            disabled={is_loading}
            className="rounded-lg bg-purple-400/15 px-4 py-2 text-sm text-gray-300 hover:bg-purple-400/20"
          >
            {is_loading ? <Spinner size="sm" /> : 'Load more'}
          </button>
        </div>
      )}

      {is_loading && posts.length > 0 && (
        <div className="flex justify-center py-4">
          <Spinner />
        </div>
      )}
    </div>
  )
}
