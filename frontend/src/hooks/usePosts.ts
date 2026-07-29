'use client'

import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getFeed, getPost, createPost as apiCreatePost, deletePost as apiDeletePost } from '@/api/posts'
import { useUI } from '@/context/UIProvider'

const FEED_PAGE_SIZE = 10

export function useFeed() {
  return useInfiniteQuery({
    queryKey: ['feed'],
    queryFn: ({ pageParam }) => getFeed(pageParam, FEED_PAGE_SIZE),
    getNextPageParam: (lastPage, allPages) => {
      if (lastPage.length < FEED_PAGE_SIZE) return undefined
      return allPages.length + 1
    },
    initialPageParam: 1,
    staleTime: 30_000,
  })
}

export function usePost(id: number) {
  return useQuery({
    queryKey: ['post', id],
    queryFn: () => getPost(id),
    enabled: !!id,
  })
}

export function useCreatePost() {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: (data: { content: string; privacy_level: string; group_id?: number }) =>
      apiCreatePost(data),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['feed'] })
      show_toast({ message: 'Post created', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to create post',
        type: 'error',
      })
    },
  })
}

export function useDeletePost() {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: (id: number) => apiDeletePost(id),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['feed'] })
      show_toast({ message: 'Post deleted', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to delete post',
        type: 'error',
      })
    },
  })
}
