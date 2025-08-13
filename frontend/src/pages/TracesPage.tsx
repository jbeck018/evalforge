import { useState, useEffect, useMemo } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Clock, CheckCircle, XCircle, AlertCircle, Search, Filter, X, ChevronLeft, ChevronRight } from 'lucide-react'
import { api } from '../utils/api'

interface Trace {
  id: string
  trace_id: string
  operation_type: string
  status: string
  start_time: string
  end_time: string
  duration_ms: number
  cost: number
  model: string
  provider: string
  span_id: string
  input?: any
  output?: any
  metadata?: Record<string, any>
  name?: string
}

export default function TracesPage() {
  const { id: projectId } = useParams<{ id: string }>()
  const [traces, setTraces] = useState<Trace[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedTrace, setSelectedTrace] = useState<Trace | null>(null)
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [operationFilter, setOperationFilter] = useState<string>('all')
  const [showFilters, setShowFilters] = useState(false)
  const [currentPage, setCurrentPage] = useState(1)
  const itemsPerPage = 20

  useEffect(() => {
    if (projectId) {
      fetchTraces()
    }
  }, [projectId])

  const fetchTraces = async () => {
    try {
      // Fetch events which have more detail than traces
      const response = await api.get(`/api/projects/${projectId}/events?limit=100`)
      setTraces(response.data.events || [])
    } catch (error) {
      console.error('Failed to fetch traces:', error)
    } finally {
      setLoading(false)
    }
  }

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="h-5 w-5 text-green-500" />
      case 'failed':
        return <XCircle className="h-5 w-5 text-red-500" />
      default:
        return <AlertCircle className="h-5 w-5 text-yellow-500" />
    }
  }

  const formatDuration = (ms: number) => {
    if (ms < 1000) return `${ms}ms`
    return `${(ms / 1000).toFixed(2)}s`
  }

  // Filter and search logic
  const filteredTraces = useMemo(() => {
    let filtered = [...traces]

    // Apply search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      filtered = filtered.filter(trace => 
        trace.operation_type?.toLowerCase().includes(query) ||
        trace.trace_id?.toLowerCase().includes(query) ||
        trace.model?.toLowerCase().includes(query) ||
        trace.provider?.toLowerCase().includes(query) ||
        trace.metadata?.model?.toLowerCase().includes(query) ||
        trace.metadata?.provider?.toLowerCase().includes(query) ||
        JSON.stringify(trace.input)?.toLowerCase().includes(query) ||
        JSON.stringify(trace.output)?.toLowerCase().includes(query)
      )
    }

    // Apply status filter
    if (statusFilter !== 'all') {
      filtered = filtered.filter(trace => trace.status === statusFilter)
    }

    // Apply operation type filter
    if (operationFilter !== 'all') {
      filtered = filtered.filter(trace => trace.operation_type === operationFilter)
    }

    return filtered
  }, [traces, searchQuery, statusFilter, operationFilter])

  // Get unique values for filters
  const uniqueStatuses = useMemo(() => {
    const statuses = new Set(traces.map(t => t.status).filter(Boolean))
    return Array.from(statuses)
  }, [traces])

  const uniqueOperations = useMemo(() => {
    const operations = new Set(traces.map(t => t.operation_type).filter(Boolean))
    return Array.from(operations)
  }, [traces])

  // Pagination logic
  const paginatedTraces = useMemo(() => {
    const startIndex = (currentPage - 1) * itemsPerPage
    const endIndex = startIndex + itemsPerPage
    return filteredTraces.slice(startIndex, endIndex)
  }, [filteredTraces, currentPage, itemsPerPage])

  const totalPages = Math.ceil(filteredTraces.length / itemsPerPage)

  // Reset to first page when filters change
  useEffect(() => {
    setCurrentPage(1)
  }, [searchQuery, statusFilter, operationFilter])

  const handlePageChange = (page: number) => {
    if (page >= 1 && page <= totalPages) {
      setCurrentPage(page)
      // Scroll to top of table
      document.querySelector('.traces-table')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    }
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
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
          <h1 className="text-2xl font-semibold text-gray-900">Traces</h1>
          <p className="mt-2 text-sm text-gray-700">
            View execution traces and debug information for your project.
          </p>
        </div>
      </div>

      {/* Search and Filter Bar */}
      <div className="bg-white shadow rounded-lg mb-6 p-4">
        <div className="flex flex-col sm:flex-row gap-4">
          {/* Search Input */}
          <div className="flex-1">
            <div className="relative">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <Search className="h-5 w-5 text-gray-400" />
              </div>
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-md leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm"
                placeholder="Search traces by operation, ID, model, or content..."
              />
              {searchQuery && (
                <button
                  onClick={() => setSearchQuery('')}
                  className="absolute inset-y-0 right-0 pr-3 flex items-center"
                >
                  <X className="h-5 w-5 text-gray-400 hover:text-gray-600" />
                </button>
              )}
            </div>
          </div>

          {/* Filter Toggle Button */}
          <button
            onClick={() => setShowFilters(!showFilters)}
            className="inline-flex items-center px-4 py-2 border border-gray-300 rounded-md shadow-sm text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
          >
            <Filter className="h-4 w-4 mr-2" />
            Filters
            {(statusFilter !== 'all' || operationFilter !== 'all') && (
              <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-indigo-100 text-indigo-800">
                {[statusFilter !== 'all', operationFilter !== 'all'].filter(Boolean).length}
              </span>
            )}
          </button>
        </div>

        {/* Filter Options */}
        {showFilters && (
          <div className="mt-4 pt-4 border-t border-gray-200">
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {/* Status Filter */}
              <div>
                <label htmlFor="status-filter" className="block text-sm font-medium text-gray-700 mb-1">
                  Status
                </label>
                <select
                  id="status-filter"
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                  className="block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                >
                  <option value="all">All Statuses</option>
                  {uniqueStatuses.map(status => (
                    <option key={status} value={status}>
                      {status}
                    </option>
                  ))}
                </select>
              </div>

              {/* Operation Type Filter */}
              <div>
                <label htmlFor="operation-filter" className="block text-sm font-medium text-gray-700 mb-1">
                  Operation Type
                </label>
                <select
                  id="operation-filter"
                  value={operationFilter}
                  onChange={(e) => setOperationFilter(e.target.value)}
                  className="block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
                >
                  <option value="all">All Operations</option>
                  {uniqueOperations.map(operation => (
                    <option key={operation} value={operation}>
                      {operation}
                    </option>
                  ))}
                </select>
              </div>

              {/* Clear Filters Button */}
              {(statusFilter !== 'all' || operationFilter !== 'all') && (
                <div className="flex items-end">
                  <button
                    onClick={() => {
                      setStatusFilter('all')
                      setOperationFilter('all')
                    }}
                    className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm leading-4 font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                  >
                    Clear Filters
                  </button>
                </div>
              )}
            </div>
          </div>
        )}

        {/* Results Count */}
        <div className="mt-4 text-sm text-gray-600">
          Showing {Math.min((currentPage - 1) * itemsPerPage + 1, filteredTraces.length)}-{Math.min(currentPage * itemsPerPage, filteredTraces.length)} of {filteredTraces.length} traces
          {searchQuery && ` matching "${searchQuery}"`}
        </div>
      </div>

      <div className="bg-white shadow rounded-lg traces-table">
        <table className="min-w-full divide-y divide-gray-200">
          <thead className="bg-gray-50">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Operation
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Model/Provider
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Cost
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                Time
              </th>
            </tr>
          </thead>
          <tbody className="bg-white divide-y divide-gray-200">
            {paginatedTraces.map((trace) => (
              <tr
                key={trace.id}
                onClick={() => setSelectedTrace(trace)}
                className="hover:bg-gray-50 cursor-pointer"
              >
                <td className="px-6 py-4 whitespace-nowrap">
                  <div>
                    <div className="text-sm font-medium text-gray-900">{trace.operation_type}</div>
                    <div className="text-xs text-gray-500 font-mono">{trace.trace_id.slice(0, 8)}...</div>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  {trace.metadata?.model || trace.model || trace.metadata?.provider || trace.provider || 'N/A'}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center">
                    {getStatusIcon(trace.status)}
                    <span className="ml-2 text-sm text-gray-900">{trace.status}</span>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  ${trace.cost.toFixed(6)}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                  <div className="flex items-center">
                    <Clock className="h-4 w-4 mr-1" />
                    {new Date(trace.start_time).toLocaleString()}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {traces.length === 0 && (
          <div className="text-center py-12">
            <p className="text-gray-500">No traces recorded yet.</p>
          </div>
        )}

        {traces.length > 0 && filteredTraces.length === 0 && (
          <div className="text-center py-12">
            <p className="text-gray-500">No traces match your search criteria.</p>
            <button
              onClick={() => {
                setSearchQuery('')
                setStatusFilter('all')
                setOperationFilter('all')
              }}
              className="mt-2 text-indigo-600 hover:text-indigo-500 text-sm"
            >
              Clear filters
            </button>
          </div>
        )}

        {/* Pagination Controls */}
        {totalPages > 1 && (
          <div className="bg-white px-4 py-3 border-t border-gray-200 sm:px-6">
            <div className="flex items-center justify-between">
              <div className="flex-1 flex justify-between sm:hidden">
                <button
                  onClick={() => handlePageChange(currentPage - 1)}
                  disabled={currentPage === 1}
                  className="relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Previous
                </button>
                <button
                  onClick={() => handlePageChange(currentPage + 1)}
                  disabled={currentPage === totalPages}
                  className="ml-3 relative inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  Next
                </button>
              </div>
              <div className="hidden sm:flex-1 sm:flex sm:items-center sm:justify-between">
                <div>
                  <p className="text-sm text-gray-700">
                    Page <span className="font-medium">{currentPage}</span> of{' '}
                    <span className="font-medium">{totalPages}</span>
                  </p>
                </div>
                <div>
                  <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
                    <button
                      onClick={() => handlePageChange(currentPage - 1)}
                      disabled={currentPage === 1}
                      className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      <span className="sr-only">Previous</span>
                      <ChevronLeft className="h-5 w-5" />
                    </button>
                    
                    {/* Page numbers */}
                    {[...Array(Math.min(5, totalPages))].map((_, i) => {
                      let pageNum: number
                      if (totalPages <= 5) {
                        pageNum = i + 1
                      } else if (currentPage <= 3) {
                        pageNum = i + 1
                      } else if (currentPage >= totalPages - 2) {
                        pageNum = totalPages - 4 + i
                      } else {
                        pageNum = currentPage - 2 + i
                      }
                      
                      return (
                        <button
                          key={pageNum}
                          onClick={() => handlePageChange(pageNum)}
                          className={`relative inline-flex items-center px-4 py-2 border text-sm font-medium ${
                            currentPage === pageNum
                              ? 'z-10 bg-indigo-50 border-indigo-500 text-indigo-600'
                              : 'bg-white border-gray-300 text-gray-500 hover:bg-gray-50'
                          }`}
                        >
                          {pageNum}
                        </button>
                      )
                    })}
                    
                    <button
                      onClick={() => handlePageChange(currentPage + 1)}
                      disabled={currentPage === totalPages}
                      className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                      <span className="sr-only">Next</span>
                      <ChevronRight className="h-5 w-5" />
                    </button>
                  </nav>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {selectedTrace && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center p-4 z-50">
          <div className="bg-white rounded-lg p-6 max-w-4xl w-full max-h-[80vh] overflow-y-auto">
            <div className="flex justify-between items-start mb-4">
              <h3 className="text-lg font-medium text-gray-900">Trace Details</h3>
              <button
                onClick={() => setSelectedTrace(null)}
                className="text-gray-400 hover:text-gray-500"
              >
                <span className="sr-only">Close</span>
                <svg className="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>

            <div className="space-y-4">
              <div>
                <h4 className="text-sm font-medium text-gray-700">General Information</h4>
                <dl className="mt-2 grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <dt className="text-gray-500">Operation</dt>
                    <dd className="font-medium">{selectedTrace.operation_type}</dd>
                  </div>
                  <div>
                    <dt className="text-gray-500">Model</dt>
                    <dd className="font-medium">{selectedTrace.metadata?.model || selectedTrace.model || 'N/A'}</dd>
                  </div>
                  <div>
                    <dt className="text-gray-500">Status</dt>
                    <dd className="font-medium">{selectedTrace.status}</dd>
                  </div>
                  <div>
                    <dt className="text-gray-500">Cost</dt>
                    <dd className="font-medium">${selectedTrace.cost.toFixed(6)}</dd>
                  </div>
                  <div>
                    <dt className="text-gray-500">Trace ID</dt>
                    <dd className="font-medium font-mono text-xs">{selectedTrace.trace_id}</dd>
                  </div>
                  <div>
                    <dt className="text-gray-500">Provider</dt>
                    <dd className="font-medium">{selectedTrace.metadata?.provider || selectedTrace.provider || 'N/A'}</dd>
                  </div>
                  <div>
                    <dt className="text-gray-500">Latency</dt>
                    <dd className="font-medium">{selectedTrace.metadata?.latency_ms || selectedTrace.duration_ms || 0}ms</dd>
                  </div>
                </dl>
              </div>

              {selectedTrace.input && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700">Input</h4>
                  <pre className="mt-2 p-3 bg-gray-50 rounded text-xs overflow-x-auto">
                    {JSON.stringify(selectedTrace.input, null, 2)}
                  </pre>
                </div>
              )}

              {selectedTrace.output && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700">Output</h4>
                  <pre className="mt-2 p-3 bg-gray-50 rounded text-xs overflow-x-auto">
                    {JSON.stringify(selectedTrace.output, null, 2)}
                  </pre>
                </div>
              )}

              {selectedTrace.error && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700">Error</h4>
                  <pre className="mt-2 p-3 bg-red-50 text-red-700 rounded text-xs overflow-x-auto">
                    {selectedTrace.error}
                  </pre>
                </div>
              )}

              {selectedTrace.metadata && Object.keys(selectedTrace.metadata).length > 0 && (
                <div>
                  <h4 className="text-sm font-medium text-gray-700">Metadata</h4>
                  <pre className="mt-2 p-3 bg-gray-50 rounded text-xs overflow-x-auto">
                    {JSON.stringify(selectedTrace.metadata, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}