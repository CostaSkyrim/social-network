'use client'

import { useEffect, useRef } from 'react'

interface EmojiSuggestion {
  id: string
  native: string
  name: string
}

interface EmojiSuggestionsProps {
  matches: EmojiSuggestion[]
  selected_index: number
  on_select: (emoji: string) => void
}

export function EmojiSuggestions({ matches, selected_index, on_select }: EmojiSuggestionsProps) {
  const list_ref = useRef<HTMLUListElement>(null)

  useEffect(() => {
    const el = list_ref.current?.children[selected_index] as HTMLElement | undefined
    el?.scrollIntoView({ block: 'nearest' })
  }, [selected_index])

  if (matches.length === 0) return null

  return (
    <div className="absolute top-full left-0 z-50 mt-1 w-72 rounded-xl border border-purple-400/20 bg-[#241748] shadow-lg">
      <ul ref={list_ref} className="max-h-48 overflow-y-auto py-1">
        {matches.map((match, i) => (
          <li
            key={match.id}
            onClick={() => on_select(match.native)}
            className={`flex cursor-pointer items-center gap-3 px-3 py-1.5 text-sm ${
              i === selected_index ? 'bg-violet-400/20' : 'hover:bg-purple-400/10'
            }`}
          >
            <span className="text-lg">{match.native}</span>
            <span className="text-gray-300">:{match.id}:</span>
            <span className="ml-auto text-xs text-gray-300">{match.name}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}
