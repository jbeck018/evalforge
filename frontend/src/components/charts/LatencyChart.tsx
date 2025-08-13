import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { useQuery } from '@tanstack/react-query'
import { api } from '../../utils/api'

interface LatencyData {
  range: string
  count: number
}

interface Project {
  id: number
  name: string
}

interface Event {
  id: number
  latency_ms: number | null
}

export default function LatencyChart() {
  const { data: projects } = useQuery({
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await api.get('/api/projects')
      return response.data.projects as Project[]
    },
  })

  const { data, isLoading } = useQuery({
    queryKey: ['latency-chart', projects],
    queryFn: async () => {
      if (!projects || projects.length === 0) {
        return []
      }

      // Fetch events from all projects to calculate latency distribution
      const eventsPromises = projects.map(project =>
        api.get(`/api/projects/${project.id}/events`, {
          params: {
            limit: 1000, // Get recent events for latency calculation
          }
        })
          .then(res => res.data.events || [])
          .catch(() => [])
      )

      const allProjectEvents = await Promise.all(eventsPromises)
      const allEvents = allProjectEvents.flat()

      if (allEvents.length === 0) {
        return []
      }

      // Calculate latency distribution from actual events
      const latencyBuckets = {
        '0-100ms': 0,
        '100-200ms': 0,
        '200-500ms': 0,
        '500-1000ms': 0,
        '1000ms+': 0,
      }

      allEvents.forEach(event => {
        if (event.latency_ms && typeof event.latency_ms === 'number') {
          const latency = event.latency_ms
          if (latency < 100) {
            latencyBuckets['0-100ms']++
          } else if (latency < 200) {
            latencyBuckets['100-200ms']++
          } else if (latency < 500) {
            latencyBuckets['200-500ms']++
          } else if (latency < 1000) {
            latencyBuckets['500-1000ms']++
          } else {
            latencyBuckets['1000ms+']++
          }
        }
      })

      return Object.entries(latencyBuckets).map(([range, count]) => ({
        range,
        count,
      }))
    },
    enabled: !!projects,
  })

  if (isLoading) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
      </div>
    )
  }

  if (!data || data.length === 0 || data.every(d => d.count === 0)) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-center text-gray-500">
          <p className="text-sm">No latency data available</p>
          <p className="text-xs mt-1">Start sending events to see latency distribution</p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-64">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart data={data}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis 
            dataKey="range" 
            tick={{ fontSize: 12 }}
          />
          <YAxis 
            tick={{ fontSize: 12 }}
          />
          <Tooltip 
            formatter={(value: number) => [`${value} requests`, 'Count']}
          />
          <Bar 
            dataKey="count" 
            fill="#10b981"
          />
        </BarChart>
      </ResponsiveContainer>
    </div>
  )
}