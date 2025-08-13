import { useQuery } from '@tanstack/react-query'
import { Clock, CheckCircle, XCircle, AlertCircle } from 'lucide-react'
import { api } from '../utils/api'

interface Event {
  id: string
  operation_type: string
  model?: string
  input?: string
  output?: string
  error?: string
  created_at: string
  project_id: number
}

interface Project {
  id: number
  name: string
}

interface Activity {
  id: string
  type: string
  message: string
  timestamp: string
  status: 'success' | 'error' | 'warning'
  project_name?: string
}

export default function RecentActivity() {
  // First fetch projects to get their names
  const { data: projects } = useQuery({
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await api.get('/api/projects')
      return response.data.projects as Project[]
    },
  })

  // Fetch recent events from all projects
  const { data: activities, isLoading } = useQuery({
    queryKey: ['recent-activity', projects],
    queryFn: async () => {
      if (!projects || projects.length === 0) {
        return []
      }

      // Fetch recent events from each project
      const eventsPromises = projects.map(project => 
        api.get(`/api/projects/${project.id}/events`, {
          params: { limit: 5 }
        })
          .then(res => {
            const events = res.data.events || []
            // Transform events to activities
            return events.map((event: Event) => ({
              id: event.id,
              type: event.operation_type || 'Unknown',
              message: `${event.operation_type || 'Operation'} ${event.model ? `using ${event.model}` : ''}`,
              timestamp: event.created_at,
              status: event.error ? 'error' : 'success',
              project_name: project.name,
            } as Activity))
          })
          .catch(() => [])
      )

      const allActivities = await Promise.all(eventsPromises)
      
      // Flatten and sort by timestamp
      const flattened = allActivities.flat()
      flattened.sort((a, b) => 
        new Date(b.timestamp).getTime() - new Date(a.timestamp).getTime()
      )
      
      // Return the 10 most recent activities
      return flattened.slice(0, 10)
    },
    enabled: !!projects,
    refetchInterval: 30000 // Refresh every 30 seconds
  })

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="h-5 w-5 text-green-500" />
      case 'error':
        return <XCircle className="h-5 w-5 text-red-500" />
      case 'warning':
        return <AlertCircle className="h-5 w-5 text-yellow-500" />
      default:
        return <Clock className="h-5 w-5 text-gray-400" />
    }
  }

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    
    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins} minutes ago`
    
    const diffHours = Math.floor(diffMins / 60)
    if (diffHours < 24) return `${diffHours} hours ago`
    
    return date.toLocaleDateString()
  }

  if (isLoading) {
    return (
      <div className="bg-white shadow rounded-lg p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Recent Activity</h3>
        <div className="flex justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
        </div>
      </div>
    )
  }

  return (
    <div className="bg-white shadow rounded-lg">
      <div className="px-4 py-5 sm:p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">Recent Activity</h3>
        <div className="flow-root">
          <ul className="-mb-8">
            {activities?.map((activity, idx) => (
              <li key={activity.id}>
                <div className="relative pb-8">
                  {idx !== (activities?.length || 0) - 1 && (
                    <span
                      className="absolute top-5 left-5 -ml-px h-full w-0.5 bg-gray-200"
                      aria-hidden="true"
                    />
                  )}
                  <div className="relative flex items-start space-x-3">
                    <div>
                      <div className="relative px-1">
                        <div className="flex items-center justify-center">
                          {getStatusIcon(activity.status)}
                        </div>
                      </div>
                    </div>
                    <div className="min-w-0 flex-1">
                      <div>
                        <p className="text-sm text-gray-900">
                          {activity.message}
                        </p>
                        <div className="mt-1 flex items-center space-x-2 text-xs text-gray-500">
                          <span>{formatTime(activity.timestamp)}</span>
                          {activity.project_name && (
                            <>
                              <span>•</span>
                              <span>{activity.project_name}</span>
                            </>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </li>
            ))}
          </ul>
          {(!activities || activities.length === 0) && (
            <p className="text-center text-gray-500 py-4">No recent activity</p>
          )}
        </div>
      </div>
    </div>
  )
}