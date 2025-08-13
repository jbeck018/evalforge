import React, { useState, useMemo } from 'react'
import { useParams, useNavigate, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'
import { ArrowLeft, Plus, Play, Eye, MoreHorizontal, Clock, CheckCircle, XCircle, AlertCircle, Search, Filter, X } from 'lucide-react'

import { api } from '../utils/api'
import MetricsDashboard from '../components/MetricsDashboard'
import SuggestionCards from '../components/SuggestionCards'

import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Dialog, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

// Types
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
}

interface EvaluationMetrics {
  overall_score: number
  pass_rate: number
  test_cases_passed: number
  test_cases_total: number
  classification_metrics?: ClassificationMetrics
  generation_metrics?: GenerationMetrics
  custom_metrics: Record<string, number>
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

// Form validation schema
const createEvaluationSchema = z.object({
  name: z.string().min(1, 'Name is required').max(100, 'Name must be less than 100 characters'),
  description: z.string().min(1, 'Description is required').max(500, 'Description must be less than 500 characters'),
  prompt: z.string().min(1, 'Prompt is required'),
})

type CreateEvaluationForm = z.infer<typeof createEvaluationSchema>

const EvaluationsPage: React.FC = () => {
  const { id: projectId } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  
  const [selectedEvaluation, setSelectedEvaluation] = useState<Evaluation | null>(null)
  const [isCreateDialogOpen, setIsCreateDialogOpen] = useState(false)
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [showFilters, setShowFilters] = useState(false)
  
  // Form setup
  const form = useForm<CreateEvaluationForm>({
    resolver: zodResolver(createEvaluationSchema),
    defaultValues: {
      name: '',
      description: '',
      prompt: '',
    },
  })

  // Fetch evaluations with React Query
  const { data: evaluations = [], isLoading, error } = useQuery({
    queryKey: ['evaluations', projectId],
    queryFn: async () => {
      const response = await api.get(`/api/projects/${projectId}/evaluations`)
      return response.data.evaluations || []
    },
    enabled: !!projectId,
  })

  // Create evaluation mutation
  const createEvaluationMutation = useMutation({
    mutationFn: async (data: CreateEvaluationForm) => {
      const response = await api.post(`/api/projects/${projectId}/evaluations`, data)
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['evaluations', projectId] })
      setIsCreateDialogOpen(false)
      form.reset()
    },
  })

  // Run evaluation mutation
  const runEvaluationMutation = useMutation({
    mutationFn: async ({ evaluationId, async: isAsync }: { evaluationId: string; async: boolean }) => {
      const response = await api.post(`/api/evaluations/${evaluationId}/run`, { async: isAsync })
      return response.data
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['evaluations', projectId] })
      if (selectedEvaluation) {
        queryClient.invalidateQueries({ queryKey: ['evaluation', selectedEvaluation.id] })
      }
    },
  })

  // Fetch evaluation details
  const { data: evaluationDetails } = useQuery({
    queryKey: ['evaluation', selectedEvaluation?.id],
    queryFn: async () => {
      const response = await api.get(`/api/evaluations/${selectedEvaluation!.id}`)
      return response.data.evaluation
    },
    enabled: !!selectedEvaluation?.id,
  })

  // Apply suggestion mutation
  const applySuggestionMutation = useMutation({
    mutationFn: async ({ evaluationId, suggestionId }: { evaluationId: string; suggestionId: string }) => {
      const response = await api.post(`/api/evaluations/${evaluationId}/suggestions/${suggestionId}/apply`)
      return response.data
    },
    onSuccess: () => {
      if (selectedEvaluation) {
        queryClient.invalidateQueries({ queryKey: ['evaluation', selectedEvaluation.id] })
      }
    },
  })

  // Form submit handler
  const onSubmit = (data: CreateEvaluationForm) => {
    createEvaluationMutation.mutate(data)
  }

  // Run evaluation handler
  const handleRunEvaluation = (evaluationId: string, async: boolean = true) => {
    runEvaluationMutation.mutate({ evaluationId, async })
  }

  // Apply suggestion handler
  const handleApplySuggestion = (evaluationId: string, suggestionId: string) => {
    applySuggestionMutation.mutate({ evaluationId, suggestionId })
  }

  // Select evaluation handler
  const handleSelectEvaluation = (evaluation: Evaluation) => {
    setSelectedEvaluation(evaluation)
  }

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

  // Filter evaluations based on search and status
  const filteredEvaluations = useMemo(() => {
    let filtered = [...evaluations]

    // Apply search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      filtered = filtered.filter(evaluation => 
        evaluation.name?.toLowerCase().includes(query) ||
        evaluation.description?.toLowerCase().includes(query) ||
        evaluation.id?.toLowerCase().includes(query)
      )
    }

    // Apply status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter(evaluation => evaluation.status === statusFilter)
    }

    return filtered
  }, [evaluations, searchQuery, statusFilter])

  // Get unique statuses for filter dropdown
  const uniqueStatuses = useMemo(() => {
    const statuses = new Set(evaluations.map(e => e.status).filter(Boolean))
    return Array.from(statuses)
  }, [evaluations])

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-center">
        <XCircle className="h-12 w-12 text-destructive mb-4" />
        <p className="text-lg font-medium">Failed to load evaluations</p>
        <p className="text-sm text-muted-foreground">Please try refreshing the page</p>
      </div>
    )
  }

  return (
    <div className="page-content">
      {/* Header */}
      <div className="mb-6">
        <Link
          to={`/projects/${projectId}`}
          className="inline-flex items-center text-sm text-muted-foreground hover:text-foreground mb-4"
        >
          <ArrowLeft className="h-4 w-4 mr-1" />
          Back to Project
        </Link>
        
        <div className="flex justify-between items-start">
          <div>
            <h1 className="text-3xl font-bold tracking-tight">Evaluations</h1>
            <p className="text-muted-foreground mt-2">
              Automatic prompt evaluation and optimization for your project
            </p>
          </div>
          
          <Dialog open={isCreateDialogOpen} onOpenChange={setIsCreateDialogOpen}>
            <DialogTrigger asChild>
              <Button>
                <Plus className="h-4 w-4 mr-2" />
                New Evaluation
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-2xl">
              <DialogHeader>
                <DialogTitle>Create New Evaluation</DialogTitle>
                <DialogDescription>
                  Set up a new evaluation to test and optimize your prompt performance.
                </DialogDescription>
              </DialogHeader>
              
              <Form {...form}>
                <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
                  <FormField
                    control={form.control}
                    name="name"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Name</FormLabel>
                        <FormControl>
                          <Input placeholder="Enter evaluation name" {...field} />
                        </FormControl>
                        <FormDescription>
                          A descriptive name for this evaluation.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  
                  <FormField
                    control={form.control}
                    name="description"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Description</FormLabel>
                        <FormControl>
                          <Input placeholder="Brief description of what you're testing" {...field} />
                        </FormControl>
                        <FormDescription>
                          Explain the purpose of this evaluation.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  
                  <FormField
                    control={form.control}
                    name="prompt"
                    render={({ field }) => (
                      <FormItem>
                        <FormLabel>Prompt to Evaluate</FormLabel>
                        <FormControl>
                          <Textarea 
                            placeholder="Enter the prompt you want to evaluate..."
                            className="min-h-[120px]"
                            {...field} 
                          />
                        </FormControl>
                        <FormDescription>
                          The prompt that will be evaluated against your test cases.
                        </FormDescription>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  
                  <div className="flex justify-end space-x-3">
                    <Button 
                      type="button" 
                      variant="outline" 
                      onClick={() => setIsCreateDialogOpen(false)}
                    >
                      Cancel
                    </Button>
                    <Button 
                      type="submit" 
                      disabled={createEvaluationMutation.isPending}
                    >
                      {createEvaluationMutation.isPending ? 'Creating...' : 'Create Evaluation'}
                    </Button>
                  </div>
                </form>
              </Form>
            </DialogContent>
          </Dialog>
        </div>
      </div>

      {/* Search and Filter Bar */}
      <Card className="mb-6">
        <CardContent className="p-4">
          <div className="flex flex-col sm:flex-row gap-4">
            {/* Search Input */}
            <div className="flex-1">
              <div className="relative">
                <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                  <Search className="h-5 w-5 text-muted-foreground" />
                </div>
                <Input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10 pr-10"
                  placeholder="Search evaluations by name or description..."
                />
                {searchQuery && (
                  <button
                    onClick={() => setSearchQuery('')}
                    className="absolute inset-y-0 right-0 pr-3 flex items-center"
                  >
                    <X className="h-5 w-5 text-muted-foreground hover:text-foreground" />
                  </button>
                )}
              </div>
            </div>

            {/* Filter Toggle */}
            <Button
              variant="outline"
              onClick={() => setShowFilters(!showFilters)}
            >
              <Filter className="h-4 w-4 mr-2" />
              Filters
              {statusFilter !== 'all' && (
                <Badge variant="secondary" className="ml-2">
                  1
                </Badge>
              )}
            </Button>
          </div>

          {/* Filter Options */}
          {showFilters && (
            <div className="mt-4 pt-4 border-t">
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                {/* Status Filter */}
                <div>
                  <label htmlFor="status-filter" className="block text-sm font-medium mb-2">
                    Status
                  </label>
                  <select
                    id="status-filter"
                    value={statusFilter}
                    onChange={(e) => setStatusFilter(e.target.value)}
                    className="w-full px-3 py-2 border border-input bg-background rounded-md focus:outline-none focus:ring-2 focus:ring-ring"
                  >
                    <option value="all">All Statuses</option>
                    {uniqueStatuses.map(status => (
                      <option key={status} value={status}>
                        {status.charAt(0).toUpperCase() + status.slice(1)}
                      </option>
                    ))}
                  </select>
                </div>

                {/* Clear Filters */}
                {statusFilter !== 'all' && (
                  <div className="flex items-end">
                    <Button
                      variant="outline"
                      onClick={() => setStatusFilter('all')}
                      className="w-full sm:w-auto"
                    >
                      Clear Filters
                    </Button>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Results Count */}
          <div className="mt-4 text-sm text-muted-foreground">
            Showing {filteredEvaluations.length} of {evaluations.length} evaluations
            {searchQuery && ` matching "${searchQuery}"`}
          </div>
        </CardContent>
      </Card>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Evaluations List */}
        <div className="lg:col-span-1">
          <Card>
            <CardHeader>
              <CardTitle>Evaluations</CardTitle>
              <CardDescription>
                {evaluations.length} evaluation{evaluations.length !== 1 ? 's' : ''} in this project
              </CardDescription>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {filteredEvaluations.map((evaluation) => (
                  <Card
                    key={evaluation.id}
                    className={`cursor-pointer transition-all hover:shadow-md ${
                      selectedEvaluation?.id === evaluation.id 
                        ? 'ring-2 ring-primary shadow-md' 
                        : 'hover:bg-accent/50'
                    }`}
                    onClick={() => handleSelectEvaluation(evaluation)}
                  >
                    <CardContent className="p-4">
                      <div className="flex justify-between items-start mb-2">
                        <h4 className="font-medium truncate flex-1 mr-2">{evaluation.name}</h4>
                        <Badge variant={getStatusVariant(evaluation.status)} className="flex items-center gap-1">
                          {getStatusIcon(evaluation.status)}
                          {evaluation.status}
                        </Badge>
                      </div>
                      
                      <p className="text-sm text-muted-foreground mb-3 line-clamp-2">
                        {evaluation.description}
                      </p>
                      
                      <div className="flex justify-between items-center text-xs text-muted-foreground">
                        <span>
                          {new Date(evaluation.created_at).toLocaleDateString()}
                        </span>
                        
                        <div className="flex items-center gap-2">
                          {evaluation.status === 'pending' && (
                            <Button
                              size="sm"
                              variant="ghost"
                              onClick={(e) => {
                                e.stopPropagation()
                                handleRunEvaluation(evaluation.id)
                              }}
                              disabled={runEvaluationMutation.isPending}
                            >
                              <Play className="h-3 w-3 mr-1" />
                              Run
                            </Button>
                          )}
                          
                          {evaluation.status === 'completed' && (
                            <Link to={`/projects/${projectId}/evaluations/${evaluation.id}`}>
                              <Button size="sm" variant="ghost">
                                <Eye className="h-3 w-3 mr-1" />
                                View
                              </Button>
                            </Link>
                          )}
                        </div>
                      </div>
                      
                      {evaluation.status === 'running' && (
                        <div className="mt-3">
                          <div className="flex justify-between text-xs text-muted-foreground mb-1">
                            <span>Progress</span>
                            <span>{evaluation.progress.toFixed(1)}%</span>
                          </div>
                          <div className="bg-secondary rounded-full h-2">
                            <div
                              className="bg-primary h-2 rounded-full transition-all duration-300"
                              style={{ width: `${evaluation.progress}%` }}
                            />
                          </div>
                        </div>
                      )}
                    </CardContent>
                  </Card>
                ))}
                
                {evaluations.length === 0 && (
                  <div className="text-center py-8 text-muted-foreground">
                    <AlertCircle className="h-12 w-12 mx-auto mb-4 opacity-50" />
                    <p className="font-medium">No evaluations yet</p>
                    <p className="text-sm">Create your first evaluation to get started</p>
                  </div>
                )}

                {evaluations.length > 0 && filteredEvaluations.length === 0 && (
                  <div className="text-center py-8 text-muted-foreground">
                    <Search className="h-12 w-12 mx-auto mb-4 opacity-50" />
                    <p className="font-medium">No evaluations match your criteria</p>
                    <p className="text-sm mt-2">
                      <button
                        onClick={() => {
                          setSearchQuery('')
                          setStatusFilter('all')
                        }}
                        className="text-primary hover:underline"
                      >
                        Clear filters
                      </button>
                    </p>
                  </div>
                )}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Evaluation Details */}
        <div className="lg:col-span-2">
          {selectedEvaluation ? (
            <div className="space-y-6">
              {/* Evaluation Header */}
              <Card>
                <CardHeader>
                  <div className="flex justify-between items-start">
                    <div>
                      <CardTitle>{selectedEvaluation.name}</CardTitle>
                      <CardDescription className="mt-1">
                        {selectedEvaluation.description}
                      </CardDescription>
                    </div>
                    <div className="flex items-center gap-2">
                      <Badge variant={getStatusVariant(selectedEvaluation.status)} className="flex items-center gap-1">
                        {getStatusIcon(selectedEvaluation.status)}
                        {selectedEvaluation.status}
                      </Badge>
                      
                      {selectedEvaluation.status === 'completed' && (
                        <Link to={`/projects/${projectId}/evaluations/${selectedEvaluation.id}`}>
                          <Button size="sm">
                            <Eye className="h-4 w-4 mr-2" />
                            View Details
                          </Button>
                        </Link>
                      )}
                    </div>
                  </div>
                </CardHeader>
                
                {selectedEvaluation.status === 'running' && (
                  <CardContent>
                    <div className="space-y-2">
                      <div className="flex justify-between text-sm">
                        <span>Progress</span>
                        <span>{selectedEvaluation.progress.toFixed(1)}%</span>
                      </div>
                      <div className="bg-secondary rounded-full h-3">
                        <div
                          className="bg-primary h-3 rounded-full transition-all duration-300"
                          style={{ width: `${selectedEvaluation.progress}%` }}
                        />
                      </div>
                    </div>
                  </CardContent>
                )}
              </Card>

              {/* Metrics Overview */}
              {(evaluationDetails?.metrics || selectedEvaluation.metrics) && (
                <MetricsDashboard 
                  metrics={evaluationDetails?.metrics || selectedEvaluation.metrics!} 
                />
              )}

              {/* Optimization Suggestions */}
              {selectedEvaluation.status === 'completed' && (
                <Card>
                  <CardHeader>
                    <CardTitle>Optimization Suggestions</CardTitle>
                    <CardDescription>
                      AI-generated recommendations to improve your prompt performance
                    </CardDescription>
                  </CardHeader>
                  <CardContent>
                    <SuggestionCards
                      suggestions={(evaluationDetails?.suggestions || selectedEvaluation.suggestions) || []}
                      onApplySuggestion={(suggestionId) => handleApplySuggestion(selectedEvaluation.id, suggestionId)}
                    />
                  </CardContent>
                </Card>
              )}
            </div>
          ) : (
            <Card>
              <CardContent className="py-16">
                <div className="text-center text-muted-foreground">
                  <Eye className="h-12 w-12 mx-auto mb-4 opacity-50" />
                  <h3 className="text-lg font-medium mb-2">Select an evaluation</h3>
                  <p className="text-sm">
                    Choose an evaluation from the list to view detailed metrics and optimization suggestions.
                  </p>
                </div>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  )
}

export default EvaluationsPage