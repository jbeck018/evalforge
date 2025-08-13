import React, { useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Clock, CheckCircle, XCircle, AlertCircle, TrendingUp, TrendingDown, Play, RefreshCw, FileText, Zap } from 'lucide-react'

import { api } from '../utils/api'
import MetricsDashboard from '../components/MetricsDashboard'
import SuggestionCards from '../components/SuggestionCards'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'

// Types (reusing from EvaluationsPage)
interface Evaluation {
  id: string
  name: string
  description: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  progress: number
  created_at: string
  completed_at?: string
  metrics?: EvaluationMetrics
  suggestions?: OptimizationSuggestion[]
  test_results?: TestResult[]
  prompt?: string
  cost_analysis?: CostAnalysis
}

interface EvaluationMetrics {
  overall_score: number
  pass_rate: number
  test_cases_passed: number
  test_cases_total: number
  classification_metrics?: ClassificationMetrics
  generation_metrics?: GenerationMetrics
  custom_metrics: Record<string, number>
  latency_stats?: {
    avg_latency: number
    min_latency: number
    max_latency: number
    p95_latency: number
  }
}

interface ClassificationMetrics {
  accuracy: number
  precision: Record<string, number>
  recall: Record<string, number>
  f1_score: Record<string, number>
  macro_f1: number
  weighted_f1: number
  confusion_matrix: Record<string, Record<string, number>>
}

interface GenerationMetrics {
  bleu: number
  rouge_1: number
  rouge_2: number
  rouge_l: number
  bert_score: number
  diversity: number
  coherence: number
  relevance: number
}

interface OptimizationSuggestion {
  id: string
  type: string
  title: string
  description: string
  old_prompt: string
  new_prompt: string
  expected_impact: number
  confidence: number
  priority: 'high' | 'medium' | 'low'
  status: 'pending' | 'accepted' | 'rejected' | 'applied'
  reasoning: string
}

interface TestResult {
  id: string
  input: Record<string, any>
  expected_output: Record<string, any>
  actual_output: Record<string, any>
  passed: boolean
  score: number
  latency: number
  cost: number
  error_message?: string
  feedback?: string
}

interface CostAnalysis {
  total_cost: number
  avg_cost_per_test: number
  cost_breakdown: {
    input_tokens: number
    output_tokens: number
    total_tokens: number
    cost_per_token: number
  }
}

const EvaluationDetailPage: React.FC = () => {
  const { projectId, evaluationId } = useParams<{ projectId: string; evaluationId: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  
  const [selectedTab, setSelectedTab] = useState('overview')

  // Fetch evaluation details
  const { data: evaluation, isLoading, error } = useQuery({
    queryKey: ['evaluation-detail', evaluationId],
    queryFn: async () => {
      const response = await api.get(`/api/evaluations/${evaluationId}`)
      return response.data.evaluation as Evaluation
    },
    enabled: !!evaluationId,
    refetchInterval: (data) => {
      // Auto-refresh if evaluation is still running
      return data?.status === 'running' ? 5000 : false
    },
  })

  // Fetch evaluation metrics
  const { data: metrics } = useQuery({
    queryKey: ['evaluation-metrics', evaluationId],
    queryFn: async () => {
      const response = await api.get(`/api/evaluations/${evaluationId}/metrics`)
      return response.data
    },
    enabled: !!evaluationId && evaluation?.status === 'completed',
  })

  // Fetch suggestions
  const { data: suggestions = [] } = useQuery({
    queryKey: ['evaluation-suggestions', evaluationId],
    queryFn: async () => {
      const response = await api.get(`/api/evaluations/${evaluationId}/suggestions`)
      return response.data.suggestions || []
    },
    enabled: !!evaluationId && evaluation?.status === 'completed',
  })

  // Run evaluation mutation
  const runEvaluationMutation = useMutation({
    mutationFn: async ({ async: isAsync }: { async: boolean }) => {
      const response = await api.post(`/api/evaluations/${evaluationId}/run`, { async: isAsync })
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['evaluation-detail', evaluationId] })
    },
  })

  // Apply suggestion mutation
  const applySuggestionMutation = useMutation({
    mutationFn: async (suggestionId: string) => {
      const response = await api.post(`/api/evaluations/${evaluationId}/suggestions/${suggestionId}/apply`)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['evaluation-suggestions', evaluationId] })
      queryClient.invalidateQueries({ queryKey: ['evaluation-detail', evaluationId] })
    },
  })

  // Status variants for badges
  const getStatusVariant = (status: string): 'default' | 'secondary' | 'destructive' | 'outline' => {
    switch (status) {
      case 'completed': return 'default'
      case 'running': return 'secondary'
      case 'failed': return 'destructive'
      default: return 'outline'
    }
  }

  // Status icons
  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return <CheckCircle className="h-4 w-4" />
      case 'running': return <Clock className="h-4 w-4" />
      case 'failed': return <XCircle className="h-4 w-4" />
      default: return <AlertCircle className="h-4 w-4" />
    }
  }

  // Format currency
  const formatCurrency = (amount: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 4,
    }).format(amount)
  }

  // Format latency
  const formatLatency = (latency: number) => {
    return `${latency.toFixed(0)}ms`
  }

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (error || !evaluation) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-center">
        <XCircle className="h-12 w-12 text-destructive mb-4" />
        <p className="text-lg font-medium">Failed to load evaluation details</p>
        <p className="text-sm text-muted-foreground">Please try refreshing the page</p>
      </div>
    )
  }

  const evaluationMetrics = metrics || evaluation.metrics
  const evaluationSuggestions = suggestions || evaluation.suggestions || []

  return (
    <div className="page-content">
      {/* Header */}
      <div className="mb-6">
        <Link
          to={`/projects/${projectId}/evaluations`}
          className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground mb-4"
        >
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Evaluations
        </Link>
        
        <div className="flex justify-between items-start">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{evaluation.name}</h1>
            <p className="text-muted-foreground mt-2">{evaluation.description}</p>
          </div>
          
          <div className="flex items-center gap-3">
            <Badge variant={getStatusVariant(evaluation.status)} className="flex items-center gap-1 h-10 px-4">
              {getStatusIcon(evaluation.status)}
              <span className="capitalize">{evaluation.status}</span>
            </Badge>
            
            {evaluation.status === 'pending' && (
              <Button
                onClick={() => runEvaluationMutation.mutate({ async: true })}
                disabled={runEvaluationMutation.isPending}
              >
                <Play className="h-4 w-4 mr-2" />
                Run Evaluation
              </Button>
            )}
            
            {evaluation.status === 'completed' && (
              <Button
                variant="outline"
                onClick={() => runEvaluationMutation.mutate({ async: true })}
                disabled={runEvaluationMutation.isPending}
              >
                <RefreshCw className="h-4 w-4 mr-2" />
                Re-run
              </Button>
            )}
          </div>
        </div>

        {/* Progress for running evaluations */}
        {evaluation.status === 'running' && (
          <Card className="mt-6">
            <CardContent className="pt-6">
              <div className="space-y-2">
                <div className="flex justify-between text-sm">
                  <span>Evaluation Progress</span>
                  <span>{(evaluation.progress || 0).toFixed(1)}%</span>
                </div>
                <div className="bg-secondary rounded-full h-3">
                  <div
                    className="bg-primary h-3 rounded-full transition-all duration-300"
                    style={{ width: `${evaluation.progress || 0}%` }}
                  />
                </div>
                <p className="text-xs text-muted-foreground">
                  Running tests and analyzing results...
                </p>
              </div>
            </CardContent>
          </Card>
        )}
      </div>

      {/* Content Tabs */}
      <Tabs value={selectedTab} onValueChange={setSelectedTab}>
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="results" disabled={evaluation.status !== 'completed'}>Test Results</TabsTrigger>
          <TabsTrigger value="suggestions" disabled={evaluation.status !== 'completed'}>Suggestions</TabsTrigger>
          <TabsTrigger value="cost" disabled={evaluation.status !== 'completed'}>Cost Analysis</TabsTrigger>
        </TabsList>

        {/* Overview Tab */}
        <TabsContent value="overview" className="space-y-6">
          {/* Key Metrics */}
          {evaluationMetrics && (
            <>
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <Card>
                  <CardHeader className="pb-2">
                    <CardDescription>Overall Score</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {isNaN(evaluationMetrics.overall_score) || evaluationMetrics.overall_score == null 
                        ? '—' 
                        : `${(evaluationMetrics.overall_score * 100).toFixed(1)}%`}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {isNaN(evaluationMetrics.overall_score) || evaluationMetrics.overall_score == null
                        ? 'Not available'
                        : evaluationMetrics.overall_score >= 0.8 ? 'Excellent' : 
                          evaluationMetrics.overall_score >= 0.6 ? 'Good' : 'Needs Improvement'}
                    </p>
                  </CardContent>
                </Card>
                
                <Card>
                  <CardHeader className="pb-2">
                    <CardDescription>Pass Rate</CardDescription>
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">
                      {isNaN(evaluationMetrics.pass_rate) || evaluationMetrics.pass_rate == null
                        ? '—'
                        : `${(evaluationMetrics.pass_rate * 100).toFixed(1)}%`}
                    </div>
                    <p className="text-xs text-muted-foreground">
                      {evaluationMetrics.test_cases_total 
                        ? `${evaluationMetrics.test_cases_passed || 0} of ${evaluationMetrics.test_cases_total} tests`
                        : 'No tests run'}
                    </p>
                  </CardContent>
                </Card>
                
                {evaluationMetrics.latency_stats && (
                  <Card>
                    <CardHeader className="pb-2">
                      <CardDescription>Avg Latency</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {formatLatency(evaluationMetrics.latency_stats.avg_latency)}
                      </div>
                      <p className="text-xs text-muted-foreground">
                        P95: {formatLatency(evaluationMetrics.latency_stats.p95_latency)}
                      </p>
                    </CardContent>
                  </Card>
                )}
                
                {evaluation.cost_analysis && (
                  <Card>
                    <CardHeader className="pb-2">
                      <CardDescription>Total Cost</CardDescription>
                    </CardHeader>
                    <CardContent>
                      <div className="text-2xl font-bold">
                        {formatCurrency(evaluation.cost_analysis.total_cost)}
                      </div>
                      <p className="text-xs text-muted-foreground">
                        {formatCurrency(evaluation.cost_analysis.avg_cost_per_test)} per test
                      </p>
                    </CardContent>
                  </Card>
                )}
              </div>

              {/* Detailed Metrics */}
              <MetricsDashboard metrics={evaluationMetrics} />
            </>
          )}

          {/* Prompt */}
          {evaluation.prompt && (
            <Card>
              <CardHeader>
                <CardTitle className="flex items-center gap-2">
                  <FileText className="h-5 w-5" />
                  Evaluated Prompt
                </CardTitle>
              </CardHeader>
              <CardContent>
                <div className="bg-muted rounded-lg p-4">
                  <pre className="text-sm whitespace-pre-wrap font-mono">
                    {evaluation.prompt}
                  </pre>
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>

        {/* Test Results Tab */}
        <TabsContent value="results" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle>Individual Test Results</CardTitle>
              <CardDescription>
                Detailed results for each test case in this evaluation
              </CardDescription>
            </CardHeader>
            <CardContent>
              {evaluation.test_cases && evaluation.test_cases.length > 0 ? (
                <div className="space-y-4">
                  {evaluation.test_cases.map((result, index) => (
                    <Card key={result.id} className={`${result.status === 'passed' ? 'border-green-200' : 'border-red-200'}`}>
                      <CardHeader className="pb-3">
                        <div className="flex justify-between items-center">
                          <CardTitle className="text-base">{result.name || `Test ${index + 1}`}</CardTitle>
                          <div className="flex items-center gap-2">
                            <Badge variant={result.status === 'passed' ? 'default' : 'destructive'}>
                              {result.status === 'passed' ? 'Passed' : 'Failed'}
                            </Badge>
                            <Badge variant="outline">Score: {((result.score || 0) * 100).toFixed(0)}%</Badge>
                          </div>
                        </div>
                        {result.description && (
                          <p className="text-sm text-muted-foreground mt-1">{result.description}</p>
                        )}
                      </CardHeader>
                      <CardContent className="space-y-4">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                          <div>
                            <h5 className="font-medium mb-2">Input</h5>
                            <div className="bg-muted rounded p-3 text-sm">
                              <pre>{JSON.stringify(result.input, null, 2)}</pre>
                            </div>
                          </div>
                          <div>
                            <h5 className="font-medium mb-2">Expected Output</h5>
                            <div className="bg-muted rounded p-3 text-sm">
                              <pre>{JSON.stringify(result.expected_output, null, 2)}</pre>
                            </div>
                          </div>
                        </div>
                        
                        <div>
                          <h5 className="font-medium mb-2">Actual Output</h5>
                          <div className={`rounded p-3 text-sm ${result.passed ? 'bg-green-50' : 'bg-red-50'}`}>
                            <pre>{JSON.stringify(result.actual_output, null, 2)}</pre>
                          </div>
                        </div>
                        
                        {(result.execution_time_ms || result.category) && (
                          <div className="flex justify-between text-sm text-muted-foreground">
                            {result.execution_time_ms && <span>Execution Time: {result.execution_time_ms}ms</span>}
                            {result.category && <span>Category: {result.category}</span>}
                          </div>
                        )}
                        
                        {result.error_message && (
                          <div className="bg-red-50 border border-red-200 rounded p-3">
                            <p className="text-sm text-red-800">{result.error_message}</p>
                          </div>
                        )}
                      </CardContent>
                    </Card>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8 text-muted-foreground">
                  <AlertCircle className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <p>No test results available</p>
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>

        {/* Suggestions Tab */}
        <TabsContent value="suggestions" className="space-y-6">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Zap className="h-5 w-5" />
                Optimization Suggestions
              </CardTitle>
              <CardDescription>
                AI-generated recommendations to improve your prompt performance
              </CardDescription>
            </CardHeader>
            <CardContent>
              <SuggestionCards
                suggestions={evaluationSuggestions}
                onApplySuggestion={(suggestionId) => applySuggestionMutation.mutate(suggestionId)}
              />
            </CardContent>
          </Card>
        </TabsContent>

        {/* Cost Analysis Tab */}
        <TabsContent value="cost" className="space-y-6">
          {evaluation.cost_analysis ? (
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <Card>
                <CardHeader>
                  <CardTitle>Cost Overview</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex justify-between">
                    <span>Total Cost</span>
                    <span className="font-bold">{formatCurrency(evaluation.cost_analysis.total_cost)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Average Cost per Test</span>
                    <span>{formatCurrency(evaluation.cost_analysis.avg_cost_per_test)}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Cost per Token</span>
                    <span>{formatCurrency(evaluation.cost_analysis.cost_breakdown.cost_per_token)}</span>
                  </div>
                </CardContent>
              </Card>
              
              <Card>
                <CardHeader>
                  <CardTitle>Token Usage</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <div className="flex justify-between">
                    <span>Input Tokens</span>
                    <span>{evaluation.cost_analysis.cost_breakdown.input_tokens.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between">
                    <span>Output Tokens</span>
                    <span>{evaluation.cost_analysis.cost_breakdown.output_tokens.toLocaleString()}</span>
                  </div>
                  <div className="flex justify-between font-bold">
                    <span>Total Tokens</span>
                    <span>{evaluation.cost_analysis.cost_breakdown.total_tokens.toLocaleString()}</span>
                  </div>
                </CardContent>
              </Card>
            </div>
          ) : (
            <Card>
              <CardContent className="py-16">
                <div className="text-center text-muted-foreground">
                  <AlertCircle className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <p>Cost analysis not available</p>
                </div>
              </CardContent>
            </Card>
          )}
        </TabsContent>
      </Tabs>
    </div>
  )
}

export default EvaluationDetailPage