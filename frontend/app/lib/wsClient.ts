import { getConnectionInfo } from './storage'

export interface WSEvent {
  type: 'status' | 'patch' | 'exit' | 'error' | 'output'
  data?: any
  message?: string
}

export class WebSocketClient {
  private ws: WebSocket | null = null
  private eventHandlers: Map<string, ((event: WSEvent) => void)[]> = new Map()
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000

  constructor(private endpoint: string) {}

  async connect(): Promise<void> {
    try {
      const connectionInfo = await getConnectionInfo()
      if (!connectionInfo?.sessionId || !connectionInfo?.token) {
        throw new Error('No connection info available')
      }

      const url = `${this.endpoint}?sessionId=${connectionInfo.sessionId}&token=${connectionInfo.token}`
      
      this.ws = new WebSocket(url)

      this.ws.onopen = () => {
        console.log('WebSocket connected')
        this.reconnectAttempts = 0
      }

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          this.handleEvent(message)
        } catch (error) {
          console.error('Error parsing WebSocket message:', error)
        }
      }

      this.ws.onclose = () => {
        console.log('WebSocket disconnected')
        this.handleEvent({ type: 'exit', message: 'Connection closed' })
        this.attemptReconnect()
      }

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error)
        this.handleEvent({ type: 'error', message: 'WebSocket connection error' })
      }

    } catch (error) {
      console.error('Failed to connect WebSocket:', error)
      throw error
    }
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('Max reconnection attempts reached')
      return
    }

    this.reconnectAttempts++
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
    
    console.log(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts})`)
    
    setTimeout(() => {
      this.connect().catch(console.error)
    }, delay)
  }

  private handleEvent(event: WSEvent): void {
    const handlers = this.eventHandlers.get(event.type) || []
    handlers.forEach(handler => {
      try {
        handler(event)
      } catch (error) {
        console.error('Error in event handler:', error)
      }
    })
  }

  on(event: string, handler: (event: WSEvent) => void): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, [])
    }
    this.eventHandlers.get(event)!.push(handler)
  }

  off(event: string, handler: (event: WSEvent) => void): void {
    const handlers = this.eventHandlers.get(event)
    if (handlers) {
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }

  send(data: any): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    } else {
      console.error('WebSocket is not connected')
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.eventHandlers.clear()
  }

  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }
}

// Factory function to create WebSocket clients
export function createWebSocketClient(endpoint: string): WebSocketClient {
  return new WebSocketClient(endpoint)
}
