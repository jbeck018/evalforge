import React from 'react';
import ConfusionMatrix from './ConfusionMatrix';

interface MetricsDashboardProps {
  metrics: EvaluationMetrics;
  className?: string;
}

interface EvaluationMetrics {
  overall_score: number;
  pass_rate: number;
  test_cases_passed: number;
  test_cases_total: number;
  classification_metrics?: ClassificationMetrics;
  generation_metrics?: GenerationMetrics;
  custom_metrics: Record<string, number>;
}

interface ClassificationMetrics {
  accuracy: number;
  precision: Record<string, number>;
  recall: Record<string, number>;
  f1_score: Record<string, number>;
  macro_f1: number;
  weighted_f1: number;
  confusion_matrix: Record<string, Record<string, number>>;
}

interface GenerationMetrics {
  bleu: number;
  rouge_1: number;
  rouge_2: number;
  rouge_l: number;
  bert_score: number;
  diversity: number;
  coherence: number;
  relevance: number;
}

const MetricsDashboard: React.FC<MetricsDashboardProps> = ({ metrics, className = '' }) => {
  const getScoreColor = (score: number) => {
    if (score >= 0.8) return 'text-green-600';
    if (score >= 0.6) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getScoreBgColor = (score: number) => {
    if (score >= 0.8) return 'bg-green-100';
    if (score >= 0.6) return 'bg-yellow-100';
    return 'bg-red-100';
  };

  const formatPercentage = (value: number) => {
    if (value == null || isNaN(value)) return '—';
    return `${(value * 100).toFixed(1)}%`;
  };

  const formatScore = (value: number) => {
    if (value == null || isNaN(value)) return '—';
    return value.toFixed(3);
  };

  // Extract classes from confusion matrix or precision/recall data
  const classes = metrics.classification_metrics 
    ? Object.keys(metrics.classification_metrics.confusion_matrix || {})
    : [];

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Overall Performance */}
      <div className="bg-white rounded-lg shadow border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Overall Performance</h3>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            <div className="text-center">
              <div className={`text-3xl font-bold ${getScoreColor(metrics.overall_score)}`}>
                {formatPercentage(metrics.overall_score)}
              </div>
              <div className="text-sm text-gray-600 mt-1">Overall Score</div>
              <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                <div 
                  className={`h-2 rounded-full transition-all duration-300 ${
                    metrics.overall_score >= 0.8 ? 'bg-green-500' :
                    metrics.overall_score >= 0.6 ? 'bg-yellow-500' : 'bg-red-500'
                  }`}
                  style={{ width: `${(metrics.overall_score || 0) * 100}%` }}
                ></div>
              </div>
            </div>
            
            <div className="text-center">
              <div className={`text-3xl font-bold ${getScoreColor(metrics.pass_rate)}`}>
                {formatPercentage(metrics.pass_rate)}
              </div>
              <div className="text-sm text-gray-600 mt-1">Pass Rate</div>
              <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                <div 
                  className={`h-2 rounded-full transition-all duration-300 ${
                    metrics.pass_rate >= 0.8 ? 'bg-green-500' :
                    metrics.pass_rate >= 0.6 ? 'bg-yellow-500' : 'bg-red-500'
                  }`}
                  style={{ width: `${(metrics.pass_rate || 0) * 100}%` }}
                ></div>
              </div>
            </div>
            
            <div className="text-center">
              <div className="text-3xl font-bold text-blue-600">
                {metrics.test_cases_passed}
              </div>
              <div className="text-sm text-gray-600 mt-1">Tests Passed</div>
              <div className="text-xs text-gray-500 mt-2">
                of {metrics.test_cases_total} total
              </div>
            </div>
            
            <div className="text-center">
              <div className="text-3xl font-bold text-gray-900">
                {metrics.test_cases_total}
              </div>
              <div className="text-sm text-gray-600 mt-1">Total Tests</div>
              <div className="text-xs text-gray-500 mt-2">
                {metrics.test_cases_total - metrics.test_cases_passed} failed
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Classification Metrics */}
      {metrics.classification_metrics && (
        <>
          <div className="bg-white rounded-lg shadow border border-gray-200">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Classification Performance</h3>
            </div>
            <div className="p-6">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div className="text-center">
                  <div className={`text-2xl font-bold ${getScoreColor(metrics.classification_metrics?.accuracy || 0)}`}>
                    {formatPercentage(metrics.classification_metrics?.accuracy || 0)}
                  </div>
                  <div className="text-sm text-gray-600 mt-1">Accuracy</div>
                  <div className="w-full bg-gray-200 rounded-full h-2 mt-2">
                    <div 
                      className={`h-2 rounded-full ${
                        (metrics.classification_metrics?.accuracy || 0) >= 0.8 ? 'bg-green-500' :
                        (metrics.classification_metrics?.accuracy || 0) >= 0.6 ? 'bg-yellow-500' : 'bg-red-500'
                      }`}
                      style={{ width: `${(metrics.classification_metrics?.accuracy || 0) * 100}%` }}
                    ></div>
                  </div>
                </div>
                
                <div className="text-center">
                  <div className={`text-2xl font-bold ${getScoreColor(metrics.classification_metrics?.macro_f1 || 0)}`}>
                    {formatScore(metrics.classification_metrics?.macro_f1 || 0)}
                  </div>
                  <div className="text-sm text-gray-600 mt-1">Macro F1</div>
                  <div className="text-xs text-gray-500 mt-2">Unweighted average</div>
                </div>
                
                <div className="text-center">
                  <div className={`text-2xl font-bold ${getScoreColor(metrics.classification_metrics?.weighted_f1 || 0)}`}>
                    {formatScore(metrics.classification_metrics?.weighted_f1 || 0)}
                  </div>
                  <div className="text-sm text-gray-600 mt-1">Weighted F1</div>
                  <div className="text-xs text-gray-500 mt-2">Support-weighted</div>
                </div>
              </div>

              {/* Per-class metrics */}
              {classes.length > 0 && (
                <div className="mt-8">
                  <h4 className="text-md font-medium text-gray-900 mb-4">Per-Class Performance</h4>
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200">
                      <thead className="bg-gray-50">
                        <tr>
                          <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Class
                          </th>
                          <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Precision
                          </th>
                          <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                            Recall
                          </th>
                          <th className="px-4 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                            F1-Score
                          </th>
                        </tr>
                      </thead>
                      <tbody className="bg-white divide-y divide-gray-200">
                        {classes.map(className => {
                          const precision = metrics.classification_metrics?.precision?.[className] || 0;
                          const recall = metrics.classification_metrics?.recall?.[className] || 0;
                          const f1 = metrics.classification_metrics?.f1_score?.[className] || 0;
                          
                          return (
                            <tr key={className}>
                              <td className="px-4 py-3 text-sm font-medium text-gray-900 capitalize">
                                {className}
                              </td>
                              <td className="px-4 py-3 text-center">
                                <span className={`text-sm font-medium ${getScoreColor(precision)}`}>
                                  {formatScore(precision)}
                                </span>
                              </td>
                              <td className="px-4 py-3 text-center">
                                <span className={`text-sm font-medium ${getScoreColor(recall)}`}>
                                  {formatScore(recall)}
                                </span>
                              </td>
                              <td className="px-4 py-3 text-center">
                                <span className={`text-sm font-medium ${getScoreColor(f1)}`}>
                                  {formatScore(f1)}
                                </span>
                              </td>
                            </tr>
                          );
                        })}
                      </tbody>
                    </table>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Confusion Matrix */}
          {metrics.classification_metrics?.confusion_matrix && classes.length > 0 && (
            <ConfusionMatrix
              data={metrics.classification_metrics?.confusion_matrix || {}}
              classes={classes}
              title="Confusion Matrix"
            />
          )}
        </>
      )}

      {/* Generation Metrics */}
      {metrics.generation_metrics && (
        <div className="bg-white rounded-lg shadow border border-gray-200">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-medium text-gray-900">Generation Quality</h3>
          </div>
          <div className="p-6">
            <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
              <div className="text-center">
                <div className={`text-2xl font-bold ${getScoreColor(metrics.generation_metrics?.bleu || 0)}`}>
                  {formatScore(metrics.generation_metrics?.bleu || 0)}
                </div>
                <div className="text-sm text-gray-600 mt-1">BLEU Score</div>
                <div className="text-xs text-gray-500 mt-1">N-gram overlap</div>
              </div>
              
              <div className="text-center">
                <div className={`text-2xl font-bold ${getScoreColor(metrics.generation_metrics?.rouge_1 || 0)}`}>
                  {formatScore(metrics.generation_metrics?.rouge_1 || 0)}
                </div>
                <div className="text-sm text-gray-600 mt-1">ROUGE-1</div>
                <div className="text-xs text-gray-500 mt-1">Unigram overlap</div>
              </div>
              
              <div className="text-center">
                <div className={`text-2xl font-bold ${getScoreColor(metrics.generation_metrics?.diversity || 0)}`}>
                  {formatScore(metrics.generation_metrics?.diversity || 0)}
                </div>
                <div className="text-sm text-gray-600 mt-1">Diversity</div>
                <div className="text-xs text-gray-500 mt-1">Lexical variety</div>
              </div>
              
              <div className="text-center">
                <div className={`text-2xl font-bold ${getScoreColor(metrics.generation_metrics?.coherence || 0)}`}>
                  {formatScore(metrics.generation_metrics?.coherence || 0)}
                </div>
                <div className="text-sm text-gray-600 mt-1">Coherence</div>
                <div className="text-xs text-gray-500 mt-1">Text flow</div>
              </div>
            </div>

            <div className="mt-6 grid grid-cols-2 md:grid-cols-4 gap-6">
              <div className="text-center">
                <div className={`text-xl font-semibold ${getScoreColor(metrics.generation_metrics?.rouge_2 || 0)}`}>
                  {formatScore(metrics.generation_metrics?.rouge_2 || 0)}
                </div>
                <div className="text-sm text-gray-600 mt-1">ROUGE-2</div>
              </div>
              
              <div className="text-center">
                <div className={`text-xl font-semibold ${getScoreColor(metrics.generation_metrics?.rouge_l || 0)}`}>
                  {formatScore(metrics.generation_metrics?.rouge_l || 0)}
                </div>
                <div className="text-sm text-gray-600 mt-1">ROUGE-L</div>
              </div>
              
              <div className="text-center">
                <div className={`text-xl font-semibold ${getScoreColor(metrics.generation_metrics?.bert_score || 0)}`}>
                  {formatScore(metrics.generation_metrics?.bert_score || 0)}
                </div>
                <div className="text-sm text-gray-600 mt-1">BERTScore</div>
              </div>
              
              <div className="text-center">
                <div className={`text-xl font-semibold ${getScoreColor(metrics.generation_metrics?.relevance || 0)}`}>
                  {formatScore(metrics.generation_metrics?.relevance || 0)}
                </div>
                <div className="text-sm text-gray-600 mt-1">Relevance</div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Custom Metrics */}
      {metrics.custom_metrics && Object.keys(metrics.custom_metrics).length > 0 && (
        <div className="bg-white rounded-lg shadow border border-gray-200">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-medium text-gray-900">Custom Metrics</h3>
          </div>
          <div className="p-6">
            <div className="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
              {Object.entries(metrics.custom_metrics).map(([key, value]) => (
                <div key={key} className="text-center">
                  <div className={`text-xl font-semibold ${getScoreColor(value)}`}>
                    {typeof value === 'number' && value <= 1 && value >= 0 
                      ? formatPercentage(value)
                      : formatScore(value)
                    }
                  </div>
                  <div className="text-sm text-gray-600 mt-1 capitalize">
                    {key.replace(/_/g, ' ')}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Performance Summary */}
      <div className="bg-white rounded-lg shadow border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200">
          <h3 className="text-lg font-medium text-gray-900">Performance Summary</h3>
        </div>
        <div className="p-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-3">Strengths</h4>
              <div className="space-y-2">
                {metrics.overall_score >= 0.8 && (
                  <div className="flex items-center text-sm text-green-600">
                    <svg className="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                    High overall performance ({formatPercentage(metrics.overall_score)})
                  </div>
                )}
                {metrics.pass_rate >= 0.8 && (
                  <div className="flex items-center text-sm text-green-600">
                    <svg className="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                    Excellent pass rate ({formatPercentage(metrics.pass_rate)})
                  </div>
                )}
                {metrics.classification_metrics?.accuracy && metrics.classification_metrics.accuracy >= 0.8 && (
                  <div className="flex items-center text-sm text-green-600">
                    <svg className="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clipRule="evenodd" />
                    </svg>
                    High classification accuracy ({formatPercentage(metrics.classification_metrics.accuracy)})
                  </div>
                )}
              </div>
            </div>
            
            <div>
              <h4 className="text-sm font-medium text-gray-900 mb-3">Areas for Improvement</h4>
              <div className="space-y-2">
                {metrics.overall_score < 0.6 && (
                  <div className="flex items-center text-sm text-red-600">
                    <svg className="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                    </svg>
                    Low overall performance needs attention
                  </div>
                )}
                {metrics.pass_rate < 0.7 && (
                  <div className="flex items-center text-sm text-yellow-600">
                    <svg className="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                    </svg>
                    Pass rate could be improved ({formatPercentage(metrics.pass_rate)})
                  </div>
                )}
                {metrics.classification_metrics?.accuracy !== undefined && metrics.classification_metrics.accuracy < 0.7 && (
                  <div className="flex items-center text-sm text-red-600">
                    <svg className="w-4 h-4 mr-2" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                    </svg>
                    Classification accuracy needs improvement
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default MetricsDashboard;