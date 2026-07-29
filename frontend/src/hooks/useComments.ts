'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { getComments, createComment as apiCreateComment, deleteComment as apiDeleteComment, editComment as apiEditComment } from '@/api/comments'
import { useUI } from '@/context/UIProvider'

export function useComments(postId: number) {
  return useQuery({
    queryKey: ['comments', postId],
    queryFn: () => getComments(postId),
    enabled: !!postId,
  })
}

export function useCreateComment() {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: (data: { post_id: number; content: string; image_path?: string }) =>
      apiCreateComment(data),
    onSuccess: (_data, variables) => {
      query_client.invalidateQueries({ queryKey: ['comments', variables.post_id] })
      query_client.invalidateQueries({ queryKey: ['feed'] })
      show_toast({ message: 'Comment added', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to add comment',
        type: 'error',
      })
    },
  })
}

export function useDeleteComment() {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: (id: number) => apiDeleteComment(id),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['comments'] })
      show_toast({ message: 'Comment deleted', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to delete comment',
        type: 'error',
      })
    },
  })
}

export function useEditComment() {
  const query_client = useQueryClient()
  const { show_toast } = useUI()

  return useMutation({
    mutationFn: ({ id, content, image_path }: { id: number; content: string; image_path?: string }) =>
      apiEditComment(id, { content, image_path }),
    onSuccess: () => {
      query_client.invalidateQueries({ queryKey: ['comments'] })
      show_toast({ message: 'Comment updated', type: 'success' })
    },
    onError: (err: any) => {
      show_toast({
        message: err?.response?.data?.error || 'Failed to update comment',
        type: 'error',
      })
    },
  })
}
