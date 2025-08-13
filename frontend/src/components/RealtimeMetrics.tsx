import React, { useMemo } from 'react'
import { Activity, TrendingUp, TrendingDown, AlertCircle, CheckCircle, Clock, Zap, Users } from 'lucide-react'
import { LineChart, Line, AreaChart, Area, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend, PieChart, Pie, Cell } from 'recharts'
import { useRealtimeMetrics } from '../hooks/useRealtimeMetrics.js'
import { format } from 'date-fns'

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']

export function RealtimeMetrics() {
  const { metrics, isConnected, connectionStatus, lastUpdateTime } = useRealtimeMetrics()

  const formattedUpdateTime = useMemo(() => {
    if (!lastUpdateTime) return 'Never'
    return format(lastUpdateTime, 'HH:mm:ss')
  }, [lastUpdateTime])

  const operationData = useMemo(() => {
    if (!metrics?.operationCounts) return []
    return Object.entries(metrics.operationCounts)
      .map(([name, count]) => ({ name, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 5)
  }, [metrics?.operationCounts])

  const statusData = useMemo(() => {
    if (!metrics?.statusCounts) return []
    return Object.entries(metrics.statusCounts)
      .map(([name, value]) => ({ name, value }))
  }, [metrics?.statusCounts])

  const errorTrend = useMemo(() => {
    if (!metrics?.errorRate) return 'stable'
    if (metrics.errorRate > 5) return 'high'
    if (metrics.errorRate > 2) return 'moderate'
    return 'low'
  }, [metrics?.errorRate])

  if (!metrics) {
    return (
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-lg p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Real-time Metrics</h3>
          <div className="flex items-center gap-2">
            <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${
              isConnected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
            }`}>
              <span className={`w-2 h-2 rounded-full mr-1 ${
                isConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'
              }`} />
              {connectionStatus}
            </span>
          </div>
        </div>
        <div className="flex items-center justify-center h-64">
          <div className="text-center">
            <Activity className="w-12 h-12 text-gray-400 mx-auto mb-3 animate-pulse" />
            <p className="text-gray-500">Connecting to real-time metrics...</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Connection Status Bar */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <h2 className="text-xl font-bold text-gray-900 dark:text-white">Real-time Analytics</h2>
            <span className={`inline-flex items-center px-3 py-1 rounded-full text-xs font-medium ${
              isConnected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
            }`}>
              <span className={`w-2 h-2 rounded-full mr-2 ${
                isConnected ? 'bg-green-500 animate-pulse' : 'bg-red-500'
              }`} />
              {connectionStatus}
            </span>
          </div>
          <div className="text-sm text-gray-500">
            Last update: {formattedUpdateTime}
          </div>
        </div>
      </div>

      {/* Key Metrics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Events/Min</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white">
                {metrics.eventsPerMinute.toFixed(1)}
              </p>
            </div>
            <div className="p-3 bg-blue-100 rounded-lg">
              <Zap className="w-6 h-6 text-blue-600" />
            </div>
          </div>
          <div className="mt-2 flex items-center text-sm">
            {metrics.eventsPerMinute > 10 ? (
              <>
                <TrendingUp className="w-4 h-4 text-green-500 mr-1" />
                <span className="text-green-600">High activity</span>
              </>
            ) : (
              <>
                <TrendingDown className="w-4 h-4 text-gray-500 mr-1" />
                <span className="text-gray-600">Normal activity</span>
              </>
            )}
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Error Rate</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white">
                {metrics.errorRate.toFixed(2)}%
              </p>
            </div>
            <div className={`p-3 rounded-lg ${
              errorTrend === 'high' ? 'bg-red-100' : errorTrend === 'moderate' ? 'bg-yellow-100' : 'bg-green-100'
            }`}>
              <AlertCircle className={`w-6 h-6 ${
                errorTrend === 'high' ? 'text-red-600' : errorTrend === 'moderate' ? 'text-yellow-600' : 'text-green-600'
              }`} />
            </div>
          </div>
          <div className="mt-2 flex items-center text-sm">
            {errorTrend === 'high' ? (
              <>
                <AlertCircle className="w-4 h-4 text-red-500 mr-1" />
                <span className="text-red-600">High error rate</span>
              </>
            ) : errorTrend === 'moderate' ? (
              <>
                <AlertCircle className="w-4 h-4 text-yellow-500 mr-1" />
                <span className="text-yellow-600">Moderate errors</span>
              </>
            ) : (
              <>
                <CheckCircle className="w-4 h-4 text-green-500 mr-1" />
                <span className="text-green-600">Healthy</span>
              </>
            )}
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Avg Latency</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white">
                {metrics.avgLatency ? `${metrics.avgLatency.toFixed(0)}ms` : 'N/A'}
              </p>
            </div>
            <div className="p-3 bg-purple-100 rounded-lg">
              <Clock className="w-6 h-6 text-purple-600" />
            </div>
          </div>
          <div className="mt-2 flex items-center text-sm">
            {metrics.avgLatency && metrics.avgLatency < 100 ? (
              <>
                <CheckCircle className="w-4 h-4 text-green-500 mr-1" />
                <span className="text-green-600">Excellent</span>
              </>
            ) : metrics.avgLatency && metrics.avgLatency < 300 ? (
              <>
                <Activity className="w-4 h-4 text-yellow-500 mr-1" />
                <span className="text-yellow-600">Good</span>
              </>
            ) : (
              <>
                <AlertCircle className="w-4 h-4 text-red-500 mr-1" />
                <span className="text-red-600">Slow</span>
              </>
            )}
          </div>
        </div>

        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600 dark:text-gray-400">Active Projects</p>
              <p className="text-2xl font-bold text-gray-900 dark:text-white">
                {metrics.activeProjects}
              </p>
            </div>
            <div className="p-3 bg-indigo-100 rounded-lg">
              <Users className="w-6 h-6 text-indigo-600" />
            </div>
          </div>
          <div className="mt-2 flex items-center text-sm">
            <Activity className="w-4 h-4 text-blue-500 mr-1" />
            <span className="text-blue-600">{metrics.recentEvaluations} recent evaluations</span>
          </div>
        </div>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Top Operations Chart */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Top Operations</h3>
          {operationData.length > 0 ? (
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={operationData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Bar dataKey="count" fill="#3b82f6" />
              </BarChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex items-center justify-center h-[250px] text-gray-500">
              No operation data available
            </div>
          )}
        </div>

        {/* Status Distribution Chart */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow-sm p-6">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Status Distribution</h3>
          {statusData.length > 0 ? (
            <ResponsiveContainer width="100%" height={250}>
              <PieChart>
                <Pie
                  data={statusData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name}: ${(percent * 100).toFixed(0)}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {statusData.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <div className="flex items-center justify-center h-[250px] text-gray-500">
              No status data available
            </div>
          )}
        </div>
      </div>

      {/* Summary Stats */}
      <div className="bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg shadow-lg p-6 text-white">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="text-center">
            <p className="text-3xl font-bold">{metrics.totalEvents.toLocaleString()}</p>
            <p className="text-sm opacity-90">Total Events</p>
          </div>
          <div className="text-center">
            <p className="text-3xl font-bold">{metrics.eventsPerMinute.toFixed(1)}</p>
            <p className="text-sm opacity-90">Events/Min</p>
          </div>
          <div className="text-center">
            <p className="text-3xl font-bold">{metrics.errorRate.toFixed(1)}%</p>
            <p className="text-sm opacity-90">Error Rate</p>
          </div>
          <div className="text-center">
            <p className="text-3xl font-bold">{metrics.activeProjects}</p>
            <p className="text-sm opacity-90">Active Projects</p>
          </div>
        </div>
      </div>
    </div>
  )
}