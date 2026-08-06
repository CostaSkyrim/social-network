'use client'

import { useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { useCreateGroup } from '@/hooks/useGroups'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { ImageUpload } from '@/components/common/ImageUpload'
import { Card, CardContent, CardHeader } from '@/components/ui/Card'

export default function CreateGroupPage() {
  const router = useRouter()
  const create_group = useCreateGroup()
  const [title, set_title] = useState('')
  const [description, set_description] = useState('')
  const [image, set_image] = useState<File | null>(null)
  const [preview_url, set_preview_url] = useState<string | null>(null)
  const [errors, set_errors] = useState<Record<string, string>>({})

  const clear_image = useCallback(() => {
    if (preview_url) URL.revokeObjectURL(preview_url)
    set_image(null)
    set_preview_url(null)
  }, [preview_url])

  const handle_select_image = (file: File) => {
    if (preview_url) URL.revokeObjectURL(preview_url)
    set_image(file)
    set_preview_url(URL.createObjectURL(file))
  }

  async function handle_submit(e: React.FormEvent) {
    e.preventDefault()
    const new_errors: Record<string, string> = {}

    if (!title.trim()) new_errors.title = 'Title is required'

    if (Object.keys(new_errors).length > 0) {
      set_errors(new_errors)
      return
    }

    try {
      const id = await create_group.mutateAsync({
        title: title.trim(),
        description: description.trim() || undefined,
        image: image ?? undefined,
      })
      router.push(`/groups/${id}`)
    } catch {
      // error toast handled by mutation
    }
  }

  return (
    <div className="space-y-6">
      <h2 className="text-xl font-semibold text-gray-900">Create Group</h2>

      <Card>
        <CardHeader>
          <h3 className="text-sm font-semibold text-gray-900">Group details</h3>
        </CardHeader>
        <CardContent>
          <form onSubmit={handle_submit} className="space-y-4">
            <div className="flex items-center gap-4">
              <ImageUpload on_select={handle_select_image}>
                <button
                  type="button"
                  className="group relative block h-20 w-20 overflow-hidden rounded-full border-2 border-dashed border-gray-300 hover:border-gray-400"
                  title="Add group photo"
                >
                  {preview_url ? (
                    /* eslint-disable-next-line @next/next/no-img-element */
                    <img
                      src={preview_url}
                      alt="Group avatar preview"
                      className="h-full w-full object-cover"
                    />
                  ) : (
                    <span className="flex h-full w-full items-center justify-center text-2xl text-gray-400">
                      📷
                    </span>
                  )}
                  <span className="absolute inset-0 flex items-center justify-center bg-black/40 text-xs font-medium text-white opacity-0 transition-opacity group-hover:opacity-100">
                    {image ? 'Change' : 'Add photo'}
                  </span>
                </button>
              </ImageUpload>
              {image && (
                <button
                  type="button"
                  onClick={clear_image}
                  className="text-sm text-red-600 hover:text-red-700"
                >
                  Remove photo
                </button>
              )}
            </div>
            <Input
              id="group-title"
              label="Title"
              value={title}
              onChange={(e) => set_title(e.target.value)}
              placeholder="e.g. Golang Enthusiasts"
              error={errors.title}
              maxLength={300}
            />
            <Input
              id="group-description"
              label="Description (optional)"
              value={description}
              onChange={(e) => set_description(e.target.value)}
              placeholder="What is this group about?"
            />
            <div className="flex items-center gap-2">
              <Button type="submit" loading={create_group.isPending} disabled={!title.trim()}>
                Create group
              </Button>
              <Button type="button" variant="ghost" onClick={() => router.push('/groups')}>
                Cancel
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
