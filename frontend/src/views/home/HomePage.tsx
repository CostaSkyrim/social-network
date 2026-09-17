'use client'

import { useMemo } from 'react'
import { PostForm } from '@/components/post/PostForm'
import { useExplorePosts, useFollowingPosts } from '@/hooks/usePosts'
import { useUserGroupEvents } from '@/hooks/useGroups'
import { EventsCarousel } from '@/components/home/EventsCarousel'
import { PostGrid } from '@/components/home/PostGrid'

export default function HomePage() {
  const events = useUserGroupEvents()
  const following = useFollowingPosts()
  const explore = useExplorePosts()

  const all_events = useMemo(
    () => events.data?.pages.flatMap((page) => page) ?? [],
    [events.data],
  )
  const following_posts = useMemo(
    () => following.data?.pages.flatMap((page) => page) ?? [],
    [following.data],
  )
  const explore_posts = useMemo(
    () => explore.data?.pages.flatMap((page) => page) ?? [],
    [explore.data],
  )

  return (
    <div className="space-y-8">
      {/* 1) Recent events from the user's groups */}
      <section>
        <h2 className="mb-3 text-lg font-semibold text-gray-100">Upcoming events</h2>
        <EventsCarousel
          events={all_events}
          is_loading={events.isLoading}
          has_next={events.hasNextPage}
          on_load_more={() => events.fetchNextPage()}
        />
      </section>

      {/* 2) New post form */}
      <section>
        <PostForm />
      </section>

      {/* 3) Recent posts from followed users */}
      <section>
        <h2 className="mb-3 text-lg font-semibold text-gray-100">From people you follow</h2>
        <PostGrid
          posts={following_posts}
          is_loading={following.isLoading}
          has_next={following.hasNextPage}
          on_load_more={() => following.fetchNextPage()}
          empty_title="No posts from people you follow"
          empty_description="Posts from accounts you follow will show up here."
        />
      </section>

      {/* 4) Public posts from accounts you don't follow */}
      <section>
        <h2 className="mb-3 text-lg font-semibold text-gray-100">Explore</h2>
        <PostGrid
          posts={explore_posts}
          is_loading={explore.isLoading}
          has_next={explore.hasNextPage}
          on_load_more={() => explore.fetchNextPage()}
          empty_title="No posts to explore"
          empty_description="Public posts from accounts you don't follow will appear here."
        />
      </section>
    </div>
  )
}
