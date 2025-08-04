import { useQuery } from '@tanstack/react-query'
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
import CostChart from '../components/charts/CostChart'
import LatencyChart from '../components/charts/LatencyChart'
import ErrorRateChart from '../components/charts/ErrorRateChart'

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
  const { data: stats, isLoading } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      const response = await api.get('/api/dashboard/stats')
      return response.data as DashboardStats
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  })

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
          <h3 className="text-lg font-semibold text-gray-900 mb-4">Recent Activity</h3>
          <RecentActivity />
        </div>
      </div>

      {/* Quick Actions */}
      <div className="card p-6">
        <h3 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h3>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <button className="btn-secondary btn-md text-left">
            <div>
              <p className="font-medium">Create Project</p>
              <p className="text-sm text-gray-500">Set up a new monitoring project</p>
            </div>
          </button>
          
          <button className="btn-secondary btn-md text-left">
            <div>
              <p className="font-medium">View Analytics</p>
              <p className="text-sm text-gray-500">Explore detailed performance metrics</p>
            </div>
          </button>
          
          <button className="btn-secondary btn-md text-left">
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