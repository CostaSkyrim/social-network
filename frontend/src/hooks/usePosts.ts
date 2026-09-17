'use client'

import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { getFeed, getFollowingPosts, getExplorePosts, getGroupPosts, getUserPosts, getPost, createPost as apiCreatePost, deletePost as apiDeletePost, editPost as apiEditPost } from '@/api/posts'
import { useUI } from '@/context/UIProvider'

const FEED_PAGE_SIZE = 10
const PROFILE_PAGE_SIZE = 5

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

export function useFollowingPosts() {
  return useInfiniteQuery({
    queryKey: ['following-posts'],
    queryFn: ({ pageParam }) => getFollowingPosts(pageParam, FEED_PAGE_SIZE),
    getNextPageParam: (lastPage, allPages) => {
      if (lastPage.length < FEED_PAGE_SIZE) return undefined
      return allPages.length + 1
    },
    initialPageParam: 1,
    staleTime: 30_000,
  })
}

export function useExplorePosts() {
  return useInfiniteQuery({
    queryKey: ['explore-posts'],
    queryFn: ({ pageParam }) => getExplorePosts(pageParam, FEED_PAGE_SIZE),
    getNextPageParam: (lastPage, allPages) => {
      if (lastPage.length < FEED_PAGE_SIZE) return undefined
      return allPages.length + 1
    },
    initialPageParam: 1,
    staleTime: 30_000,
  })
}

export function useGroupPosts(groupId: string) {
  return useInfiniteQuery({
    queryKey: ['group-posts', groupId],
    queryFn: ({ pageParam }) => getGroupPosts(groupId, pageParam, FEED_PAGE_SIZE),
    getNextPageParam: (lastPage, allPages) => {
      if (lastPage.length < FEED_PAGE_SIZE) return undefined
      return allPages.length + 1
    },
    initialPageParam: 1,
    enabled: !!groupId,
    staleTime: 30_000,
  })
}

export function useUserPosts(uuid: string) {
  return useInfiniteQuery({
    queryKey: ['user-posts', uuid],
    queryFn: ({ pageParam }) => getUserPosts(uuid, pageParam, PROFILE_PAGE_SIZE),
    getNextPageParam: (lastPage, allPages) => {
      if (lastPage.length < PROFILE_PAGE_SIZE) return undefined
      return allPages.length + 1
    },
    initialPageParam: 1,
    enabled: !!uuid,
    staleTime: 30_000,
  })
}

export function usePost(id: string) {
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
    mutationFn: (data: { content: string; privacy_level: string; group_id?: string; image?: File; visible_user_ids?: string[] }) =>
      apiCreatePost(data),
    onSuccess: (_data, variables) => {
      query_client.invalidateQueries({ queryKey: ['feed'] })
      if (variables.group_id) {
        query_client.invalidateQueries({ queryKey: ['group-posts', variables.group_id] })
      }
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
    mutationFn: (id: string) => apiDeletePost(id),
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

export function useEditPost() {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: ({ id, content, privacy_level, image, remove_image, visible_user_ids }: { id: string; content: string; privacy_level: string; image?: File; remove_image?: boolean; visible_user_ids?: string[] }) =>
      apiEditPost(id, { content, privacy_level, image, remove_image, visible_user_ids }),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['feed'] })
      query_client.invalidateQueries({ queryKey: ['post'] })
      query_client.invalidateQueries({ queryKey: ['user-posts'] })
      show_toast({ message: 'Post updated', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to update post',
        type: 'error',
      })
    },
  })
}
