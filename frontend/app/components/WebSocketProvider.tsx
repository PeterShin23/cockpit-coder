import React, { createContext, useContext, useEffect, useRef, useState } from 'react'
import { getConnectionInfo } from '../lib/storage'

interface WebSocketContextType {
  sendMessage: (data: any) => void
  isConnected: boolean
  messages: any[]
}

const WebSocketContext = createContext<WebSocketContextType | null>(null)

interface WebSocketProviderProps {
  children: React.ReactNode
  endpoint: string
  onMessage?: (message: any) => void
}

export function WebSocketProvider({ children, endpoint, onMessage }: WebSocketProviderProps) {
  const wsRef = useRef<WebSocket | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [messages, setMessages] = useState<any[]>([])
  const reconnectAttemptsRef = useRef(0)
  const maxReconnectAttempts = 5

  useEffect(() => {
    let reconnectTimeout: any

    const connect = async () => {
      try {
        const connectionInfo = await getConnectionInfo()
        if (!connectionInfo?.sessionId || !connectionInfo?.token) {
          throw new Error('No connection info available')
        }

        const url = `${endpoint}?sessionId=${connectionInfo.sessionId}&token=${connectionInfo.token}`
        
        const ws = new WebSocket(url)
        wsRef.current = ws

        ws.onopen = () => {
          console.log('WebSocket connected')
          setIsConnected(true)
          reconnectAttemptsRef.current = 0
        }

        ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data)
            setMessages(prev => [...prev, message])
            if (onMessage) {
              onMessage(message)
            }
          } catch (error) {
            console.error('Error parsing WebSocket message:', error)
          }
        }

        ws.onclose = () => {
          console.log('WebSocket disconnected')
          setIsConnected(false)
          attemptReconnect()
        }

        ws.onerror = (error) => {
          console.error('WebSocket error:', error)
          setIsConnected(false)
        }

      } catch (error) {
        console.error('Failed to connect WebSocket:', error)
        attemptReconnect()
      }
    }

    const attemptReconnect = () => {
      if (reconnectAttemptsRef.current >= maxReconnectAttempts) {
        console.log('Max reconnection attempts reached')
        return
      }

      reconnectAttemptsRef.current++
      const delay = 1000 * Math.pow(2, reconnectAttemptsRef.current - 1)
      
      console.log(`Attempting to reconnect in ${delay}ms (attempt ${reconnectAttemptsRef.current})`)
      
      reconnectTimeout = setTimeout(() => {
        connect()
      }, delay)
    }

    connect()

    return () => {
      if (wsRef.current) {
        wsRef.current.close()
      }
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout)
      }
    }
  }, [endpoint, onMessage])

  const sendMessage = (data: any) => {
    if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(data))
    } else {
      console.error('WebSocket is not connected')
    }
  }

  return (
    <WebSocketContext.Provider value={{ sendMessage, isConnected, messages }}>
      {children}
    </WebSocketContext.Provider>
  )
}

export function useWebSocket() {
  const context = useContext(WebSocketContext)
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider')
  }
  return context
}
