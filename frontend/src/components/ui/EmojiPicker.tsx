'use client'

import { useState, useRef, useEffect } from 'react'
import EmojiPickerReact, { type EmojiClickData } from 'emoji-picker-react'
import { cn } from '@/lib/cn'

interface EmojiPickerProps {
  on_select: (emoji: string) => void
  /** Open above the trigger (use in bottom-anchored composers). */
  direction?: 'up' | 'down'
}

export function EmojiPicker({ on_select, direction = 'down' }: EmojiPickerProps) {
  const [open, set_open] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    function on_click(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        set_open(false)
      }
    }
    document.addEventListener('mousedown', on_click)
    return () => document.removeEventListener('mousedown', on_click)
  }, [])

  function handle_emoji_click(data: EmojiClickData) {
    on_select(data.emoji)
    set_open(false)
  }

  return (
    <div ref={ref} className="relative">
      <button
        type="button"
        onClick={() => set_open(!open)}
        className="rounded-lg p-1.5 text-lg text-gray-300 hover:bg-purple-400/15"
        title="Add emoji"
      >
        😊
      </button>
      {open && (
        <div
          className={cn(
            'absolute z-50 -left-2',
            direction === 'up' ? 'bottom-full mb-2' : 'top-10',
          )}
          style={{ width: '320px' }}
        >
          <EmojiPickerReact
            onEmojiClick={handle_emoji_click}
            searchPlaceholder="Search emojis..."
            width={320}
            height={380}
          />
        </div>
      )}
    </div>
  )
}
