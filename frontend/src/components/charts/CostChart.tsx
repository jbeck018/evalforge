import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'
import { useQuery } from '@tanstack/react-query'
import { api } from '../../utils/api'

interface CostData {
  date: string
  cost: number
}

interface Project {
  id: number
  name: string
}

interface Event {
  id: number
  cost: string | null
  created_at: string
}

export default function CostChart() {
  const { data: projects } = useQuery({
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await api.get('/api/projects')
      return response.data.projects as Project[]
    },
  })

  const { data, isLoading } = useQuery({
    queryKey: ['cost-chart', projects],
    queryFn: async () => {
      if (!projects || projects.length === 0) {
        return []
      }

      // Fetch events from all projects to calculate daily costs
      const eventsPromises = projects.map(project =>
        api.get(`/api/projects/${project.id}/events`, {
          params: {
            limit: 1000, // Get recent events for cost calculation
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

      // Group events by date and sum costs
      const costsByDate = new Map<string, number>()
      
      allEvents.forEach(event => {
        if (event.cost && event.created_at) {
          const date = new Date(event.created_at).toISOString().split('T')[0]
          const currentCost = costsByDate.get(date) || 0
          costsByDate.set(date, currentCost + parseFloat(event.cost))
        }
      })

      // Convert to array and sort by date
      const costData = Array.from(costsByDate.entries())
        .map(([date, cost]) => ({ date, cost }))
        .sort((a, b) => new Date(a.date).getTime() - new Date(b.date).getTime())
        .slice(-7) // Show last 7 days only

      return costData
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
          <p className="text-sm">No cost data available</p>
          <p className="text-xs mt-1">Start sending events to see cost trends</p>
        </div>
      </div>
    )
  }

  return (
    <div className="h-64">
      <ResponsiveContainer width="100%" height="100%">
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis 
            dataKey="date" 
            tick={{ fontSize: 12 }}
            tickFormatter={(value) => new Date(value).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
          />
          <YAxis 
            tick={{ fontSize: 12 }}
            tickFormatter={(value) => `$${value.toFixed(4)}`}
          />
          <Tooltip 
            formatter={(value: number) => `$${value.toFixed(4)}`}
            labelFormatter={(label) => new Date(label).toLocaleDateString()}
          />
          <Line 
            type="monotone" 
            dataKey="cost" 
            stroke="#6366f1" 
            strokeWidth={2}
            dot={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}