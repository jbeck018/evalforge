import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  Bot,
  Activity,
  AlertCircle,
  CheckCircle,
  Clock,
  Zap,
  TrendingUp,
  TrendingDown,
  BarChart3,
  Cpu,
  Database,
  Network,
  Settings,
  Play,
  Pause,
  RefreshCw,
  Filter,
  Download,
  ArrowUp,
  ArrowDown,
  Minus,
  Info,
  XCircle
} from 'lucide-react';
import { api } from '../utils/api';
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer
} from 'recharts';

interface Agent {
  id: string;
  name: string;
  type: 'ingestion' | 'evaluation' | 'analytics' | 'monitoring' | 'custom';
  status: 'active' | 'idle' | 'error' | 'offline';
  health: 'healthy' | 'degraded' | 'unhealthy';
  version: string;
  host: string;
  metrics: {
    events_processed: number;
    events_per_second: number;
    error_rate: number;
    avg_processing_time: number;
    memory_usage: number;
    cpu_usage: number;
    queue_size: number;
    last_heartbeat: string;
  };
  configuration: {
    batch_size: number;
    flush_interval: number;
    retry_limit: number;
    timeout: number;
  };
  errors: Array<{
    timestamp: string;
    message: string;
    count: number;
  }>;
  last_activity: string;
  uptime: number;
}

interface AgentPerformance {
  timestamp: string;
  events_per_second: number;
  error_rate: number;
  avg_latency: number;
  memory_mb: number;
  cpu_percent: number;
}

const AGENT_TYPE_COLORS = {
  ingestion: '#10b981',
  evaluation: '#6366f1',
  analytics: '#f59e0b',
  monitoring: '#8b5cf6',
  custom: '#6b7280'
};

const STATUS_COLORS = {
  active: '#10b981',
  idle: '#f59e0b',
  error: '#ef4444',
  offline: '#6b7280'
};

const AgentMonitoringPage: React.FC = () => {
  const navigate = useNavigate();
  const [selectedAgent, setSelectedAgent] = useState<Agent | null>(null);
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [timeRange, setTimeRange] = useState('1h');
  const [filterType, setFilterType] = useState<string>('all');
  const [filterStatus, setFilterStatus] = useState<string>('all');

  // Fetch agents
  const { data: agents = [], isLoading } = useQuery({
    queryKey: ['agents', filterType, filterStatus],
    queryFn: async () => {
      try {
        const response = await api.get('/api/agents');
        let agentList = response.data.agents || [];
        
        // Apply filters
        if (filterType !== 'all') {
          agentList = agentList.filter((a: Agent) => a.type === filterType);
        }
        if (filterStatus !== 'all') {
          agentList = agentList.filter((a: Agent) => a.status === filterStatus);
        }
        
        return agentList;
      } catch {
        // Return mock data if endpoint doesn't exist
        return generateMockAgents();
      }
    },
    refetchInterval: autoRefresh ? 5000 : undefined
  });

  // Fetch performance data for selected agent
  const { data: performanceData = [] } = useQuery({
    queryKey: ['agent-performance', selectedAgent?.id, timeRange],
    queryFn: async () => {
      if (!selectedAgent) return [];
      
      try {
        const response = await api.get(`/api/agents/${selectedAgent.id}/performance`, {
          params: { range: timeRange }
        });
        return response.data.performance || [];
      } catch {
        // Return mock performance data
        return generateMockPerformance();
      }
    },
    enabled: !!selectedAgent,
    refetchInterval: autoRefresh ? 10000 : undefined
  });

  // Calculate aggregate metrics
  const aggregateMetrics = React.useMemo(() => {
    const total = agents.length;
    const active = agents.filter(a => a.status === 'active').length;
    const errors = agents.filter(a => a.status === 'error').length;
    const totalEvents = agents.reduce((sum, a) => sum + a.metrics.events_processed, 0);
    const avgErrorRate = agents.reduce((sum, a) => sum + a.metrics.error_rate, 0) / (total || 1);
    const totalMemory = agents.reduce((sum, a) => sum + a.metrics.memory_usage, 0);
    const avgCpu = agents.reduce((sum, a) => sum + a.metrics.cpu_usage, 0) / (total || 1);
    
    return {
      total,
      active,
      errors,
      totalEvents,
      avgErrorRate,
      totalMemory,
      avgCpu
    };
  }, [agents]);

  // Agent type distribution for pie chart
  const typeDistribution = React.useMemo(() => {
    const distribution = agents.reduce((acc, agent) => {
      acc[agent.type] = (acc[agent.type] || 0) + 1;
      return acc;
    }, {} as Record<string, number>);
    
    return Object.entries(distribution).map(([type, count]) => ({
      name: type.charAt(0).toUpperCase() + type.slice(1),
      value: count,
      color: AGENT_TYPE_COLORS[type as keyof typeof AGENT_TYPE_COLORS]
    }));
  }, [agents]);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'active':
        return <Play className="h-4 w-4 text-green-600" />;
      case 'idle':
        return <Pause className="h-4 w-4 text-yellow-600" />;
      case 'error':
        return <XCircle className="h-4 w-4 text-red-600" />;
      case 'offline':
        return <Minus className="h-4 w-4 text-gray-600" />;
      default:
        return null;
    }
  };

  const getHealthBadge = (health: string) => {
    const colors = {
      healthy: 'bg-green-100 text-green-800',
      degraded: 'bg-yellow-100 text-yellow-800',
      unhealthy: 'bg-red-100 text-red-800'
    };
    
    return (
      <span className={`px-2 py-1 text-xs rounded-full ${colors[health as keyof typeof colors]}`}>
        {health}
      </span>
    );
  };

  const formatUptime = (seconds: number) => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    
    if (days > 0) return `${days}d ${hours}h`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  };

  const handleRestartAgent = async (agentId: string) => {
    try {
      await api.post(`/api/agents/${agentId}/restart`);
      // Refresh data
    } catch (error) {
      console.error('Failed to restart agent:', error);
    }
  };

  const handleExportMetrics = () => {
    const data = {
      timestamp: new Date().toISOString(),
      agents,
      aggregateMetrics,
      performanceData
    };
    
    const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `agent-metrics-${Date.now()}.json`;
    a.click();
  };

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
      </div>
    );
  }

  return (
    <div className="p-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="bg-white shadow rounded-lg p-6 mb-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold text-gray-900 flex items-center">
              <Bot className="h-6 w-6 mr-2 text-indigo-600" />
              Agent Monitoring
            </h1>
            <p className="mt-1 text-sm text-gray-500">
              Monitor and manage your LLM observability agents
            </p>
          </div>
          <div className="flex items-center space-x-4">
            <select
              value={timeRange}
              onChange={(e) => setTimeRange(e.target.value)}
              className="px-3 py-2 border border-gray-300 rounded-md text-sm"
            >
              <option value="1h">Last 1 hour</option>
              <option value="6h">Last 6 hours</option>
              <option value="24h">Last 24 hours</option>
              <option value="7d">Last 7 days</option>
            </select>
            <button
              onClick={() => setAutoRefresh(!autoRefresh)}
              className={`flex items-center px-3 py-2 rounded-md text-sm font-medium ${
                autoRefresh 
                  ? 'bg-green-100 text-green-700' 
                  : 'bg-gray-100 text-gray-700'
              }`}
            >
              <RefreshCw className={`h-4 w-4 mr-2 ${autoRefresh ? 'animate-spin' : ''}`} />
              {autoRefresh ? 'Auto-refresh ON' : 'Auto-refresh OFF'}
            </button>
            <button
              onClick={handleExportMetrics}
              className="bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-indigo-700 flex items-center"
            >
              <Download className="h-4 w-4 mr-2" />
              Export Metrics
            </button>
          </div>
        </div>
      </div>

      {/* Aggregate Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Total Agents</p>
              <p className="text-2xl font-bold text-gray-900">{aggregateMetrics.total}</p>
              <p className="text-sm text-green-600">
                {aggregateMetrics.active} active
              </p>
            </div>
            <Bot className="h-8 w-8 text-indigo-600" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Events Processed</p>
              <p className="text-2xl font-bold text-gray-900">
                {aggregateMetrics.totalEvents.toLocaleString()}
              </p>
              <p className="text-sm text-gray-500">Total all-time</p>
            </div>
            <Activity className="h-8 w-8 text-green-600" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Avg Error Rate</p>
              <p className="text-2xl font-bold text-gray-900">
                {(aggregateMetrics.avgErrorRate * 100).toFixed(2)}%
              </p>
              <p className={`text-sm ${
                aggregateMetrics.avgErrorRate > 0.05 ? 'text-red-600' : 'text-green-600'
              }`}>
                {aggregateMetrics.avgErrorRate > 0.05 ? (
                  <span className="flex items-center">
                    <ArrowUp className="h-3 w-3 mr-1" />
                    Above threshold
                  </span>
                ) : (
                  <span className="flex items-center">
                    <ArrowDown className="h-3 w-3 mr-1" />
                    Within limits
                  </span>
                )}
              </p>
            </div>
            <AlertCircle className="h-8 w-8 text-yellow-600" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Resource Usage</p>
              <p className="text-2xl font-bold text-gray-900">
                {aggregateMetrics.avgCpu.toFixed(1)}%
              </p>
              <p className="text-sm text-gray-500">
                {(aggregateMetrics.totalMemory / 1024).toFixed(1)} GB RAM
              </p>
            </div>
            <Cpu className="h-8 w-8 text-purple-600" />
          </div>
        </div>
      </div>

      {/* Filters */}
      <div className="bg-white rounded-lg shadow p-4 mb-6">
        <div className="flex items-center space-x-4">
          <Filter className="h-5 w-5 text-gray-500" />
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md text-sm"
          >
            <option value="all">All Types</option>
            <option value="ingestion">Ingestion</option>
            <option value="evaluation">Evaluation</option>
            <option value="analytics">Analytics</option>
            <option value="monitoring">Monitoring</option>
            <option value="custom">Custom</option>
          </select>
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="px-3 py-2 border border-gray-300 rounded-md text-sm"
          >
            <option value="all">All Status</option>
            <option value="active">Active</option>
            <option value="idle">Idle</option>
            <option value="error">Error</option>
            <option value="offline">Offline</option>
          </select>
          {aggregateMetrics.errors > 0 && (
            <div className="ml-auto flex items-center text-red-600">
              <AlertCircle className="h-5 w-5 mr-2" />
              <span className="text-sm font-medium">
                {aggregateMetrics.errors} agents with errors
              </span>
            </div>
          )}
        </div>
      </div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Agent List */}
        <div className="lg:col-span-2">
          <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900">Active Agents</h2>
            </div>
            <div className="divide-y divide-gray-200 max-h-96 overflow-y-auto">
              {agents.length > 0 ? (
                agents.map((agent) => (
                  <div
                    key={agent.id}
                    className={`p-4 hover:bg-gray-50 cursor-pointer ${
                      selectedAgent?.id === agent.id ? 'bg-indigo-50' : ''
                    }`}
                    onClick={() => setSelectedAgent(agent)}
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center">
                        <div className={`h-10 w-10 rounded-lg flex items-center justify-center mr-3`}
                             style={{ backgroundColor: `${AGENT_TYPE_COLORS[agent.type]}20` }}>
                          <Bot className="h-6 w-6" style={{ color: AGENT_TYPE_COLORS[agent.type] }} />
                        </div>
                        <div>
                          <div className="flex items-center">
                            <h3 className="font-medium text-gray-900">{agent.name}</h3>
                            {getStatusIcon(agent.status)}
                            <span className="ml-2 text-xs text-gray-500">v{agent.version}</span>
                          </div>
                          <p className="text-sm text-gray-500">{agent.host}</p>
                        </div>
                      </div>
                      <div className="text-right">
                        {getHealthBadge(agent.health)}
                        <p className="text-xs text-gray-500 mt-1">
                          Uptime: {formatUptime(agent.uptime)}
                        </p>
                      </div>
                    </div>
                    
                    <div className="mt-3 grid grid-cols-4 gap-2 text-xs">
                      <div>
                        <span className="text-gray-500">Events:</span>
                        <span className="ml-1 font-medium">{agent.metrics.events_processed.toLocaleString()}</span>
                      </div>
                      <div>
                        <span className="text-gray-500">Rate:</span>
                        <span className="ml-1 font-medium">{agent.metrics.events_per_second}/s</span>
                      </div>
                      <div>
                        <span className="text-gray-500">Errors:</span>
                        <span className={`ml-1 font-medium ${
                          agent.metrics.error_rate > 0.05 ? 'text-red-600' : 'text-green-600'
                        }`}>
                          {(agent.metrics.error_rate * 100).toFixed(2)}%
                        </span>
                      </div>
                      <div>
                        <span className="text-gray-500">Queue:</span>
                        <span className="ml-1 font-medium">{agent.metrics.queue_size}</span>
                      </div>
                    </div>
                  </div>
                ))
              ) : (
                <div className="p-8 text-center text-gray-500">
                  <Bot className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  <p>No agents found</p>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Agent Distribution */}
        <div className="space-y-6">
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Agent Distribution</h3>
            <ResponsiveContainer width="100%" height={200}>
              <PieChart>
                <Pie
                  data={typeDistribution}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ name, percent }) => `${name} ${(percent * 100).toFixed(0)}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="value"
                >
                  {typeDistribution.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip />
              </PieChart>
            </ResponsiveContainer>
          </div>

          {/* Recent Errors */}
          {selectedAgent && selectedAgent.errors.length > 0 && (
            <div className="bg-white rounded-lg shadow p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Recent Errors</h3>
              <div className="space-y-2">
                {selectedAgent.errors.slice(0, 5).map((error, index) => (
                  <div key={index} className="p-3 bg-red-50 rounded-md">
                    <div className="flex items-start">
                      <AlertCircle className="h-4 w-4 text-red-600 mt-0.5 mr-2" />
                      <div className="flex-1">
                        <p className="text-sm text-red-800">{error.message}</p>
                        <div className="flex justify-between mt-1">
                          <span className="text-xs text-red-600">
                            {new Date(error.timestamp).toLocaleTimeString()}
                          </span>
                          <span className="text-xs text-red-600 font-medium">
                            {error.count}x
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Performance Charts */}
      {selectedAgent && (
        <div className="mt-6 grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">
              Events & Errors - {selectedAgent.name}
            </h3>
            <ResponsiveContainer width="100%" height={250}>
              <LineChart data={performanceData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  dataKey="timestamp" 
                  tickFormatter={(ts) => new Date(ts).toLocaleTimeString()}
                />
                <YAxis yAxisId="left" />
                <YAxis yAxisId="right" orientation="right" />
                <Tooltip 
                  labelFormatter={(ts) => new Date(ts).toLocaleString()}
                />
                <Legend />
                <Line 
                  yAxisId="left"
                  type="monotone" 
                  dataKey="events_per_second" 
                  stroke="#10b981" 
                  name="Events/s"
                  strokeWidth={2}
                />
                <Line 
                  yAxisId="right"
                  type="monotone" 
                  dataKey="error_rate" 
                  stroke="#ef4444" 
                  name="Error Rate %"
                  strokeWidth={2}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>

          <div className="bg-white rounded-lg shadow p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">
              Resource Usage - {selectedAgent.name}
            </h3>
            <ResponsiveContainer width="100%" height={250}>
              <AreaChart data={performanceData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  dataKey="timestamp" 
                  tickFormatter={(ts) => new Date(ts).toLocaleTimeString()}
                />
                <YAxis />
                <Tooltip 
                  labelFormatter={(ts) => new Date(ts).toLocaleString()}
                />
                <Legend />
                <Area 
                  type="monotone" 
                  dataKey="cpu_percent" 
                  stackId="1"
                  stroke="#8b5cf6" 
                  fill="#8b5cf6"
                  name="CPU %"
                />
                <Area 
                  type="monotone" 
                  dataKey="memory_mb" 
                  stackId="2"
                  stroke="#f59e0b" 
                  fill="#f59e0b"
                  name="Memory MB"
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      )}

      {/* Agent Details */}
      {selectedAgent && (
        <div className="mt-6 bg-white rounded-lg shadow p-6">
          <div className="flex justify-between items-center mb-4">
            <h3 className="text-lg font-semibold text-gray-900">
              Agent Configuration - {selectedAgent.name}
            </h3>
            <button
              onClick={() => handleRestartAgent(selectedAgent.id)}
              className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-50 flex items-center"
            >
              <RefreshCw className="h-4 w-4 mr-2" />
              Restart Agent
            </button>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="p-4 bg-gray-50 rounded-md">
              <p className="text-sm text-gray-600 mb-1">Batch Size</p>
              <p className="text-lg font-semibold text-gray-900">
                {selectedAgent.configuration.batch_size}
              </p>
            </div>
            <div className="p-4 bg-gray-50 rounded-md">
              <p className="text-sm text-gray-600 mb-1">Flush Interval</p>
              <p className="text-lg font-semibold text-gray-900">
                {selectedAgent.configuration.flush_interval}s
              </p>
            </div>
            <div className="p-4 bg-gray-50 rounded-md">
              <p className="text-sm text-gray-600 mb-1">Retry Limit</p>
              <p className="text-lg font-semibold text-gray-900">
                {selectedAgent.configuration.retry_limit}
              </p>
            </div>
            <div className="p-4 bg-gray-50 rounded-md">
              <p className="text-sm text-gray-600 mb-1">Timeout</p>
              <p className="text-lg font-semibold text-gray-900">
                {selectedAgent.configuration.timeout}s
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

// Mock data generators
function generateMockAgents(): Agent[] {
  return [
    {
      id: 'agent-1',
      name: 'Primary Trace Collector',
      type: 'ingestion',
      status: 'active',
      health: 'healthy',
      version: '1.2.3',
      host: 'collector-01.evalforge.io',
      metrics: {
        events_processed: 1523456,
        events_per_second: 245,
        error_rate: 0.002,
        avg_processing_time: 12,
        memory_usage: 512,
        cpu_usage: 35,
        queue_size: 128,
        last_heartbeat: new Date().toISOString()
      },
      configuration: {
        batch_size: 100,
        flush_interval: 5,
        retry_limit: 3,
        timeout: 30
      },
      errors: [],
      last_activity: new Date().toISOString(),
      uptime: 864000
    },
    {
      id: 'agent-2',
      name: 'Auto Evaluator',
      type: 'evaluation',
      status: 'active',
      health: 'healthy',
      version: '1.1.0',
      host: 'evaluator-01.evalforge.io',
      metrics: {
        events_processed: 892134,
        events_per_second: 120,
        error_rate: 0.01,
        avg_processing_time: 45,
        memory_usage: 1024,
        cpu_usage: 55,
        queue_size: 256,
        last_heartbeat: new Date().toISOString()
      },
      configuration: {
        batch_size: 50,
        flush_interval: 10,
        retry_limit: 5,
        timeout: 60
      },
      errors: [
        {
          timestamp: new Date(Date.now() - 3600000).toISOString(),
          message: 'Failed to evaluate trace: timeout exceeded',
          count: 3
        }
      ],
      last_activity: new Date().toISOString(),
      uptime: 432000
    },
    {
      id: 'agent-3',
      name: 'Cost Analyzer',
      type: 'analytics',
      status: 'idle',
      health: 'healthy',
      version: '2.0.1',
      host: 'analytics-01.evalforge.io',
      metrics: {
        events_processed: 564321,
        events_per_second: 0,
        error_rate: 0,
        avg_processing_time: 8,
        memory_usage: 256,
        cpu_usage: 5,
        queue_size: 0,
        last_heartbeat: new Date(Date.now() - 60000).toISOString()
      },
      configuration: {
        batch_size: 200,
        flush_interval: 30,
        retry_limit: 3,
        timeout: 30
      },
      errors: [],
      last_activity: new Date(Date.now() - 3600000).toISOString(),
      uptime: 1728000
    },
    {
      id: 'agent-4',
      name: 'Metrics Aggregator',
      type: 'monitoring',
      status: 'active',
      health: 'degraded',
      version: '1.0.5',
      host: 'monitor-01.evalforge.io',
      metrics: {
        events_processed: 2134567,
        events_per_second: 380,
        error_rate: 0.08,
        avg_processing_time: 6,
        memory_usage: 768,
        cpu_usage: 70,
        queue_size: 512,
        last_heartbeat: new Date().toISOString()
      },
      configuration: {
        batch_size: 500,
        flush_interval: 1,
        retry_limit: 2,
        timeout: 10
      },
      errors: [
        {
          timestamp: new Date(Date.now() - 300000).toISOString(),
          message: 'High memory usage detected',
          count: 15
        },
        {
          timestamp: new Date(Date.now() - 600000).toISOString(),
          message: 'Queue overflow warning',
          count: 8
        }
      ],
      last_activity: new Date().toISOString(),
      uptime: 259200
    }
  ];
}

function generateMockPerformance(): AgentPerformance[] {
  const data: AgentPerformance[] = [];
  const now = Date.now();
  
  for (let i = 23; i >= 0; i--) {
    data.push({
      timestamp: new Date(now - i * 3600000).toISOString(),
      events_per_second: Math.floor(Math.random() * 300) + 100,
      error_rate: Math.random() * 0.1,
      avg_latency: Math.floor(Math.random() * 50) + 10,
      memory_mb: Math.floor(Math.random() * 500) + 200,
      cpu_percent: Math.floor(Math.random() * 60) + 20
    });
  }
  
  return data;
}

export default AgentMonitoringPage;