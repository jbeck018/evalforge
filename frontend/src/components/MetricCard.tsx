import { ComponentType } from 'react'
import { TrendingUp, TrendingDown, LucideProps } from 'lucide-react'

interface MetricCardProps {
  title: string
  value: string | number
  icon: ComponentType<LucideProps>
  change?: number
  changeLabel?: string
  trend?: 'up' | 'down' | 'neutral'
  color?: 'blue' | 'green' | 'yellow' | 'red'
}

export default function MetricCard({ 
  title, 
  value, 
  icon: Icon, 
  change, 
  changeLabel,
  trend,
  color = 'blue' 
}: MetricCardProps) {
  const colorClasses = {
    blue: 'bg-blue-500',
    green: 'bg-green-500',
    yellow: 'bg-yellow-500',
    red: 'bg-red-500'
  }

  return (
    <div className="bg-white overflow-hidden shadow rounded-lg">
      <div className="p-5">
        <div className="flex items-center">
          <div className="flex-shrink-0">
            <div className={`${colorClasses[color]} rounded-md p-3`}>
              <Icon className="h-6 w-6 text-white" />
            </div>
          </div>
          <div className="ml-5 w-0 flex-1">
            <dl>
              <dt className="text-sm font-medium text-gray-500 truncate">
                {title}
              </dt>
              <dd className="flex items-baseline">
                <div className="text-2xl font-semibold text-gray-900">
                  {value}
                </div>
                {change !== undefined && (
                  <div className={`ml-2 flex items-baseline text-sm font-semibold ${
                    change >= 0 ? 'text-green-600' : 'text-red-600'
                  }`}>
                    {change >= 0 ? (
                      <TrendingUp className="self-center flex-shrink-0 h-4 w-4 text-green-500" />
                    ) : (
                      <TrendingDown className="self-center flex-shrink-0 h-4 w-4 text-red-500" />
                    )}
                    <span className="ml-1">
                      {Math.abs(change)}%
                    </span>
                    {changeLabel && (
                      <span className="text-gray-500 ml-1">{changeLabel}</span>
                    )}
                  </div>
                )}
              </dd>
            </dl>
          </div>
        </div>
      </div>
    </div>
  )
}