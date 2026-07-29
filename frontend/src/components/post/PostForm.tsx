'use client'

import { useState } from 'react'
import { Button } from '@/components/ui/Button'
import { Card, CardContent } from '@/components/ui/Card'
import { PrivacySelector } from './PrivacySelector'
import { useCreatePost } from '@/hooks/usePosts'
import { useAuth } from '@/context/AuthProvider'
import { Avatar } from '@/components/ui/Avatar'

export function PostForm() {
  const [content, set_content] = useState('')
  const [privacy, set_privacy] = useState('public')
  const create_post = useCreatePost()
  const { user } = useAuth()

  async function handle_submit(e: React.FormEvent) {
    e.preventDefault()
    if (!content.trim()) return
    await create_post.mutateAsync({ content: content.trim(), privacy_level: privacy })
    set_content('')
    set_privacy('public')
  }

  return (
    <Card>
      <CardContent>
        <form onSubmit={handle_submit} className="space-y-3">
          <div className="flex items-start gap-3">
            <Avatar
              src={user?.avatar_path}
              alt={`${user?.first_name} ${user?.last_name}`}
              size="md"
            />
            <textarea
              value={content}
              onChange={(e) => set_content(e.target.value)}
              placeholder="What's on your mind?"
              rows={3}
              className="flex-1 resize-none rounded-lg border border-gray-300 p-3 text-sm outline-none focus:border-blue-500 focus:ring-1 focus:ring-blue-200"
              maxLength={7000}
            />
          </div>

          <div className="flex items-center justify-between">
            <PrivacySelector value={privacy} onChange={set_privacy} />
            <Button
              type="submit"
              size="sm"
              loading={create_post.isPending}
              disabled={!content.trim()}
            >
              Post
            </Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
