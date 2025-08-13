import { useState, useEffect } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { ArrowLeft, Activity, BarChart, Settings, Zap } from 'lucide-react'
import { api } from '../utils/api'

interface Project {
  id: string
  name: string
  description: string
  created_at: string
  api_key: string
}

export default function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [project, setProject] = useState<Project | null>(null)
  const [loading, setLoading] = useState(true)
  const [showApiKey, setShowApiKey] = useState(false)

  useEffect(() => {
    if (id) {
      fetchProject()
    }
  }, [id])

  const fetchProject = async () => {
    try {
      const response = await api.get(`/api/projects/${id}`)
      setProject(response.data.project || response.data)
    } catch (error) {
      console.error('Failed to fetch project:', error)
      navigate('/projects')
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

  if (!project) {
    return null
  }

  return (
    <div className="page-content">
      <div className="mb-6">
        <Link
          to="/projects"
          className="inline-flex items-center text-sm text-gray-500 hover:text-gray-700"
        >
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Projects
        </Link>
      </div>

      <div className="bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h1 className="text-2xl font-semibold text-gray-900">{project.name}</h1>
          <p className="mt-2 text-sm text-gray-600">{project.description}</p>
          
          <div className="mt-6 border-t border-gray-200 pt-6">
            <h2 className="text-lg font-medium text-gray-900 mb-4">API Configuration</h2>
            <div className="bg-gray-50 rounded-md p-4">
              <div className="flex items-center justify-between">
                <div className="flex-1">
                  <p className="text-sm font-medium text-gray-700">Project API Key</p>
                  <p className="mt-1 text-xs text-gray-500">Use this key to authenticate SDK requests</p>
                </div>
                <button
                  onClick={() => setShowApiKey(!showApiKey)}
                  className="ml-4 text-sm text-indigo-600 hover:text-indigo-500"
                >
                  {showApiKey ? 'Hide' : 'Show'}
                </button>
              </div>
              {showApiKey && (
                <div className="mt-3">
                  <code className="block p-2 bg-gray-900 text-green-400 rounded text-xs font-mono break-all">
                    {project.api_key}
                  </code>
                </div>
              )}
            </div>
          </div>

          <div className="mt-8 grid gap-4 sm:grid-cols-4">
            <Link
              to={`/projects/${id}/evaluations`}
              className="relative rounded-lg border border-gray-300 bg-white px-6 py-5 shadow-sm flex items-center space-x-3 hover:border-gray-400 focus-within:ring-2 focus-within:ring-offset-2 focus-within:ring-indigo-500"
            >
              <div className="flex-shrink-0">
                <Zap className="h-6 w-6 text-gray-600" />
              </div>
              <div className="flex-1 min-w-0">
                <span className="absolute inset-0" aria-hidden="true" />
                <p className="text-sm font-medium text-gray-900">Evaluations</p>
                <p className="text-sm text-gray-500">Test and optimize prompts</p>
              </div>
            </Link>

            <Link
              to={`/projects/${id}/traces`}
              className="relative rounded-lg border border-gray-300 bg-white px-6 py-5 shadow-sm flex items-center space-x-3 hover:border-gray-400 focus-within:ring-2 focus-within:ring-offset-2 focus-within:ring-indigo-500"
            >
              <div className="flex-shrink-0">
                <Activity className="h-6 w-6 text-gray-600" />
              </div>
              <div className="flex-1 min-w-0">
                <span className="absolute inset-0" aria-hidden="true" />
                <p className="text-sm font-medium text-gray-900">Traces</p>
                <p className="text-sm text-gray-500">View execution traces</p>
              </div>
            </Link>

            <Link
              to={`/projects/${id}/analytics`}
              className="relative rounded-lg border border-gray-300 bg-white px-6 py-5 shadow-sm flex items-center space-x-3 hover:border-gray-400 focus-within:ring-2 focus-within:ring-offset-2 focus-within:ring-indigo-500"
            >
              <div className="flex-shrink-0">
                <BarChart className="h-6 w-6 text-gray-600" />
              </div>
              <div className="flex-1 min-w-0">
                <span className="absolute inset-0" aria-hidden="true" />
                <p className="text-sm font-medium text-gray-900">Analytics</p>
                <p className="text-sm text-gray-500">View project metrics</p>
              </div>
            </Link>

            <Link
              to={`/projects/${id}/settings`}
              className="relative rounded-lg border border-gray-300 bg-white px-6 py-5 shadow-sm flex items-center space-x-3 hover:border-gray-400 focus-within:ring-2 focus-within:ring-offset-2 focus-within:ring-indigo-500"
            >
              <div className="flex-shrink-0">
                <Settings className="h-6 w-6 text-gray-600" />
              </div>
              <div className="flex-1 min-w-0">
                <span className="absolute inset-0" aria-hidden="true" />
                <p className="text-sm font-medium text-gray-900">Settings</p>
                <p className="text-sm text-gray-500">Configure project</p>
              </div>
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}