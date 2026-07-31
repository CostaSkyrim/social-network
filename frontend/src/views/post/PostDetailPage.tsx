'use client'

import { useParams } from 'next/navigation'
import { usePost } from '@/hooks/usePosts'
import { useComments } from '@/hooks/useComments'
import { PostCard } from '@/components/post/PostCard'
import { CommentList } from '@/components/comment/CommentList'
import { Card, CardContent, CardHeader } from '@/components/ui/Card'
import { Spinner } from '@/components/ui/Spinner'

export default function PostDetailPage() {
  const params = useParams()
  const id = (params?.uuid as string) || ''
  const { data: post, isLoading: post_loading, isError } = usePost(id)
  const { data: comments, isLoading: comments_loading } = useComments(id)

  if (post_loading) {
    return (
      <div className="flex justify-center py-16">
        <Spinner size="lg" />
      </div>
    )
  }

  if (isError || !post) {
    return (
      <div className="flex flex-col items-center justify-center py-16 text-center">
        <h2 className="text-lg font-semibold text-gray-900">Post not found</h2>
        <p className="mt-1 text-sm text-gray-500">This post may have been deleted.</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <PostCard post={post} />
      <Card>
        <CardHeader>
          <h3 className="text-sm font-semibold text-gray-900">
            Comments {comments ? `(${comments.length})` : ''}
          </h3>
        </CardHeader>
        <CardContent>
          <CommentList
            comments={comments ?? []}
            is_loading={comments_loading}
          />
        </CardContent>
      </Card>
    </div>
  )
}
