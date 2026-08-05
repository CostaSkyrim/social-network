'use client'

import { useEffect, useRef, useCallback, useState } from 'react'

const API_URL =
  process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

function getWsURL(): string {
  const url = new URL(API_URL)
  url.protocol = url.protocol === 'https:' ? 'wss:' : 'ws:'
  url.pathname = '/api/ws'
  return url.toString()
}

interface WSMessage {
  type: string
  payload?: unknown
  sender_id?: number
  timestamp: string
}

type MessageHandler = (msg: WSMessage) => void

let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
const handlers = new Map<string, Set<MessageHandler>>()

function connect() {
  if (ws?.readyState === WebSocket.OPEN) return

  ws = new WebSocket(getWsURL())

  ws.onopen = () => {
    listeners.forEach((l) => l(true))
  }

  ws.onmessage = (event) => {
    try {
      const msg: WSMessage = JSON.parse(event.data)
      const h = handlers.get(msg.type)
      if (h) {
        h.forEach((fn) => fn(msg))
      }
    } catch {
      // ignore parse errors
    }
  }

  ws.onclose = () => {
    listeners.forEach((l) => l(false))
    ws = null
    reconnectTimer = setTimeout(() => {
      reconnectTimer = null
      connect()
    }, 3000)
  }

  ws.onerror = () => {
    ws?.close()
  }
}

function disconnect() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
  ws?.close()
  ws = null
}

const listeners = new Set<(connected: boolean) => void>()

export function useWebSocket() {
  const [is_connected, set_is_connected] = useState(false)

  useEffect(() => {
    const listener = (connected: boolean) => set_is_connected(connected)
    listeners.add(listener)
    if (ws?.readyState === WebSocket.OPEN) {
      set_is_connected(true)
    } else if (!ws || ws.readyState === WebSocket.CLOSED) {
      connect()
    }
    return () => {
      listeners.delete(listener)
    }
  }, [])

  const subscribe = useCallback(
    (type: string, handler: MessageHandler) => {
      if (!handlers.has(type)) {
        handlers.set(type, new Set())
      }
      handlers.get(type)!.add(handler)

      return () => {
        handlers.get(type)?.delete(handler)
      }
    },
    [],
  )

  return { is_connected, subscribe }
}

export { connect as connectWebSocket, disconnect as disconnectWebSocket }
