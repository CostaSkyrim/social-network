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

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const handlersRef = useRef<Map<string, Set<MessageHandler>>>(new Map())
  const reconnectTimeoutRef = useRef<ReturnType<typeof setTimeout>>()
  const [is_connected, set_is_connected] = useState(false)

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    const ws = new WebSocket(getWsURL())
    wsRef.current = ws

    ws.onopen = () => {
      set_is_connected(true)
    }

    ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        const handlers = handlersRef.current.get(msg.type)
        if (handlers) {
          handlers.forEach((h) => h(msg))
        }

        if (msg.type === 'notification') {
          const notifHandlers = handlersRef.current.get('notification')
          if (notifHandlers) {
            notifHandlers.forEach((h) => h(msg))
          }
        }
      } catch {
        // ignore parse errors
      }
    }

    ws.onclose = () => {
      set_is_connected(false)
      wsRef.current = null
      reconnectTimeoutRef.current = setTimeout(() => {
        connect()
      }, 3000)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [])

  useEffect(() => {
    connect()
    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current)
      }
      wsRef.current?.close()
    }
  }, [connect])

  const subscribe = useCallback(
    (type: string, handler: MessageHandler) => {
      if (!handlersRef.current.has(type)) {
        handlersRef.current.set(type, new Set())
      }
      handlersRef.current.get(type)!.add(handler)

      return () => {
        handlersRef.current.get(type)?.delete(handler)
      }
    },
    [],
  )

  return { is_connected, subscribe }
}
