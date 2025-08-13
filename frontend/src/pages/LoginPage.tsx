import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { toast } from 'sonner'
import { BarChart3, Eye, EyeOff } from 'lucide-react'

import { useAuthStore } from '../stores/auth'
import { Button } from '../components/ui/button'
import { Input } from '../components/ui/input'
import { Label } from '../components/ui/label'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '../components/ui/card'

const loginSchema = z.object({
  email: z.string().email('Please enter a valid email address'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
})

type LoginForm = z.infer<typeof loginSchema>

export default function LoginPage() {
  const [showPassword, setShowPassword] = useState(false)
  const { login, isLoading } = useAuthStore()
  const navigate = useNavigate()

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
  })

  const onSubmit = async (data: LoginForm) => {
    try {
      await login(data.email, data.password)
      toast.success('Welcome back!')
      navigate('/dashboard')
    } catch (error: any) {
      const message = error.response?.data?.error || 'Login failed'
      toast.error(message)
    }
  }

  const loading = isLoading || isSubmitting

  return (
    <div className="min-h-screen flex">
      {/* Left side - Hero */}
      <div className="hidden lg:flex lg:flex-1 lg:flex-col lg:justify-center lg:px-12 xl:px-24 bg-gradient-to-br from-blue-600 to-blue-800">
        <div className="max-w-md">
          <div className="flex items-center space-x-3 mb-8">
            <div className="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center">
              <BarChart3 className="w-7 h-7 text-white" />
            </div>
            <div className="text-3xl font-bold text-white">EvalForge</div>
          </div>
          
          <h1 className="text-4xl font-bold text-white mb-6">
            Zero-overhead LLM observability
          </h1>
          
          <p className="text-xl text-blue-100 mb-8">
            Monitor, evaluate, and optimize your LLM applications with real-time insights 
            and comprehensive analytics.
          </p>
          
          <div className="space-y-4">
            <div className="flex items-center space-x-3 text-blue-100">
              <div className="w-2 h-2 bg-blue-300 rounded-full" />
              <span>Real-time cost and performance tracking</span>
            </div>
            <div className="flex items-center space-x-3 text-blue-100">
              <div className="w-2 h-2 bg-blue-300 rounded-full" />
              <span>One-line SDK integration</span>
            </div>
            <div className="flex items-center space-x-3 text-blue-100">
              <div className="w-2 h-2 bg-blue-300 rounded-full" />
              <span>OpenAI and Anthropic support</span>
            </div>
          </div>
        </div>
      </div>

      {/* Right side - Login form */}
      <div className="flex-1 flex flex-col justify-center px-8 sm:px-12 lg:px-24 xl:px-32">
        <div className="w-full max-w-md mx-auto">
          {/* Mobile logo */}
          <div className="lg:hidden flex items-center justify-center space-x-2 mb-8">
            <div className="w-10 h-10 bg-blue-600 rounded-lg flex items-center justify-center">
              <BarChart3 className="w-6 h-6 text-white" />
            </div>
            <div className="text-2xl font-bold text-gray-900">EvalForge</div>
          </div>

          <div className="text-center mb-8">
            <h2 className="text-3xl font-bold text-gray-900 mb-2">
              Welcome back
            </h2>
            <p className="text-gray-600">
              Sign in to your EvalForge account
            </p>
          </div>

          <form onSubmit={handleSubmit(onSubmit)} className="space-y-6">
            <div className="space-y-2">
              <Label htmlFor="email">
                Email address
              </Label>
              <Input
                {...register('email')}
                type="email"
                id="email"
                placeholder="Enter your email"
                disabled={loading}
              />
              {errors.email && (
                <p className="text-sm text-red-600">{errors.email.message}</p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">
                Password
              </Label>
              <div className="relative">
                <Input
                  {...register('password')}
                  type={showPassword ? 'text' : 'password'}
                  id="password"
                  placeholder="Enter your password"
                  disabled={loading}
                  className="pr-10"
                />
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  className="absolute right-0 top-0 h-full px-3 py-2 hover:bg-transparent"
                  onClick={() => setShowPassword(!showPassword)}
                  disabled={loading}
                >
                  {showPassword ? (
                    <EyeOff className="h-4 w-4 text-gray-400" />
                  ) : (
                    <Eye className="h-4 w-4 text-gray-400" />
                  )}
                </Button>
              </div>
              {errors.password && (
                <p className="text-sm text-red-600">{errors.password.message}</p>
              )}
            </div>

            <Button
              type="submit"
              disabled={loading}
              className="w-full"
              size="lg"
            >
              {loading ? (
                <div className="flex items-center justify-center space-x-2">
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white" />
                  <span>Signing in...</span>
                </div>
              ) : (
                'Sign in'
              )}
            </Button>
          </form>

          <div className="mt-6 text-center">
            <p className="text-sm text-gray-600">
              Don't have an account?{' '}
              <Link
                to="/register"
                className="font-medium text-blue-600 hover:text-blue-500 transition-colors"
              >
                Sign up for free
              </Link>
            </p>
          </div>

          {/* Demo credentials */}
          <Card className="mt-8">
            <CardHeader className="pb-3">
              <CardTitle className="text-sm">Demo Account</CardTitle>
              <CardDescription className="text-xs">
                Try EvalForge with these demo credentials:
              </CardDescription>
            </CardHeader>
            <CardContent className="pt-0">
              <div className="text-xs font-mono text-gray-700 space-y-1">
                <div>Email: demo@evalforge.dev</div>
                <div>Password: demo123</div>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}