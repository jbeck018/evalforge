import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import {
  Activity,
  TrendingUp,
  AlertCircle,
  CheckCircle,
  Clock,
  DollarSign,
  Zap,
  Users,
  GitBranch,
  BarChart3,
  Brain,
  Bot,
  FileText,
  ArrowRight,
  Play,
  Pause,
  RefreshCw,
  Settings,
  Plus
} from 'lucide-react';
import { api } from '../utils/api';

interface DashboardStats {
  totalProjects: number;
  activeEvaluations: number;
  runningAgents: number;
  totalEvents: number;
  successRate: number;
  avgLatency: number;
  totalCost: number;
  errorCount: number;
  activeABTests: number;
  modelsInUse: number;
}

interface Evaluation {
  id: number;
  name: string;
  project_name: string;
  status: 'running' | 'completed' | 'failed' | 'pending';
  progress: number;
  score?: number;
  last_run: string;
}

interface Agent {
  id: string;
  name: string;
  type: string;
  status: 'active' | 'idle' | 'error';
  events_processed: number;
  last_activity: string;
  error_rate: number;
}

interface ABTest {
  id: number;
  name: string;
  status: 'running' | 'completed' | 'draft';
  control_samples: number;
  variant_samples: number;
  winner?: string;
  improvement?: number;
}

interface Model {
  model: string;
  provider: string;
  total_events: number;
  avg_latency: number;
  cost: number;
  error_rate: number;
}

const ImprovedDashboard: React.FC = () => {
  const navigate = useNavigate();
  const [autoRefresh, setAutoRefresh] = useState(true);

  // Fetch dashboard stats
  const { data: stats, isLoading: statsLoading } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: async () => {
      try {
        // Fetch projects
        const projectsRes = await api.get('/api/projects');
        const projects = projectsRes.data.projects || [];
        
        // Aggregate stats from projects
        let totalEvents = 0;
        let totalCost = 0;
        let totalLatency = 0;
        let errorCount = 0;
        
        for (const project of projects) {
          try {
            const analyticsRes = await api.get(`/api/projects/${project.id}/analytics/summary`);
            const summary = analyticsRes.data.summary || {};
            totalEvents += summary.total_events || 0;
            totalCost += summary.total_cost || 0;
            totalLatency += summary.average_latency || 0;
            errorCount += summary.error_count || 0;
          } catch (e) {
            // Skip projects with no data
          }
        }
        
        return {
          totalProjects: projects.length,
          activeEvaluations: 0, // Will be fetched separately
          runningAgents: 3, // Simulated
          totalEvents,
          successRate: errorCount > 0 ? (totalEvents - errorCount) / totalEvents : 1,
          avgLatency: projects.length > 0 ? totalLatency / projects.length : 0,
          totalCost,
          errorCount,
          activeABTests: 0, // Will be fetched separately
          modelsInUse: 0, // Will be fetched separately
        };
      } catch (error) {
        console.error('Failed to fetch stats:', error);
        return {
          totalProjects: 0,
          activeEvaluations: 0,
          runningAgents: 0,
          totalEvents: 0,
          successRate: 0,
          avgLatency: 0,
          totalCost: 0,
          errorCount: 0,
          activeABTests: 0,
          modelsInUse: 0,
        };
      }
    },
    refetchInterval: autoRefresh ? 10000 : undefined,
  });

  // Fetch evaluations
  const { data: evaluations = [] } = useQuery({
    queryKey: ['evaluations'],
    queryFn: async () => {
      try {
        const response = await api.get('/api/evaluations');
        return (response.data.evaluations || []).slice(0, 5);
      } catch {
        return [];
      }
    },
    refetchInterval: autoRefresh ? 10000 : undefined,
  });

  // Simulated agents data
  const agents: Agent[] = [
    {
      id: 'agent-1',
      name: 'Trace Collector',
      type: 'ingestion',
      status: 'active',
      events_processed: 15234,
      last_activity: new Date().toISOString(),
      error_rate: 0.02
    },
    {
      id: 'agent-2',
      name: 'Auto Evaluator',
      type: 'evaluation',
      status: 'active',
      events_processed: 8921,
      last_activity: new Date().toISOString(),
      error_rate: 0.01
    },
    {
      id: 'agent-3',
      name: 'Cost Analyzer',
      type: 'analytics',
      status: 'idle',
      events_processed: 5643,
      last_activity: new Date(Date.now() - 3600000).toISOString(),
      error_rate: 0
    }
  ];

  // Fetch A/B tests
  const { data: abTests = [] } = useQuery({
    queryKey: ['ab-tests'],
    queryFn: async () => {
      try {
        const projectsRes = await api.get('/api/projects');
        const projects = projectsRes.data.projects || [];
        
        const allTests = [];
        for (const project of projects.slice(0, 3)) {
          try {
            const testsRes = await api.get(`/api/projects/${project.id}/abtests`);
            if (testsRes.data.tests) {
              allTests.push(...testsRes.data.tests);
            }
          } catch {
            // Skip projects with no tests
          }
        }
        
        return allTests.slice(0, 3);
      } catch {
        return [];
      }
    },
    refetchInterval: autoRefresh ? 10000 : undefined,
  });

  // Fetch models in use
  const { data: models = [] } = useQuery({
    queryKey: ['models'],
    queryFn: async () => {
      try {
        const projectsRes = await api.get('/api/projects');
        const projects = projectsRes.data.projects || [];
        
        if (projects.length === 0) return [];
        
        // Get model comparison for first project
        const project = projects[0];
        const comparisonRes = await api.get(`/api/projects/${project.id}/model-comparison`);
        return comparisonRes.data.comparison?.models || [];
      } catch {
        return [];
      }
    },
    refetchInterval: autoRefresh ? 30000 : undefined,
  });

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'running':
      case 'active':
      case 'success':
        return 'text-green-600';
      case 'pending':
      case 'idle':
      case 'info':
        return 'text-blue-600';
      case 'failed':
      case 'error':
        return 'text-red-600';
      case 'warning':
        return 'text-yellow-600';
      default:
        return 'text-gray-600';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running':
      case 'active':
        return <Play className="h-4 w-4" />;
      case 'completed':
      case 'success':
        return <CheckCircle className="h-4 w-4" />;
      case 'failed':
      case 'error':
        return <AlertCircle className="h-4 w-4" />;
      case 'pending':
      case 'idle':
        return <Clock className="h-4 w-4" />;
      case 'paused':
        return <Pause className="h-4 w-4" />;
      default:
        return null;
    }
  };

  if (statsLoading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="bg-white shadow rounded-lg p-6">
        <div className="flex justify-between items-center">
          <div>
            <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
            <p className="mt-1 text-sm text-gray-500">
              Real-time monitoring and insights for your LLM operations
            </p>
          </div>
          <div className="flex items-center space-x-4">
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
              onClick={() => navigate('/llm-config')}
              className="bg-indigo-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-indigo-700 flex items-center"
            >
              <Settings className="h-4 w-4 mr-2" />
              Configure LLMs
            </button>
          </div>
        </div>
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Active Evaluations</p>
              <p className="text-2xl font-bold text-gray-900">{evaluations.length}</p>
            </div>
            <Brain className="h-8 w-8 text-indigo-600" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Running Agents</p>
              <p className="text-2xl font-bold text-gray-900">{agents.filter(a => a.status === 'active').length}</p>
            </div>
            <Bot className="h-8 w-8 text-green-600" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Success Rate</p>
              <p className="text-2xl font-bold text-gray-900">
                {((stats?.successRate || 0) * 100).toFixed(1)}%
              </p>
            </div>
            <TrendingUp className="h-8 w-8 text-blue-600" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Avg Latency</p>
              <p className="text-2xl font-bold text-gray-900">
                {(stats?.avgLatency || 0).toFixed(0)}ms
              </p>
            </div>
            <Zap className="h-8 w-8 text-yellow-600" />
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Total Cost</p>
              <p className="text-2xl font-bold text-gray-900">
                ${(stats?.totalCost || 0).toFixed(2)}
              </p>
            </div>
            <DollarSign className="h-8 w-8 text-red-600" />
          </div>
        </div>
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Evaluations Panel */}
        <div className="bg-white rounded-lg shadow">
          <div className="p-6 border-b border-gray-200">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-gray-900 flex items-center">
                <Brain className="h-5 w-5 mr-2 text-indigo-600" />
                Active Evaluations
              </h2>
              <button
                onClick={() => navigate('/evaluations')}
                className="text-sm text-indigo-600 hover:text-indigo-800 flex items-center"
              >
                View all
                <ArrowRight className="h-4 w-4 ml-1" />
              </button>
            </div>
          </div>
          <div className="p-6 space-y-4 max-h-96 overflow-y-auto">
            {evaluations.length > 0 ? (
              evaluations.map((evaluation: Evaluation) => (
                <div key={evaluation.id} className="border rounded-lg p-4">
                  <div className="flex items-center justify-between mb-2">
                    <h3 className="font-medium text-gray-900">{evaluation.name}</h3>
                    <span className={`flex items-center ${getStatusColor(evaluation.status)}`}>
                      {getStatusIcon(evaluation.status)}
                      <span className="ml-1 text-sm">{evaluation.status}</span>
                    </span>
                  </div>
                  <p className="text-sm text-gray-500 mb-2">{evaluation.project_name}</p>
                  {evaluation.status === 'running' && (
                    <div className="w-full bg-gray-200 rounded-full h-2">
                      <div 
                        className="bg-indigo-600 h-2 rounded-full transition-all duration-300"
                        style={{ width: `${evaluation.progress}%` }}
                      />
                    </div>
                  )}
                  {evaluation.score !== undefined && (
                    <p className="text-sm mt-2">
                      Score: <span className="font-semibold">{evaluation.score.toFixed(2)}%</span>
                    </p>
                  )}
                </div>
              ))
            ) : (
              <div className="text-center py-8 text-gray-500">
                <Brain className="h-12 w-12 mx-auto mb-3 text-gray-300" />
                <p>No active evaluations</p>
                <button
                  onClick={() => navigate('/evaluations')}
                  className="mt-3 text-indigo-600 hover:text-indigo-800 text-sm flex items-center mx-auto"
                >
                  <Plus className="h-4 w-4 mr-1" />
                  Create evaluation
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Agents Panel */}
        <div className="bg-white rounded-lg shadow">
          <div className="p-6 border-b border-gray-200">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-semibold text-gray-900 flex items-center">
                <Bot className="h-5 w-5 mr-2 text-green-600" />
                Agent Status
              </h2>
              <button
                onClick={() => navigate('/agents')}
                className="text-sm text-indigo-600 hover:text-indigo-800 flex items-center"
              >
                Manage
                <ArrowRight className="h-4 w-4 ml-1" />
              </button>
            </div>
          </div>
          <div className="p-6 space-y-4 max-h-96 overflow-y-auto">
            {agents.map((agent) => (
              <div key={agent.id} className="border rounded-lg p-4">
                <div className="flex items-center justify-between mb-2">
                  <h3 className="font-medium text-gray-900">{agent.name}</h3>
                  <span className={`flex items-center ${getStatusColor(agent.status)}`}>
                    {getStatusIcon(agent.status)}
                    <span className="ml-1 text-sm">{agent.status}</span>
                  </span>
                </div>
                <div className="grid grid-cols-2 gap-2 text-sm">
                  <div>
                    <span className="text-gray-500">Type:</span>
                    <span className="ml-2 font-medium">{agent.type}</span>
                  </div>
                  <div>
                    <span className="text-gray-500">Events:</span>
                    <span className="ml-2 font-medium">{agent.events_processed.toLocaleString()}</span>
                  </div>
                  <div className="col-span-2">
                    <span className="text-gray-500">Error Rate:</span>
                    <span className={`ml-2 font-medium ${
                      agent.error_rate > 0.05 ? 'text-red-600' : 'text-green-600'
                    }`}>
                      {(agent.error_rate * 100).toFixed(2)}%
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Models & A/B Tests Panel */}
        <div className="space-y-6">
          {/* Models in Use */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-gray-900 flex items-center">
                  <Settings className="h-5 w-5 mr-2 text-blue-600" />
                  Models in Use
                </h2>
                <button
                  onClick={() => navigate('/models')}
                  className="text-sm text-indigo-600 hover:text-indigo-800 flex items-center"
                >
                  Compare
                  <ArrowRight className="h-4 w-4 ml-1" />
                </button>
              </div>
            </div>
            <div className="p-6 space-y-3 max-h-64 overflow-y-auto">
              {models.length > 0 ? (
                models.slice(0, 3).map((model: Model, index: number) => (
                  <div key={index} className="border rounded-lg p-3">
                    <div className="flex items-center justify-between mb-2">
                      <h3 className="font-medium text-gray-900 text-sm">{model.model}</h3>
                      <span className="text-xs text-gray-500">{model.provider}</span>
                    </div>
                    <div className="grid grid-cols-2 gap-2 text-xs text-gray-500">
                      <span>Events: {model.total_events}</span>
                      <span>Cost: ${(model.cost || 0).toFixed(4)}</span>
                      <span>Latency: {(model.avg_latency || 0).toFixed(0)}ms</span>
                      <span>Errors: {((model.error_rate || 0) * 100).toFixed(1)}%</span>
                    </div>
                  </div>
                ))
              ) : (
                <p className="text-center text-gray-500 text-sm py-4">No models configured</p>
              )}
            </div>
          </div>

          {/* A/B Tests */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-gray-900 flex items-center">
                  <GitBranch className="h-5 w-5 mr-2 text-purple-600" />
                  Active A/B Tests
                </h2>
                <button
                  onClick={() => navigate('/ab-tests')}
                  className="text-sm text-indigo-600 hover:text-indigo-800 flex items-center"
                >
                  View all
                  <ArrowRight className="h-4 w-4 ml-1" />
                </button>
              </div>
            </div>
            <div className="p-6 space-y-3">
              {abTests.length > 0 ? (
                abTests.slice(0, 3).map((test: ABTest) => (
                  <div key={test.id} className="border rounded-lg p-3">
                    <div className="flex items-center justify-between mb-2">
                      <h3 className="font-medium text-gray-900 text-sm">{test.name}</h3>
                      <span className={`text-xs ${getStatusColor(test.status)}`}>
                        {test.status}
                      </span>
                    </div>
                    <div className="flex justify-between text-xs text-gray-500">
                      <span>Control: {test.control_samples}</span>
                      <span>Variant: {test.variant_samples}</span>
                    </div>
                    {test.winner && (
                      <p className="text-xs mt-2 text-green-600">
                        Winner: {test.winner} (+{test.improvement?.toFixed(1)}%)
                      </p>
                    )}
                  </div>
                ))
              ) : (
                <p className="text-center text-gray-500 text-sm py-4">No active A/B tests</p>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
        <div className="grid grid-cols-2 md:grid-cols-5 gap-4">
          <button
            onClick={() => navigate('/evaluations')}
            className="flex flex-col items-center p-4 border rounded-lg hover:bg-gray-50"
          >
            <Brain className="h-8 w-8 text-indigo-600 mb-2" />
            <span className="text-sm font-medium">New Evaluation</span>
          </button>
          <button
            onClick={() => navigate('/agents')}
            className="flex flex-col items-center p-4 border rounded-lg hover:bg-gray-50"
          >
            <Bot className="h-8 w-8 text-green-600 mb-2" />
            <span className="text-sm font-medium">Deploy Agent</span>
          </button>
          <button
            onClick={() => navigate('/ab-tests')}
            className="flex flex-col items-center p-4 border rounded-lg hover:bg-gray-50"
          >
            <GitBranch className="h-8 w-8 text-purple-600 mb-2" />
            <span className="text-sm font-medium">Start A/B Test</span>
          </button>
          <button
            onClick={() => navigate('/analytics')}
            className="flex flex-col items-center p-4 border rounded-lg hover:bg-gray-50"
          >
            <BarChart3 className="h-8 w-8 text-blue-600 mb-2" />
            <span className="text-sm font-medium">View Analytics</span>
          </button>
          <button
            onClick={() => navigate('/llm-config')}
            className="flex flex-col items-center p-4 border rounded-lg hover:bg-gray-50"
          >
            <Settings className="h-8 w-8 text-gray-600 mb-2" />
            <span className="text-sm font-medium">Configure LLMs</span>
          </button>
        </div>
      </div>
    </div>
  );
};

export default ImprovedDashboard;