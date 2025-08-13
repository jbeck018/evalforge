import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Clock, TrendingUp, TrendingDown, AlertTriangle, CheckCircle } from 'lucide-react'
import { api } from '../utils/api'

interface Evaluation {
  id: number
  name: string
  status: 'running' | 'completed' | 'failed'
  overall_score: number | null
  created_at: string
  project_id: number
  project_name: string
}

interface Project {
  id: number
  name: string
}

export default function RecentEvaluations() {
  const navigate = useNavigate()

  // Fetch all projects first
  const { data: projects } = useQuery({
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await api.get('/api/projects')
      return response.data.projects as Project[]
    },
  })

  // Fetch evaluations from all projects
  const { data: evaluations, isLoading } = useQuery({
    queryKey: ['recent-evaluations', projects],
    queryFn: async () => {
      if (!projects || projects.length === 0) {
        return []
      }

      // Fetch evaluations from all projects
      const evaluationPromises = projects.map(project =>
        api.get(`/api/projects/${project.id}/evaluations`)
          .then(res => {
            const evals = res.data.evaluations || []
            return evals.map((evaluation: any) => ({
              ...evaluation,
              project_id: project.id,
              project_name: project.name,
            }))
          })
          .catch(() => [])
      )

      const allEvaluations = await Promise.all(evaluationPromises)
      const flatEvaluations = allEvaluations.flat() as Evaluation[]
      
      // Sort by created_at and take the 5 most recent
      return flatEvaluations
        .sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
        .slice(0, 5)
    },
    enabled: !!projects,
    refetchInterval: 30000,
  })

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-4 w-4 text-success-600" />
      case 'failed':
        return <AlertTriangle className="h-4 w-4 text-error-600" />
      case 'running':
        return <Clock className="h-4 w-4 text-warning-600 animate-pulse" />
      default:
        return <Clock className="h-4 w-4 text-gray-400" />
    }
  }

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'text-success-700 bg-success-100'
      case 'failed':
        return 'text-error-700 bg-error-100'
      case 'running':
        return 'text-warning-700 bg-warning-100'
      default:
        return 'text-gray-700 bg-gray-100'
    }
  }

  const getScoreIcon = (score: number | null) => {
    if (score === null) return null
    if (score >= 0.8) return <TrendingUp className="h-4 w-4 text-success-600" />
    if (score <= 0.6) return <TrendingDown className="h-4 w-4 text-error-600" />
    return null
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / (1000 * 60))
    const diffHours = Math.floor(diffMins / 60)
    const diffDays = Math.floor(diffHours / 24)

    if (diffMins < 60) {
      return `${diffMins}m ago`
    } else if (diffHours < 24) {
      return `${diffHours}h ago`
    } else if (diffDays < 7) {
      return `${diffDays}d ago`
    } else {
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' })
    }
  }

  const handleEvaluationClick = (evaluation: Evaluation) => {
    navigate(`/projects/${evaluation.project_id}/evaluations/${evaluation.id}`)
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        {[...Array(3)].map((_, i) => (
          <div key={i} className="flex items-center justify-between p-3 border border-gray-200 rounded-lg">
            <div className="flex-1">
              <div className="loading-skeleton h-4 w-32 mb-2" />
              <div className="loading-skeleton h-3 w-24" />
            </div>
            <div className="loading-skeleton h-6 w-16" />
          </div>
        ))}
      </div>
    )
  }

  if (!evaluations || evaluations.length === 0) {
    return (
      <div className="text-center py-8">
        <div className="text-gray-400 mb-2">
          <Clock className="h-8 w-8 mx-auto" />
        </div>
        <p className="text-sm text-gray-500">No evaluations yet</p>
        <p className="text-xs text-gray-400 mt-1">
          Run your first evaluation to see results here
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {evaluations.map((evaluation) => (
        <div
          key={`${evaluation.project_id}-${evaluation.id}`}
          onClick={() => handleEvaluationClick(evaluation)}
          className="flex items-center justify-between p-3 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer transition-colors"
        >
          <div className="flex-1 min-w-0">
            <div className="flex items-center space-x-2 mb-1">
              {getStatusIcon(evaluation.status)}
              <h4 className="text-sm font-medium text-gray-900 truncate">
                {evaluation.name}
              </h4>
            </div>
            <div className="flex items-center space-x-2 text-xs text-gray-500">
              <span>{evaluation.project_name}</span>
              <span>•</span>
              <span>{formatDate(evaluation.created_at)}</span>
            </div>
          </div>
          
          <div className="flex items-center space-x-3 ml-4">
            {evaluation.overall_score !== null && (
              <div className="flex items-center space-x-1">
                {getScoreIcon(evaluation.overall_score)}
                <span className="text-sm font-medium text-gray-900">
                  {(evaluation.overall_score * 100).toFixed(0)}%
                </span>
              </div>
            )}
            
            <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(evaluation.status)}`}>
              {evaluation.status}
            </span>
          </div>
        </div>
      ))}
      
      {evaluations.length === 5 && (
        <div className="text-center pt-2">
          <button
            onClick={() => navigate('/evaluations')}
            className="text-sm text-primary-600 hover:text-primary-700 font-medium"
          >
            View all evaluations →
          </button>
        </div>
      )}
    </div>
  )
}