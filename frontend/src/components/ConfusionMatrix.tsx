import React from 'react';

interface ConfusionMatrixProps {
  data: Record<string, Record<string, number>>;
  classes: string[];
  title?: string;
}

const ConfusionMatrix: React.FC<ConfusionMatrixProps> = ({ 
  data, 
  classes, 
  title = "Confusion Matrix" 
}) => {
  // Calculate totals and percentages
  const getTotal = () => {
    let total = 0;
    classes.forEach(actualClass => {
      classes.forEach(predictedClass => {
        total += data[actualClass]?.[predictedClass] || 0;
      });
    });
    return total;
  };

  const total = getTotal();
  
  const getPercentage = (actualClass: string, predictedClass: string) => {
    const value = data[actualClass]?.[predictedClass] || 0;
    return total > 0 ? (value / total * 100) : 0;
  };

  const getRowTotal = (actualClass: string) => {
    return classes.reduce((sum, predictedClass) => {
      return sum + (data[actualClass]?.[predictedClass] || 0);
    }, 0);
  };

  const getColumnTotal = (predictedClass: string) => {
    return classes.reduce((sum, actualClass) => {
      return sum + (data[actualClass]?.[predictedClass] || 0);
    }, 0);
  };

  const getCellColor = (actualClass: string, predictedClass: string) => {
    const value = data[actualClass]?.[predictedClass] || 0;
    const percentage = getPercentage(actualClass, predictedClass);
    
    if (actualClass === predictedClass) {
      // Diagonal (correct predictions) - green scale
      if (percentage > 20) return 'bg-green-500';
      if (percentage > 10) return 'bg-green-400';
      if (percentage > 5) return 'bg-green-300';
      if (percentage > 1) return 'bg-green-200';
      return 'bg-green-100';
    } else {
      // Off-diagonal (incorrect predictions) - red scale
      if (percentage > 20) return 'bg-red-500';
      if (percentage > 10) return 'bg-red-400';
      if (percentage > 5) return 'bg-red-300';
      if (percentage > 1) return 'bg-red-200';
      if (value > 0) return 'bg-red-100';
      return 'bg-gray-50';
    }
  };

  const getTextColor = (actualClass: string, predictedClass: string) => {
    const percentage = getPercentage(actualClass, predictedClass);
    return percentage > 10 ? 'text-white' : 'text-gray-800';
  };

  if (classes.length === 0 || !data) {
    return (
      <div className="bg-white rounded-lg border border-gray-200 p-6">
        <h3 className="text-lg font-medium text-gray-900 mb-4">{title}</h3>
        <div className="text-center text-gray-500 py-8">
          <p>No confusion matrix data available</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-6">
      <h3 className="text-lg font-medium text-gray-900 mb-4">{title}</h3>
      
      <div className="overflow-x-auto">
        <div className="inline-block min-w-full">
          {/* Table container */}
          <div className="relative">
            {/* Predicted label */}
            <div className="text-center mb-2">
              <span className="text-sm font-medium text-gray-700">Predicted Class</span>
            </div>
            
            <table className="min-w-full">
              <thead>
                <tr>
                  <th className="w-16"></th> {/* Empty corner */}
                  {classes.map(predictedClass => (
                    <th
                      key={predictedClass}
                      className="px-3 py-2 text-center text-xs font-medium text-gray-700 uppercase tracking-wider border-b border-gray-200 min-w-20"
                    >
                      {predictedClass}
                    </th>
                  ))}
                  <th className="px-3 py-2 text-center text-xs font-medium text-gray-700 uppercase tracking-wider border-b border-gray-200">
                    Total
                  </th>
                </tr>
              </thead>
              <tbody>
                {classes.map((actualClass, rowIndex) => (
                  <tr key={actualClass}>
                    {/* Actual class label */}
                    {rowIndex === Math.floor(classes.length / 2) && (
                      <td
                        rowSpan={classes.length}
                        className="px-2 py-1 text-center text-xs font-medium text-gray-700 uppercase tracking-wider border-r border-gray-200 vertical-text"
                        style={{ 
                          writingMode: 'vertical-rl',
                          textOrientation: 'mixed',
                          width: '2rem'
                        }}
                      >
                        Actual Class
                      </td>
                    )}
                    {rowIndex !== Math.floor(classes.length / 2) && (
                      <td className="w-16 border-r border-gray-200"></td>
                    )}
                    
                    {/* Matrix cells */}
                    {classes.map(predictedClass => {
                      const value = data[actualClass]?.[predictedClass] || 0;
                      const percentage = getPercentage(actualClass, predictedClass);
                      
                      return (
                        <td
                          key={predictedClass}
                          className={`px-3 py-4 text-center border border-gray-200 ${getCellColor(actualClass, predictedClass)} ${getTextColor(actualClass, predictedClass)}`}
                        >
                          <div className="font-semibold text-sm">{value}</div>
                          <div className="text-xs opacity-90">
                            {percentage.toFixed(1)}%
                          </div>
                        </td>
                      );
                    })}
                    
                    {/* Row total */}
                    <td className="px-3 py-4 text-center border border-gray-200 bg-gray-100 font-medium">
                      <div className="text-sm">{getRowTotal(actualClass)}</div>
                    </td>
                  </tr>
                ))}
                
                {/* Column totals */}
                <tr>
                  <td className="px-3 py-2 text-center text-xs font-medium text-gray-700 uppercase tracking-wider border-t border-gray-200">
                    Total
                  </td>
                  {classes.map(predictedClass => (
                    <td
                      key={predictedClass}
                      className="px-3 py-2 text-center border border-gray-200 bg-gray-100 font-medium text-sm"
                    >
                      {getColumnTotal(predictedClass)}
                    </td>
                  ))}
                  <td className="px-3 py-2 text-center border border-gray-200 bg-gray-200 font-semibold text-sm">
                    {total}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      {/* Legend */}
      <div className="mt-4 flex items-center justify-center space-x-6 text-xs">
        <div className="flex items-center space-x-2">
          <div className="w-4 h-4 bg-green-400 rounded"></div>
          <span className="text-gray-600">Correct Predictions</span>
        </div>
        <div className="flex items-center space-x-2">
          <div className="w-4 h-4 bg-red-400 rounded"></div>
          <span className="text-gray-600">Incorrect Predictions</span>
        </div>
        <div className="flex items-center space-x-2">
          <div className="w-4 h-4 bg-gray-100 border border-gray-300 rounded"></div>
          <span className="text-gray-600">No Predictions</span>
        </div>
      </div>

      {/* Summary Statistics */}
      <div className="mt-4 pt-4 border-t border-gray-200">
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-center">
          <div>
            <div className="text-lg font-semibold text-green-600">
              {classes.reduce((sum, cls) => {
                return sum + (data[cls]?.[cls] || 0);
              }, 0)}
            </div>
            <div className="text-xs text-gray-600">Correct</div>
          </div>
          <div>
            <div className="text-lg font-semibold text-red-600">
              {total - classes.reduce((sum, cls) => {
                return sum + (data[cls]?.[cls] || 0);
              }, 0)}
            </div>
            <div className="text-xs text-gray-600">Incorrect</div>
          </div>
          <div>
            <div className="text-lg font-semibold text-blue-600">
              {total > 0 ? (classes.reduce((sum, cls) => {
                return sum + (data[cls]?.[cls] || 0);
              }, 0) / total * 100).toFixed(1) : 0}%
            </div>
            <div className="text-xs text-gray-600">Accuracy</div>
          </div>
          <div>
            <div className="text-lg font-semibold text-gray-700">
              {classes.length}
            </div>
            <div className="text-xs text-gray-600">Classes</div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ConfusionMatrix;