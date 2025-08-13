import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft } from 'lucide-react'
import { api } from '../utils/api'
import MetricsDashboard from '../components/MetricsDashboard'
import { RealtimeMetrics } from '../components/RealtimeMetrics'

interface Analytics {
  total_traces: number
  success_rate: number
  avg_duration_ms: number
  error_rate: number
  traces_by_type: Record<string, number>
  traces_by_status: Record<string, number>
  traces_over_time: Array<{
    date: string
    count: number
  }>
}

export default function AnalyticsPage() {
  const { id: projectId } = useParams<{ id: string }>()
  const [analytics, setAnalytics] = useState<Analytics | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (projectId) {
      fetchAnalytics()
    }
  }, [projectId])

  const fetchAnalytics = async () => {
    try {
      const response = await api.get(`/api/projects/${projectId}/analytics`)
      setAnalytics(response.data)
    } catch (error) {
      console.error('Failed to fetch analytics:', error)
    } finally {
      setLoading(false)
    }
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
      </div>
    )
  }

  return (
    <div className="page-content">
      <div className="mb-6">
        <Link
          to={`/projects/${projectId}`}
          className="inline-flex items-center text-sm text-gray-500 hover:text-gray-700"
        >
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Project
        </Link>
      </div>

      <div className="sm:flex sm:items-center mb-6">
        <div className="sm:flex-auto">
          <h1 className="text-2xl font-semibold text-gray-900">Analytics</h1>
          <p className="mt-2 text-sm text-gray-700">
            View metrics and insights for your project.
          </p>
        </div>
      </div>

      {/* Real-time Metrics Section */}
      <div className="mb-8">
        <RealtimeMetrics />
      </div>

      {/* Historical Analytics */}
      {analytics && (
        <div className="mt-8">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Historical Analytics</h2>
          <MetricsDashboard analytics={analytics} />
        </div>
      )}

      {!analytics && !loading && (
        <div className="bg-white rounded-lg shadow p-6">
          <p className="text-center text-gray-500">No analytics data available yet.</p>
        </div>
      )}
    </div>
  )
}