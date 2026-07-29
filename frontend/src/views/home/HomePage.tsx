'use client'

import { useMemo } from 'react'
import { PostForm } from '@/components/post/PostForm'
import { PostList } from '@/components/post/PostList'
import { useFeed } from '@/hooks/usePosts'

export default function HomePage() {
  const feed = useFeed()

  const all_posts = useMemo(
    () => feed.data?.pages.flatMap((page) => page) ?? [],
    [feed.data],
  )

  return (
    <div className="space-y-6">
      <PostForm />
      <PostList
        posts={all_posts}
        is_loading={feed.isLoading}
        has_next={feed.hasNextPage}
        on_load_more={() => feed.fetchNextPage()}
      />
    </div>
  )
}
