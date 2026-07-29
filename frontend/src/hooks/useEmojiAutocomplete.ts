'use client'

import { useMemo, useState, useCallback } from 'react'
import data from '@emoji-mart/data'

interface EmojiMatch {
  id: string
  native: string
  name: string
}

const emoji_data = data as {
  emojis: Record<string, { id: string; name: string; keywords: string[]; skins: { native: string }[] }>
}

function build_index(): EmojiMatch[] {
  const list: EmojiMatch[] = []
  for (const key of Object.keys(emoji_data.emojis)) {
    const e = emoji_data.emojis[key]
    list.push({
      id: e.id,
      native: e.skins[0]?.native ?? '',
      name: e.name,
    })
  }
  return list
}

let index: EmojiMatch[] | null = null

function get_index(): EmojiMatch[] {
  if (!index) index = build_index()
  return index
}

export function useEmojiAutocomplete(text: string, cursor_pos: number) {
  const [selected_index, set_selected_index] = useState(0)

  const { word, start } = useMemo(() => {
    if (cursor_pos <= 0) return { word: null, start: -1 }

    const before = text.slice(0, cursor_pos)
    const last_colon = before.lastIndexOf(':')

    if (last_colon === -1 || last_colon === cursor_pos - 1) {
      return { word: null, start: -1 }
    }

    const candidate = before.slice(last_colon + 1)
    if (!candidate || candidate.length < 1 || candidate.includes(' ')) {
      return { word: null, start: -1 }
    }

    return { word: candidate.toLowerCase(), start: last_colon }
  }, [text, cursor_pos])

  const matches = useMemo(() => {
    if (!word) return []
    const all = get_index()
    const results = all.filter(
      (e) =>
        e.id.includes(word) ||
        e.name.toLowerCase().includes(word) ||
        e.native === word,
    )
    return results.slice(0, 10)
  }, [word])

  const reset = useCallback(() => set_selected_index(0), [])

  if (word && matches.length > 0 && selected_index >= matches.length) {
    set_selected_index(0)
  }

  return {
    word,
    start,
    matches,
    selected_index,
    set_selected_index,
    reset,
  }
}
