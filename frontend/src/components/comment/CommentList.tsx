'use client'

import { CommentItem } from './CommentItem'
import { Spinner } from '@/components/ui/Spinner'
import type { Comment } from '@/types/comment'

interface CommentNode extends Comment {
  children: CommentNode[]
}

interface CommentListProps {
  comments: Comment[]
  is_loading: boolean
}

function buildCommentTree(comments: Comment[]): CommentNode[] {
  const map = new Map<string, CommentNode>()
  const roots: CommentNode[] = []

  for (const c of comments) {
    map.set(c.id, { ...c, children: [] })
  }

  for (const c of comments) {
    const node = map.get(c.id)!
    if (c.parent_comment_id && map.has(c.parent_comment_id)) {
      map.get(c.parent_comment_id)!.children.push(node)
    } else {
      roots.push(node)
    }
  }

  return roots
}

function renderTree(nodes: CommentNode[], depth = 0) {
  return nodes.map((node) => (
    <div key={node.id}>
      <CommentItem comment={node} depth={depth} />
      {node.children.length > 0 && (
        <div className="mt-2 space-y-2">
          {renderTree(node.children, depth + 1)}
        </div>
      )}
    </div>
  ))
}

export function CommentList({ comments, is_loading }: CommentListProps) {
  if (is_loading) {
    return (
      <div className="flex justify-center py-4">
        <Spinner />
      </div>
    )
  }

  if (comments.length === 0) {
    return (
      <p className="py-4 text-center text-sm text-gray-300">No comments yet.</p>
    )
  }

  const tree = buildCommentTree(comments)

  return (
    <div className="space-y-2">
      {renderTree(tree)}
    </div>
  )
}
