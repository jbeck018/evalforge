import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { 
  TrendingUp, 
  TrendingDown, 
  DollarSign, 
  Clock, 
  AlertTriangle,
  Activity,
  Users,
  Database
} from 'lucide-react'

import { api } from '../utils/api'
import MetricCard from '../components/MetricCard'
import RecentActivity from '../components/RecentActivity'
import RecentEvaluations from '../components/RecentEvaluations'
import CostChart from '../components/charts/CostChart'
import LatencyChart from '../components/charts/LatencyChart'
import ErrorRateChart from '../components/charts/ErrorRateChart'

interface ProjectAnalytics {
  total_events: number
  total_cost: number
  average_latency: number
  error_rate: number
}

interface Project {
  id: number
  name: string
  api_key: string
  created_at: string
}

interface DashboardStats {
  total_events: number
  total_cost: number
  average_latency: number
  error_rate: number
  active_projects: number
  events_24h: number
  cost_24h: number
  events_change: number
  cost_change: number
  latency_change: number
  error_rate_change: number
}

export default function DashboardPage() {
  const navigate = useNavigate()
  
  // First fetch all projects
  const { data: projects, isLoading: projectsLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await api.get('/api/projects')
      return response.data.projects as Project[]
    },
    refetchInterval: 30000,
  })

  // Then fetch analytics for each project and aggregate
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['dashboard-stats', projects],
    queryFn: async () => {
      if (!projects || projects.length === 0) {
        return {
          total_events: 0,
          total_cost: 0,
          average_latency: 0,
          error_rate: 0,
          active_projects: 0,
          events_24h: 0,
          cost_24h: 0,
          events_change: 0,
          cost_change: 0,
          latency_change: 0,
          error_rate_change: 0,
        } as DashboardStats
      }

      // Fetch analytics for each project
      const analyticsPromises = projects.map(project => 
        api.get(`/api/projects/${project.id}/analytics/summary`)
          .then(res => res.data.summary || res.data)
          .catch(() => ({
            total_events: 0,
            total_cost: 0,
            average_latency: 0,
            error_rate: 0,
          }))
      )

      const projectAnalytics = await Promise.all(analyticsPromises)

      // Aggregate the data
      const aggregated = projectAnalytics.reduce((acc, curr) => ({
        total_events: acc.total_events + (curr.total_events || 0),
        total_cost: acc.total_cost + (curr.total_cost || 0),
        average_latency: acc.average_latency + (curr.average_latency || 0),
        error_rate: acc.error_rate + (curr.error_rate || 0),
      }), {
        total_events: 0,
        total_cost: 0,
        average_latency: 0,
        error_rate: 0,
      })

      // Calculate averages
      const projectCount = projectAnalytics.filter(p => p.total_events > 0).length || 1
      
      return {
        total_events: aggregated.total_events,
        total_cost: aggregated.total_cost,
        average_latency: aggregated.average_latency / projectCount,
        error_rate: aggregated.error_rate / projectCount,
        active_projects: projects.length,
        events_24h: 0, // Would need time-series data to calculate 24h metrics
        cost_24h: 0, // Would need time-series data to calculate 24h metrics
        events_change: 0, // Would need historical data to calculate changes
        cost_change: 0,
        latency_change: 0,
        error_rate_change: 0,
      } as DashboardStats
    },
    enabled: !!projects,
    refetchInterval: 30000,
  })

  const isLoading = projectsLoading || statsLoading

  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 4,
    }).format(amount)
  }

  const formatNumber = (num: number) => {
    return new Intl.NumberFormat('en-US').format(num)
  }

  const formatLatency = (ms: number) => {
    return `${ms.toFixed(0)}ms`
  }

  const formatPercentage = (rate: number) => {
    return `${rate.toFixed(2)}%`
  }

  if (isLoading || !stats) {
    return (
      <div className="page-content">
        <div className="mb-8">
          <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
          <p className="text-gray-600">Overview of your LLM observability metrics</p>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="metric-card">
              <div className="loading-skeleton h-4 w-20 mb-2" />
              <div className="loading-skeleton h-8 w-24 mb-2" />
              <div className="loading-skeleton h-3 w-16" />
            </div>
          ))}
        </div>
      </div>
    )
  }

  return (
    <div className="page-content">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600">Overview of your LLM observability metrics</p>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        <MetricCard
          title="Total Events"
          value={formatNumber(stats.total_events)}
          change={stats.events_change}
          changeLabel="vs last 24h"
          icon={Activity}
          trend={stats.events_change > 0 ? 'up' : stats.events_change < 0 ? 'down' : 'neutral'}
        />
        
        <MetricCard
          title="Total Cost"
          value={formatCurrency(stats.total_cost)}
          change={stats.cost_change}
          changeLabel="vs last 24h"
          icon={DollarSign}
          trend={stats.cost_change > 0 ? 'up' : stats.cost_change < 0 ? 'down' : 'neutral'}
        />
        
        <MetricCard
          title="Avg Latency"
          value={formatLatency(stats.average_latency)}
          change={stats.latency_change}
          changeLabel="vs last 24h"
          icon={Clock}
          trend={stats.latency_change > 0 ? 'down' : stats.latency_change < 0 ? 'up' : 'neutral'}
        />
        
        <MetricCard
          title="Error Rate"
          value={formatPercentage(stats.error_rate)}
          change={stats.error_rate_change}
          changeLabel="vs last 24h"
          icon={AlertTriangle}
          trend={stats.error_rate_change > 0 ? 'down' : stats.error_rate_change < 0 ? 'up' : 'neutral'}
        />
      </div>

      {/* Secondary Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="card p-6">
          <div className="flex items-center">
            <Users className="h-8 w-8 text-primary-600" />
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Active Projects</p>
              <p className="text-2xl font-bold text-gray-900">{stats.active_projects}</p>
            </div>
          </div>
        </div>
        
        <div className="card p-6">
          <div className="flex items-center">
            <Database className="h-8 w-8 text-success-600" />
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Events (24h)</p>
              <p className="text-2xl font-bold text-gray-900">{formatNumber(stats.events_24h)}</p>
            </div>
          </div>
        </div>
        
        <div className="card p-6">
          <div className="flex items-center">
            <TrendingUp className="h-8 w-8 text-warning-600" />
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-600">Cost (24h)</p>
              <p className="text-2xl font-bold text-gray-900">{formatCurrency(stats.cost_24h)}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Charts Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Cost Over Time</h3>
          <CostChart />
        </div>
        
        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Latency Distribution</h3>
          <LatencyChart />
        </div>
        
        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Error Rate Trend</h3>
          <ErrorRateChart />
        </div>
        
        <div className="card p-6">
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Recent Evaluations</h3>
          <RecentEvaluations />
        </div>
      </div>

      {/* Quick Actions */}
      <div className="card p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <button 
            onClick={() => navigate('/projects')}
            className="btn-secondary btn-md text-left"
          >
            <div>
              <p className="font-medium">Create Project</p>
              <p className="text-sm text-gray-500">Set up a new monitoring project</p>
            </div>
          </button>
          
          <button 
            onClick={() => projects?.[0] && navigate(`/projects/${projects[0].id}/analytics`)}
            className="btn-secondary btn-md text-left"
          >
            <div>
              <p className="font-medium">View Analytics</p>
              <p className="text-sm text-gray-500">Explore detailed performance metrics</p>
            </div>
          </button>
          
          <button 
            onClick={() => navigate('/settings')}
            className="btn-secondary btn-md text-left"
          >
            <div>
              <p className="font-medium">Setup SDK</p>
              <p className="text-sm text-gray-500">Integrate with your application</p>
            </div>
          </button>
        </div>
      </div>
    </div>
  )
}