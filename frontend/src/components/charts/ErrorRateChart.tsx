import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { useQuery } from '@tanstack/react-query'
import { api } from '../../utils/api'

interface ErrorRateData {
  date: string
  error_rate: number
  error_count: number
  total_count: number
}

interface Project {
  id: number
  name: string
}

interface Event {
  id: number
  status: string | null
  error: string | null
  error_message: string | null
  created_at: string
}

export default function ErrorRateChart() {
  const { data: projects } = useQuery({
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await api.get('/api/projects')
      return response.data.projects as Project[]
    },
  })

  const { data, isLoading } = useQuery({
    queryKey: ['error-rate-chart', projects],
    queryFn: async () => {
      if (!projects || projects.length === 0) {
        return []
      }

      // Fetch events from all projects to calculate error rates by day
      const eventsPromises = projects.map(project =>
        api.get(`/api/projects/${project.id}/events`, {
          params: {
            limit: 1000, // Get recent events for error calculation
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

      // Group events by date and calculate error rates
      const eventsByDate = new Map<string, { total: number, errors: number }>()
      
      allEvents.forEach(event => {
        if (event.created_at) {
          const date = new Date(event.created_at).toISOString().split('T')[0]
          const current = eventsByDate.get(date) || { total: 0, errors: 0 }
          
          current.total++
          // Consider an event an error if it has an error status or error field
          if (event.status === 'error' || event.error || event.error_message) {
            current.errors++
          }
          
          eventsByDate.set(date, current)
        }
      })

      // Convert to array and sort by date
      const errorData = Array.from(eventsByDate.entries())
        .map(([date, { total, errors }]) => ({
          date,
          error_rate: total > 0 ? errors / total : 0,
          error_count: errors,
          total_count: total,
        }))
        .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
        .slice(-7) // Show last 7 days only

      return errorData
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

  if (!data || data.length === 0) {
    return (
      <div className="h-64 flex items-center justify-center">
        <div className="text-center text-gray-500">
          <p className="text-sm">No error data available</p>
          <p className="text-xs mt-1">Start sending events to see error trends</p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-64">
      <ResponsiveContainer width="100%" height="100%">
        <AreaChart data={data}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis 
            dataKey="date" 
            tick={{ fontSize: 12 }}
            tickFormatter={(value) => new Date(value).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
          />
          <YAxis 
            tick={{ fontSize: 12 }}
            tickFormatter={(value) => `${(value * 100).toFixed(1)}%`}
          />
          <Tooltip 
            formatter={(value: number, name: string) => {
              if (name === 'error_rate') {
                return `${(value * 100).toFixed(2)}%`
              }
              return value
            }}
            labelFormatter={(label) => new Date(label).toLocaleDateString()}
            content={({ active, payload, label }) => {
              if (active && payload && payload.length) {
                const data = payload[0].payload
                return (
                  <div className="bg-white p-2 border border-gray-200 rounded shadow-sm">
                    <p className="text-sm font-medium">{new Date(label).toLocaleDateString()}</p>
                    <p className="text-sm text-red-600">Error Rate: {(data.error_rate * 100).toFixed(2)}%</p>
                    <p className="text-xs text-gray-500">
                      {data.error_count} errors / {data.total_count} total
                    </p>
                  </div>
                )
              }
              return null
            }}
          />
          <Area 
            type="monotone" 
            dataKey="error_rate" 
            stroke="#ef4444" 
            fill="#fee2e2"
            strokeWidth={2}
          />
        </AreaChart>
      </ResponsiveContainer>
    </div>
  )
}