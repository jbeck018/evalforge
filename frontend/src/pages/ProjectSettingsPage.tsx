import { useState, useEffect } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { ArrowLeft, Save, AlertTriangle, Copy, Eye, EyeOff, RefreshCw, Trash2 } from 'lucide-react'
import { api } from '../utils/api'
import toast from 'react-hot-toast'
import { Button } from '../components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../components/ui/tabs'
import { Label } from '../components/ui/label'

interface ProjectSettings {
  id: number
  name: string
  description: string
  api_key: string
  created_at: string
  settings: {
    auto_evaluation_enabled: boolean
    evaluation_threshold: number
    max_traces_per_day: number
    retention_days: number
    alert_email: string
    webhook_url: string
  }
}

export default function ProjectSettingsPage() {
  const { id: projectId } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [project, setProject] = useState<ProjectSettings | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [showApiKey, setShowApiKey] = useState(false)
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)

  // Form state
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    auto_evaluation_enabled: false,
    evaluation_threshold: 0.8,
    max_traces_per_day: 10000,
    retention_days: 30,
    alert_email: '',
    webhook_url: ''
  })

  useEffect(() => {
    if (projectId) {
      fetchProjectSettings()
    }
  }, [projectId])

  const fetchProjectSettings = async () => {
    try {
      const response = await api.get(`/api/projects/${projectId}/settings`)
      const projectData = response.data.project
      setProject(projectData)
      setFormData({
        name: projectData.name,
        description: projectData.description,
        auto_evaluation_enabled: projectData.settings?.auto_evaluation_enabled || false,
        evaluation_threshold: projectData.settings?.evaluation_threshold || 0.8,
        max_traces_per_day: projectData.settings?.max_traces_per_day || 10000,
        retention_days: projectData.settings?.retention_days || 30,
        alert_email: projectData.settings?.alert_email || '',
        webhook_url: projectData.settings?.webhook_url || ''
      })
    } catch (error) {
      console.error('Failed to fetch project settings:', error)
      toast.error('Failed to load project settings')
    } finally {
      setLoading(false)
    }
  }

  const handleSave = async () => {
    setSaving(true)
    try {
      await api.put(`/api/projects/${projectId}/settings`, formData)
      toast.success('Project settings updated successfully')
      fetchProjectSettings()
    } catch (error) {
      console.error('Failed to save settings:', error)
      toast.error('Failed to save settings')
    } finally {
      setSaving(false)
    }
  }

  const handleRegenerateApiKey = async () => {
    try {
      const response = await api.post(`/api/projects/${projectId}/regenerate-key`)
      setProject(prev => prev ? { ...prev, api_key: response.data.api_key } : null)
      toast.success('API key regenerated successfully')
    } catch (error) {
      console.error('Failed to regenerate API key:', error)
      toast.error('Failed to regenerate API key')
    }
  }

  const handleCopyApiKey = () => {
    if (project?.api_key) {
      navigator.clipboard.writeText(project.api_key)
      toast.success('API key copied to clipboard')
    }
  }

  const handleDeleteProject = async () => {
    try {
      await api.delete(`/api/projects/${projectId}`)
      toast.success('Project deleted successfully')
      navigate('/projects')
    } catch (error) {
      console.error('Failed to delete project:', error)
      toast.error('Failed to delete project')
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
    return (
      <div className="page-content">
        <div className="text-center py-12">
          <p className="text-gray-500">Project not found</p>
          <Link to="/projects" className="text-indigo-600 hover:text-indigo-500 mt-4 inline-block">
            Back to projects
          </Link>
        </div>
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
          <h1 className="text-2xl font-semibold text-gray-900">Project Settings</h1>
          <p className="mt-2 text-sm text-gray-700">
            Configure your project settings and manage API keys.
          </p>
        </div>
      </div>

      <Card>
        <Tabs defaultValue="general" className="w-full">
          <CardHeader>
            <TabsList className="grid w-full grid-cols-4">
              <TabsTrigger value="general">General</TabsTrigger>
              <TabsTrigger value="api">API Keys</TabsTrigger>
              <TabsTrigger value="evaluation">Evaluation</TabsTrigger>
              <TabsTrigger value="alerts">Alerts & Webhooks</TabsTrigger>
            </TabsList>
          </CardHeader>

          <CardContent>
            <TabsContent value="general" className="mt-0">
              <div className="space-y-6">
                <div className="space-y-4">
                  <div>
                    <Label htmlFor="name">Project Name</Label>
                    <input
                      id="name"
                      type="text"
                      value={formData.name}
                      onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                    />
                  </div>

                  <div>
                    <Label htmlFor="description">Description</Label>
                    <textarea
                      id="description"
                      rows={3}
                      value={formData.description}
                      onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                    />
                  </div>

                  <div>
                    <Label htmlFor="retention">Data Retention (days)</Label>
                    <input
                      id="retention"
                      type="number"
                      min="1"
                      max="365"
                      value={formData.retention_days}
                      onChange={(e) => setFormData({ ...formData, retention_days: parseInt(e.target.value) })}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                    />
                    <p className="mt-1 text-sm text-gray-500">
                      How long to keep trace data (1-365 days)
                    </p>
                  </div>

                  <div>
                    <Label htmlFor="max-traces">Max Traces Per Day</Label>
                    <input
                      id="max-traces"
                      type="number"
                      min="100"
                      max="1000000"
                      value={formData.max_traces_per_day}
                      onChange={(e) => setFormData({ ...formData, max_traces_per_day: parseInt(e.target.value) })}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                    />
                    <p className="mt-1 text-sm text-gray-500">
                      Maximum number of traces to ingest per day
                    </p>
                  </div>
                </div>

                <div className="flex justify-between items-center pt-6 border-t">
                  <Button onClick={handleSave} disabled={saving}>
                    <Save className="h-4 w-4 mr-2" />
                    {saving ? 'Saving...' : 'Save Changes'}
                  </Button>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="api" className="mt-0">
              <div className="space-y-6">
                <div>
                  <CardTitle className="text-lg mb-4">API Key</CardTitle>
                  <CardDescription className="mb-4">
                    Use this key to authenticate your SDK or API requests.
                  </CardDescription>
                  
                  <div className="space-y-4">
                    <div className="flex items-center space-x-2">
                      <div className="flex-1">
                        <div className="relative">
                          <input
                            type={showApiKey ? 'text' : 'password'}
                            value={project.api_key || 'No API key generated'}
                            readOnly
                            className="block w-full rounded-md border-gray-300 bg-gray-50 pr-20 font-mono text-sm"
                          />
                          <div className="absolute inset-y-0 right-0 flex items-center pr-3 space-x-1">
                            <button
                              onClick={() => setShowApiKey(!showApiKey)}
                              className="text-gray-400 hover:text-gray-600"
                            >
                              {showApiKey ? (
                                <EyeOff className="h-4 w-4" />
                              ) : (
                                <Eye className="h-4 w-4" />
                              )}
                            </button>
                            <button
                              onClick={handleCopyApiKey}
                              className="text-gray-400 hover:text-gray-600"
                            >
                              <Copy className="h-4 w-4" />
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>

                    <div className="flex items-center space-x-4">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={handleRegenerateApiKey}
                      >
                        <RefreshCw className="h-4 w-4 mr-2" />
                        Regenerate Key
                      </Button>
                      <p className="text-sm text-gray-500">
                        Last regenerated: {project.created_at ? new Date(project.created_at).toLocaleDateString() : 'Never'}
                      </p>
                    </div>

                    <div className="bg-yellow-50 border border-yellow-200 rounded-md p-4">
                      <div className="flex">
                        <AlertTriangle className="h-5 w-5 text-yellow-400 mr-2" />
                        <div className="text-sm text-yellow-800">
                          <p className="font-medium">Keep your API key secure</p>
                          <p className="mt-1">
                            Never share your API key or commit it to version control. Regenerating the key will invalidate the old one.
                          </p>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="evaluation" className="mt-0">
              <div className="space-y-6">
                <CardTitle className="text-lg">Evaluation Settings</CardTitle>
                
                <div className="space-y-4">
                  <div className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      id="auto-eval"
                      checked={formData.auto_evaluation_enabled}
                      onChange={(e) => setFormData({ ...formData, auto_evaluation_enabled: e.target.checked })}
                      className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                    />
                    <Label htmlFor="auto-eval">
                      Enable automatic evaluation
                    </Label>
                  </div>

                  <div>
                    <Label htmlFor="threshold">Evaluation Threshold</Label>
                    <div className="mt-1 flex items-center space-x-2">
                      <input
                        id="threshold"
                        type="range"
                        min="0"
                        max="1"
                        step="0.1"
                        value={formData.evaluation_threshold}
                        onChange={(e) => setFormData({ ...formData, evaluation_threshold: parseFloat(e.target.value) })}
                        className="flex-1"
                        disabled={!formData.auto_evaluation_enabled}
                      />
                      <span className="text-sm font-medium text-gray-700 w-12">
                        {(formData.evaluation_threshold * 100).toFixed(0)}%
                      </span>
                    </div>
                    <p className="mt-1 text-sm text-gray-500">
                      Minimum confidence score to trigger automatic evaluations
                    </p>
                  </div>

                  <div className="bg-blue-50 border border-blue-200 rounded-md p-4">
                    <p className="text-sm text-blue-800">
                      When enabled, evaluations will automatically run when new traces match your configured criteria.
                    </p>
                  </div>
                </div>

                <div className="pt-6 border-t">
                  <Button onClick={handleSave} disabled={saving}>
                    <Save className="h-4 w-4 mr-2" />
                    {saving ? 'Saving...' : 'Save Changes'}
                  </Button>
                </div>
              </div>
            </TabsContent>

            <TabsContent value="alerts" className="mt-0">
              <div className="space-y-6">
                <CardTitle className="text-lg">Alerts & Webhooks</CardTitle>
                
                <div className="space-y-4">
                  <div>
                    <Label htmlFor="alert-email">Alert Email</Label>
                    <input
                      id="alert-email"
                      type="email"
                      value={formData.alert_email}
                      onChange={(e) => setFormData({ ...formData, alert_email: e.target.value })}
                      placeholder="alerts@example.com"
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                    />
                    <p className="mt-1 text-sm text-gray-500">
                      Email address for receiving alerts about errors and anomalies
                    </p>
                  </div>

                  <div>
                    <Label htmlFor="webhook-url">Webhook URL</Label>
                    <input
                      id="webhook-url"
                      type="url"
                      value={formData.webhook_url}
                      onChange={(e) => setFormData({ ...formData, webhook_url: e.target.value })}
                      placeholder="https://example.com/webhook"
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                    />
                    <p className="mt-1 text-sm text-gray-500">
                      URL to receive real-time notifications about events and evaluations
                    </p>
                  </div>

                  <div className="bg-gray-50 border border-gray-200 rounded-md p-4">
                    <p className="text-sm text-gray-700">
                      <strong>Webhook Events:</strong> evaluation_complete, error_threshold_exceeded, anomaly_detected
                    </p>
                  </div>
                </div>

                <div className="pt-6 border-t">
                  <Button onClick={handleSave} disabled={saving}>
                    <Save className="h-4 w-4 mr-2" />
                    {saving ? 'Saving...' : 'Save Changes'}
                  </Button>
                </div>
              </div>
            </TabsContent>
          </CardContent>
        </Tabs>
      </Card>

      {/* Danger Zone */}
      <Card className="mt-8 border-red-200">
        <CardHeader>
          <CardTitle className="text-red-600">Danger Zone</CardTitle>
          <CardDescription>
            Irreversible actions that affect your project.
          </CardDescription>
        </CardHeader>
        <CardContent>
          {!showDeleteConfirm ? (
            <Button
              variant="destructive"
              onClick={() => setShowDeleteConfirm(true)}
            >
              <Trash2 className="h-4 w-4 mr-2" />
              Delete Project
            </Button>
          ) : (
            <div className="space-y-4">
              <p className="text-sm text-red-600">
                Are you sure you want to delete this project? This action cannot be undone.
              </p>
              <div className="flex space-x-2">
                <Button
                  variant="destructive"
                  onClick={handleDeleteProject}
                >
                  Yes, Delete Project
                </Button>
                <Button
                  variant="outline"
                  onClick={() => setShowDeleteConfirm(false)}
                >
                  Cancel
                </Button>
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}