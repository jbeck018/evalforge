import React, { useState } from 'react';

interface OptimizationSuggestion {
  id: string;
  type: string;
  title: string;
  description: string;
  old_prompt: string;
  new_prompt: string;
  expected_impact: number;
  confidence: number;
  priority: 'high' | 'medium' | 'low';
  status: 'pending' | 'accepted' | 'rejected' | 'applied';
  reasoning: string;
  examples?: Array<{
    input: Record<string, any>;
    output: Record<string, any>;
    explanation: string;
  }>;
}

interface SuggestionCardsProps {
  suggestions: OptimizationSuggestion[];
  onApplySuggestion: (suggestionId: string) => void;
  onRejectSuggestion?: (suggestionId: string) => void;
  className?: string;
}

const SuggestionCards: React.FC<SuggestionCardsProps> = ({
  suggestions,
  onApplySuggestion,
  onRejectSuggestion,
  className = ''
}) => {
  const [expandedSuggestions, setExpandedSuggestions] = useState<Set<string>>(new Set());

  const toggleExpanded = (suggestionId: string) => {
    const newExpanded = new Set(expandedSuggestions);
    if (newExpanded.has(suggestionId)) {
      newExpanded.delete(suggestionId);
    } else {
      newExpanded.add(suggestionId);
    }
    setExpandedSuggestions(newExpanded);
  };

  const getPriorityColor = (priority: string) => {
    switch (priority) {
      case 'high': return 'text-red-700 bg-red-100 border-red-200';
      case 'medium': return 'text-yellow-700 bg-yellow-100 border-yellow-200';
      case 'low': return 'text-green-700 bg-green-100 border-green-200';
      default: return 'text-gray-700 bg-gray-100 border-gray-200';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'applied': return 'text-green-700 bg-green-100 border-green-200';
      case 'accepted': return 'text-blue-700 bg-blue-100 border-blue-200';
      case 'rejected': return 'text-red-700 bg-red-100 border-red-200';
      default: return 'text-gray-700 bg-gray-100 border-gray-200';
    }
  };

  const getTypeIcon = (type: string) => {
    switch (type.toLowerCase()) {
      case 'clarity':
        return (
          <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
          </svg>
        );
      case 'examples':
        return (
          <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <path d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
        );
      case 'format':
        return (
          <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M4 4a2 2 0 012-2h8a2 2 0 012 2v12a2 2 0 01-2 2H6a2 2 0 01-2-2V4zm2 0v12h8V4H6z" clipRule="evenodd" />
          </svg>
        );
      case 'accuracy':
        return (
          <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
          </svg>
        );
      default:
        return (
          <svg className="w-5 h-5" fill="currentColor" viewBox="0 0 20 20">
            <path fillRule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clipRule="evenodd" />
          </svg>
        );
    }
  };

  const getImpactLevel = (impact: number) => {
    if (impact >= 0.2) return { label: 'High', color: 'text-green-600' };
    if (impact >= 0.1) return { label: 'Medium', color: 'text-yellow-600' };
    return { label: 'Low', color: 'text-gray-600' };
  };

  if (suggestions.length === 0) {
    return (
      <div className={`bg-white rounded-lg shadow border border-gray-200 ${className}`}>
        <div className="px-6 py-12 text-center">
          <svg className="mx-auto h-12 w-12 text-gray-400 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z" />
          </svg>
          <h3 className="text-lg font-medium text-gray-900 mb-2">No Optimization Suggestions</h3>
          <p className="text-gray-600">
            Your prompt is performing well, or evaluation hasn't completed yet.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {suggestions.map((suggestion) => {
        const isExpanded = expandedSuggestions.has(suggestion.id);
        const impactLevel = getImpactLevel(suggestion.expected_impact);
        
        return (
          <div key={suggestion.id} className="bg-white rounded-lg shadow border border-gray-200">
            <div className="px-6 py-4">
              {/* Header */}
              <div className="flex items-start justify-between">
                <div className="flex items-start space-x-3 flex-1">
                  <div className="flex-shrink-0 mt-1">
                    <div className="text-blue-600">
                      {getTypeIcon(suggestion.type)}
                    </div>
                  </div>
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center space-x-2 mb-1">
                      <h4 className="text-lg font-medium text-gray-900 truncate">
                        {suggestion.title}
                      </h4>
                      <span className={`px-2 py-1 text-xs font-medium rounded-full border ${getPriorityColor(suggestion.priority)}`}>
                        {suggestion.priority}
                      </span>
                      <span className={`px-2 py-1 text-xs font-medium rounded-full border ${getStatusColor(suggestion.status)}`}>
                        {suggestion.status}
                      </span>
                    </div>
                    <p className="text-gray-600 text-sm mb-3">
                      {suggestion.description}
                    </p>
                    
                    {/* Metrics */}
                    <div className="flex items-center space-x-6 text-sm">
                      <div className="flex items-center space-x-1">
                        <span className="text-gray-500">Expected Impact:</span>
                        <span className={`font-medium ${impactLevel.color}`}>
                          {(suggestion.expected_impact * 100).toFixed(1)}% ({impactLevel.label})
                        </span>
                      </div>
                      <div className="flex items-center space-x-1">
                        <span className="text-gray-500">Confidence:</span>
                        <span className="font-medium text-gray-900">
                          {(suggestion.confidence * 100).toFixed(1)}%
                        </span>
                      </div>
                      <div className="flex items-center space-x-1">
                        <span className="text-gray-500">Type:</span>
                        <span className="font-medium text-gray-900 capitalize">
                          {suggestion.type}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
                
                {/* Actions */}
                <div className="flex items-center space-x-2 ml-4">
                  {suggestion.status === 'pending' && (
                    <>
                      <button
                        onClick={() => onApplySuggestion(suggestion.id)}
                        className="px-3 py-1 bg-blue-600 text-white text-sm rounded hover:bg-blue-700 transition-colors"
                      >
                        Apply
                      </button>
                      {onRejectSuggestion && (
                        <button
                          onClick={() => onRejectSuggestion(suggestion.id)}
                          className="px-3 py-1 bg-gray-200 text-gray-700 text-sm rounded hover:bg-gray-300 transition-colors"
                        >
                          Reject
                        </button>
                      )}
                    </>
                  )}
                  <button
                    onClick={() => toggleExpanded(suggestion.id)}
                    className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
                  >
                    <svg
                      className={`w-5 h-5 transform transition-transform ${isExpanded ? 'rotate-180' : ''}`}
                      fill="currentColor"
                      viewBox="0 0 20 20"
                    >
                      <path fillRule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clipRule="evenodd" />
                    </svg>
                  </button>
                </div>
              </div>

              {/* Expanded Content */}
              {isExpanded && (
                <div className="mt-4 pt-4 border-t border-gray-200">
                  {/* Reasoning */}
                  <div className="mb-4">
                    <h5 className="text-sm font-medium text-gray-900 mb-2">Reasoning</h5>
                    <p className="text-sm text-gray-700 leading-relaxed">
                      {suggestion.reasoning}
                    </p>
                  </div>

                  {/* Prompt Comparison */}
                  <div className="mb-4">
                    <h5 className="text-sm font-medium text-gray-900 mb-3">Prompt Changes</h5>
                    <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
                      <div>
                        <div className="flex items-center mb-2">
                          <span className="text-xs font-medium text-gray-500 uppercase tracking-wider">Current Prompt</span>
                        </div>
                        <div className="bg-red-50 border border-red-200 rounded-lg p-3 text-sm font-mono text-gray-800 max-h-40 overflow-y-auto">
                          {suggestion.old_prompt}
                        </div>
                      </div>
                      <div>
                        <div className="flex items-center mb-2">
                          <span className="text-xs font-medium text-gray-500 uppercase tracking-wider">Suggested Prompt</span>
                        </div>
                        <div className="bg-green-50 border border-green-200 rounded-lg p-3 text-sm font-mono text-gray-800 max-h-40 overflow-y-auto">
                          {suggestion.new_prompt}
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Examples */}
                  {suggestion.examples && suggestion.examples.length > 0 && (
                    <div>
                      <h5 className="text-sm font-medium text-gray-900 mb-3">Examples</h5>
                      <div className="space-y-3">
                        {suggestion.examples.map((example, index) => (
                          <div key={index} className="bg-gray-50 border border-gray-200 rounded-lg p-3">
                            <div className="grid grid-cols-1 md:grid-cols-2 gap-3 mb-2">
                              <div>
                                <span className="text-xs font-medium text-gray-500 uppercase tracking-wider">Input</span>
                                <div className="mt-1 text-sm text-gray-800">
                                  {JSON.stringify(example.input, null, 2)}
                                </div>
                              </div>
                              <div>
                                <span className="text-xs font-medium text-gray-500 uppercase tracking-wider">Expected Output</span>
                                <div className="mt-1 text-sm text-gray-800">
                                  {JSON.stringify(example.output, null, 2)}
                                </div>
                              </div>
                            </div>
                            {example.explanation && (
                              <div className="pt-2 border-t border-gray-200">
                                <span className="text-xs font-medium text-gray-500 uppercase tracking-wider">Explanation</span>
                                <p className="mt-1 text-sm text-gray-700">{example.explanation}</p>
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );
};

export default SuggestionCards;