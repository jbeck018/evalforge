import { useEffect, useState, useCallback } from 'react'
import useWebSocket, { ReadyState } from 'react-use-websocket'
import toast from 'react-hot-toast'

interface RealtimeMetrics {
  totalEvents: number
  eventsPerMinute: number
  errorRate: number
  avgLatency: number
  activeProjects: number
  recentEvaluations: number
  operationCounts: Record<string, number>
  statusCounts: Record<string, number>
  timestamp: Date
}

interface WebSocketMessage {
  type: 'metrics' | 'evaluation' | 'new_event'
  timestamp: string
  data: any
}

export function useRealtimeMetrics() {
  const [metrics, setMetrics] = useState<RealtimeMetrics | null>(null)
  const [isConnected, setIsConnected] = useState(false)
  const [lastUpdateTime, setLastUpdateTime] = useState<Date | null>(null)

  const socketUrl = `ws://localhost:8088/ws`

  const { sendMessage, lastMessage, readyState } = useWebSocket(socketUrl, {
    onOpen: () => {
      console.log('WebSocket connection established')
      setIsConnected(true)
    },
    onClose: () => {
      console.log('WebSocket connection closed')
      setIsConnected(false)
    },
    onError: (error) => {
      console.error('WebSocket error:', error)
      toast.error('Real-time connection error')
    },
    shouldReconnect: () => true,
    reconnectAttempts: 10,
    reconnectInterval: 3000,
  })

  useEffect(() => {
    if (lastMessage !== null) {
      try {
        const message: WebSocketMessage = JSON.parse(lastMessage.data)
        
        if (message.type === 'metrics') {
          setMetrics(message.data as RealtimeMetrics)
          setLastUpdateTime(new Date())
        } else if (message.type === 'new_event') {
          // Handle new event notifications
          console.log('New event received:', message.data)
        } else if (message.type === 'evaluation') {
          // Handle new evaluation notifications
          console.log('New evaluation received:', message.data)
        }
      } catch (error) {
        console.error('Error parsing WebSocket message:', error)
      }
    }
  }, [lastMessage])

  const connectionStatus = {
    [ReadyState.CONNECTING]: 'Connecting',
    [ReadyState.OPEN]: 'Connected',
    [ReadyState.CLOSING]: 'Closing',
    [ReadyState.CLOSED]: 'Disconnected',
    [ReadyState.UNINSTANTIATED]: 'Uninstantiated',
  }[readyState]

  const sendMetricRequest = useCallback((request: any) => {
    if (readyState === ReadyState.OPEN) {
      sendMessage(JSON.stringify(request))
    }
  }, [readyState, sendMessage])

  return {
    metrics,
    isConnected,
    connectionStatus,
    lastUpdateTime,
    sendMetricRequest,
  }
}